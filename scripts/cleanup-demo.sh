#!/bin/bash

# MME Demo Cleanup Script

echo "🧹 Cleaning up MME Demo..."

# Stop services
docker-compose down

# Remove volumes (optional - uncomment to remove all data)
# docker-compose down -v

echo "✅ Cleanup completed!"
