#!/bin/bash

# Push Service Images Script
# This script builds and pushes all microservice images to a Docker registry

set -e  # Exit on error

# Configuration
AWS_ACCOUNT_ID="${AWS_ACCOUNT_ID}"
AWS_REGION="${AWS_REGION:-us-east-1}"
ECR_REPO_NAME="${ECR_REPO_NAME:-english-app}"
TAG="${IMAGE_TAG:-latest}"

# Construct ECR registry URL
REGISTRY="${AWS_ACCOUNT_ID}.dkr.ecr.${AWS_REGION}.amazonaws.com"
NAMESPACE="${ECR_REPO_NAME}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Services to build and push
SERVICES=(
  "user-services"
  "lesson-services"
  "content-services"
  "notification-services"
  "bff-services"
)

# Function to print colored messages
print_message() {
  local color=$1
  local message=$2
  echo -e "${color}${message}${NC}"
}

# Function to build and push a service
build_and_push() {
  local service=$1
  local image_name="${REGISTRY}/${NAMESPACE}/${service}:${TAG}"
  
  print_message "$YELLOW" "Building ${service}..."
  
  if [ ! -d "${service}" ]; then
    print_message "$RED" "Error: Directory ${service} not found!"
    return 1
  fi
  
  # Build the image
  docker build -t "${image_name}" -f "${service}/Dockerfile" "${service}"
  
  if [ $? -ne 0 ]; then
    print_message "$RED" "Failed to build ${service}"
    return 1
  fi
  
  print_message "$GREEN" "Successfully built ${image_name}"
  
  # Push the image
  print_message "$YELLOW" "Pushing ${image_name}..."
  docker push "${image_name}"
  
  if [ $? -ne 0 ]; then
    print_message "$RED" "Failed to push ${service}"
    return 1
  fi
  
  print_message "$GREEN" "Successfully pushed ${image_name}"
  echo ""
}

# Main execution
main() {
  print_message "$GREEN" "========================================="
  print_message "$GREEN" "Push Images to AWS ECR"
  print_message "$GREEN" "========================================="
  echo ""
  
  # Validate required variables
  if [ -z "${AWS_ACCOUNT_ID}" ]; then
    print_message "$RED" "Error: AWS_ACCOUNT_ID not set"
    exit 1
  fi
  
  print_message "$YELLOW" "AWS Account: ${AWS_ACCOUNT_ID}"
  print_message "$YELLOW" "AWS Region: ${AWS_REGION}"
  print_message "$YELLOW" "ECR Repository: ${ECR_REPO_NAME}"
  print_message "$YELLOW" "Registry: ${REGISTRY}"
  print_message "$YELLOW" "Tag: ${TAG}"
  echo ""
  
  # Check if Docker is running
  if ! docker info > /dev/null 2>&1; then
    print_message "$RED" "Error: Docker is not running!"
    exit 1
  fi
  
  # Check if AWS CLI is installed
  if ! command -v aws &> /dev/null; then
    print_message "$RED" "Error: AWS CLI is not installed!"
    exit 1
  fi
  
  # Login to ECR
  print_message "$YELLOW" "Logging in to ECR..."
  aws ecr get-login-password --region "${AWS_REGION}" | docker login --username AWS --password-stdin "${REGISTRY}" || {
    print_message "$RED" "Failed to login to ECR"
    exit 1
  }
  print_message "$GREEN" "Successfully logged in to ECR"
  echo ""
  
  # Build and push each service
  for service in "${SERVICES[@]}"; do
    if ! build_and_push "$service"; then
      print_message "$RED" "Failed to process ${service}, stopping..."
      exit 1
    fi
  done
  
  print_message "$GREEN" "========================================="
  print_message "$GREEN" "All images built and pushed successfully!"
  print_message "$GREEN" "========================================="
}

# Run main function
main
