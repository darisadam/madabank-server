# 📡 MadaBank API Reference

**Base URL (Production):** `https://api.madabank.art/api/v1`
**Version:** `v1`

## 🔐 Authentication

### Register User
Create a new user account.

- **Endpoint:** `POST /auth/register`
- **Auth Required:** No
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123",
    "first_name": "John",
    "last_name": "Doe",
    "phone": "+1234567890",
    "date_of_birth": "1990-01-01"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    "kyc_status": "pending",
    "is_active": true,
    "created_at": "2024-01-01T00:00:00Z"
  }
  ```

### Login
Authenticate user and receive JWT tokens.

- **Endpoint:** `POST /auth/login`
- **Auth Required:** No
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "password": "securepassword123"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "token": "jwt_access_token",
    "refresh_token": "long_lived_refresh_token",
    "expires_at": "2024-01-02T00:00:00Z",
    "user": { ... }
  }
  ```

### Refresh Token
Get a new access token using a valid refresh token.

- **Endpoint:** `POST /auth/refresh`
- **Auth Required:** No
- **Request Body:**
  ```json
  {
    "refresh_token": "long_lived_refresh_token"
  }
  ```
- **Response (200 OK):** (Same as Login response)

### Forgot Password
Initiate password reset flow (sends OTP).

- **Endpoint:** `POST /auth/forgot-password`
- **Auth Required:** No
- **Request Body:**
  ```json
  {
    "email": "user@example.com"
  }
  ```
- **Response (200 OK):**
  ```json
  { "message": "If this email exists, an OTP has been sent." }
  ```

### Reset Password
Reset password using OTP.

- **Endpoint:** `POST /auth/reset-password`
- **Auth Required:** No
- **Request Body:**
  ```json
  {
    "email": "user@example.com",
    "otp": "123456",
    "new_password": "newsecurepassword123"
  }
  ```
- **Response (200 OK):**
  ```json
  { "message": "Password reset successfully" }
  ```

---

## 👤 Users
*Requires Bearer Token*

### Get Profile
- **Endpoint:** `GET /users/profile`
- **Response (200 OK):**
  ```json
  {
    "id": "uuid",
    "email": "user@example.com",
    "first_name": "John",
    "last_name": "Doe",
    ...
  }
  ```

### Update Profile
- **Endpoint:** `PUT /users/profile`
- **Request Body:**
  ```json
  {
    "first_name": "Johnny",
    "phone": "+9876543210"
  }
  ```
- **Response (200 OK):** Updated user object.

### Delete Account
Soft delete the user account.
- **Endpoint:** `DELETE /users/profile`
- **Response (204 No Content)**

---

## 🏦 Accounts
*Requires Bearer Token*

### Create Account
- **Endpoint:** `POST /accounts`
- **Request Body:**
  ```json
  {
    "account_type": "checking", // or "savings"
    "currency": "USD"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "id": "uuid",
    "account_number": "1234567890",
    "balance": 0,
    "status": "active",
    ...
  }
  ```

### List Accounts
- **Endpoint:** `GET /accounts`
- **Response (200 OK):**
  ```json
  {
    "accounts": [ { ... }, { ... } ],
    "total": 2
  }
  ```

### Get Account Details
- **Endpoint:** `GET /accounts/:id`
- **Response (200 OK):** Single account object.

### Get Account Balance
- **Endpoint:** `GET /accounts/:id/balance`
- **Response (200 OK):**
  ```json
  {
    "account_id": "uuid",
    "account_number": "123...",
    "balance": 100.50,
    "currency": "USD",
    "as_of_date": "2024-..."
  }
  ```

### Update Account
Update status (e.g., freeze account).
- **Endpoint:** `PATCH /accounts/:id`
- **Request Body:**
  ```json
  {
    "status": "frozen"
  }
  ```
- **Response (200 OK):** Updated account object.

