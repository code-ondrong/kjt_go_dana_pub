package dana

import (
	"encoding/json"
	"time"
)

// ============================================================
// REQUEST & RESPONSE TYPES
// ============================================================

// CreateQRRequest adalah request untuk membuat QR Code pembayaran DANA
type CreateQRRequest struct {
	// Merchant order ID yang unik dari sistem partner
	PartnerReferenceNo string `json:"partnerReferenceNo"`

	// Jumlah pembayaran dalam IDR
	Amount Amount `json:"amount"`

	// Merchant ID (required untuk QRIS)
	MerchantID string `json:"merchantId,omitempty"`

	// Expiry time untuk QR (format: YYYY-MM-DDTHH:mm:ss+07:00)
	ExpiredTime string `json:"expiredTime,omitempty"`

	// Additional info
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// Amount representasi jumlah uang
type Amount struct {
	Value    string `json:"value"`    // Format: "10000.00"
	Currency string `json:"currency"` // "IDR"
}

// CreateQRResponse adalah response dari pembuatan QR Code
type CreateQRResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`

	// Data QR yang dihasilkan
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`

	// QR Code content (string yang di-encode menjadi QR)
	QRContent string `json:"qrContent"`

	// URL gambar QR (opsional, tergantung implementasi)
	QRUrl string `json:"qrUrl,omitempty"`

	// Waktu expired QR
	ExpiredTime string `json:"expiredTime,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// QueryQRRequest untuk mengecek status pembayaran
type QueryQRRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo,omitempty"`
	MerchantID         string `json:"merchantId,omitempty"`
}

