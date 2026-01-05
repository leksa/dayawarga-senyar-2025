#!/bin/bash
# ===========================================
# Dayawarga Senyar - Deployment Script
# ===========================================

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  Dayawarga Senyar - Deployment${NC}"
echo -e "${GREEN}========================================${NC}"

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${RED}Error: .env file not found!${NC}"
    echo "Please copy .env.example to .env and configure it."
    exit 1
fi

# Pull latest changes
echo -e "\n${YELLOW}Pulling latest changes...${NC}"
git pull origin main

# Build services
echo -e "\n${YELLOW}Building Docker images...${NC}"
docker compose build

# Start services
echo -e "\n${YELLOW}Starting services...${NC}"
docker compose up -d

# Wait for health checks
echo -e "\n${YELLOW}Waiting for services to be healthy...${NC}"
sleep 10

# Show status
echo -e "\n${YELLOW}Service status:${NC}"
docker compose ps

# Cleanup
echo -e "\n${YELLOW}Cleaning up old images...${NC}"
docker image prune -f

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}  Deployment completed!${NC}"
echo -e "${GREEN}========================================${NC}"

echo -e "\nServices available at:"
echo -e "  - Frontend:  https://dayawarga.com"
echo -e "  - API:       https://api.dayawarga.com"
echo -e "  - Blog:      https://stories.dayawarga.com"
