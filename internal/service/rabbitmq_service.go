package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"log"
	"smat/iot/simulation/iot-inventory-management/internal/config"
)

type rabbitMQService struct {
	conn            *amqp.Connection
	channel         *amqp.Channel
	queue           amqp.Queue
	config          *config.Config
	mu              sync.RWMutex
	reconnecting    atomic.Bool
	closeChan       chan struct{}
	closeOnce       sync.Once
	notifyConnClose chan *amqp.Error
	consumerTag     string
}

func NewRabbitMQService(cfg *config.Config) RabbitMQService {
	return &rabbitMQService{
		config:      cfg,
		closeChan:   make(chan struct{}),
		consumerTag: fmt.Sprintf("consumer-%d", time.Now().Unix()),
	}
}

func (s *rabbitMQService) Connect() error {
	if err := s.connectWithRetry(); err != nil {
		return err
	}

	go s.handleReconnect()

	return nil
}

func (s *rabbitMQService) connectWithRetry() error {
	maxRetries := 30
	baseDelay := time.Second
	maxDelay := 30 * time.Second

	for attempt := 0; attempt < maxRetries; attempt++ {
		if err := s.establishConnection(); err == nil {
			log.Println("Successfully connected to RabbitMQ")
			return nil
		} else if attempt == maxRetries-1 {
			return fmt.Errorf("failed to connect to RabbitMQ after %d attempts: %w", maxRetries, err)
		} else {
			delay := s.calculateBackoff(attempt, baseDelay, maxDelay)
			log.Printf("Failed to connect to RabbitMQ (attempt %d/%d): %v. Retrying in %v...",
				attempt+1, maxRetries, err, delay)

			select {
			case <-time.After(delay):
				continue
			case <-s.closeChan:
				return fmt.Errorf("connection cancelled")
			}
		}
	}

	return fmt.Errorf("connection failed after all retries")
}

func (s *rabbitMQService) calculateBackoff(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := baseDelay * time.Duration(1<<uint(attempt))

	// Cap at maxDelay
	if delay > maxDelay {
		delay = maxDelay
	}

	// Add jitter (±25% of delay)
	jitter := time.Duration(rand.Float64() * float64(delay) * 0.5)
	if rand.Intn(2) == 0 {
		delay = delay + jitter
	} else {
		delay = delay - jitter
	}

	return delay
}

func (s *rabbitMQService) establishConnection() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn != nil && !s.conn.IsClosed() {
		s.conn.Close()
	}

	conn, err := amqp.Dial(s.config.RabbitMQURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	err = channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to set QoS: %w", err)
	}

	queue, err := channel.QueueDeclare(
		s.config.RabbitMQQueue,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-message-ttl": 3600000, // 1 hour TTL for messages
		},
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	s.conn = conn
	s.channel = channel
	s.queue = queue
	s.reconnecting.Store(false)

	s.notifyConnClose = make(chan *amqp.Error, 1)
	s.conn.NotifyClose(s.notifyConnClose)

	return nil
}

func (s *rabbitMQService) handleReconnect() {
	for {
		select {
		case <-s.closeChan:
			return
		case err := <-s.notifyConnClose:
			if err != nil {
				log.Printf("RabbitMQ connection lost: %v", err)
				s.reconnect()
			}
		}
	}
}

func (s *rabbitMQService) reconnect() {
	s.reconnecting.Store(true)
	defer s.reconnecting.Store(false)

	for {
		select {
		case <-s.closeChan:
			return
		default:
			log.Println("Attempting to reconnect to RabbitMQ...")

			if err := s.connectWithRetry(); err != nil {
				log.Printf("Reconnection failed: %v. Retrying in 5 seconds...", err)
				time.Sleep(5 * time.Second)
				continue
			}

			log.Println("Successfully reconnected to RabbitMQ")
			return
		}
	}
}

func (s *rabbitMQService) PublishMessage(message []byte) error {
	return s.PublishMessageWithContext(context.Background(), message)
}

