# 📡 MadaBank API Reference

Base URL (Production): `https://api.madabank.art/api/v1`

## Authentication

### Register
**POST** `/auth/register`
- **Body**: `{ "email": "user@example.com", "password": "...", "full_name": "..." }`
- **Response**: `201 Created`

### Login
**POST** `/auth/login`
- **Body**: `{ "email": "..." , "password": "..." }`
- **Response**: `200 OK`
  ```json
  {
    "token": "jwt_access_token",
    "refresh_token": "long_lived_token",
    "expires_at": "2024-..."
  }
  ```

### Refresh Token
**POST** `/auth/refresh`
- **Headers**: none
- **Body**: `{ "refresh_token": "..." }`
- **Response**: `200 OK` (New Access Token)

## Users
*Requires Bearer Token*

- **GET** `/users/profile` - Get current user profile
- **PUT** `/users/profile` - Update profile
- **DELETE** `/users/profile` - Delete account

## Accounts
*Requires Bearer Token*

- **POST** `/accounts` - Create new bank account (`checking` or `savings`)
- **GET** `/accounts` - List my accounts
- **GET** `/accounts/:id` - Get details
- **GET** `/accounts/:id/balance` - Get balance
- **DELETE** `/accounts/:id` - Close account

## Transactions
*Requires Bearer Token*

### Transfer
**POST** `/transactions/transfer`
- **Body**: `{ "from_account_id": "...", "to_account_id": "...", "amount": 100.00 }`

### Deposit
**POST** `/transactions/deposit`
- **Body**: `{ "account_id": "...", "amount": 50.00 }`

### Withdraw
**POST** `/transactions/withdraw`
- **Body**: `{ "account_id": "...", "amount": 20.00 }`

### History
**GET** `/transactions/history`
- **Query Params**: `?account_id=...&page=1&limit=10`

## Cards
- **POST** `/cards` - Issue new card
- **GET** `/cards` - List cards
- **POST** `/cards/:id/block` - Block card

## Error Handling
All errors return consistent JSON:
```json
{
  "error": "Short error code",
  "message": "Human readable description"
}
```
