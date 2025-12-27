#!/bin/bash

# Simple Asynq Demo Runner
# Starts Redis (if needed) and runs the demo

set -e

echo "🚀 Starting Asynq Demo..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go first."
    exit 1
fi

# Check if Redis is running
if ! redis-cli ping &> /dev/null; then
    echo "⚠️  Redis is not running. Starting Redis with Docker..."
    if command -v docker &> /dev/null; then
        docker run -d --name asynq-redis -p 6380:6379 redis:7-alpine
        echo "✅ Redis started with Docker"
        sleep 2
    else
        echo "❌ Docker is not available. Please start Redis manually on localhost:6380"
        exit 1
    fi
else
    echo "✅ Redis is already running"
fi

# Install/update dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Build the demo
echo "🔨 Building demo..."
go build -o bin/asynq-demo main.go

# Run the demo
echo "🎯 Running demo..."
./bin/asynq-demo

# Cleanup
if docker ps -a | grep -q asynq-redis; then
    echo "🧹 Cleaning up Redis container..."
    docker stop asynq-redis
    docker rm asynq-redis
fi

echo "🎉 Demo completed!"
