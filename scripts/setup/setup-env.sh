#!/bin/bash
# MME Environment Setup Script
# Generates secure passwords and creates .env file

echo "🔧 MME Environment Setup"
echo "======================="

# Check if .env already exists
if [ -f ".env" ]; then
    echo "⚠️  .env file already exists!"
    read -p "Do you want to overwrite it? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Setup cancelled."
        exit 1
    fi
fi

echo "🔐 Generating secure passwords..."

# Generate secure passwords
MONGODB_ROOT_PASSWORD=$(openssl rand -base64 32)
JWT_SECRET=$(openssl rand -base64 64)
REDIS_PASSWORD=$(openssl rand -base64 16)

echo "✅ Passwords generated successfully!"
echo

# Create .env file
cat > .env << EOF
# MME (Multi-Modal Memory Extractor) Environment Configuration
# Generated on $(date)

# =============================================================================
# SECURITY CONFIGURATION (REQUIRED)
# =============================================================================

# MongoDB Configuration
MONGODB_ROOT_PASSWORD=${MONGODB_ROOT_PASSWORD}

# JWT Secret for authentication
JWT_SECRET=${JWT_SECRET}

# Redis Password (Optional - for caching)
REDIS_PASSWORD=${REDIS_PASSWORD}

# =============================================================================
# API KEYS (REQUIRED)
# =============================================================================

# OpenAI API Key (Required for AI features)
# Get from: https://platform.openai.com/api-keys
OPENAI_API_KEY=your_openai_api_key_here

# =============================================================================
# SERVICE CONFIGURATION (OPTIONAL)
# =============================================================================

# CORS Configuration
CORS_ALLOWED_ORIGINS=*

ENABLE_TAGGING_SERVICE=true
LOG_LEVEL=INFO

# =============================================================================
# MME ALGORITHM PARAMETERS (OPTIONAL - Has defaults)
# =============================================================================

# Memory extraction parameters
MME_MAX_EDGES_PER_TAG=32
MME_MAX_DEPTH=2
MME_BEAM_WIDTH=128
MME_MAX_SEED_TAGS=5
MME_DECAY_ALPHA=0.85
MME_MIN_ACTIVATION=0.05
MME_TOKEN_BUDGET=2048
MME_DIVERSITY_LAMBDA=0.15
MME_RECENCY_TAU_DAYS=60
MME_CACHE_TTL_SECS=300

# Edge learning parameters
MME_LEARN_ETA=0.05
MME_LEARN_R=0.03
MME_LEARN_D=0.01
MME_LEARN_WMAX=1.0
MME_LEARN_WINDOW_HOURS=24
EOF

echo "📄 .env file created successfully!"
echo
echo "🔑 Next steps:"
echo "1. Edit .env and replace 'your_openai_api_key_here' with your actual OpenAI API key"
echo "2. Get your OpenAI API key from: https://platform.openai.com/api-keys"
echo "3. Run: docker-compose up -d"
echo
echo "🎉 Setup complete!"