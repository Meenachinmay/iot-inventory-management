package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"smat/iot/simulation/iot-inventory-management/internal/config"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"sync"
	"syscall"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type DeviceSimulator struct {
	deviceID      string
	currentWeight float64
	itemWeight    float64
	mqttClient    mqtt.Client
	mu            sync.Mutex
}

func NewDeviceSimulator(deviceID string, initialWeight, itemWeight float64, mqttClient mqtt.Client) *DeviceSimulator {
	return &DeviceSimulator{
		deviceID:      deviceID,
		currentWeight: initialWeight,
		itemWeight:    itemWeight,
		mqttClient:    mqttClient,
	}
}

func (d *DeviceSimulator) SimulateWeightChange() {
	d.mu.Lock()
	defer d.mu.Unlock()

	itemsSold := rand.Intn(5) + 1
	weightReduction := float64(itemsSold) * d.itemWeight

	if d.currentWeight-weightReduction >= 0 {
		d.currentWeight -= weightReduction

		message := domain.DeviceMessage{
			DeviceID:  d.deviceID,
			Weight:    d.currentWeight,
			Timestamp: time.Now(),
		}

		payload, _ := json.Marshal(message)
		topic := fmt.Sprintf("devices/%s/weight", d.deviceID)

		token := d.mqttClient.Publish(topic, 1, false, payload)
		if token.Wait() && token.Error() != nil {
			log.Printf("Failed to publish for device %s: %v", d.deviceID, token.Error())
		} else {
			log.Printf("Device %s: Weight changed to %.2f kg (sold %d items)",
				d.deviceID, d.currentWeight, itemsSold)
		}
	}
}

func (d *DeviceSimulator) Run(stopCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(rand.Intn(10)+5) * time.Second)
	defer ticker.Stop()

	d.SendCurrentWeight()

	for {
		select {
		case <-ticker.C:
			d.SimulateWeightChange()
		case <-stopCh:
			return
		}
	}
}

func (d *DeviceSimulator) SendCurrentWeight() {
	message := domain.DeviceMessage{
		DeviceID:  d.deviceID,
		Weight:    d.currentWeight,
		Timestamp: time.Now(),
	}

	payload, _ := json.Marshal(message)
	topic := fmt.Sprintf("devices/%s/weight", d.deviceID)

	token := d.mqttClient.Publish(topic, 1, false, payload)
	if token.Wait() && token.Error() != nil {
		log.Printf("Failed to send initial weight for device %s: %v", d.deviceID, token.Error())
	} else {
		log.Printf("Device %s: Initial weight %.2f kg", d.deviceID, d.currentWeight)
	}
}

func createTLSConfig(certPath string) (*tls.Config, error) {
	caCert, err := ioutil.ReadFile(certPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	return &tls.Config{
		RootCAs:            caCertPool,
		ClientAuth:         tls.NoClientCert,
		InsecureSkipVerify: false,
	}, nil
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	opts := mqtt.NewClientOptions()
	opts.AddBroker(cfg.MQTTBroker)
	opts.SetClientID(fmt.Sprintf("device-simulator-%d", time.Now().Unix()))
	opts.SetUsername(cfg.MQTTUsername)
	opts.SetPassword(cfg.MQTTPassword)

	if cfg.MQTTUseTLS {
		tlsConfig, err := createTLSConfig(cfg.MQTTCACertPath)
		if err != nil {
			log.Fatal("Failed to create TLS config:", err)
		}
		opts.SetTLSConfig(tlsConfig)
	}

	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)

	opts.SetOnConnectHandler(func(c mqtt.Client) {
		log.Println("Simulator connected to MQTT broker")
	})

	opts.SetConnectionLostHandler(func(c mqtt.Client, err error) {
		log.Printf("Simulator connection lost: %v", err)
	})

	mqttClient := mqtt.NewClient(opts)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal("Failed to connect to MQTT:", token.Error())
	}
	defer mqttClient.Disconnect(250)

	log.Printf("Connected to EMQX Cloud broker at %s", cfg.MQTTBroker)

	var simulators []*DeviceSimulator

	for i := 0; i < cfg.SimulationClients; i++ {
		for j := 0; j < cfg.SimulationDevicesPerClient; j++ {
			deviceID := fmt.Sprintf("device-%d-%03d", i+1, j+1)
			initialWeight := 50.0 + rand.Float64()*50.0
			simulator := NewDeviceSimulator(deviceID, initialWeight, 1.0, mqttClient)
			simulators = append(simulators, simulator)
		}
	}

	log.Printf("Created %d device simulators", len(simulators))

	stopCh := make(chan struct{})
	var wg sync.WaitGroup

	for _, sim := range simulators {
		wg.Add(1)
		go sim.Run(stopCh, &wg)
	}

	log.Println("Simulation started. Press Ctrl+C to stop.")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	log.Println("Shutting down simulators...")
	close(stopCh)
	wg.Wait()
	log.Println("All simulators stopped")
}