### Close Account
- **Endpoint:** `DELETE /accounts/:id`
- **Response (204 No Content)**

---

## 💸 Transactions
*Requires Bearer Token*

### Transfer Money
- **Endpoint:** `POST /transactions/transfer`
- **Request Body:**
  ```json
  {
    "from_account_id": "uuid",
    "to_account_id": "uuid",
    "amount": 50.00,
    "description": "Lunch money",
    "idempotency_key": "unique-uuid"
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "id": "uuid",
    "status": "completed",
    "amount": 50.00,
    ...
  }
  ```

### Deposit
- **Endpoint:** `POST /transactions/deposit`
- **Request Body:**
  ```json
  {
    "account_id": "uuid",
    "amount": 100.00,
    "idempotency_key": "unique-uuid"
  }
  ```

### Withdraw
- **Endpoint:** `POST /transactions/withdraw`
- **Request Body:**
  ```json
  {
    "account_id": "uuid",
    "amount": 20.00,
    "idempotency_key": "unique-uuid"
  }
  ```

### Resolve QR Code
Resolve a QR string to account details before transfer.
- **Endpoint:** `POST /transactions/qr/resolve`
- **Request Body:**
  ```json
  {
    "qr_code": "encoded_qr_string"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "account_id": "uuid",
    "owner_name": "John Doe",
    "currency": "USD"
  }
  ```

### Get History
- **Endpoint:** `GET /transactions/history`
- **Query Params:**
  - `account_id` (required)
  - `limit` (default 20)
  - `offset` (default 0)
  - `start_date` (YYYY-MM-DD)
  - `end_date` (YYYY-MM-DD)
  - `type` (transfer, deposit, etc.)
- **Response (200 OK):**
  ```json
  {
    "transactions": [ ... ],
    "total": 50,
    "limit": 20,
    "offset": 0
  }
  ```

### Get Transaction Details
- **Endpoint:** `GET /transactions/:id`
- **Response (200 OK):** Single transaction object.

---

## 💳 Cards
*Requires Bearer Token*

### Issue Card
- **Endpoint:** `POST /cards`
- **Request Body:**
  ```json
  {
    "account_id": "uuid",
    "card_holder_name": "JOHN DOE",
    "card_type": "debit",
    "daily_limit": 1000.00
  }
  ```
- **Response (201 Created):**
  ```json
  {
    "id": "uuid",
    "card_number_masked": "************1234",
    "status": "active",
    ...
  }
  ```

### List Cards
- **Endpoint:** `GET /cards`
- **Query Params:** `account_id` (required)
- **Response (200 OK):** `{ "cards": [ ... ] }`

### Get Card Details
Get full PAN and CVV (sensitive).
- **Endpoint:** `POST /cards/details`
- **Request Body:**
  ```json
  {
    "card_id": "uuid",
    "password": "user_password_for_verification"
  }
  ```
- **Response (200 OK):**
  ```json
  {
    "card_number": "1234567812345678",
    "cvv": "123",
    "expiry_month": 12,
    "expiry_year": 2028
  }
  ```

### Update Card
- **Endpoint:** `PATCH /cards/:id`
- **Request Body:**
  ```json
  {
    "status": "blocked", // or active
    "daily_limit": 2000.00
  }
  ```

### Block Card
Quick freeze.
- **Endpoint:** `POST /cards/:id/block`
- **Response (200 OK):** Message success.

### Delete Card
- **Endpoint:** `DELETE /cards/:id`
- **Response (204 No Content)**

---

## 🛡️ Security

### Get Public Key
Get the server's public key for E2EE (frontend encryption).
- **Endpoint:** `GET /security/public-key`
- **Auth Required:** No
- **Response (200 OK):** `text/plain` (PEM format)

---

## 🩺 System Endpoints

- `GET /health` - Health check
- `GET /ready` - Readiness check (DB connection)
- `GET /version` - Build version info
- `GET /metrics` - Prometheus metrics
