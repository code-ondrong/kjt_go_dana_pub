# DANA Payment Gateway API - Implementation Notes

## Payment Gateway API Status

**Current Status**: ❌ Payment Gateway API (Gapura) returning 500 Internal Server Error in Sandbox

### Issue Details

The DANA Payment Gateway API (`/payment-gateway/v1.0/debit/payment-host-to-host.htm`) is returning HTTP 500 when called from the sandbox environment. This appears to be a sandbox limitation rather than a code issue.

**Possible Causes:**
1. Sandbox account not activated for Payment Gateway (Gapura)
2. Different authentication requirements for Payment Gateway vs Shop Management
3. Payment Gateway requires additional setup/configuration in DANA dashboard

### Request Being Sent:
```json
{
  "additionalInfo": {
    "envInfo": {
      "sourcePlatform": "IPG",
      "terminalType": "WEB"
    },
    "mcc": "0000",
    "order": {
      "goods": [...],
      "orderTitle": "...",
      "scenario": "API"
    }
  },
  "amount": {
    "currency": "IDR",
    "value": "10000.00"
  },
  "merchantId": "216620060009037857198",
  "partnerReferenceNo": "TEST-ORDER-001",
  "payOptionDetails": [
    {
      "payMethod": "NETWORK_PAY",
      "payOption": "NETWORK_PAY_PG_QRIS",
      "transAmount": {...}
    }
  ],
  "urlParams": []
}
```

### Recommendations:

1. **Use QRIS API for payments** instead of Payment Gateway
   - QRIS API: `/v1.0/qr/qr-mpm-generate.htm`
   - Returns QR code that users can scan
   - Works with Bearer authentication (ClientSecret)

2. **Contact DANA Support** to activate Payment Gateway in sandbox
   - Check DANA dashboard for Payment Gateway activation
   - Verify merchant account has Payment Gateway enabled

3. **Use Production credentials** if available
   - Sandbox may have limited Payment Gateway features
   - Production might work differently

## Working APIs

### ✅ Shop Management API (SDK)
```
POST /api/v1/shop/create  - Create new shop
GET  /api/v1/shop/query   - Query shop information
POST /api/v1/shop/update  - Update shop information
```

### ✅ QRIS Payment API (Manual)
```
POST /api/qris/create         - Create QR code
GET  /api/qris/status/:ref    - Check QR payment status
```

## Code References

- SDK Client: `internal/dana/sdk_client.go`
- SDK Handlers: `internal/api/sdk_handlers.go`
- Routes: `internal/api/routes.go`
