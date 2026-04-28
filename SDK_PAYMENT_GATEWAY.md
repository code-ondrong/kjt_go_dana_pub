# Payment Gateway API Implementation (SNAP)

## Overview

This document describes the Payment Gateway API implementation using **DANA SNAP (Standard National API Payment)** protocol with manual RSA signature.

## Architecture

Payment Gateway menggunakan implementasi SNAP manual (bukan SDK) karena:
1. SDK `dana-go` tidak menyediakan modul Payment Gateway
2. SNAP memerlukan header khusus: `X-TIMESTAMP`, `X-SIGNATURE`, `X-PARTNER-ID`, `CHANNEL-ID`, `ORIGIN`
3. Signature menggunakan RSA SHA256 dengan format khusus

## API Endpoints

### 1. Create Payment Order (Hosted Checkout)
**Endpoint:** `POST /api/v1/payment/create`

Creates a payment order and returns a checkout URL for Hosted Checkout.

**Request Body:**
```json
{
  "partnerReferenceNo": "PAY-20240223-001",
  "merchantId": "YOUR_MERCHANT_ID",
  "amount": "15000.00",
  "currency": "IDR",
  "orderTitle": "Coffee Shop Payment",
  "validUpTo": "2024-02-23T15:30:00+07:00"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "2005400",
    "responseMessage": "SUCCESS",
    "partnerReferenceNo": "PAY-20240223-001",
    "referenceNo": "20240223123456789",
    "checkoutUrl": "https://checkout-sandbox.dana.id/v1/payment/index.html?...",
    "paymentStatus": "01"
  }
}
```

### 2. Query Payment Status
**Endpoint:** `GET /api/v1/payment/query?partnerReferenceNo=PAY-xxx`

Queries the status of a payment order.

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "2005400",
    "responseMessage": "SUCCESS",
    "partnerReferenceNo": "PAY-20240223-001",
    "referenceNo": "20240223123456789",
    "paymentStatus": "00",
    "paymentAmount": "15000.00",
    "currency": "IDR",
    "paidTime": "2024-02-23T14:25:30+07:00"
  }
}
```

**Payment Status Codes:**
- `00` - SUCCESS (Payment completed)
- `01` - INITIATED (Order created, waiting for payment)
- `02` - PAYING (Payment is being processed)
- `05` - CANCELLED (Order was cancelled)

### 3. Cancel Payment
**Endpoint:** `POST /api/v1/payment/cancel`

Cancels an unpaid payment order.

**Request Body:**
```json
{
  "partnerReferenceNo": "PAY-20240223-001",
  "originalReferenceNo": "20240223123456789",
  "reason": "Customer request"
}
```

### 4. Refund Payment
**Endpoint:** `POST /api/v1/payment/refund`

Refunds a completed payment.

**Request Body:**
```json
{
  "partnerReferenceNo": "REFUND-20240223-001",
  "originalReferenceNo": "20240223123456789",
  "refundAmount": "15000.00",
  "currency": "IDR",
  "reason": "Product out of stock"
}
```

### 5. Webhook Notification
**Endpoint:** `POST /webhook/dana`

Receives payment status notifications from DANA.

**Payload:**
```json
{
  "partnerReferenceNo": "PAY-20240223-001",
  "referenceNo": "20240223123456789",
  "merchantId": "YOUR_MERCHANT_ID",
  "transactionStatus": "00",
  "amount": {
    "value": "15000.00",
    "currency": "IDR"
  },
  "paidTime": "2024-02-23T14:25:30+07:00"
}
```

## SNAP Header Requirements

| Header | Value | Source |
|--------|-------|--------|
| `Content-Type` | `application/json` | Fixed |
| `X-TIMESTAMP` | `2024-02-23T14:25:30+07:00` | GMT+7 format |
| `X-SIGNATURE` | Base64(RSA-SHA256) | Generated |
| `X-PARTNER-ID` | Partner ID | `.env` DANA_PARTNER_ID |
| `X-EXTERNAL-ID` | partnerReferenceNo | Request |
| `CHANNEL-ID` | `95221` | `.env` CHANNEL_ID |
| `ORIGIN` | `http://localhost:8888` | `.env` ORIGIN |

## Signature Generation

```
StringToSign = SHA256(BodyJSON + Timestamp)
Signature = Base64(RSA-Sign-PKCS1v15(StringToSign))
```

**Important:**
- Body JSON must be minified (no extra spaces)
- Timestamp format: `2006-01-02T15:04:05+07:00` (GMT+7)
- Private key must be in PEM format

## Environment Variables

```env
# Required for Payment Gateway
DANA_PARTNER_ID=your_partner_id
DANA_MERCHANT_ID=your_merchant_id
DANA_PRIVATE_KEY=-----BEGIN PRIVATE KEY-----
...
-----END PRIVATE KEY-----
CHANNEL_ID=95221
ORIGIN=http://localhost:8888
DANA_ENV=sandbox

# Optional
DANA_CLIENT_ID=your_client_id
DANA_CLIENT_SECRET=your_client_secret
```

