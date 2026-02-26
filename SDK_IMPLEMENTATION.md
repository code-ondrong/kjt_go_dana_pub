# Official DANA SDK Implementation

## Overview

This project now includes the Official DANA SDK (`github.com/dana-id/dana-go` v1.2.11) for Shop Management API operations. The SDK provides a more robust and maintained way to interact with DANA's Merchant Management APIs.

## What's New

### 1. SDK Client Implementation

**File**: `internal/dana/sdk_client.go`

The `SDKClient` wraps the official DANA SDK and provides:

- **Automatic RSA Signature Authentication** - Handles signature generation automatically
- **Proper Request/Response Mapping** - Converts between internal types and SDK types
- **Environment Configuration** - Automatically configures sandbox/production URLs
- **ORIGIN Header Support** - Properly sets the ORIGIN header required by DANA

Key Features:
```go
// Initialize SDK client
sdkClient, err := dana.NewSDKClient(cfg)

// Create Shop
resp, err := sdkClient.CreateShop(ctx, &dana.CreateShopRequest{
    ShopParentId:   "216620060009037857198",
    ShopAlias:      "SHOP_001",
    ShopName:       "My Shop",
    ShopParentType: "MERCHANT",
    SizeType:       "UMI",
})

// Query Shop
resp, err := sdkClient.QueryShop(ctx, &dana.QueryShopRequest{
    ShopParentId: "216620060009037857198",
    ShopID:       "SHOP_001",
    ShopIdType:   "EXTERNAL_ID",
})

// Update Shop
resp, err := sdkClient.UpdateShop(ctx, &dana.UpdateShopRequest{
    ShopID:       "SHOP_001",
    ShopIdType:   "EXTERNAL_ID",
    ShopParentId: "216620060009037857198",
    ShopName:     "Updated Shop Name",
})
```

### 2. SDK-Based HTTP Handlers

**File**: `internal/api/sdk_handlers.go`

Provides REST API endpoints using the Official SDK:

- **POST /api/v1/sdk/shop/create** - Create a new shop
- **GET /api/v1/sdk/shop/query** - Query shop information
- **POST /api/v1/sdk/shop/update** - Update shop information
- **GET /api/v1/sdk/health** - Health check endpoint

### 3. Dual Implementation

The project now supports **two implementations**:

1. **Manual Implementation** (`/api/shop/*`) - Custom HTTP client implementation
2. **SDK Implementation** (`/api/v1/sdk/shop/*`) - Official DANA SDK

Both implementations are available simultaneously, allowing you to:
- Test and compare results
- Migrate gradually from manual to SDK
- Use SDK for production while manual for debugging

## SDK vs Manual Implementation

| Feature | Manual Implementation | SDK Implementation |
|---------|---------------------|-------------------|
| Authentication | Manual RSA signature generation | Automatic RSA signature handling |
| Type Safety | Custom types | Official SDK types |
| Maintenance | Custom maintenance | Official DANA updates |
| Error Handling | Custom error parsing | SDK error handling |
| Endpoints | `/api/shop/*` | `/api/v1/sdk/shop/*` |

## Environment Configuration

The SDK uses the same `.env` configuration as the manual implementation:

```env
DANA_PARTNER_ID=2025061317571934535528
DANA_CLIENT_ID=2025061317571934535528
DANA_MERCHANT_ID=216620060009037857198
DANA_CLIENT_SECRET=6b73bcb111ee9fdb68b5a7698db764d40851f2709816bde85ec31239cc53c48f
DANA_PUBLIC_KEY=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
DANA_PRIVATE_KEY=MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQD...
SERVER_HOST=0.0.0.0
SERVER_PORT=8888
DANA_ENV=sandbox
ORIGIN=https://www.waasisten.com
```

## API Usage Examples

### 1. Create Shop (SDK)

```bash
curl -X POST http://localhost:8888/api/v1/sdk/shop/create \
  -H "Content-Type: application/json" \
  -d '{
    "shopParentId": "216620060009037857198",
    "externalShopId": "SHOP_001",
    "shopName": "My Shop",
    "shopParentType": "MERCHANT",
    "sizeType": "UMI",
    "shopDesc": "Test shop description"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "00000000",
    "responseMessage": "Success",
    "shopId": "216620060009037857199",
    "merchantId": "216620060009037857198",
    "shopName": "My Shop",
    "shopStatus": "ACTIVE"
  }
}
```

### 2. Query Shop (SDK)

```bash
curl -X GET "http://localhost:8888/api/v1/sdk/shop/query?shopParentId=216620060009037857198&shopId=SHOP_001&shopIdType=EXTERNAL_ID"
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "00000000",
    "responseMessage": "Success",
    "shopDetailInfoList": [
      {
        "shopId": "SHOP_001",
        "shopName": "My Shop",
        "shopStatus": "ACTIVE",
        "shopAlias": "SHOP_001",
        "shopParentId": "216620060009037857198",
        "shopParentType": "MERCHANT",
        "sizeType": "UMI",
        "shopAddress": "123 Main St",
        "shopCity": "Jakarta",
        "shopProvince": "DKI Jakarta"
      }
    ]
  }
}
```

### 3. Update Shop (SDK)

```bash
curl -X POST http://localhost:8888/api/v1/sdk/shop/update \
  -H "Content-Type: application/json" \
  -d '{
    "shopId": "SHOP_001",
    "shopIdType": "EXTERNAL_ID",
    "shopParentId": "216620060009037857198",
    "shopName": "Updated Shop Name",
    "shopAlias": "SHOP_001",
    "shopDesc": "Updated description"
  }'
```

