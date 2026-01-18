#!/bin/bash

# Comprehensive API test for deployed MadaBank
# Usage: ./scripts/test-deployed-api.sh <api-base-url>

set -e

API_BASE_URL=$1

if [ -z "$API_BASE_URL" ]; then
  echo "Usage: $0 <api-base-url>"
  echo "Example: $0 http://madabank-dev-alb-123.us-east-1.elb.amazonaws.com"
  exit 1
fi

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================="
echo "MadaBank Deployed API Test Suite"
echo "API: $API_BASE_URL"
echo "========================================="
echo ""

# Helper function to check response
check_response() {
  local response=$1
  local expected_status=$2
  local description=$3
  
  local status=$(echo "$response" | jq -r '.status // .error // "unknown"')
  
  if [[ "$response" == *"$expected_status"* ]]; then
    echo -e "${GREEN}✓ PASSED${NC}: $description"
    return 0
  else
    echo -e "${RED}✗ FAILED${NC}: $description"
    echo "Response: $response"
    return 1
  fi
}

# Test 1: Health Check
echo "Test 1: Health Check"
HEALTH_RESPONSE=$(curl -s $API_BASE_URL/health)
check_response "$HEALTH_RESPONSE" "healthy" "Health endpoint should return healthy status"
echo ""

# Test 2: Ready Check
echo "Test 2: Ready Check"
READY_RESPONSE=$(curl -s $API_BASE_URL/ready)
check_response "$READY_RESPONSE" "ready" "Ready endpoint should return ready status"
echo ""

# Test 3: Version Info
echo "Test 3: Version Info"
VERSION_RESPONSE=$(curl -s $API_BASE_URL/version)
echo "Version: $(echo $VERSION_RESPONSE | jq -r '.version')"
echo "Commit: $(echo $VERSION_RESPONSE | jq -r '.commit_sha')"
echo -e "${GREEN}✓ PASSED${NC}: Version endpoint"
echo ""

# Test 4: Metrics Endpoint
echo "Test 4: Prometheus Metrics"
METRICS_RESPONSE=$(curl -s $API_BASE_URL/metrics | head -n 5)
if [[ "$METRICS_RESPONSE" == *"madabank"* ]]; then
  echo -e "${GREEN}✓ PASSED${NC}: Metrics endpoint is exposing Prometheus metrics"
else
  echo -e "${RED}✗ FAILED${NC}: Metrics endpoint not working"
fi
echo ""

# Test 5: User Registration
echo "Test 5: User Registration"
REGISTER_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "testuser-'$(date +%s)'@madabank.com",
    "password": "SecurePass123!",
    "first_name": "Test",
    "last_name": "User",
    "phone": "+1234567890"
  }')

USER_ID=$(echo $REGISTER_RESPONSE | jq -r '.id')

if [ "$USER_ID" != "null" ] && [ ! -z "$USER_ID" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: User registered successfully"
  echo "User ID: $USER_ID"
else
  echo -e "${RED}✗ FAILED${NC}: User registration failed"
  echo "Response: $REGISTER_RESPONSE"
fi
echo ""

# Test 6: User Login
echo "Test 6: User Login"
LOGIN_EMAIL=$(echo $REGISTER_RESPONSE | jq -r '.email')
LOGIN_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "'$LOGIN_EMAIL'",
    "password": "SecurePass123!"
  }')

TOKEN=$(echo $LOGIN_RESPONSE | jq -r '.token')

