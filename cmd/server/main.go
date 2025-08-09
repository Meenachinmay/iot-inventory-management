package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"smat/iot/simulation/iot-inventory-management/internal/config"
	"smat/iot/simulation/iot-inventory-management/internal/database"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"smat/iot/simulation/iot-inventory-management/internal/handler"
	"smat/iot/simulation/iot-inventory-management/internal/repository"
	"smat/iot/simulation/iot-inventory-management/internal/router"
	"smat/iot/simulation/iot-inventory-management/internal/service"
	"sync"
	"syscall"
	"time"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatal("Failed to connect to PostgreSQL:", err)
	}
	defer db.Close()

	//if os.Getenv("CLEAN_DB") == "true" {
	//	log.Println("CLEAN_DB=true detected, cleaning database...")
	//	if err := database.CleanDatabase(db); err != nil {
	//		log.Fatal("Failed to clean database:", err)
	//	}
	//}

	if err := database.InitializeDatabase(db); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	log.Println("Database initialized successfully")

	redisClient, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
	defer redisClient.Close()

	deviceRepo := repository.NewDeviceRepository(db)
	inventoryRepo := repository.NewInventoryRepository(db, redisClient)

	wsHub := service.NewWebSocketHub()
	go wsHub.Run()

	rabbitMQ := service.NewRabbitMQService(cfg)
	log.Println("Connecting to RabbitMQ at", cfg.RabbitMQURL)
	if err := rabbitMQ.Connect(); err != nil {
		log.Printf("Warning: Initial connection to RabbitMQ failed: %v", err)
		log.Println("Application will continue and retry connecting to RabbitMQ in the background")
		// Start a goroutine to keep trying to connect in the background
		go func() {
			for {
				time.Sleep(10 * time.Second)
				if err := rabbitMQ.Connect(); err == nil {
					log.Println("Successfully connected to RabbitMQ in background")
					break
				} else {
					log.Printf("Background connection attempt to RabbitMQ failed: %v", err)
				}
			}
		}()
	} else {
		log.Println("Successfully connected to RabbitMQ")
	}
	defer rabbitMQ.Close()

	deviceService := service.NewDeviceService(deviceRepo)
	inventoryService := service.NewInventoryService(deviceRepo, inventoryRepo, rabbitMQ, wsHub)

	mqttService := service.NewMQTTService(cfg, rabbitMQ)
	if err := mqttService.Connect(); err != nil {
		log.Fatal("Failed to connect to MQTT:", err)
	}
	defer mqttService.Disconnect()

	if err := mqttService.Subscribe(cfg.MQTTTopic); err != nil {
		log.Fatal("Failed to subscribe to MQTT topic:", err)
	}

	// Start consuming messages with better error handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		consumeRabbitMQMessages(ctx, rabbitMQ, inventoryService)
	}()

	ctx = context.Background()
	if err := deviceService.InitializeDevices(ctx); err != nil {
		log.Printf("Warning: Failed to initialize devices: %v", err)
	}

	deviceHandler := handler.NewDeviceHandler(deviceService)
	inventoryHandler := handler.NewInventoryHandler(inventoryService)
	wsHandler := handler.NewWebSocketHandler(wsHub)
	healthHandler := handler.NewHealthHandler(rabbitMQ)

	r := router.SetupRouter(deviceHandler, inventoryHandler, wsHandler, healthHandler)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.ServerPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	log.Printf("Server started on port %s", cfg.ServerPort)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Cancel background operations
	cancel()

	// Wait for goroutines with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Println("All goroutines finished gracefully")
	case <-time.After(5 * time.Second):
		log.Println("Timeout waiting for goroutines")
	}

	log.Println("Server shutdown complete")
}

func consumeRabbitMQMessages(ctx context.Context, rabbitMQ service.RabbitMQService, inventoryService service.InventoryService) {
	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping RabbitMQ consumer")
			return
		default:
			messages, err := rabbitMQ.ConsumeMessages()
			if err != nil {
				log.Printf("Failed to start consuming messages: %v", err)

				// Check health and retry
				if healthErr := rabbitMQ.HealthCheck(); healthErr != nil {
					log.Printf("RabbitMQ unhealthy: %v, retrying in 5 seconds", healthErr)
					time.Sleep(5 * time.Second)
					continue
				}
			}

			log.Println("Started consuming messages from RabbitMQ")

			// Process messages
			for {
				select {
				case <-ctx.Done():
					return
				case msg, ok := <-messages:
					if !ok {
						log.Println("Message channel closed, reconnecting...")
						break // Break inner loop to reconnect
					}

					var deviceMsg domain.DeviceMessage
					if err := json.Unmarshal(msg, &deviceMsg); err != nil {
						log.Printf("Failed to unmarshal message: %v", err)
						continue
					}

					// Process with timeout
					processCtx, processCancel := context.WithTimeout(ctx, 30*time.Second)
					if err := inventoryService.ProcessWeightUpdate(processCtx, &deviceMsg); err != nil {
						log.Printf("Failed to process weight update: %v", err)
					}
					processCancel()
				}
			}
		}
	}
}
