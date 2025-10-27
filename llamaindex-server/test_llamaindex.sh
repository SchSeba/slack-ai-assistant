#!/bin/bash
# Test script for LlamaIndex server
# Usage: ./test_llamaindex.sh

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== LlamaIndex Server Test Suite ===${NC}\n"

# Check if server is running
SERVER_URL="${LLAMAINDEX_HOST:-http://localhost:5000}"
echo -e "${YELLOW}Testing server at: ${SERVER_URL}${NC}"

# Test 1: Health Check
echo -e "\n${GREEN}[1/4] Testing health endpoint...${NC}"
curl -s "${SERVER_URL}/health" | jq '.'
if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Health check passed${NC}"
else
    echo -e "${RED}✗ Health check failed${NC}"
    exit 1
fi

# Test 2: Answer endpoint (will fail without index, but tests the endpoint)
echo -e "\n${GREEN}[2/4] Testing /v1/answer endpoint...${NC}"
ANSWER_RESPONSE=$(curl -s -X POST "${SERVER_URL}/v1/answer" \
  -H "Content-Type: application/json" \
  -d '{
    "project": "sriov",
    "version": "4.16",
    "thread_slug": "test-thread-'$(date +%s)'",
    "message": "How do I configure SR-IOV?"
  }')

echo "$ANSWER_RESPONSE" | jq '.'

if echo "$ANSWER_RESPONSE" | jq -e '.textResponse' > /dev/null; then
    echo -e "${GREEN}✓ Answer endpoint returned a response${NC}"
else
    echo -e "${YELLOW}⚠ Answer endpoint returned error (may need indexes)${NC}"
fi

# Test 3: Elaborate endpoint
echo -e "\n${GREEN}[3/4] Testing /v1/elaborate endpoint...${NC}"
ELABORATE_RESPONSE=$(curl -s -X POST "${SERVER_URL}/v1/elaborate" \
  -H "Content-Type: application/json" \
  -d '{
    "thread_slug": "test-elaborate-'$(date +%s)'",
    "message": "SR-IOV is a virtualization technology that allows a single PCIe device to appear as multiple devices. It provides near-native performance for virtual machines."
  }')

echo "$ELABORATE_RESPONSE" | jq '.'

if echo "$ELABORATE_RESPONSE" | jq -e '.textResponse' > /dev/null; then
    echo -e "${GREEN}✓ Elaborate endpoint working${NC}"
else
    echo -e "${RED}✗ Elaborate endpoint failed${NC}"
fi

# Test 4: Inject endpoint
echo -e "\n${GREEN}[4/4] Testing /v1/inject endpoint...${NC}"
INJECT_RESPONSE=$(curl -s -X POST "${SERVER_URL}/v1/inject" \
  -H "Content-Type: application/json" \
  -d '{
    "project": "sriov",
    "version": "4.16",
    "textContent": "Test injection: SR-IOV requires hardware support and must be enabled in the BIOS.",
    "metadata": {"source": "test-script"}
  }')

echo "$INJECT_RESPONSE" | jq '.'

if echo "$INJECT_RESPONSE" | jq -e '.status' > /dev/null; then
    echo -e "${GREEN}✓ Inject endpoint working${NC}"
else
    echo -e "${RED}✗ Inject endpoint failed${NC}"
fi

echo -e "\n${GREEN}=== Test Suite Complete ===${NC}"