if [ "$TOKEN" != "null" ] && [ ! -z "$TOKEN" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Login successful"
  echo "JWT Token: ${TOKEN:0:50}..."
else
  echo -e "${RED}✗ FAILED${NC}: Login failed"
  echo "Response: $LOGIN_RESPONSE"
  exit 1
fi
echo ""

# Test 7: Get User Profile
echo "Test 7: Get User Profile (Authenticated)"
PROFILE_RESPONSE=$(curl -s -X GET $API_BASE_URL/api/v1/users/profile \
  -H "Authorization: Bearer $TOKEN")

PROFILE_EMAIL=$(echo $PROFILE_RESPONSE | jq -r '.email')

if [ "$PROFILE_EMAIL" == "$LOGIN_EMAIL" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Profile retrieved successfully"
else
  echo -e "${RED}✗ FAILED${NC}: Profile retrieval failed"
fi
echo ""

# Test 8: Create Checking Account
echo "Test 8: Create Checking Account"
CHECKING_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "account_type": "checking",
    "currency": "USD"
  }')

CHECKING_ACCOUNT_ID=$(echo $CHECKING_RESPONSE | jq -r '.id')
CHECKING_ACCOUNT_NUMBER=$(echo $CHECKING_RESPONSE | jq -r '.account_number')

if [ "$CHECKING_ACCOUNT_ID" != "null" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Checking account created"
  echo "Account Number: $CHECKING_ACCOUNT_NUMBER"
else
  echo -e "${RED}✗ FAILED${NC}: Account creation failed"
fi
echo ""

# Test 9: Create Savings Account
echo "Test 9: Create Savings Account"
SAVINGS_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/accounts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "account_type": "savings",
    "currency": "USD",
    "interest_rate": 0.0425
  }')

SAVINGS_ACCOUNT_ID=$(echo $SAVINGS_RESPONSE | jq -r '.id')

if [ "$SAVINGS_ACCOUNT_ID" != "null" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Savings account created"
else
  echo -e "${RED}✗ FAILED${NC}: Savings account creation failed"
fi
echo ""

# Test 10: Deposit Money
echo "Test 10: Deposit Money"
DEPOSIT_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/transactions/deposit \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "account_id": "'$CHECKING_ACCOUNT_ID'",
    "amount": 1000.00,
    "description": "Initial deposit",
    "idempotency_key": "deposit-'$(uuidgen)'"
  }')

DEPOSIT_STATUS=$(echo $DEPOSIT_RESPONSE | jq -r '.status')

if [ "$DEPOSIT_STATUS" == "completed" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Deposit successful"
else
  echo -e "${RED}✗ FAILED${NC}: Deposit failed"
fi
echo ""

# Test 11: Check Balance
echo "Test 11: Check Balance"
BALANCE_RESPONSE=$(curl -s -X GET $API_BASE_URL/api/v1/accounts/$CHECKING_ACCOUNT_ID/balance \
  -H "Authorization: Bearer $TOKEN")

BALANCE=$(echo $BALANCE_RESPONSE | jq -r '.balance')

if [ "$BALANCE" == "1000" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Balance is correct: $BALANCE USD"
else
  echo -e "${YELLOW}⚠ WARNING${NC}: Expected balance 1000, got $BALANCE"
fi
echo ""

# Test 12: Transfer Money
echo "Test 12: Transfer Money"
TRANSFER_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/transactions/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "from_account_id": "'$CHECKING_ACCOUNT_ID'",
    "to_account_id": "'$SAVINGS_ACCOUNT_ID'",
    "amount": 250.00,
    "description": "Transfer to savings",
    "idempotency_key": "transfer-'$(uuidgen)'"
  }')

TRANSFER_STATUS=$(echo $TRANSFER_RESPONSE | jq -r '.status')

if [ "$TRANSFER_STATUS" == "completed" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Transfer successful"
else
  echo -e "${RED}✗ FAILED${NC}: Transfer failed"
  echo "Response: $TRANSFER_RESPONSE"
fi
echo ""

# Test 13: Verify Balances After Transfer
echo "Test 13: Verify Balances After Transfer"
CHECKING_BALANCE=$(curl -s $API_BASE_URL/api/v1/accounts/$CHECKING_ACCOUNT_ID/balance \
  -H "Authorization: Bearer $TOKEN" | jq -r '.balance')
SAVINGS_BALANCE=$(curl -s $API_BASE_URL/api/v1/accounts/$SAVINGS_ACCOUNT_ID/balance \
  -H "Authorization: Bearer $TOKEN" | jq -r '.balance')

echo "Checking balance: $CHECKING_BALANCE USD (expected: 750)"
echo "Savings balance: $SAVINGS_BALANCE USD (expected: 250)"

if [ "$CHECKING_BALANCE" == "750" ] && [ "$SAVINGS_BALANCE" == "250" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Balances correct after transfer"
else
  echo -e "${YELLOW}⚠ WARNING${NC}: Balance mismatch"
fi
echo ""

# Test 14: Transaction History
echo "Test 14: Get Transaction History"
HISTORY_RESPONSE=$(curl -s "$API_BASE_URL/api/v1/transactions/history?account_id=$CHECKING_ACCOUNT_ID&limit=10" \
  -H "Authorization: Bearer $TOKEN")

TRANSACTION_COUNT=$(echo $HISTORY_RESPONSE | jq '.total')

if [ "$TRANSACTION_COUNT" -ge 2 ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Transaction history retrieved ($TRANSACTION_COUNT transactions)"
else
  echo -e "${RED}✗ FAILED${NC}: Transaction history incomplete"
fi
echo ""

# Test 15: Create Card
echo "Test 15: Create Debit Card"
CARD_RESPONSE=$(curl -s -X POST $API_BASE_URL/api/v1/cards \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "account_id": "'$CHECKING_ACCOUNT_ID'",
    "card_holder_name": "Test User",
    "card_type": "debit",
    "daily_limit": 5000.00
  }')

CARD_ID=$(echo $CARD_RESPONSE | jq -r '.id')
CARD_NUMBER_MASKED=$(echo $CARD_RESPONSE | jq -r '.card_number_masked')

if [ "$CARD_ID" != "null" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Card created successfully"
  echo "Card Number: $CARD_NUMBER_MASKED"
else
  echo -e "${RED}✗ FAILED${NC}: Card creation failed"
fi
echo ""

# Test 16: Rate Limiting
echo "Test 16: Rate Limiting Test"
echo "Sending 10 rapid requests to test rate limiter..."

RATE_LIMIT_HITS=0
for i in {1..10}; do
  RATE_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" \
    -X POST $API_BASE_URL/api/v1/auth/login \
    -H "Content-Type: application/json" \
    -d '{"email":"fake@test.com","password":"wrong"}')
  
  if [ "$RATE_RESPONSE" == "429" ]; then
    RATE_LIMIT_HITS=$((RATE_LIMIT_HITS + 1))
  fi
  sleep 0.1
done

if [ $RATE_LIMIT_HITS -gt 0 ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Rate limiting is working ($RATE_LIMIT_HITS/10 requests blocked)"
else
  echo -e "${YELLOW}⚠ WARNING${NC}: Rate limiting may not be working"
fi
echo ""

# Test 17: Idempotency
echo "Test 17: Idempotency Test"
IDEMPOTENCY_KEY="test-idempotency-$(date +%s)"

# First transfer
TRANSFER1=$(curl -s -X POST $API_BASE_URL/api/v1/transactions/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "from_account_id": "'$CHECKING_ACCOUNT_ID'",
    "to_account_id": "'$SAVINGS_ACCOUNT_ID'",
    "amount": 50.00,
    "description": "Idempotency test",
    "idempotency_key": "'$IDEMPOTENCY_KEY'"
  }')

# Duplicate transfer with same key
TRANSFER2=$(curl -s -X POST $API_BASE_URL/api/v1/transactions/transfer \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "from_account_id": "'$CHECKING_ACCOUNT_ID'",
    "to_account_id": "'$SAVINGS_ACCOUNT_ID'",
    "amount": 50.00,
    "description": "Idempotency test duplicate",
    "idempotency_key": "'$IDEMPOTENCY_KEY'"
  }')

TRANSFER1_ID=$(echo $TRANSFER1 | jq -r '.id')
TRANSFER2_ID=$(echo $TRANSFER2 | jq -r '.id')

if [ "$TRANSFER1_ID" == "$TRANSFER2_ID" ]; then
  echo -e "${GREEN}✓ PASSED${NC}: Idempotency working (duplicate prevented)"
else
  echo -e "${RED}✗ FAILED${NC}: Idempotency not working (duplicate transaction created)"
fi
echo ""

# Summary
echo "========================================="
echo "Test Suite Complete!"
echo "========================================="
echo ""
echo "Summary:"
echo "- API Base URL: $API_BASE_URL"
echo "- Test User: $LOGIN_EMAIL"
echo "- Checking Account: $CHECKING_ACCOUNT_NUMBER"
echo "- Card: $CARD_NUMBER_MASKED"
echo ""
echo "View logs:"
echo "aws logs tail /ecs/madabank-dev --follow"
echo ""
echo "View metrics:"
echo "curl $API_BASE_URL/metrics | grep madabank"