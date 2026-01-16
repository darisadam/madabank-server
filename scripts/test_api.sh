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
echo "========================================="
echo "All tests completed!"
echo "========================================="