// QueryQRResponse response status pembayaran
type QueryQRResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`

	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`

	// Status: INITIATED, WAITING_PAYMENT, SUCCESS, FAILED, EXPIRED, CANCELLED
	TransactionStatus string `json:"transactionStatus"`

	Amount     Amount `json:"amount,omitempty"`
	PaidAmount Amount `json:"paidAmount,omitempty"`
	PaidTime   string `json:"paidTime,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// CancelQRRequest untuk membatalkan QR
type CancelQRRequest struct {
	PartnerReferenceNo  string `json:"partnerReferenceNo"`
	OriginalReferenceNo string `json:"originalReferenceNo"`
	MerchantID          string `json:"merchantId,omitempty"`
	Reason              string `json:"reason,omitempty"`
}

// CancelQRResponse response pembatalan QR
type CancelQRResponse struct {
	ResponseCode       string `json:"responseCode"`
	ResponseMessage    string `json:"responseMessage"`
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
}

// NotificationPayload adalah payload yang diterima dari DANA saat pembayaran berhasil
type NotificationPayload struct {
	PartnerReferenceNo string                 `json:"partnerReferenceNo"`
	ReferenceNo        string                 `json:"referenceNo"`
	TransactionStatus  string                 `json:"transactionStatus"`
	Amount             Amount                 `json:"amount"`
	PaidAmount         Amount                 `json:"paidAmount"`
	PaidTime           string                 `json:"paidTime"`
	MerchantID         string                 `json:"merchantId"`
	AdditionalInfo     map[string]interface{} `json:"additionalInfo,omitempty"`
	Signature          string                 `json:"signature"`
}

// ============================================================
// INTERNAL TYPES
// ============================================================

// QRData menyimpan informasi lengkap QR yang dibuat
type QRData struct {
	PartnerReferenceNo string
	ReferenceNo        string
	QRContent          string
	QRImageBase64      string // Base64 encoded QR image
	ExpiredAt          time.Time
	Status             string
	Amount             string
	Currency           string
	CreatedAt          time.Time
}

// ============================================================
// RESPONSE CODES
// ============================================================

const (
	ResponseCodeSuccess = "2005400"
	ResponseCodePending = "0235400"

	TransactionStatusInitiated      = "INITIATED"
	TransactionStatusWaitingPayment = "WAITING_PAYMENT"
	TransactionStatusSuccess        = "SUCCESS"
	TransactionStatusFailed         = "FAILED"
	TransactionStatusExpired        = "EXPIRED"
	TransactionStatusCancelled      = "CANCELLED"
)

// Shop ID Type constants untuk QueryShop
const (
	ShopIdTypeExternalID = "EXTERNAL_ID" // ID dari sistem partner (SDK enum)
	ShopIdTypeDanaID     = "INNER_ID"    // ID dari DANA (SDK enum)
)

// ============================================================
// SHOP MANAGEMENT TYPES
// ============================================================

// CreateShopRequest request untuk membuat shop baru
type CreateShopRequest struct {
	// Parent type: MERCHANT atau DIVISION
	ShopParentType string `json:"shopParentType" binding:"required"`

	// ID parent (merchantId atau divisionId)
	ShopParentId string `json:"shopParentId" binding:"required"`

	// Nama shop
	ShopName string `json:"shopName" binding:"required"`

	// Alias shop (untuk URL/identifier)
	ShopAlias string `json:"shopAlias,omitempty"`

	// Tipe shop: RETAIL, F&B, dll
	ShopType string `json:"shopType,omitempty"`

	// Kategori shop (MCC Code)
	ShopCategoryCode string `json:"shopCategoryCode,omitempty"`

	// Ukuran bisnis: MICRO, SMALL, MEDIUM, LARGE
	SizeType string `json:"sizeType,omitempty"`

	// Alamat shop
	ShopAddress     string `json:"shopAddress,omitempty"`
	ShopCity        string `json:"shopCity,omitempty"`
	ShopProvince    string `json:"shopProvince,omitempty"`
	ShopPostalCode  string `json:"shopPostalCode,omitempty"`
	ShopCountryCode string `json:"shopCountryCode,omitempty"`
	ShopLat         string `json:"shopLat,omitempty"`
	ShopLong        string `json:"shopLong,omitempty"`

	// Kontak
	ShopPhoneNo  string `json:"shopPhoneNo,omitempty"`
	ShopMobileNo string `json:"shopMobileNo,omitempty"`
	ShopEmail    string `json:"shopEmail,omitempty"`

	// Jam operasional (format: HH:MM)
	ShopOpenTime  string `json:"shopOpenTime,omitempty"`
	ShopCloseTime string `json:"shopCloseTime,omitempty"`

	// Program loyalty: YES atau NO
	Loyalty string `json:"loyalty,omitempty"`

	// Entitas bisnis: INDIVIDUAL atau COMPANY
	BusinessEntity string `json:"businessEntity,omitempty"`

	// Identitas pemilik
	OwnerIdType    string `json:"ownerIdType,omitempty"`    // KTP, PASPOR, dll
	OwnerId        string `json:"ownerId,omitempty"`        // Nomor identitas
	OwnerName      string `json:"ownerName,omitempty"`      // Nama pemilik
	OwnerBirthDate string `json:"ownerBirthDate,omitempty"` // Format: YYYY-MM-DD

	// Tipe kepemilikan: OWNER atau RENTER
	ShopOwning string `json:"shopOwning,omitempty"`

	// Dokumen bisnis (KTP, NPWP, dll) - base64 encoded
	BusinessDocs []BusinessDoc `json:"businessDocs,omitempty"`

	// Informasi nomor HP yang terverifikasi
	MobileNoInfo []MobileNoInfo `json:"mobileNoInfo,omitempty"`

	// Resource shop (logo, dll)
	MerchantResourceInformation []MerchantResourceInfo `json:"merchantResourceInformation,omitempty"`

	// Nomor rekening bank untuk settlement
	BankAccountNo   string `json:"bankAccountNo,omitempty"`
	BankAccountName string `json:"bankAccountName,omitempty"`
	BankCode        string `json:"bankCode,omitempty"`

	// Informasi tambahan
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// BusinessDoc dokumen bisnis
type BusinessDoc struct {
	DocType string `json:"docType"` // KTP, NPWP, dll
	DocFile string `json:"docFile"` // Base64 encoded file
}

// MobileNoInfo informasi nomor HP
type MobileNoInfo struct {
	MobileNo string `json:"mobileNo"`
	Verified string `json:"verified"` // TRUE atau FALSE
}

// MerchantResourceInfo resource shop (logo, dll)
type MerchantResourceInfo struct {
	ResourceType string `json:"resourceType"` // LOGO, dll
	ResourceName string `json:"resourceName"`
	ResourceFile string `json:"resourceFile"` // Base64 encoded
}

// CreateShopResponse response dari pembuatan shop
type CreateShopResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`

	// Shop ID yang baru dibuat
	ShopID string `json:"shopId"`

	// Merchant ID
	MerchantID string `json:"merchantId"`

	// Sub Merchant ID
	SubMerchantID string `json:"subMerchantId,omitempty"`

	// Nama shop
	ShopName string `json:"shopName"`

	// Status shop
	ShopStatus string `json:"shopStatus"`

	// Waktu dibuat
	CreatedAt string `json:"createdAt,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// UpdateShopRequest request untuk mengupdate shop
type UpdateShopRequest struct {
	// Shop ID yang akan diupdate
	ShopID string `json:"shopId" binding:"required"`

	// Shop ID Type (INNER_ID atau EXTERNAL_ID)
	ShopIdType string `json:"shopIdType,omitempty"`

	// Merchant ID (required untuk update via SDK)
	ShopParentId string `json:"shopParentId,omitempty"`

	// Nama shop baru (opsional)
	ShopName string `json:"shopName,omitempty"`

	// Alias shop (opsional)
	ShopAlias string `json:"shopAlias,omitempty"`

	// Tipe shop (opsional)
	ShopType string `json:"shopType,omitempty"`

	// Kategori shop (opsional)
	ShopCategoryCode string `json:"shopCategoryCode,omitempty"`

	// Ukuran bisnis (opsional)
	SizeType string `json:"sizeType,omitempty"`

	// Alamat shop baru (opsional)
	ShopAddress     string `json:"shopAddress,omitempty"`
	ShopCity        string `json:"shopCity,omitempty"`
	ShopProvince    string `json:"shopProvince,omitempty"`
	ShopPostalCode  string `json:"shopPostalCode,omitempty"`
	ShopCountryCode string `json:"shopCountryCode,omitempty"`
	ShopLat         string `json:"shopLat,omitempty"`
	ShopLong        string `json:"shopLong,omitempty"`

	// Kontak (opsional)
	ShopPhoneNo  string `json:"shopPhoneNo,omitempty"`
	ShopMobileNo string `json:"shopMobileNo,omitempty"`
	ShopEmail    string `json:"shopEmail,omitempty"`

	// Jam operasional (opsional)
	ShopOpenTime  string `json:"shopOpenTime,omitempty"`
	ShopCloseTime string `json:"shopCloseTime,omitempty"`

	// Status shop (opsional) - ACTIVE, INACTIVE, SUSPENDED
	ShopStatus string `json:"shopStatus,omitempty"`

	// Program loyalty (opsional)
	Loyalty string `json:"loyalty,omitempty"`

	// Resource shop (opsional)
	MerchantResourceInformation []MerchantResourceInfo `json:"merchantResourceInformation,omitempty"`

	// Informasi tambahan
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// UpdateShopResponse response dari update shop
type UpdateShopResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`

	// Shop ID
	ShopID string `json:"shopId"`

	// Merchant ID
	MerchantID string `json:"merchantId"`

	// Nama shop
	ShopName string `json:"shopName"`

	// Status shop
	ShopStatus string `json:"shopStatus"`

	// Waktu diupdate
	UpdatedAt string `json:"updatedAt,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// QueryShopRequest request untuk query shop
type QueryShopRequest struct {
	// Shop ID spesifik (opsional)
	// - Jika diisi: akan return detail shop spesifik
	// - Jika kosong: akan return list shop berdasarkan shopParentId
	ShopID string `json:"shopId,omitempty"`

	// BUGFIX #3: ShopIdType WAJIB diisi
	// EXTERNAL_SHOP_ID = ID dari sistem kamu
	// DANA_SHOP_ID     = ID dari DANA
	ShopIdType string `json:"shopIdType,omitempty"`

	// Parent type: MERCHANT atau DIVISION
	ShopParentType string `json:"shopParentType,omitempty"`

	// Parent ID (merchantId atau divisionId)
	// - Untuk query LIST: isi field ini, kosongkan shopId
	ShopParentId string `json:"shopParentId,omitempty"`

	// Status shop untuk filter (opsional) - ACTIVE, INACTIVE, SUSPENDED
	ShopStatus string `json:"shopStatus,omitempty"`

	// Pagination
	PageNo   int32 `json:"pageNo,omitempty"`
	PageSize int32 `json:"pageSize,omitempty"`
}

// QueryShopResponse response dari query shop
type QueryShopResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`

	// List shop detail
	ShopDetailInfoList []ShopInfo `json:"shopDetailInfoList,omitempty"`

	// Pagination info
	TotalCount *int64 `json:"totalCount,omitempty"`
	PageNo     *int32 `json:"pageNo,omitempty"`
	PageSize   *int32 `json:"pageSize,omitempty"`

	RawDANA        interface{}            `json:"rawDana,omitempty"` // Full raw response from DANA
	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// ShopInfo informasi lengkap shop
