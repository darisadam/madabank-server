# Frontend Security & Feature Guide

This guide explains how to secure user data and implement advanced features like QR payments in the frontend application.

## 1. End-to-End Encryption (E2EE)

We use **RSA-2048** encryption to secure sensitive data (passwords, PINs, card numbers) before it leaves the client device. This ensures that even if TLS is breached, the data remains secure.

### How it works
1.  **Fetch Public Key**: On app launch (or before sensitive action), fetch the backend's public key.
2.  **Encrypt Data**: Use the public key to encrypt the sensitive payload using **RSA-OAEP** with **SHA-256**.
3.  **Send Request**: Send the Base64-encoded ciphertext to the backend.

### API Endpoint
-   **GET** `/api/v1/security/public-key`
-   **Response**: `PEM Encoded String`

### Frontend Implementation Example (JavaScript/TypeScript)

You can use a library like `node-forge` or the native `Web Crypto API`.

```javascript
// Example using Web Crypto API
async function encryptData(plaintext, publicKeyPem) {
  // 1. Import the PEM key
  const binaryDerString = window.atob(publicKeyPem.replace(/-----BEGIN PUBLIC KEY-----|\n|-----END PUBLIC KEY-----/g, ''));
  const binaryDer = new Uint8Array(binaryDerString.length);
  for (let i = 0; i < binaryDerString.length; i++) {
    binaryDer[i] = binaryDerString.charCodeAt(i);
  }
  
  const key = await window.crypto.subtle.importKey(
    "spki",
    binaryDer,
    {
      name: "RSA-OAEP",
      hash: "SHA-256"
    },
    true,
    ["encrypt"]
  );

  // 2. Encrypt
  const encoder = new TextEncoder();
  const data = encoder.encode(plaintext);
  const encrypted = await window.crypto.subtle.encrypt(
    {
      name: "RSA-OAEP"
    },
    key,
    data
  );

  // 3. Return Base64
  return window.btoa(String.fromCharCode(...new Uint8Array(encrypted)));
}
```

## 2. Card Management

-   **List Cards**: `GET /api/v1/cards` (Returns masked numbers)
-   **Get Full Details**: `POST /api/v1/cards/details` (Requires password, returns unmasked PAN/CVV)
-   **Block Card**: `POST /api/v1/cards/:id/block`

## 3. QR/NFC Payments

The backend supports resolving QR codes that follow the `madabank:account:<uuid>` format.

### QR Code Format
```text
madabank:account:123e4567-e89b-12d3-a456-426614174000
```

### Payment Flow
1.  **Scan**: User scans QR code.
2.  **Resolve**: Frontend calls `POST /api/v1/transactions/qr/resolve` with `{ "qr_code": "..." }`.
3.  **Confirm**: Backend returns `{ "account_id": "...", "owner_name": "John Doe", "currency": "USD" }`.
4.  **Display**: Show "Pay to **John Doe**?".
5.  **Transfer**: If confirmed, call `POST /api/v1/transactions/transfer` with the `account_id` and amount.
