# IoT Inventory Management

A real-time IoT inventory management system that simulates and tracks inventory levels across multiple devices. The system uses weight sensors to monitor inventory levels, processes data through MQTT and RabbitMQ, and provides real-time updates via WebSockets.

## Project Overview

This project simulates an IoT-based inventory management system where:

- IoT devices (weight sensors) monitor inventory levels in containers
- Devices publish weight changes via MQTT when items are removed (sales)
- A server processes these updates and maintains inventory state in PostgreSQL
- Real-time inventory updates are broadcast to clients via WebSockets
- A simulation component allows testing without physical devices

## Project Structure

```
├── cmd/                      # Application entry points
│   ├── server/               # Main server application
│   └── simulator/            # Device simulator for testing
├── internal/                 # Internal packages
│   ├── config/               # Configuration loading
│   ├── database/             # Database connections (PostgreSQL, Redis)
│   ├── domain/               # Domain models
│   ├── handler/              # HTTP request handlers
│   ├── middleware/           # HTTP middleware
│   ├── repository/           # Data access layer
│   ├── router/               # API route definitions
│   └── service/              # Business logic
├── migrations/               # Database migrations
├── pkg/                      # Shared packages
└── tests/                    # Test files
```

## Key Components

### Server Application

The server application (`cmd/server/main.go`) is the core of the system:

- Connects to PostgreSQL for persistent storage
- Connects to Redis for caching (if configured)
- Processes MQTT messages from devices via RabbitMQ
- Exposes REST API endpoints for device management
- Provides WebSocket connections for real-time updates
- Handles health checks and monitoring

### Device Simulator

The simulator (`cmd/simulator/main.go`) creates virtual IoT devices that:

- Simulate weight changes (item removals) at random intervals
- Publish weight updates via MQTT
- Can be configured for different numbers of clients and devices

### Data Model

- **Clients**: Organizations or locations that own devices
- **Devices**: IoT weight sensors that track inventory
  - Track current item count and weight
  - Record total items sold
  - Have configurable item weights and maximum capacities

## API Endpoints

The system exposes the following REST API endpoints:

### Device Management

- `GET /server/v1/devices` - List all devices
- `GET /server/v1/devices/:deviceId` - Get a specific device
- `POST /server/v1/devices/initialize` - Initialize devices
- `GET /server/v1/clients/:clientId/devices` - List devices for a client

### Simulation

- `POST /server/v1/simulation/device/:deviceId/sale` - Simulate a sale for a device

### Monitoring

- `GET /server/v1/queue/stats` - Get RabbitMQ queue statistics
- `GET /health` - Health check endpoint
- `GET /ws` - WebSocket connection for real-time updates

## How the Simulation Works

1. The simulator creates virtual devices with initial inventory weights
2. Each device simulates sales by reducing weight at random intervals
3. Weight changes are published to MQTT topics (`devices/{deviceId}/weight`)
4. The server consumes these messages via RabbitMQ
5. The server updates device state in the database
6. Updates are broadcast to connected clients via WebSockets
7. The API also allows manual simulation of sales

## Setup and Installation

### Prerequisites

- Docker and Docker Compose
- Go 1.16 or higher (for development)
- PostgreSQL
- RabbitMQ
- MQTT broker (e.g., EMQX)

### Running with Docker Compose

```bash
# Start all services
docker compose -f docker-local.compose.yml up -d

# Or use the test script for sequential startup
./test_rabbitmq_connection.sh
```

## RabbitMQ Connection Troubleshooting

If you encounter RabbitMQ connection issues like:

```
Failed to connect to RabbitMQ: dial tcp 172.23.0.2:5672: connect: connection refused
```

This is typically caused by timing issues during container startup. The application tries to connect to RabbitMQ before the RabbitMQ service is fully initialized.

### Solution

The following improvements have been implemented to address this issue:

1. **Increased retry attempts**: The maximum number of connection retry attempts has been increased from 10 to 30, with exponential backoff.

2. **Non-blocking startup**: The application now continues to start even if the initial RabbitMQ connection fails, and will retry connecting in the background.

3. **Improved error logging**: More detailed error messages are now logged to help diagnose connection issues.

4. **Sequential service startup**: The test script now starts RabbitMQ first and waits 15 seconds before starting other services.

### Testing the Connection

Use the provided test script to verify the connection:

```bash
./test_rabbitmq_connection.sh
```

This script:
1. Stops any existing containers
2. Starts RabbitMQ first and waits for it to initialize
3. Starts the remaining services
4. Shows the application logs to observe connection attempts

### Verifying the Connection

To verify that the connection is working:

1. Check the logs for successful connection messages
2. Access the health endpoint: http://localhost:8080/health
3. Check the RabbitMQ management UI: http://localhost:15672 (login with guest/guest)

### Manual Startup Procedure

If you prefer to start services manually:

```bash
# Stop existing containers
docker compose -f docker-local.compose.yml down

# Start RabbitMQ first
docker compose -f docker-local.compose.yml up -d rabbitmq

# Wait for RabbitMQ to initialize (at least 15 seconds)
sleep 15

# Start the remaining services
docker compose -f docker-local.compose.yml up -d
```

This ensures that RabbitMQ has enough time to initialize before the application tries to connect to it.