# Payment Gateway API Implementation using Official DANA SDK

## Overview

This document describes the Payment Gateway API implementation using the official DANA SDK (`github.com/dana-id/dana-go`).

## Key Differences: Payment Gateway vs QRIS API

### Payment Gateway (Gapura)
- **Output**: Checkout URL that redirects users to DANA payment page
- **User Flow**: User is redirected to DANA → Completes payment → Redirected back
- **Payment Methods**: Supports all DANA payment methods (Balance, Linked Banks, QRIS, etc.)
- **Use Cases**: E-commerce checkout, Mobile app payments, Web payments
- **Authentication**: RSA Signature (handled automatically by SDK)

### QRIS API (Manual)
- **Output**: QR Code string/image
- **User Flow**: User scans QR with DANA app → Pays
- **Payment Methods**: QRIS only
- **Use Cases**: In-store payments, In-person transactions
- **Authentication**: Bearer token (ClientSecret)

## SDK API Endpoints

### 1. Create Payment Order
**Endpoint:** `POST /api/v1/sdk/payment/create`

Creates a payment order and returns a checkout URL.

**Request Body:**
```json
{
  "orderId": "ORDER-20240223-001",
  "merchantId": "YOUR_MERCHANT_ID",
  "amount": 15000,
  "orderTitle": "Coffee Shop Payment"
}
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "2005400",
    "responseMessage": "SUCCESS",
    "orderId": "ORDER-20240223-001",
    "checkoutUrl": "https://checkout-sandbox.dana.id/v1/payment/index.html?...",
    "referenceNo": "20240223123456789"
  }
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
