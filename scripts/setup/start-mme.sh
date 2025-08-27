#!/bin/bash

# MME Services Startup Script
# This script starts the standalone MME (Memory Management Engine) services

set -e

echo "🧠 Starting MME (Memory Management Engine) Services..."

# Check if required environment file exists
if [ ! -f ".env" ]; then
    echo "❌ Environment file not found. Creating from template..."
    cp .env.example .env
    echo "⚠️  Please edit .env with your configuration:"
    echo "   - OPENAI_API_KEY (required)"
    echo "   - JWT_SECRET (required)"
    echo ""
    echo "Then run this script again."
    exit 1
fi

# Check if Docker is running
if ! docker info >/dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

# Check if OpenAI API key is set
if ! grep -q "OPENAI_API_KEY=sk-" .env 2>/dev/null; then
    echo "⚠️  Warning: OPENAI_API_KEY not found in .env"
    echo "   Tag extraction will fail without a valid OpenAI API key"
fi

echo "🐳 Starting MME services with Docker Compose..."

# Start MME services
docker-compose up -d

echo "⏳ Waiting for services to be healthy..."

# Wait for services to be healthy
max_attempts=30
attempt=0

while [ $attempt -lt $max_attempts ]; do
    if docker-compose ps | grep -q "healthy"; then
        # Check if both services are healthy
        tagging_health=$(docker inspect --format='{{.State.Health.Status}}' mme-tagging-service 2>/dev/null || echo "starting")
        tagmaker_health=$(docker inspect --format='{{.State.Health.Status}}' mme-tagmaker-service 2>/dev/null || echo "starting")
        
        if [ "$tagging_health" = "healthy" ] && [ "$tagmaker_health" = "healthy" ]; then
            echo "✅ All MME services are healthy!"
            break
        fi
    fi
    
    echo "   Still waiting... (attempt $((attempt + 1))/$max_attempts)"
    sleep 10
    attempt=$((attempt + 1))
done

if [ $attempt -eq $max_attempts ]; then
    echo "⚠️  Services may still be starting. Check status with:"
    echo "   docker-compose ps"
    echo "   docker-compose logs"
fi

echo ""
echo "🎉 MME Services Status:"
echo "   🏷️  MME Tagging Service:  http://localhost:8081/health"
echo "   🤖 MME Tagmaker Service: http://localhost:8000/health"
echo "   📊 Tagmaker API Docs:    http://localhost:8000/docs"
echo "   🗄️  MongoDB:             localhost:27017"
echo "   🔴 Redis:               localhost:6379"
echo ""
echo "📖 Quick API Test:"
echo "   curl http://localhost:8081/health"
echo "   curl http://localhost:8000/health"
echo ""
echo "🛑 To stop services:"
echo "   docker-compose down"
echo ""
echo "📚 See README.md for detailed usage instructions."
