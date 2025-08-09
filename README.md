# IoT Inventory Management

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