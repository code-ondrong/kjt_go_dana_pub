# DANA Payment Gateway API (Go + Gin)

✅ **Payment Gateway UAT Ready**: Official DANA SDK integration for Payment Gateway + Shop Management API.

## 📦 API Implementation

This project uses the **Official DANA SDK** (`github.com/dana-id/dana-go`) for Shop Management and custom SNAP implementation for Payment Gateway with proper RSA Signature authentication.

### API Endpoints

All endpoints are under `/api/v1`:

#### Payment Gateway (SNAP Implementation)
```
POST /api/v1/payment/create  - Create payment order (Hosted Checkout)
GET  /api/v1/payment/query   - Query payment status
POST /api/v1/payment/cancel  - Cancel payment order
POST /api/v1/payment/refund  - Refund payment
POST /webhook/dana           - Payment webhook notification
GET  /sse/payment            - SSE real-time payment updates
```

#### Shop Management (Official SDK)
```
POST /api/v1/shop/create  - Create new shop
GET  /api/v1/shop/query   - Query shop information
POST /api/v1/shop/update  - Update shop information
```

#### Division Management (Official SDK)
```
POST /api/v1/division/create  - Create new division
GET  /api/v1/division/query   - Query division information
POST /api/v1/division/update  - Update division information
```

#### Disbursement (Official SDK)
```
POST /api/v1/disbursement/account-inquiry - Account inquiry
POST /api/v1/disbursement/transfer-to-dana - Disbursement to DANA balance
POST /api/v1/disbursement/transfer-to-dana/status - Query transfer status
```

#### Common
```
GET  /api/v1/health        - Health check
GET  /                     - Demo page
```

---

## 🚀 Quick Start

### 1. Create a Shop

**Request:**
```bash
curl -X POST http://localhost:8888/api/v1/shop/create \
  -H "Content-Type: application/json" \
  -d '{
    "shopParentId": "YOUR_MERCHANT_ID",
    "externalShopId": "SHOP-001",
    "shopName": "Toko Jakarta",
    "shopParentType": "MERCHANT",
    "sizeType": "SMALL"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "00000000",
    "responseMessage": "SUCCESS",
    "shopId": "216660000003283886722",
    "shopName": "Toko Jakarta",
    "shopStatus": "ACTIVE"
  }
}
```

### 2. Query Shop

**Smart Detection:** The system automatically detects if you are using an `INNER_ID` (DANA ID) or `EXTERNAL_ID` (your ID). `shopIdType` is now optional.

**Request:**
```bash
# Automatic detection (Recommended)
curl "http://localhost:8888/api/v1/shop/query?shopId=SHOP-001&shopParentId=YOUR_MERCHANT_ID"

# Or force a specific type
curl "http://localhost:8888/api/v1/shop/query?shopId=21666000000123&shopParentId=YOUR_MERCHANT_ID&shopIdType=INNER_ID"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "00000000",
    "responseMessage": "SUCCESS",
    "shopDetailInfoList": [
      {
        "shopId": "SHOP-001",
        "shopName": "Toko Jakarta",
        "shopAlias": "SHOP-001",
        "shopStatus": "ACTIVE",
        "sizeType": "UKE",
        "shopParentId": "YOUR_MERCHANT_ID",
        "nmid": "93600...X",
        "shopAddress": "Jl. Sudirman No. 1"
      }
    ],
    "rawDana": {
       "resultInfo": { "resultStatus": "S", ... },
       "shopResourceInfo": { "merchantId": "...", "mainName": "...", "nmid": "..." }
    }
  }
}
```

### 3. Update Shop

**Request:**
```bash
curl -X POST http://localhost:8888/api/v1/shop/update \
  -H "Content-Type: application/json" \
  -d '{
    "shopId": "SHOP-001",
    "shopIdType": "EXTERNAL_ID",
    "shopParentId": "YOUR_MERCHANT_ID",
    "shopName": "Toko Jakarta - Updated"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "00000000",
    "responseMessage": "SUCCESS",
    "shopId": "SHOP-001",
    "shopName": "Toko Jakarta - Updated",
    "shopStatus": "ACTIVE"
  }
}
```

---

## 🏢 Division Management (SDK Implementation)

### 1. Create a Division

**Request:**
```bash
curl -X POST http://localhost:8888/api/v1/division/create \
  -H "Content-Type: application/json" \
  -d '{
    "merchantId": "YOUR_MERCHANT_ID",
    "externalDivisionId": "DIV-001",
    "mainName": "Divisi Jakarta",
    "divisionDesc": "Divisi Operasional Jakarta"
  }'
```

### 2. Query Division

**Smart Detection:** The system automatically detects if you are using an `INNER_ID` (DANA ID) or `EXTERNAL_ID` (your ID). `divisionIdType` is now optional.

**Request:**
```bash
curl "http://localhost:8888/api/v1/division/query?divisionId=DIV-001&merchantId=YOUR_MERCHANT_ID"
```

### 3. Update Division

