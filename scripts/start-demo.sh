#!/bin/bash

# MME Demo Startup Script

set -e

echo "🚀 Starting MME Demo..."

# Check if .env exists
if [ ! -f .env ]; then
    echo "⚠️  .env file not found. Creating from template..."
    cp .env.demo .env
    echo "📝 Please edit .env with your OpenAI API key and other settings"
    echo "   Then run this script again."
    exit 1
fi

# Check Docker
if ! command -v docker &> /dev/null; then
    echo "❌ Docker not found. Please install Docker first."
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose not found. Please install Docker Compose first."
    exit 1
fi

# Start services
echo "🐳 Starting services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 30

# Check service health
echo "🔍 Checking service health..."
docker-compose ps

# Test basic functionality
echo "🧪 Testing basic functionality..."
if curl -s -H "Host: mme.localhost" -H "X-User-ID: demo_user" "http://localhost/health" | grep -q "ok"; then
    echo "✅ MME is running and healthy!"
else
    echo "❌ MME health check failed"
    exit 1
fi

echo ""
echo "🎉 MME Demo is ready!"
echo ""
echo "📊 Demo URLs:"
echo "   MME API: http://localhost/"
echo "   Grafana: http://localhost:3000 (admin/admin)"
echo "   Prometheus: http://localhost:9090"
echo "   Traefik Dashboard: http://localhost:9000"
echo ""
echo "🧪 Quick test:"
echo "   curl -H \"Host: mme.localhost\" -H \"X-User-ID: demo_user\" \\"
echo "     \"http://localhost/health\""
echo ""
echo "📚 Documentation: docs/README.md"