type ShopInfo struct {
	// Identitas
	ShopID       string `json:"shopId"`
	ShopName     string `json:"shopName"`
	ShopAlias    string `json:"shopAlias,omitempty"`
	ShopType     string `json:"shopType,omitempty"`
	ShopCategory string `json:"shopCategory,omitempty"`

	// Status
	ShopStatus string `json:"shopStatus"`

	// Ukuran
	SizeType string `json:"sizeType,omitempty"`

	// Alamat lengkap
	ShopAddress     string `json:"shopAddress,omitempty"`
	ShopAddress2    string `json:"shopAddress2,omitempty"`
	ShopCity        string `json:"shopCity,omitempty"`
	ShopProvince    string `json:"shopProvince,omitempty"`
	ShopSubDistrict string `json:"shopSubDistrict,omitempty"`
	ShopArea        string `json:"shopArea,omitempty"`
	ShopPostalCode  string `json:"shopPostalCode,omitempty"`
	ShopCountryCode string `json:"shopCountryCode,omitempty"`
	ShopLat         string `json:"shopLat,omitempty"`
	ShopLong        string `json:"shopLong,omitempty"`

	// Kontak
	ShopPhoneNo  string `json:"shopPhoneNo,omitempty"`
	ShopMobileNo string `json:"shopMobileNo,omitempty"`
	ShopEmail    string `json:"shopEmail,omitempty"`

	// Jam operasional
	ShopOpenTime  string `json:"shopOpenTime,omitempty"`
	ShopCloseTime string `json:"shopCloseTime,omitempty"`

	// Loyalty
	Loyalty string `json:"loyalty,omitempty"`

	// Entitas bisnis
	BusinessEntity string `json:"businessEntity,omitempty"`

	// Pemilik
	OwnerIdType    string `json:"ownerIdType,omitempty"`
	OwnerId        string `json:"ownerId,omitempty"`
	OwnerName      string `json:"ownerName,omitempty"`
	OwnerBirthDate string `json:"ownerBirthDate,omitempty"`

	// Kepemilikan
	ShopOwning string `json:"shopOwning,omitempty"`

	// Bank
	BankAccountNo   string `json:"bankAccountNo,omitempty"`
	BankAccountName string `json:"bankAccountName,omitempty"`
	BankCode        string `json:"bankCode,omitempty"`

	// Timestamp
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`

	// Parent info
	ShopParentType   string `json:"shopParentType,omitempty"`
	ShopParentId     string `json:"shopParentId,omitempty"`
	ParentDivisionId string `json:"parentDivisionId,omitempty"`

	// DANA Specific
	Nmid string `json:"nmid,omitempty"`

	// Raw maps
	LogoUrlMap map[string]string      `json:"logoUrlMap,omitempty"`
	ExtInfo    map[string]interface{} `json:"extInfo,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// ============================================================
// DIVISION MANAGEMENT TYPES
// ============================================================

// CreateDivisionRequest request untuk membuat division baru
type CreateDivisionRequest struct {
	MerchantId         string   `json:"merchantId" binding:"required"`
	ExternalDivisionId string   `json:"externalDivisionId" binding:"required"`
	MainName           string   `json:"mainName" binding:"required"`
	DivisionDesc       string   `json:"divisionDesc,omitempty"`
	MccCodes           []string `json:"mccCodes,omitempty"`
}

// CreateDivisionResponse response dari pembuatan division
type CreateDivisionResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	DivisionID      string `json:"divisionId"`
	MerchantID      string `json:"merchantId"`
	MainName        string `json:"mainName"`
}

