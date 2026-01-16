#!/bin/bash

# MadaBank API Test Script
# This script tests the basic API functionality

BASE_URL="http://localhost:8080/api/v1"

echo "========================================="
echo "MadaBank API Test Script"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Helper function to make requests
make_request() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    
    local headers=(-H "Content-Type: application/json")
    if [ -n "$token" ]; then
        headers+=(-H "Authorization: Bearer $token")
    fi

    if [ -n "$data" ]; then
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" "${headers[@]}" -d "$data")
    else
        response=$(curl -s -w "\n%{http_code}" -X "$method" "$BASE_URL$endpoint" "${headers[@]}")
    fi
    
    echo "$response"
}

check_status() {
    local response=$1
    local expected_code=$2
    local operation=$3
    
    local http_code=$(echo "$response" | tail -n1)
    local body=$(echo "$response" | sed '$d')
    
    if [[ "$http_code" == "$expected_code" ]]; then
        echo -e "${GREEN}✓ $operation successful ($http_code)${NC}"
        return 0
    else
        echo -e "${RED}✗ $operation failed. Expected $expected_code, got $http_code${NC}"
        echo "Response: $body"
        return 1
    fi
}

extract_json_field() {
    local response=$1
    local field=$2
    echo "$response" | sed '$d' | jq -r ".$field"
}

# Test 1: Register User
echo "Test 1: Register User"
# Generate random email to avoid conflicts
RANDOM_ID=$((RANDOM % 10000))
EMAIL="user${RANDOM_ID}@madabank.com"

echo "Registering user: $EMAIL"
REGISTER_DATA='{
    "email": "'"$EMAIL"'",
    "password": "SecurePass123!",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890"
}'

RESPONSE=$(make_request "POST" "/auth/register" "$REGISTER_DATA")
check_status "$RESPONSE" "201" "User registration" || exit 1
USER_ID=$(extract_json_field "$RESPONSE" "id")
echo "User ID: $USER_ID"
echo ""

# Test 2: Login
echo "Test 2: Login User"
LOGIN_DATA='{
    "email": "'"$EMAIL"'",
    "password": "SecurePass123!"
}'

RESPONSE=$(make_request "POST" "/auth/login" "$LOGIN_DATA")
check_status "$RESPONSE" "200" "Login" || exit 1
TOKEN=$(extract_json_field "$RESPONSE" "token")
# echo "Token: $TOKEN"
echo ""

# Test 3: Create Checking Account
echo "Test 3: Create Checking Account"
CHECKING_DATA='{
    "account_type": "checking",
    "currency": "USD"
}'

RESPONSE=$(make_request "POST" "/accounts" "$CHECKING_DATA" "$TOKEN")
check_status "$RESPONSE" "201" "Checking account creation" || exit 1
CHECKING_ID=$(extract_json_field "$RESPONSE" "id")
CHECKING_NUMBER=$(extract_json_field "$RESPONSE" "account_number")
echo "Checking Account: $CHECKING_NUMBER ($CHECKING_ID)"
echo ""

# Test 4: Create Savings Account
echo "Test 4: Create Savings Account"
SAVINGS_DATA='{
    "account_type": "savings",
    "currency": "USD",
    "interest_rate": 0.045
}'

RESPONSE=$(make_request "POST" "/accounts" "$SAVINGS_DATA" "$TOKEN")
check_status "$RESPONSE" "201" "Savings account creation" || exit 1
SAVINGS_ID=$(extract_json_field "$RESPONSE" "id")
echo "Savings Account ID: $SAVINGS_ID"
echo ""

# Test 5: Get All Accounts
echo "Test 5: Get All Accounts"
RESPONSE=$(make_request "GET" "/accounts" "" "$TOKEN")
check_status "$RESPONSE" "200" "Get all accounts" || exit 1

TOTAL=$(echo "$RESPONSE" | sed '$d' | jq '.total')
if [ "$TOTAL" -ge 2 ]; then
    echo -e "${GREEN}✓ Account count verified: $TOTAL${NC}"
else
    echo -e "${RED}✗ Expected at least 2 accounts, got $TOTAL${NC}"
fi
echo ""

# Test 6: Get Balance
echo "Test 6: Get Account Balance"
RESPONSE=$(make_request "GET" "/accounts/$CHECKING_ID/balance" "" "$TOKEN")
check_status "$RESPONSE" "200" "Get balance"
BALANCE=$(extract_json_field "$RESPONSE" "balance")
echo "Balance: $BALANCE"
echo ""

# Test 7: Freeze Account
echo "Test 7: Freeze Account"
FREEZE_DATA='{
    "status": "frozen"
}'

RESPONSE=$(make_request "PATCH" "/accounts/$CHECKING_ID" "$FREEZE_DATA" "$TOKEN")
check_status "$RESPONSE" "200" "Freeze account"
STATUS=$(extract_json_field "$RESPONSE" "status")
if [ "$STATUS" == "frozen" ]; then
    echo -e "${GREEN}✓ Status verified: $STATUS${NC}"
else
    echo -e "${RED}✗ Expected frozen, got $STATUS${NC}"
fi
echo ""

# Test 8: Reactivate Account
echo "Test 8: Reactivate Account"
ACTIVATE_DATA='{
    "status": "active"
}'

RESPONSE=$(make_request "PATCH" "/accounts/$CHECKING_ID" "$ACTIVATE_DATA" "$TOKEN")
check_status "$RESPONSE" "200" "Reactivate account"
STATUS=$(extract_json_field "$RESPONSE" "status")
if [ "$STATUS" == "active" ]; then
    echo -e "${GREEN}✓ Status verified: $STATUS${NC}"