**Request:**
```bash
curl -X POST http://localhost:8888/api/v1/division/update \
  -H "Content-Type: application/json" \
  -d '{
    "merchantId": "YOUR_MERCHANT_ID",
    "divisionId": "DIV-001",
    "mainName": "Divisi Jakarta - Updated",
    "divisionDesc": "Divisi Operasional Jakarta Updated"
  }'
```

---

## 💸 Disbursement Management (SDK Implementation)

### 1. Disbursement to DANA Balance

**Request:**
```bash
curl -X POST http://localhost:8888/api/v1/disbursement/transfer-to-dana \
  -H "Content-Type: application/json" \
  -d '{
    "partnerReferenceNo": "TRANS-001",
    "amount": "1000.00",
    "currency": "IDR",
    "customerNumber": "08123456789",
    "notes": "Hadiah untuk pelanggan"
  }'
```

---

## ⚙️ Configuration (.env)

```env
# DANA Credentials (REQUIRED for Shop Management API)
DANA_CLIENT_ID=your_client_id
DANA_CLIENT_SECRET=your_client_secret
DANA_PRIVATE_KEY=your_private_key_base64_or_pem_format
DANA_ENV=sandbox
ORIGIN=http://localhost:8888
SERVER_PORT=8888
```

### Required Credentials

1. **DANA_CLIENT_ID**: Partner ID provided by DANA
2. **DANA_CLIENT_SECRET**: Client Secret for authentication (REQUIRED!)
3. **DANA_PRIVATE_KEY**: RSA Private Key for signature authentication
   - Can be in base64 format (single line) or PEM format with headers
   - Example PEM: `-----BEGIN PRIVATE KEY-----\nMIIEvwIBADAN...\n-----END PRIVATE KEY-----`
4. **ORIGIN**: Your application domain (required by DANA for Merchant Management API)

---

## 📋 Reference Fields

### Shop Parent Type
- **MERCHANT**: Shop under direct merchant
- **DIVISION**: Shop under sub-merchant/division

### Shop/Division ID Type
- **EXTERNAL_ID**: ID from your system (default/detected automatically)
- **INNER_ID**: ID from DANA system (detected automatically for numeric IDs >= 16 chars)
- **Automatic Detection**: Applicable for both Shop and Division queries/updates.

### Size Type (auto-mapped to DANA codes)
You can use common names and they will be auto-mapped:
- **MICRO** → `UMI` (Usaha Mikro)
- **SMALL** → `UKE` (Usaha Kecil)
- **MEDIUM** → `UME` (Usaha Menengah)
- **LARGE** → `UBE` (Usaha Besar)

### MCC Codes
Merchant Category Codes are auto-filled with default value `5734` (Computer Software Stores).
You can modify this in the code if needed.

---

## 🏃 Running the Server

```bash
# Install dependencies
go mod download

# Run server
go run main.go
```

Server will start on `http://localhost:8888`

---

## 📚 Official Documentation

- [DANA API Documentation](https://dashboard.dana.id/api-docs-v2/api/merchant-management/overview)
- [Official Go SDK](https://github.com/dana-id/dana-go)
- [UAT Script Test](https://github.com/dana-id/uat-script) - Run this first to validate your setup!

---

## 🏗️ Project Structure

```
internal/
├── api/
│   ├── routes.go       # Route definitions
│   └── sdk_handlers.go # HTTP handlers for SDK API
├── config/
│   └── config.go       # Configuration management
├── dana/
│   ├── sdk_client.go   # SDK client wrapper
│   └── types.go        # Request/Response types
└── sse/
    └── broker.go       # SSE support (for future use)
```

---

## ⚠️ Important Notes

1. **Client Secret is REQUIRED**: Unlike some DANA APIs, Shop Management API requires both Client Secret AND Private Key for authentication.

2. **Environment Handling**: The SDK automatically handles sandbox vs production URLs. Do NOT manually override the server URLs in your code.

3. **MCC Codes**: DANA API requires MCC (Merchant Category Code) for shop creation. The default is `5734` (Computer Software Stores).

4. **Size Type Mapping**: Common size type names (SMALL, MEDIUM, etc.) are automatically mapped to DANA's internal codes (UKE, UME, etc.).

5. **Sandbox Environment**: Uses HTTPS for sandbox by default in the SDK. The SDK handles protocol selection automatically.

---

## 🔧 Troubleshooting

### Common Errors

1. **OAUTH_FAILED (00000016)**
   - Cause: Missing or invalid CLIENT_SECRET
   - Fix: Ensure DANA_CLIENT_SECRET is set in .env

2. **PARAM_ILLEGAL - invalid sizeType**
   - Cause: SizeType value not recognized
   - Fix: Use SMALL, MEDIUM, LARGE, or MICRO (will be auto-mapped)

3. **PARAM_ILLEGAL - Mcc codes can't be empty**
   - Cause: MCC codes not provided
   - Fix: The SDK now auto-fills default MCC code `5734`

4. **MSG_PARSE_ERROR (00000015)**
   - Cause: Request format incorrect or authentication failure
   - Fix: Ensure all credentials are correct and don't override SDK server URLs