// UpdateDivisionRequest request untuk mengupdate division
type UpdateDivisionRequest struct {
	DivisionId     string  `json:"divisionId" binding:"required"`
	DivisionIdType string  `json:"divisionIdType" binding:"required"` // INNER_ID or EXTERNAL_ID
	MerchantId     string  `json:"merchantId" binding:"required"`
	MainName       *string `json:"mainName,omitempty"`
	DivisionDesc   *string `json:"divisionDesc,omitempty"`
}

// UpdateDivisionResponse response dari update division
type UpdateDivisionResponse struct {
	ResponseCode    string `json:"responseCode"`
	ResponseMessage string `json:"responseMessage"`
	DivisionID      string `json:"divisionId"`
}

// QueryDivisionRequest request untuk query division
type QueryDivisionRequest struct {
	MerchantId     string `form:"merchantId" binding:"required"`
	DivisionId     string `form:"divisionId" binding:"required"`
	DivisionIdType string `form:"divisionIdType" binding:"required"` // INNER_ID or EXTERNAL_ID
}

// QueryDivisionResponse response dari query division
type QueryDivisionResponse struct {
	ResponseCode    string        `json:"responseCode"`
	ResponseMessage string        `json:"responseMessage"`
	DivisionDetail  *DivisionInfo `json:"divisionDetail,omitempty"`
}