else
    echo -e "${RED}✗ Expected active, got $STATUS${NC}"
fi

echo ""

# Test 9: Deposit Money
echo "Test 9: Deposit Money to Checking Account"
IDEMPOTENCY_KEY=$(uuidgen)
PAYLOAD="{
    \"account_id\": \"$CHECKING_ID\",
    \"amount\": 1000.00,
    \"description\": \"Initial deposit\",
    \"idempotency_key\": \"$IDEMPOTENCY_KEY\"
}"
# echo "Sending Payload: $PAYLOAD"

DEPOSIT_RESPONSE=$(make_request "POST" "/transactions/deposit" "$PAYLOAD" "$TOKEN")
check_status "$DEPOSIT_RESPONSE" "201" "Deposit"

echo ""

# Test 10: Check Balance After Deposit
echo "Test 10: Check Balance After Deposit"
BALANCE_RESPONSE=$(curl -s -X GET "$BASE_URL/accounts/$CHECKING_ID/balance" \
  -H "Authorization: Bearer $TOKEN")

BALANCE=$(echo "$BALANCE_RESPONSE" | jq -r '.balance')
echo "Balance: $BALANCE USD"

if [ "$BALANCE" == "1000" ]; then
    echo -e "${GREEN}✓ Balance updated correctly${NC}"
else
    echo -e "${RED}✗ Balance mismatch${NC}"
fi

echo ""

# Test 11: Transfer Money Between Accounts
echo "Test 11: Transfer Money (Checking → Savings)"
TRANSFER_RESPONSE=$(curl -s -X POST "$BASE_URL/transactions/transfer" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"from_account_id\": \"$CHECKING_ID\",
    \"to_account_id\": \"$SAVINGS_ID\",
    \"amount\": 250.00,
    \"description\": \"Transfer to savings\",
    \"idempotency_key\": \"$(uuidgen)\"
  }")

echo "$TRANSFER_RESPONSE" | jq .
if [ "$(echo "$TRANSFER_RESPONSE" | jq -r '.status')" == "completed" ]; then
    echo -e "${GREEN}✓ Transfer successful${NC}"
else
    echo -e "${RED}✗ Transfer failed${NC}"
fi

echo ""

# Test 12: Verify Balances After Transfer
echo "Test 12: Verify Balances After Transfer"
CHECKING_BALANCE=$(curl -s -X GET "$BASE_URL/accounts/$CHECKING_ID/balance" -H "Authorization: Bearer $TOKEN" | jq -r '.balance')
SAVINGS_BALANCE=$(curl -s -X GET "$BASE_URL/accounts/$SAVINGS_ID/balance" -H "Authorization: Bearer $TOKEN" | jq -r '.balance')

echo "Checking balance: $CHECKING_BALANCE USD (expected: 750)"
echo "Savings balance: $SAVINGS_BALANCE USD (expected: 250)"

if [ "$CHECKING_BALANCE" == "750" ] && [ "$SAVINGS_BALANCE" == "250" ]; then
    echo -e "${GREEN}✓ Balances correct after transfer${NC}"
else
    echo -e "${RED}✗ Balance mismatch after transfer${NC}"
fi

echo ""

# Test 13: Get Transaction History
echo "Test 13: Get Transaction History"
HISTORY_RESPONSE=$(curl -s -X GET "$BASE_URL/transactions/history?account_id=$CHECKING_ID&limit=10" \
  -H "Authorization: Bearer $TOKEN")

echo "$HISTORY_RESPONSE" | jq .
TXN_COUNT=$(echo "$HISTORY_RESPONSE" | jq '.total')

if [ "$TXN_COUNT" -ge 2 ]; then
    echo -e "${GREEN}✓ Transaction history retrieved ($TXN_COUNT transactions)${NC}"
else
    echo -e "${RED}✗ Transaction history incomplete${NC}"
fi

echo ""

# Test 14: Test Idempotency (Duplicate Transfer)
echo "Test 14: Test Idempotency (Prevent Duplicate Transfer)"
IDEMPOTENCY_KEY="test-idempotency-$(date +%s)"

# First transfer
curl -s -X POST "$BASE_URL/transactions/transfer" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"from_account_id\": \"$CHECKING_ID\",
    \"to_account_id\": \"$SAVINGS_ID\",
    \"amount\": 50.00,
    \"description\": \"Idempotency test\",
    \"idempotency_key\": \"$IDEMPOTENCY_KEY\"
  }" > /dev/null

# Duplicate transfer with same idempotency key
DUPLICATE_RESPONSE=$(curl -s -X POST "$BASE_URL/transactions/transfer" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d "{
    \"from_account_id\": \"$CHECKING_ID\",
    \"to_account_id\": \"$SAVINGS_ID\",
    \"amount\": 50.00,
    \"description\": \"Idempotency test duplicate\",
    \"idempotency_key\": \"$IDEMPOTENCY_KEY\"
  }")

# Check that balance only changed once (should be 700, not 650)
# Check that balance only changed once (should be 700, not 650)
FINAL_BALANCE=$(curl -s -X GET "$BASE_URL/accounts/$CHECKING_ID/balance" -H "Authorization: Bearer $TOKEN" | jq -r '.balance')

if [ "$FINAL_BALANCE" == "700" ]; then
    echo -e "${GREEN}✓ Idempotency working - duplicate prevented (balance: $FINAL_BALANCE)${NC}"
else
    echo -e "${RED}✗ Idempotency failed - duplicate processed (balance: $FINAL_BALANCE)${NC}"
fi

echo ""
echo "========================================="
echo "All tests completed!"
echo "========================================="