**Response:**
```json
{
  "success": true,
  "data": {
    "responseCode": "00000000",
    "responseMessage": "Success",
    "shopId": "SHOP_001",
    "shopName": "Updated Shop Name",
    "shopStatus": "ACTIVE"
  }
}
```

## Field Mappings

### CreateShopRequest

| HTTP Field | Internal Type | SDK Type |
|------------|--------------|----------|
| shopParentId | ShopParentId | MerchantId |
| externalShopId | ShopAlias | ExternalShopId |
| shopName | ShopName | MainName |
| shopParentType | ShopParentType | ShopParentType |
| sizeType | SizeType | SizeType |
| shopDesc | ShopAddress | ShopDesc |

### QueryShopRequest

| HTTP Field | Internal Type | SDK Type |
|------------|--------------|----------|
| shopParentId | ShopParentId | MerchantId |
| shopId | ShopID | ShopId |
| shopIdType | ShopIdType | ShopIdType (EXTERNAL_ID/INNER_ID) |

### UpdateShopRequest

| HTTP Field | Internal Type | SDK Type |
|------------|--------------|----------|
| shopId | ShopID | ShopId |
| shopIdType | ShopIdType | ShopIdType |
| shopParentId | ShopParentId | MerchantId (required) |
| shopName | ShopName | MainName |
| shopAlias | ShopAlias | NewExternalShopId |
| shopDesc | ShopAddress | ShopDesc |

## Response Code Reference

| Result Code | Result Status | Description |
|-------------|---------------|-------------|
| 00000000 | S | Success |
| 10000001 | F | Invalid signature |
| 10000002 | F | Invalid timestamp |
| 20000001 | F | Invalid parameter |
| 20000002 | F | Missing required parameter |
| 20000003 | F | Shop not found |
| 20000004 | F | Shop already exists |

## Benefits of Using Official SDK

1. **Automatic Authentication** - No need to manually generate RSA signatures
2. **Type Safety** - Uses official DANA SDK types
3. **Maintenance** - Updates from DANA are automatically available
4. **Error Handling** - Standardized error responses
5. **Testing** - Tested and verified by DANA team
6. **Documentation** - Official documentation and examples

## Architecture

```
┌─────────────────┐
│   HTTP Client   │
│  (Browser/Tool) │
└────────┬────────┘
         │
         │ HTTP Request
         ▼
┌─────────────────────────────────┐
│      Gin Router (main.go)       │
│  - /api/shop/*      (Manual)    │
│  - /api/v1/sdk/shop/* (SDK)     │
└──────────────┬──────────────────┘
               │
               │ Routing
               ▼
    ┌──────────────────────────┐
    │   API Handlers            │
    │  - handlers.go (Manual)   │
    │  - sdk_handlers.go (SDK)  │
    └────────┬─────────────────┘
             │
             │ Business Logic
             ▼
    ┌──────────────────────────┐
    │   DANA Clients            │
    │  - client.go (Manual)     │
    │  - sdk_client.go (SDK)    │
    └────────┬─────────────────┘
             │
             │ HTTP Requests
             ▼
    ┌──────────────────────────┐
    │   DANA API                │
    │  (api.sandbox.dana.id)    │
    └──────────────────────────┘
```

## Migration Guide

To migrate from Manual to SDK implementation:

1. **Update endpoint URLs** - Change from `/api/shop/*` to `/api/v1/sdk/shop/*`
2. **Update field names** - Use SDK field names (see Field Mappings above)
3. **Update request/response handling** - SDK responses may have different structure

Example:
```bash
# Manual endpoint
curl -X POST http://localhost:8888/api/shop/create ...

# SDK endpoint
curl -X POST http://localhost:8888/api/v1/sdk/shop/create ...
```

## Troubleshooting

### Issue: "ORIGIN header is required"

**Solution**: Ensure `ORIGIN` is set in `.env` file:
```env
ORIGIN=https://www.waasisten.com
```

### Issue: "Invalid signature"

**Solution**: Verify RSA keys are correctly set in `.env`:
```env
DANA_PRIVATE_KEY=MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQD...
DANA_PUBLIC_KEY=MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEA...
```

### Issue: "Shop not found"

**Solution**:
- Verify `shopIdType` is correct (EXTERNAL_ID for your shop IDs)
- Verify `shopParentId` (Merchant ID) is correct
- Use QueryShop with correct parameters

### Issue: HTTP vs HTTPS

**Solution**: SDK automatically uses HTTP for sandbox and HTTPS for production based on `DANA_ENV` setting.

## Next Steps

1. **Test SDK endpoints** - Verify SDK endpoints work correctly with your DANA sandbox account
2. **Compare results** - Test both manual and SDK implementations to ensure consistency
3. **Update integration** - Begin using SDK endpoints in your applications
4. **Monitor logs** - SDK provides detailed logging for debugging

## Support

For issues specific to:
- **SDK Implementation**: Check SDK logs and DANA documentation
- **Manual Implementation**: Check manual client logs and custom error handling
- **DANA API**: Contact DANA support with request/response details

## Version Information

- **Official DANA SDK**: v1.2.11
- **Go Version**: 1.22+
- **Framework**: Gin v1.10+
- **DANA Environment**: Sandbox/Production

---

**Note**: Both manual and SDK implementations are functional. You can use either or both depending on your requirements. The SDK implementation is recommended for production use due to official support and maintenance.