## UAT Scenarios

### Mandatory Scenarios
1. **Create Payment Order** - Generate checkout URL
2. **Query Payment Status** - Check order status
3. **Webhook Notification** - Receive status update

### Additional Scenarios
4. **Cancel Payment** - Cancel unpaid order
5. **Refund Payment** - Refund completed payment

## Testing with Demo Page

Access `http://localhost:8888` for interactive testing:
- Create payment orders
- Monitor real-time status via SSE
- View checkout URLs
- Track payment status updates
}
```

### 2. Query Payment Status
**Endpoint:** `GET /api/v1/sdk/payment/query`

Query the status of a payment order.

**Query Parameters:**
- `orderId` (required): The order ID from create request
- `merchantId` (required): Your DANA Merchant ID

**Example:**
```
GET /api/v1/sdk/payment/query?orderId=ORDER-20240223-001&merchantId=YOUR_MERCHANT_ID
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "2005400",
    "responseMessage": "SUCCESS",
    "orderId": "ORDER-20240223-001",
    "paymentStatus": "00",
    "paymentAmount": "15000.00",
    "paidTime": "2026-02-23T10:30:00+07:00"
  }
}
```

## Payment Status Codes

| Code | Status | Description |
|------|--------|-------------|
| `00` | SUCCESS | Payment completed successfully |
| `01` | INITIATED | Order created, waiting for payment |
| `02` | PAYING | Payment is being processed |
| `05` | CANCELLED | Order was cancelled |
| `07` | NOT_FOUND | Order not found |

## Integration Flow

```
┌─────────────┐                    ┌─────────────┐
│   Your App  │                    │   DANA API  │
└──────┬──────┘                    └──────┬──────┘
       │                                  │
       │  1. POST /payment/create        │
       │  ─────────────────────────────►  │
       │                                  │
       │  2. checkoutUrl                  │
       │  ◄─────────────────────────────  │
       │                                  │
       │  3. Redirect User                 │
       │     to checkoutUrl                │
       │                                  │
┌──────▼──────────┐                    ┌──────▼──────┐
│  DANA Checkout  │                    │   DANA App  │
│     Page        │                    │             │
└─────────────────┘                    └─────────────┘
       │                                  │
       │  4. User pays                    │
       │  ◄─────────────────────────────  │
       │                                  │
       │  5. Webhook/Notify               │
       │  ─────────────────────────────►  │
       │                                  │
       │  6. GET /payment/query           │
       │  ─────────────────────────────►  │
       │                                  │
       │  7. Payment Status               │
       │  ◄─────────────────────────────  │
       │                                  │
```

## Implementation Details

### SDK Client
File: `internal/dana/sdk_client.go`

**Key Methods:**
- `CreatePaymentOrder()` - Creates payment order via SDK
- `QueryPayment()` - Queries payment status

**Payment Gateway Types:**
```go
// CreatePaymentOrderRequest
type CreatePaymentOrderRequest struct {
    OrderID    string  // Unique order ID
    MerchantID string  // DANA Merchant ID
    Amount      float64 // Payment amount
    OrderTitle string  // Order description
    Goods      []PaymentOrderGoods // Optional items
    NotifyURL  string  // Webhook URL
    ValidUpTo  string  // Expiry time
}

// CreatePaymentOrderResponse
type CreatePaymentOrderResponse struct {
    ResponseCode    string // DANA response code
    ResponseMessage string // Response message
    OrderID         string // Order ID
    CheckoutURL     string // Redirect URL for payment
    ReferenceNo     string // DANA reference number
}
```

### HTTP Handlers
File: `internal/api/sdk_handlers.go`

**Endpoints:**
- `CreatePaymentOrder()` - HTTP handler for create payment
- `QueryPayment()` - HTTP handler for query status

## Testing with cURL

### Create Payment Order
```bash
curl -X POST http://localhost:8888/api/v1/sdk/payment/create \
  -H "Content-Type: application/json" \
  -d '{
    "orderId": "TEST-ORDER-001",
    "merchantId": "216620060009037857198",
    "amount": 10000,
    "orderTitle": "Test Payment"
  }'
```

### Query Payment Status
```bash
curl "http://localhost:8888/api/v1/sdk/payment/query?orderId=TEST-ORDER-001&merchantId=216620060009037857198"
```

## Notes

1. **Authentication**: The SDK handles RSA Signature authentication automatically. You need to configure:
   - `DANA_CLIENT_ID` - Your Partner ID
   - `DANA_PRIVATE_KEY` - Your RSA private key
   - `DANA_PUBLIC_KEY` - DANA's public key

2. **Sandbox vs Production**:
   - Sandbox: `http://api.sandbox.dana.id`
   - Production: `https://api.dana.id`

3. **Checkout URL Expiry**: The checkout URL expires based on the `validUpTo` parameter or default expiry set by DANA.

4. **Webhook Notifications**: Set `notifyUrl` to receive payment status updates from DANA.
