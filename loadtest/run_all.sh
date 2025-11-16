#!/usr/bin/env bash

#
# run_all.sh - Automated Load Testing Runner
#
# This script runs all k6 load tests in sequence:
# 1. Seed test data
# 2. Test stats endpoint (read-only)
# 3. Test PR creation endpoint (write-heavy)
#
# Usage:
#   ./run_all.sh
#   BASE_URL=http://localhost:8080 ./run_all.sh
#

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL=${BASE_URL:-http://localhost:8080}
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  PR Reviewer Service - Load Testing${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""
echo -e "Base URL: ${GREEN}${BASE_URL}${NC}"
echo -e "Scripts directory: ${SCRIPT_DIR}"
echo ""

# Check if k6 is installed
if ! command -v k6 &> /dev/null; then
    echo -e "${RED}Error: k6 is not installed${NC}"
    echo "Please install k6 from https://k6.io/docs/get-started/installation/"
    echo ""
    echo "macOS:   brew install k6"
    echo "Linux:   See https://k6.io/docs/get-started/installation/"
    echo "Windows: choco install k6"
    exit 1
fi

echo -e "${GREEN}✓ k6 is installed${NC}"
echo ""

# Function to run a k6 script
run_test() {
    local script_name=$1
    local description=$2
    
    echo -e "${BLUE}========================================${NC}"
    echo -e "${YELLOW}Running: ${description}${NC}"
    echo -e "${BLUE}========================================${NC}"
    echo ""
    
    if [ ! -f "${SCRIPT_DIR}/${script_name}" ]; then
        echo -e "${RED}Error: Script ${script_name} not found${NC}"
        exit 1
    fi
    
    if BASE_URL="${BASE_URL}" k6 run "${SCRIPT_DIR}/${script_name}"; then
        echo ""
        echo -e "${GREEN}✓ ${description} completed successfully${NC}"
        echo ""
        return 0
    else
        echo ""
        echo -e "${RED}✗ ${description} failed${NC}"
        echo ""
        return 1
    fi
}

# Check service health
echo -e "${YELLOW}Checking service health...${NC}"
if command -v curl &> /dev/null; then
    if curl -f -s "${BASE_URL}/healthz" > /dev/null; then
        echo -e "${GREEN}✓ Service is healthy${NC}"
        echo ""
    else
        echo -e "${RED}✗ Service health check failed${NC}"
        echo "Please ensure the service is running:"
        echo "  docker compose up --build"
        exit 1
    fi
else
    echo -e "${YELLOW}⚠ curl not found, skipping health check${NC}"
    echo ""
fi

# Initialize results
FAILED_TESTS=()

# Step 1: Seed test data
if ! run_test "seed_data.js" "Seed Test Data"; then
    FAILED_TESTS+=("Seed Test Data")
fi

sleep 2

# Step 2: Test stats endpoint
if ! run_test "stats_assignments.js" "Stats Endpoint Load Test"; then
    FAILED_TESTS+=("Stats Endpoint Load Test")
fi

sleep 2

# Step 3: Test PR creation endpoint
if ! run_test "create_pr.js" "PR Creation Load Test"; then
    FAILED_TESTS+=("PR Creation Load Test")
fi

# Final summary
echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Load Testing Summary${NC}"
echo -e "${BLUE}========================================${NC}"
echo ""

if [ ${#FAILED_TESTS[@]} -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed successfully!${NC}"
    echo ""
    echo "Performance requirements met:"
    echo "  • Response time SLI: p(95) < 300ms"
    echo "  • Success rate SLI: > 99.9%"
    echo "  • Target RPS: 5 req/s"
    exit 0
else
    echo -e "${RED}✗ Some tests failed:${NC}"
    for test in "${FAILED_TESTS[@]}"; do
        echo -e "  ${RED}•${NC} ${test}"
    done
    echo ""
    echo "Please review the logs above for details."
    exit 1
fi