func (s *rabbitMQService) PublishMessageWithContext(ctx context.Context, message []byte) error {
	if s.reconnecting.Load() {
		return fmt.Errorf("connection is being reestablished")
	}

	s.mu.RLock()
	channel := s.channel
	queueName := s.queue.Name
	s.mu.RUnlock()

	if channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	if _, ok := ctx.Deadline(); !ok {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	maxRetries := 3
	for i := 0; i < maxRetries; i++ {
		err := channel.PublishWithContext(
			ctx,
			"",        // exchange
			queueName, // routing key
			false,     // mandatory
			false,     // immediate
			amqp.Publishing{
				DeliveryMode: amqp.Persistent, // Make message persistent
				ContentType:  "application/json",
				Body:         message,
				Timestamp:    time.Now(),
			},
		)

		if err == nil {
			return nil
		}

		if i < maxRetries-1 {
			select {
			case <-ctx.Done():
				return fmt.Errorf("publish cancelled: %w", ctx.Err())
			case <-time.After(time.Duration(i+1) * time.Second):
				continue
			}
		}
	}

	return fmt.Errorf("failed to publish message after %d attempts", maxRetries)
}

func (s *rabbitMQService) ConsumeMessages() (<-chan []byte, error) {
	s.mu.RLock()
	channel := s.channel
	queueName := s.queue.Name
	s.mu.RUnlock()

	if channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	msgs, err := channel.Consume(
		queueName,
		s.consumerTag, // consumer tag
		false,         // auto-ack (set to false for manual ack)
		false,         // exclusive
		false,         // no-local
		false,         // no-wait
		nil,           // args
	)
	if err != nil {
		return nil, fmt.Errorf("failed to register consumer: %w", err)
	}

	messageChan := make(chan []byte, 100) // Buffered channel

	go s.processMessages(msgs, messageChan)

	return messageChan, nil
}

func (s *rabbitMQService) processMessages(msgs <-chan amqp.Delivery, output chan<- []byte) {
	defer close(output)

	for {
		select {
		case <-s.closeChan:
			return
		case msg, ok := <-msgs:
			if !ok {
				log.Println("Consumer channel closed, attempting to re-consume...")
				s.handleConsumerReconnect(output)
				return
			}

			select {
			case output <- msg.Body:
				if err := msg.Ack(false); err != nil {
					log.Printf("Failed to acknowledge message: %v", err)
				}
			case <-time.After(5 * time.Second):
				log.Printf("Timeout sending message to output channel")
				if err := msg.Nack(false, true); err != nil {
					log.Printf("Failed to nack message: %v", err)
				}
			}
		}
	}
}

func (s *rabbitMQService) handleConsumerReconnect(output chan<- []byte) {
	for {
		select {
		case <-s.closeChan:
			return
		default:
			if s.reconnecting.Load() {
				time.Sleep(1 * time.Second)
				continue
			}

			newMsgs, err := s.ConsumeMessages()
			if err != nil {
				log.Printf("Failed to re-establish consumer: %v", err)
				time.Sleep(5 * time.Second)
				continue
			}

			go func() {
				for msg := range newMsgs {
					select {
					case output <- msg:
					case <-s.closeChan:
						return
					}
				}
			}()

			return
		}
	}
}

func (s *rabbitMQService) PublishJSON(v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	return s.PublishMessage(data)
}

func (s *rabbitMQService) HealthCheck() error {
	if s.reconnecting.Load() {
		return fmt.Errorf("connection is reconnecting")
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.conn == nil || s.conn.IsClosed() {
		return fmt.Errorf("connection is closed")
	}

	if s.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	return nil
}

func (s *rabbitMQService) GetQueueInfo() (*amqp.Queue, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.channel == nil {
		return nil, fmt.Errorf("channel is not initialized")
	}

	queue, err := s.channel.QueueInspect(s.queue.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect queue: %w", err)
	}

	return &queue, nil
}

func (s *rabbitMQService) Close() {
	s.closeOnce.Do(func() {
		log.Println("Closing RabbitMQ connection...")

		// Signal all goroutines to stop
		close(s.closeChan)

		s.mu.Lock()
		defer s.mu.Unlock()

		// Cancel consumer
		if s.channel != nil && s.consumerTag != "" {
			_ = s.channel.Cancel(s.consumerTag, false)
		}

		// Close channel
		if s.channel != nil {
			_ = s.channel.Close()
			s.channel = nil
		}

		// Close connection
		if s.conn != nil {
			_ = s.conn.Close()
			s.conn = nil
		}

		log.Println("RabbitMQ connection closed")
	})
}
