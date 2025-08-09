#!/bin/bash

echo "Testing RabbitMQ connection improvements"
echo "---------------------------------------"

# Stop any existing containers
echo "Stopping existing containers..."
docker compose -f docker-local.compose.yml down

# Start RabbitMQ first to give it time to initialize
echo "Starting RabbitMQ service first..."
docker compose -f docker-local.compose.yml up -d rabbitmq
echo "Waiting 15 seconds for RabbitMQ to initialize..."
sleep 15

# Start the rest of the services
echo "Starting remaining services..."
docker compose -f docker-local.compose.yml up -d

# Follow the logs to see connection attempts
echo "Following logs to observe connection attempts..."
echo "Press Ctrl+C to stop watching logs"
docker logs -f iot-inventory-management-local

# Instructions for verification
echo ""
echo "To verify the connection is working, check:"
echo "1. The logs should show successful connection to RabbitMQ"
echo "2. Access the health endpoint: http://localhost:8080/health"
echo "3. Check RabbitMQ management UI: http://localhost:15672 (guest/guest)"