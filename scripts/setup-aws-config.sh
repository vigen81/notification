#!/bin/bash
# File: scripts/setup-aws-config.sh
# Setup AWS Parameter Store configurations

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "🚀 Setting up AWS Parameter Store configurations..."

# Function to create or update parameter
create_parameter() {
    local env=$1
    local config_file=$2
    local parameter_name="/${env}/notification-engine"

    echo "📝 Creating parameter: $parameter_name"

    if ! aws ssm get-parameter --name "$parameter_name" --query "Parameter.Name" --output text >/dev/null 2>&1; then
        # Parameter doesn't exist, create it
        aws ssm put-parameter \
            --name "$parameter_name" \
            --value "$(cat "$config_file")" \
            --type "SecureString" \
            --description "Notification Engine Configuration for $env environment" \
            --tags "Key=Environment,Value=$env" "Key=Service,Value=notification-engine"
        echo "✅ Created parameter: $parameter_name"
    else
        # Parameter exists, update it
        echo "⚠️  Parameter $parameter_name already exists. Use --overwrite to update."
        read -p "Do you want to overwrite it? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            aws ssm put-parameter \
                --name "$parameter_name" \
                --value "$(cat "$config_file")" \
                --type "SecureString" \
                --overwrite
            echo "✅ Updated parameter: $parameter_name"
        else
            echo "⏭️  Skipped parameter: $parameter_name"
        fi
    fi
}

# Check if AWS CLI is configured
if ! aws sts get-caller-identity >/dev/null 2>&1; then
    echo "❌ AWS CLI is not configured or credentials are invalid"
    echo "Please run: aws configure"
    exit 1
fi

echo "🔍 AWS Identity: $(aws sts get-caller-identity --query "Arn" --output text)"

# Create configurations for each environment
echo
echo "📁 Setting up configurations..."

if [ -f "$SCRIPT_DIR/dev-config.json" ]; then
    create_parameter "dev" "$SCRIPT_DIR/dev-config.json"
else
    echo "⚠️  dev-config.json not found"
fi

if [ -f "$SCRIPT_DIR/staging-config.json" ]; then
    create_parameter "staging" "$SCRIPT_DIR/staging-config.json"
else
    echo "⚠️  staging-config.json not found"
fi

if [ -f "$SCRIPT_DIR/prod-config.json" ]; then
    create_parameter "prod" "$SCRIPT_DIR/prod-config.json"
else
    echo "⚠️  prod-config.json not found"
fi

echo
echo "🎉 AWS Parameter Store setup complete!"
echo
echo "📋 Next steps:"
echo "1. Update the configuration files with your actual values"
echo "2. Re-run this script to update the parameters"
echo "3. Deploy your application with POD_ENV=dev|staging|prod"

# ===================================
# File: scripts/deploy.sh
# Deployment script for different environments

#!/bin/bash

set -e

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
IMAGE_NAME="notification-engine"
AWS_REGION=${AWS_REGION:-eu-central-1}

# Function to display usage
usage() {
    echo "Usage: $0 <environment> [options]"
    echo
    echo "Environments:"
    echo "  local     - Local development with Docker Compose"
    echo "  dev       - Development environment"
    echo "  staging   - Staging environment"
    echo "  prod      - Production environment"
    echo
    echo "Options:"
    echo "  --build-only    Build image without deploying"
    echo "  --no-build      Deploy without building"
    echo "  --force         Force deployment without confirmation"
    echo
    exit 1
}

# Parse arguments
ENVIRONMENT=""
BUILD_ONLY=false
NO_BUILD=false
FORCE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        local|dev|staging|prod)
            ENVIRONMENT="$1"
            shift
            ;;
        --build-only)
            BUILD_ONLY=true
            shift
            ;;
        --no-build)
            NO_BUILD=true
            shift
            ;;
        --force)
            FORCE=true
            shift
            ;;
        -h|--help)
            usage
            ;;
        *)
            echo "Unknown option: $1"
            usage
            ;;
    esac
done

if [ -z "$ENVIRONMENT" ]; then
    echo "❌ Environment is required"
    usage
fi

echo "🚀 Deploying Notification Engine to $ENVIRONMENT environment"

# Build Docker image
if [ "$NO_BUILD" = false ]; then
    echo "📦 Building Docker image..."
    cd "$PROJECT_ROOT"

    docker build -t "$IMAGE_NAME:$ENVIRONMENT" .
    docker tag "$IMAGE_NAME:$ENVIRONMENT" "$IMAGE_NAME:latest"

    echo "✅ Image built: $IMAGE_NAME:$ENVIRONMENT"
fi

if [ "$BUILD_ONLY" = true ]; then
    echo "🎉 Build complete! Image: $IMAGE_NAME:$ENVIRONMENT"
    exit 0
fi

# Deploy based on environment
case $ENVIRONMENT in
    local)
        echo "🏠 Deploying locally with Docker Compose..."
        cd "$PROJECT_ROOT"
        POD_ENV=local docker-compose up -d
        echo "✅ Local deployment complete!"
        echo "🌐 API: http://localhost:8080"
        echo "📊 Kafka UI: http://localhost:8081"
        ;;

    dev|staging|prod)
        if [ "$FORCE" = false ]; then
            echo "⚠️  You are about to deploy to $ENVIRONMENT environment"
            read -p "Are you sure? (y/N): " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                echo "❌ Deployment cancelled"
                exit 1
            fi
        fi

        echo "☁️  Deploying to $ENVIRONMENT environment..."

        # Add your cloud deployment logic here
        # Examples:
        # - AWS ECS/Fargate deployment
        # - Kubernetes deployment
        # - AWS Lambda deployment
        # - Docker Swarm deployment

        echo "🚧 Cloud deployment not implemented yet"
        echo "💡 Add your deployment logic to scripts/deploy.sh"
        echo
        echo "Example deployment commands:"
        echo "  AWS ECS: aws ecs update-service --cluster notification-cluster --service notification-engine --force-new-deployment"
        echo "  Kubernetes: kubectl set image deployment/notification-engine notification-engine=$IMAGE_NAME:$ENVIRONMENT"
        ;;
esac

echo "🎉 Deployment complete!"

# ===================================
# File: scripts/check-config.sh
# Script to validate configurations

#!/bin/bash

set -e

echo "🔍 Checking AWS Parameter Store configurations..."

check_parameter() {
    local env=$1
    local parameter_name="/${env}/notification-engine"

    echo "📋 Checking $parameter_name..."

    if aws ssm get-parameter --name "$parameter_name" --query "Parameter.Name" --output text >/dev/null 2>&1; then
        echo "✅ Parameter exists: $parameter_name"

        # Validate JSON format
        if aws ssm get-parameter --name "$parameter_name" --with-decryption --query "Parameter.Value" --output text | jq . >/dev/null 2>&1; then
            echo "✅ Valid JSON format"
        else
            echo "❌ Invalid JSON format"
        fi
    else
        echo "❌ Parameter not found: $parameter_name"
    fi
    echo
}

# Check each environment
for env in dev staging prod; do
    check_parameter "$env"
done

echo "🎉 Configuration check complete!"