package config

import (
	"github.com/joho/godotenv"
	"os"
	"strconv"
)

type Config struct {
	ServerPort string

	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
	DBSSLMode  string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	RabbitMQURL   string
	RabbitMQQueue string

	MQTTBroker     string
	MQTTClientID   string
	MQTTUsername   string
	MQTTPassword   string
	MQTTTopic      string
	MQTTUseTLS     bool
	MQTTCACertPath string

	SimulationDevicesPerClient int
	SimulationClients          int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", ""))
	devicesPerClient, _ := strconv.Atoi(getEnv("SIMULATION_DEVICES_PER_CLIENT", ""))
	clients, _ := strconv.Atoi(getEnv("SIMULATION_CLIENTS", ""))
	useTLS, _ := strconv.ParseBool(getEnv("MQTT_USE_TLS", ""))

	return &Config{
		ServerPort: getEnv("SERVER_PORT", ""),

		DBHost:     getEnv("DB_HOST", ""),
		DBPort:     getEnv("DB_PORT", ""),
		DBUser:     getEnv("DB_USER", ""),
		DBPassword: getEnv("DB_PASSWORD", ""),
		DBName:     getEnv("DB_NAME", ""),
		DBSSLMode:  getEnv("DB_SSLMODE", ""),

		RedisHost:     getEnv("REDIS_HOST", ""),
		RedisPort:     getEnv("REDIS_PORT", ""),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       redisDB,

		RabbitMQURL:   getEnv("RABBITMQ_URL", ""),
		RabbitMQQueue: getEnv("RABBITMQ_QUEUE", ""),

		MQTTBroker:     getEnv("MQTT_BROKER", ""),
		MQTTClientID:   getEnv("MQTT_CLIENT_ID", ""),
		MQTTUsername:   getEnv("MQTT_USERNAME", ""),
		MQTTPassword:   getEnv("MQTT_PASSWORD", ""),
		MQTTTopic:      getEnv("MQTT_TOPIC", ""),
		MQTTUseTLS:     useTLS,
		MQTTCACertPath: getEnv("MQTT_CA_CERT_PATH", ""),

		SimulationDevicesPerClient: devicesPerClient,
		SimulationClients:          clients,
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