// DivisionInfo informasi lengkap division
type DivisionInfo struct {
	DivisionID         string `json:"divisionId"`
	MerchantID         string `json:"merchantId"`
	ExternalDivisionId string `json:"externalDivisionId"`
	MainName           string `json:"mainName"`
	DivisionDesc       string `json:"divisionDesc,omitempty"`
	Status             string `json:"status"`
}

// ============================================================
// DISBURSEMENT TYPES
// ============================================================

// TransferToDanaRequest request untuk transfer ke DANA balance
type TransferToDanaRequest struct {
	PartnerReferenceNo string      `json:"partnerReferenceNo" binding:"required"`
	Amount             json.Number `json:"amount" binding:"required"`
	Currency           string      `json:"currency" binding:"required"` // IDR
	CustomerNumber     string      `json:"customerNumber" binding:"required"`
	Notes              string      `json:"notes,omitempty"`
}

// TransferToDanaResponse response dari transfer ke DANA balance
type TransferToDanaResponse struct {
	ResponseCode    string      `json:"responseCode"`
	ResponseMessage string      `json:"responseMessage"`
	TransactionID   string      `json:"transactionId,omitempty"`
	ReferenceNo     string      `json:"referenceNo,omitempty"`
	TransactionDate string      `json:"transactionDate,omitempty"`
	RawDana         interface{} `json:"rawDana,omitempty"`
}

// TransferToDanaInquiryStatusRequest request untuk cek status transfer
type TransferToDanaInquiryStatusRequest struct {
	OriginalPartnerReferenceNo string `json:"originalPartnerReferenceNo" binding:"required"`
	OriginalReferenceNo        string `json:"originalReferenceNo,omitempty"`
}

// TransferToDanaInquiryStatusResponse response dari cek status transfer
type TransferToDanaInquiryStatusResponse struct {
	ResponseCode            string      `json:"responseCode"`
	ResponseMessage         string      `json:"responseMessage"`
	LatestTransactionStatus string      `json:"latestTransactionStatus,omitempty"`
	TransactionStatusDesc   string      `json:"transactionStatusDesc,omitempty"`
	OriginalReferenceNo     string      `json:"originalReferenceNo,omitempty"`
	RawDana                 interface{} `json:"rawDana,omitempty"`
}
