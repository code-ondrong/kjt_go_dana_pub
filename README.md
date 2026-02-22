# DANA Payment Gateway & Shop Management API (Go + Gin)

⚠️ **IMPORTANT**: Shop Management API requires RSA Signature authentication. Current implementation uses Bearer/ClientSecret for testing. For production, you must implement RSA Signature.

## 🚨 Known Issues & Fixes

### Bug Fixes Applied (2026-02-23)
- ✅ **Bug #1**: Fixed Sandbox Base URL from `https://` to `http://api.sandbox.dana.id`
- ✅ **Bug #3**: Added missing `ShopIdType` field (default: `EXTERNAL_SHOP_ID`)

### ⚠️ Bug #2: RSA Signature Required
**Current Status**: Using Bearer/ClientSecret authentication (for testing only)

**Issue**: Merchant Management API requires RSA Signature, not Bearer token. ClientSecret is only for Disbursement API.

**To Fix**: You have two options:

#### Option 1: Implement RSA Signature (Recommended)
Implement proper RSA Signature following [DANA documentation](https://dashboard.dana.id/api-docs-v2/guide/authentication).

Required changes in `internal/dana/client.go`:
```go
// In doRequest method, replace Bearer with RSA Signature
signature := c.signer.SignRequest(timestamp, method, path, bodyBytes)
req.Header.Set("X-SIGNATURE", signature)
// Remove: req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.ClientSecret))
```

#### Option 2: Use Official SDK
Switch to official Go SDK: `github.com/dana-id/dana-go`

```bash
go get github.com/dana-id/dana-go
```

---

## 🚀 Fitur Utama

### QRIS Payment
*   **Generate QRIS**: Membuat QR Code DANA secara dinamis.
*   **Status Query**: Pengecekan status transaksi secara manual via API.
*   **Real-time Updates**: Notifikasi status pembayaran otomatis ke frontend menggunakan SSE.
*   **Signature RSA-SHA256**: Autentikasi aman sesuai standar integrasi DANA.

### Shop Management
*   **Create Shop**: Membuat outlet/shop baru dalam organisasi merchant.
*   **Update Shop**: Mengupdate informasi shop yang sudah ada.
*   **Query Shop**: Mendapatkan daftar shop dengan berbagai filter dan pagination.
*   **Division Support**: Manajemen sub-merchant (division) untuk organisasi yang kompleks.

## 🛠️ Instalasi

1.  Clone repository dan masuk ke direktori project.
2.  Instal dependensi:
    ```bash
    go mod tidy
    ```
3.  Konfigurasi environment variables di `.env`.

## ⚙️ Konfigurasi (.env)

```env
# DANA Credentials
X_PARTNER_ID=your_partner_id
DANA_CLIENT_ID=your_partner_id
DANA_MERCHANT_ID=your_merchant_id
DANA_CLIENT_SECRET=your_client_secret

# RSA Keys (Required for Shop Management in production)
DANA_PRIVATE_KEY=your_private_key
DANA_PUBLIC_KEY=dana_public_key

# Environment
DANA_ENV=sandbox
SERVER_PORT=8888
```

### ⚠️ Authentication Notes

| API Type | Authentication Method | Status |
|----------|---------------------|---------|
| QRIS Payment | Bearer/ClientSecret | ✅ Working |
| Shop Management | RSA Signature | ⚠️ Needs Implementation |

---

## 📡 API Endpoints

### QRIS Payment APIs

#### 1. Create QRIS
Membuat QR Code pembayaran.
*   **URL**: `POST /api/qris/create`
*   **Payload**:
    ```json
    {
      "partnerReferenceNo": "INV-2024001",
      "amount": 1000,
      "description": "Pembayaran Kopi",
      "expiryMinutes": 15
    }
    ```

#### 2. Check Status
Mengecek status transaksi tertentu.
*   **URL**: `GET /api/qris/status/:partnerReferenceNo`

#### 3. SSE Notification (Real-time)
Mendapatkan update status secara real-time.
*   **URL**: `GET /sse/payment?channel=:partnerReferenceNo`
*   **Event Name**: `payment_update`

#### 4. Webhook DANA
Callback otomatis dari server DANA.
*   **URL**: `POST /webhook/dana`

---

### Shop Management APIs

#### 1. Create Shop
Membuat outlet/shop baru dengan informasi lengkap.

*   **URL**: `POST /api/shop/create`

**Minimal Required Fields:**
```json
{
  "shopParentType": "MERCHANT",
  "shopParentId": "YOUR_MERCHANT_ID",
  "shopName": "Toko Saya Cabang Jakarta",
  "shopAddress": "Jl. Sudirman No. 1",
  "shopCity": "Jakarta Selatan",
  "shopProvince": "DKI Jakarta"
}
```

**Complete Request with All Fields:**
```json
{
  "shopParentType": "MERCHANT",
  "shopParentId": "YOUR_MERCHANT_ID",

  "shopName": "Toko Saya Cabang Jakarta",
  "shopAlias": "toko-jakarta",
  "shopType": "RETAIL",
  "shopCategoryCode": "5462",
  "sizeType": "SMALL",

  "shopAddress": "Jl. Sudirman No. 1",
  "shopCity": "Jakarta Selatan",
  "shopProvince": "DKI Jakarta",
  "shopPostalCode": "12190",
  "shopCountryCode": "ID",
  "shopLat": "-6.2088",
  "shopLong": "106.8456",

  "shopPhoneNo": "02112345678",
  "shopMobileNo": "628123456789",
  "shopEmail": "toko@example.com",

  "shopOpenTime": "08:00",
  "shopCloseTime": "22:00",

  "loyalty": "YES",
  "businessEntity": "INDIVIDUAL",

  "ownerIdType": "KTP",
  "ownerId": "1234567890123456",
  "ownerName": "Budi Santoso",
  "ownerBirthDate": "1990-01-15",

  "shopOwning": "OWNER",

  "businessDocs": [
    {
      "docType": "KTP",
      "docFile": "BASE64_ENCODED_FILE_CONTENT"
    }
  ],

  "mobileNoInfo": [
    {
      "mobileNo": "628123456789",
      "verified": "TRUE"
    }
  ],

  "merchantResourceInformation": [
    {
      "resourceType": "LOGO",
      "resourceName": "logo.png",
      "resourceFile": "BASE64_ENCODED_LOGO"
    }
  ],

  "bankAccountNo": "1234567890",
  "bankAccountName": "Budi Santoso",
  "bankCode": "014"
}
```

**Response:**
```json
{
  "responseCode": "2005400",
  "responseMessage": "SUCCESS",
  "shopId": "SHOP_ID",
  "shopName": "Toko Saya Cabang Jakarta",
  "shopStatus": "ACTIVE",
  "createdAt": "2026-02-23T10:00:00+07:00"
}
```

---

#### 2. Update Shop
Mengupdate informasi shop yang sudah ada. Hanya perlu mengirim field yang berubah.

*   **URL**: `POST /api/shop/update`

**Payload (hanya field yang berubah):**
```json
{
  "shopId": "SHOP_ID",
  "shopName": "Toko Saya Cabang Jakarta — Updated",
  "shopPhoneNo": "02198765432",
  "shopOpenTime": "09:00",
  "shopCloseTime": "21:00",
  "shopAddress": "Jl. Sudirman No. 10, Lt. 2",
  "shopStatus": "ACTIVE"
}
```

**Response:**
```json
{
  "responseCode": "2005400",
  "responseMessage": "SUCCESS",
  "shopId": "SHOP_ID",
  "shopName": "Toko Saya Cabang Jakarta — Updated",
  "shopStatus": "ACTIVE",
  "updatedAt": "2026-02-23T11:00:00+07:00"
}
```

---

#### 3. Query Shop

**Cara 1: Query List Semua Shop (by Parent ID)**

Untuk mendapatkan **LIST semua shop** di bawah merchant/division tertentu:

*   **URL**: `POST /api/shop/query`

**Payload:**
```json
{
  "shopParentType": "MERCHANT",
  "shopParentId": "YOUR_MERCHANT_ID",
  "pageNo": 1,
  "pageSize": 10
}
```

**Atau gunakan GET:**
```
GET /api/shop/query?shopParentType=MERCHANT&shopParentId=YOUR_MERCHANT_ID&pageNo=1&pageSize=10
```

---

**Cara 2: Query Shop Spesifik (by Shop ID)**

Untuk mendapatkan **DETAIL satu shop** spesifik:

*   **URL**: `POST /api/shop/query`

**Payload:**
```json
{
  "shopId": "SPECIFIC_SHOP_ID",
  "shopIdType": "EXTERNAL_SHOP_ID",
  "pageNo": 1,
  "pageSize": 1
}
```

**Atau gunakan GET:**
```
GET /api/shop/query?shopId=SPECIFIC_SHOP_ID
```

---

**Response:**
```json
{
  "responseCode": "2005400",
  "responseMessage": "SUCCESS",
  "totalCount": 45,
  "pageNo": 1,
  "pageSize": 10,
  "shops": [
    {
      "shopId": "SHOP_ID",
      "shopName": "Toko Saya Cabang Jakarta",
      "shopAlias": "toko-jakarta",
      "shopType": "RETAIL",
      "shopCategory": "5462",
      "shopStatus": "ACTIVE",
      "sizeType": "SMALL",
      "shopAddress": "Jl. Sudirman No. 1",
      "shopCity": "Jakarta Selatan",
      "shopProvince": "DKI Jakarta",
      "shopPostalCode": "12190",
      "shopCountryCode": "ID",
      "shopLat": "-6.2088",
      "shopLong": "106.8456",
      "shopPhoneNo": "02112345678",
      "shopMobileNo": "628123456789",
      "shopEmail": "toko@example.com",
      "shopOpenTime": "08:00",
      "shopCloseTime": "22:00",
      "loyalty": "YES",
      "businessEntity": "INDIVIDUAL",
      "ownerIdType": "KTP",
      "ownerId": "1234567890123456",
      "ownerName": "Budi Santoso",
      "ownerBirthDate": "1990-01-15",
      "shopOwning": "OWNER",
      "bankAccountNo": "1234567890",
      "bankAccountName": "Budi Santoso",
      "bankCode": "014",
      "shopParentType": "MERCHANT",
      "shopParentId": "YOUR_MERCHANT_ID",
      "createdAt": "2026-02-23T10:00:00+07:00",
      "updatedAt": "2026-02-23T11:00:00+07:00"
    }
  ]
}
```

---

## 📋 Reference Fields

### Status Shop
- **ACTIVE**: Shop aktif dan dapat menerima pembayaran
- **INACTIVE**: Shop tidak aktif
- **SUSPENDED**: Shop ditangguhkan sementara

### Shop Parent Type
- **MERCHANT**: Shop di bawah merchant langsung
- **DIVISION**: Shop di bawah sub-merchant/division

### Shop Type
- **RETAIL**: Toko retail
- **F&B**: Food & Beverage
- **SERVICE**: Jasa
- **WHOLESALE**: Grosir
- **E-COMMERCE**: E-commerce

### Size Type
- **MICRO**: Usaha mikro
- **SMALL**: Usaha kecil
- **MEDIUM**: Usaha menengah
- **LARGE**: Usaha besar

### Shop ID Type (QueryShop)
- **EXTERNAL_SHOP_ID**: ID shop dari sistem partner (default)
- **DANA_SHOP_ID**: ID shop yang diberikan oleh DANA

### Business Entity
- **INDIVIDUAL**: Perorangan
- **COMPANY**: Perusahaan/Badan usaha

### Owner ID Type
- **KTP**: Kartu Tanda Penduduk
- **PASPOR**: Paspor
- **NPWP**: Nomor Pokok Wajib Pajak
- **SIUP**: Surat Izin Usaha Perdagangan

### Shop Owning
- **OWNER**: Milik sendiri
- **RENTER**: Sewa

### Bank Codes (Indonesia)
- **014**: BCA
- **009**: BNI
- **002**: BRI
- **008**: Mandiri
- **013**: Permata
- **022**: CIMB Niaga
- **153**: BTPN/Jenius

---

## 📝 Panduan Query Shop yang Benar

### 🔑 Kunci Penting

| Kebutuhan | Field yang Diisi | Hasil |
|-----------|------------------|------|
| **List semua shop** | `shopParentId` + `shopParentType`, **kosongkan** `shopId` | Semua shop di bawah parent |
| **Detail 1 shop** | `shopId` spesifik + `shopIdType` | Detail shop tersebut |
| **Filter status** | Tambahkan `shopStatus` | Filter ACTIVE/INACTIVE/SUSPENDED |
| **Pagination** | `pageNo` + `pageSize` | Hasil terbagi halaman |

### ✅ Contoh Query List
```json
{
  "shopParentType": "MERCHANT",
  "shopParentId": "YOUR_MERCHANT_ID",
  "pageNo": 1,
  "pageSize": 20
}
```
**Hasil**: Semua shop di bawah merchant tersebut (max 20 per halaman).

### ✅ Contoh Query Detail
```json
{
  "shopId": "SPECIFIC_SHOP_ID",
  "shopIdType": "EXTERNAL_SHOP_ID"
}
```
**Hasil**: Detail lengkap shop dengan ID tersebut.

---

## 🖥️ Demo UI
Untuk mencoba integrasi secara langsung, jalankan aplikasi dan buka browser:
👉 **[http://localhost:8888/](http://localhost:8888/)**

## 🏗️ Struktur Project
*   `internal/dana`: Client SDK untuk komunikasi ke DANA.
*   `internal/api`: Handler dan routing menggunakan Gin.
*   `internal/sse`: Pengelolaan event streaming.
*   `internal/config`: Konfigurasi environment.
*   `templates`: Halaman frontend demo.

## 🏃 Menjalankan Aplikasi
```bash
go run main.go
```

## 🧪 Contoh Penggunaan dengan cURL

### Query Shop List (Sekarang dengan Base URL yang benar)
```bash
# Menggunakan POST
curl -X POST http://localhost:8888/api/shop/query \
  -H "Content-Type: application/json" \
  -d '{
    "shopParentType": "MERCHANT",
    "shopParentId": "216620060009037857198",
    "pageNo": 1,
    "pageSize": 10
  }'

# Menggunakan GET
curl "http://localhost:8888/api/shop/query?shopParentType=MERCHANT&shopParentId=216620060009037857198&pageNo=1&pageSize=10"
```

### Create Shop (Minimal)
```bash
curl -X POST http://localhost:8888/api/shop/create \
  -H "Content-Type: application/json" \
  -d '{
    "shopParentType": "MERCHANT",
    "shopParentId": "216620060009037857198",
    "shopName": "Toko Cabang Jakarta",
    "shopAddress": "Jl. Sudirman No. 1",
    "shopCity": "Jakarta Selatan",
    "shopProvince": "DKI Jakarta"
  }'
```

---

## 📚 Referensi

### Official Documentation
- [DANA API Documentation - Merchant Management](https://dashboard.dana.id/api-docs-v2/api/merchant-management/overview)
- [DANA API Documentation - Shop API](https://dashboard.dana.id/api-docs-v2/api/merchant-management/shop-api/create-shop)
- [DANA Authentication Guide](https://dashboard.dana.id/api-docs-v2/guide/authentication)

### Official SDK
- [Official Go SDK](https://github.com/dana-id/dana-go)
- **Recommended**: Switch to official SDK for production use

---

## 🔒 Security Notes

⚠️ **IMPORTANT**: Merchant Management API requires RSA Signature authentication. Do NOT use this implementation in production without proper RSA Signature implementation.

For production, either:
1. Implement proper RSA Signature (see DANA authentication docs)
2. Use official SDK: `github.com/dana-id/dana-go`

Current implementation uses Bearer/ClientSecret which is only suitable for QRIS Payment API, NOT Shop Management.
