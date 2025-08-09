package service

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"os"
	"smat/iot/simulation/iot-inventory-management/internal/config"
	"smat/iot/simulation/iot-inventory-management/internal/domain"
	"time"
)

type mqttService struct {
	client   mqtt.Client
	config   *config.Config
	rabbitMQ RabbitMQService
}

func NewMQTTService(cfg *config.Config, rabbitMQ RabbitMQService) MQTTService {
	return &mqttService{
		config:   cfg,
		rabbitMQ: rabbitMQ,
	}
}

func (s *mqttService) Connect() error {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(s.config.MQTTBroker)
	opts.SetClientID(s.config.MQTTClientID)
	opts.SetUsername(s.config.MQTTUsername)
	opts.SetPassword(s.config.MQTTPassword)

	if s.config.MQTTUseTLS {
		tlsConfig, err := s.createTLSConfig()
		if err != nil {
			return fmt.Errorf("failed to create TLS config: %w", err)
		}
		opts.SetTLSConfig(tlsConfig)
	}

	opts.SetKeepAlive(60 * time.Second)
	opts.SetPingTimeout(10 * time.Second)
	opts.SetCleanSession(true)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(10 * time.Second)

	opts.SetDefaultPublishHandler(s.messageHandler)
	opts.SetOnConnectHandler(s.onConnect)
	opts.SetConnectionLostHandler(s.onConnectionLost)
	opts.SetReconnectingHandler(s.onReconnecting)

	s.client = mqtt.NewClient(opts)

	if token := s.client.Connect(); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to connect to MQTT broker: %w", token.Error())
	}

	log.Printf("Successfully connected to MQTT broker at %s", s.config.MQTTBroker)
	return nil
}

func (s *mqttService) createTLSConfig() (*tls.Config, error) {
	caCert, err := os.ReadFile(s.config.MQTTCACertPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		return nil, fmt.Errorf("failed to parse CA certificate")
	}

	tlsConfig := &tls.Config{
		RootCAs:            caCertPool,
		ClientAuth:         tls.NoClientCert,
		InsecureSkipVerify: false, // Set to true only for testing if having cert issues
	}

	return tlsConfig, nil
}

func (s *mqttService) Subscribe(topic string) error {
	if token := s.client.Subscribe(topic, 1, nil); token.Wait() && token.Error() != nil {
		return fmt.Errorf("failed to subscribe to topic %s: %w", topic, token.Error())
	}
	log.Printf("Successfully subscribed to topic: %s", topic)
	return nil
}

func (s *mqttService) messageHandler(client mqtt.Client, msg mqtt.Message) {
	log.Printf("Received message from topic %s: %s", msg.Topic(), msg.Payload())

	var deviceMsg domain.DeviceMessage
	if err := json.Unmarshal(msg.Payload(), &deviceMsg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.rabbitMQ.PublishMessageWithContext(ctx, msg.Payload()); err != nil {
		log.Printf("Failed to publish to RabbitMQ: %v", err)

		if healthErr := s.rabbitMQ.HealthCheck(); healthErr != nil {
			log.Printf("RabbitMQ health check failed: %v", healthErr)
		}
	} else {
		log.Printf("Successfully forwarded message to RabbitMQ for device: %s", deviceMsg.DeviceID)
	}
}

func (s *mqttService) onConnect(client mqtt.Client) {
	log.Println("Connected to MQTT broker - subscribing to topics")
	if err := s.Subscribe(s.config.MQTTTopic); err != nil {
		log.Printf("Failed to re-subscribe on connect: %v", err)
	}
}

func (s *mqttService) onConnectionLost(client mqtt.Client, err error) {
	log.Printf("Connection lost to MQTT broker: %v", err)
}

func (s *mqttService) onReconnecting(client mqtt.Client, opts *mqtt.ClientOptions) {
	log.Println("Attempting to reconnect to MQTT broker...")
}

func (s *mqttService) Disconnect() {
	if s.client != nil && s.client.IsConnected() {
		s.client.Disconnect(250)
		log.Println("Disconnected from MQTT broker")
	}
}

func (s *mqttService) IsConnected() bool {
	return s.client != nil && s.client.IsConnected()
}
