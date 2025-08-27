#!/bin/bash

# MME Staging Deployment Script
# This script deploys the MME system to the staging environment

set -e

echo "🚀 Starting MME Staging Deployment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
STATIC_IP=${STATIC_IP:-"34.58.167.157"}
ENV_FILE="deployment/.env"
COMPOSE_FILE="deployment/docker-compose.staging.yml"

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Setup environment
setup_environment() {
    print_status "Setting up environment..."
    
    if [ ! -f "$ENV_FILE" ]; then
        print_warning "Environment file not found, creating from template..."
        cp deployment/config/env.staging "$ENV_FILE"
        print_warning "Please update $ENV_FILE with your actual values before continuing"
        print_warning "Especially update OPENAI_API_KEY with a real key"
        exit 1
    fi
    
    # Source environment variables
    export $(cat "$ENV_FILE" | grep -v '^#' | xargs)
    
    print_success "Environment setup completed"
}

# Create necessary directories
create_directories() {
    print_status "Creating necessary directories..."
    
    mkdir -p deployment/logs
    mkdir -p deployment/monitoring/file_sd
    mkdir -p deployment/traefik
    mkdir -p deployment/scripts
    
    # Set proper permissions
    chmod 755 deployment/logs
    chmod 755 deployment/monitoring/file_sd
    chmod 644 deployment/traefik/acme.json 2>/dev/null || touch deployment/traefik/acme.json && chmod 644 deployment/traefik/acme.json
    
    print_success "Directories created"
}

# Build and deploy services
deploy_services() {
    print_status "Building and deploying services..."
    
    cd deployment
    
    # Pull latest images
    print_status "Pulling latest images..."
    docker-compose -f docker-compose.staging.yml pull
    
    # Build services
    print_status "Building services..."
    docker-compose -f docker-compose.staging.yml build --no-cache
    
    # Start services
    print_status "Starting services..."
    docker-compose -f docker-compose.staging.yml up -d
    
    cd ..
    
    print_success "Services deployed"
}

# Wait for services to be healthy
wait_for_services() {
    print_status "Waiting for services to be healthy..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        print_status "Health check attempt $attempt/$max_attempts"
        
        # Check if all services are running
        if docker-compose -f deployment/docker-compose.staging.yml ps | grep -q "Up"; then
            print_success "All services are running"
            break
        fi
        
        if [ $attempt -eq $max_attempts ]; then
            print_error "Services failed to start within expected time"
            docker-compose -f deployment/docker-compose.staging.yml logs
            exit 1
        fi
        
        sleep 10
        attempt=$((attempt + 1))
    done
}

# Run health checks
run_health_checks() {
    print_status "Running health checks..."
    
    # Test MME service
    print_status "Testing MME service..."
    if curl -s -H "Host: mme.${STATIC_IP}.nip.io" "http://${STATIC_IP}/health" | grep -q "healthy"; then
        print_success "MME service is healthy"
    else
        print_error "MME service health check failed"
        return 1
    fi
    
    # Test Tagmaker service
    print_status "Testing Tagmaker service..."
    if curl -s -H "Host: tagmaker.${STATIC_IP}.nip.io" "http://${STATIC_IP}/health" | grep -q "healthy"; then
        print_success "Tagmaker service is healthy"
    else
        print_error "Tagmaker service health check failed"
        return 1
    fi
    
    # Test Traefik dashboard
    print_status "Testing Traefik dashboard..."
    if curl -s -H "Host: traefik.${STATIC_IP}.nip.io" "http://${STATIC_IP}/api/http/routers" > /dev/null; then
        print_success "Traefik dashboard is accessible"
    else
        print_warning "Traefik dashboard health check failed"
    fi
    
    print_success "Health checks completed"
}

# Display deployment information
display_info() {
    print_success "🎉 MME Staging Deployment Completed Successfully!"
    echo
    echo "📋 Deployment Information:"
    echo "   Static IP: ${STATIC_IP}"
    echo "   Environment: staging"
    echo
    echo "🌐 Service URLs:"
    echo "   MME API: http://mme.${STATIC_IP}.nip.io"
    echo "   Tagmaker: http://tagmaker.${STATIC_IP}.nip.io"
    echo "   Traefik Dashboard: http://traefik.${STATIC_IP}.nip.io"
    echo "   Grafana: http://grafana.${STATIC_IP}.nip.io (admin/admin)"
    echo "   Prometheus: http://prometheus.${STATIC_IP}.nip.io"
    echo "   Alertmanager: http://alertmanager.${STATIC_IP}.nip.io"
    echo "   Documentation: http://docs.${STATIC_IP}.nip.io"
    echo
    echo "🔧 Management Commands:"
    echo "   View logs: docker-compose -f deployment/docker-compose.staging.yml logs -f"
    echo "   Stop services: docker-compose -f deployment/docker-compose.staging.yml down"
    echo "   Restart services: docker-compose -f deployment/docker-compose.staging.yml restart"
    echo "   View status: docker-compose -f deployment/docker-compose.staging.yml ps"
    echo
    echo "⚠️  Important Notes:"
    echo "   - Update OPENAI_API_KEY in deployment/.env with a real key"
    echo "   - Monitor logs for any issues"
    echo "   - Check Grafana dashboards for system metrics"
    echo
}

# Main deployment flow
main() {
    echo "=========================================="
    echo "   MME Staging Environment Deployment"
    echo "=========================================="
    echo
    
    check_prerequisites
    setup_environment
    create_directories
    deploy_services
    wait_for_services
    run_health_checks
    display_info
    
    print_success "Deployment completed successfully!"
}

# Run main function
main "$@"
