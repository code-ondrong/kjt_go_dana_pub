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
	// Merchant ID (required by DANA) - auto-filled from config if empty
	MerchantId string `json:"merchantId,omitempty"`

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
	ApiVersion         string                        `json:"apiVersion,omitempty"`
	CreateTime         string                        `json:"createTime,omitempty"`
	MerchantId         string                        `json:"merchantId" binding:"required"`
	ParentDivisionId   string                        `json:"parentDivisionId,omitempty"`
	ParentRoleType     string                        `json:"parentRoleType,omitempty"` // MERCHANT, HEAD_OFFICE, BRANCH_OFFICE
	DivisionName       string                        `json:"divisionName,omitempty"`
	DivisionDesc       string                        `json:"divisionDesc,omitempty"`
	DivisionType       string                        `json:"divisionType,omitempty"`
	DivisionAddress    *AddressInfo                  `json:"divisionAddress,omitempty"`
	ExternalDivisionId string                        `json:"externalDivisionId,omitempty"`
	SizeType           string                        `json:"sizeType,omitempty"` // UMI, UKE, UME, UBE
	MccCodes           []string                      `json:"mccCodes,omitempty"`
	ExtInfo            *CreateDivisionRequestExtInfo `json:"extInfo,omitempty"`
	BusinessEntity     string                        `json:"businessEntity,omitempty"`
	BusinessDocs       []BusinessDocs                `json:"businessDocs,omitempty"`
	OwnerName          *UserName                     `json:"ownerName,omitempty"`
	OwnerPhoneNumber   *MobileNoInfo                 `json:"ownerPhoneNumber,omitempty"`
	OwnerIdType        string                        `json:"ownerIdType,omitempty"`
	OwnerIdNo          string                        `json:"ownerIdNo,omitempty"`
	OwnerAddress       *AddressInfo                  `json:"ownerAddress,omitempty"`
	DirectorPics       []PicInfo                     `json:"directorPics,omitempty"`
	NonDirectorPics    []PicInfo                     `json:"nonDirectorPics,omitempty"`
	PgDivisionFlag     string                        `json:"pgDivisionFlag,omitempty"`
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
	DivisionId            string                 `json:"divisionId" binding:"required"`
	DivisionIdType        string                 `json:"divisionIdType" binding:"required"` // INNER_ID or EXTERNAL_ID
	MerchantId            string                 `json:"merchantId" binding:"required"`
	NewExternalDivisionId string                 `json:"newExternalDivisionId,omitempty"`
	MainName              *string                `json:"mainName,omitempty"`
	DivisionDesc          *string                `json:"divisionDesc,omitempty"`
	DivisionType          string                 `json:"divisionType,omitempty"`
	DivisionAddress       *AddressInfo           `json:"divisionAddress,omitempty"`
	MccCodes              []string               `json:"mccCodes,omitempty"`
	ExtInfo               map[string]interface{} `json:"extInfo,omitempty"`
	ApiVersion            *string                `json:"apiVersion,omitempty"`
	BusinessEntity        *string                `json:"businessEntity,omitempty"`
	BusinessEndDate       *string                `json:"businessEndDate,omitempty"`
	BusinessDocs          []BusinessDocs         `json:"businessDocs,omitempty"`
	OwnerName             *UserName              `json:"ownerName,omitempty"`
	OwnerPhoneNumber      *MobileNoInfo          `json:"ownerPhoneNumber,omitempty"`
	OwnerIdType           *string                `json:"ownerIdType,omitempty"`
	OwnerIdNo             *string                `json:"ownerIdNo,omitempty"`
	OwnerAddress          *AddressInfo           `json:"ownerAddress,omitempty"`
	DirectorPics          []PicInfo              `json:"directorPics,omitempty"`
	NonDirectorPics       []PicInfo              `json:"nonDirectorPics,omitempty"`
	SizeType              *string                `json:"sizeType,omitempty"`
	PgDivisionFlag        *string                `json:"pgDivisionFlag,omitempty"`
	LogoUrlMap            map[string]string      `json:"logoUrlMap,omitempty"`
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
	DivisionIdType string `form:"divisionIdType"` // INNER_ID or EXTERNAL_ID (auto-detected if empty)
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
	DivisionType       string `json:"divisionType,omitempty"`
	ParentRoleType     string `json:"parentRoleType,omitempty"`
	PgDivisionFlag     string `json:"pgDivisionFlag,omitempty"`
	Status             string `json:"status"`
}

// AddressInfo informasi alamat untuk division
// Digunakan untuk CreateDivision dan UpdateDivision
type AddressInfo struct {
	Country     string `json:"country,omitempty"`
	Province    string `json:"province,omitempty"`
	City        string `json:"city,omitempty"`
	Area        string `json:"area,omitempty"`
	Address1    string `json:"address1,omitempty"`
	Address2    string `json:"address2,omitempty"`
	Postcode    string `json:"postcode,omitempty"`
	SubDistrict string `json:"subDistrict,omitempty"`
}

// UserName nama pemilik/pic
type UserName struct {
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

// PicInfo informasi PIC (Person In Charge)
type PicInfo struct {
	PicName     string `json:"picName,omitempty"`
	PicPosition string `json:"picPosition,omitempty"`
}

// BusinessDocs dokumen bisnis
type BusinessDocs struct {
	DocType string `json:"docType,omitempty"`
	DocId   string `json:"docId,omitempty"`
	DocFile string `json:"docFile,omitempty"`
}

// CreateDivisionRequestExtInfo informasi tambahan untuk create division
type CreateDivisionRequestExtInfo struct {
	PIC_EMAIL       string `json:"PIC_EMAIL,omitempty"`
	PIC_PHONENUMBER string `json:"PIC_PHONENUMBER,omitempty"`
	SUBMITTER_EMAIL string `json:"SUBMITTER_EMAIL,omitempty"`
	GOODS_SOLD_TYPE string `json:"GOODS_SOLD_TYPE,omitempty"`
	USECASE         string `json:"USECASE,omitempty"`
	USER_PROFILING  string `json:"USER_PROFILING,omitempty"`
	AVG_TICKET      string `json:"AVG_TICKET,omitempty"`
	OMZET           string `json:"OMZET,omitempty"`
	EXT_URLS        string `json:"EXT_URLS,omitempty"`
	BRAND_NAME      string `json:"BRAND_NAME,omitempty"`
}

// ============================================================
// DISBURSEMENT TYPES
// ============================================================

// AccountInquiryRequest request untuk inquiry akun DANA
type AccountInquiryRequest struct {
	PartnerReferenceNo string      `json:"partnerReferenceNo,omitempty"`
	CustomerNumber     string      `json:"customerNumber" binding:"required"`
	Amount             json.Number `json:"amount,omitempty"`
	Currency           string      `json:"currency,omitempty"` // IDR
}

// AccountInquiryResponse response dari inquiry akun DANA
type AccountInquiryResponse struct {
	ResponseCode          string                 `json:"responseCode"`
	ResponseMessage       string                 `json:"responseMessage"`
	PartnerReferenceNo    string                 `json:"partnerReferenceNo,omitempty"`
	ReferenceNo           string                 `json:"referenceNo,omitempty"`
	CustomerNumber        string                 `json:"customerNumber,omitempty"`
	CustomerName          string                 `json:"customerName,omitempty"`
	CustomerMonthlyInLimit string                `json:"customerMonthlyInLimit,omitempty"`
	Amount                string                 `json:"amount,omitempty"`
	Currency              string                 `json:"currency,omitempty"`
	FeeAmount             string                 `json:"feeAmount,omitempty"`
	FeeCurrency           string                 `json:"feeCurrency,omitempty"`
	MinAmount             string                 `json:"minAmount,omitempty"`
	MinCurrency           string                 `json:"minCurrency,omitempty"`
	MaxAmount             string                 `json:"maxAmount,omitempty"`
	MaxCurrency           string                 `json:"maxCurrency,omitempty"`
	AdditionalInfo        map[string]interface{} `json:"additionalInfo,omitempty"`
	RawDana               interface{}            `json:"rawDana,omitempty"`
}

// TransferToDanaRequest request untuk transfer ke DANA balance
// Per DANA UAT script: feeAmount should be {value: "1.00", currency: "IDR"}
type TransferToDanaRequest struct {
	PartnerReferenceNo string      `json:"partnerReferenceNo" binding:"required"`
	Amount             json.Number `json:"amount" binding:"required"`
	Currency           string      `json:"currency" binding:"required"` // IDR
	FeeAmount          json.Number `json:"feeAmount,omitempty"`         // Fee amount, default "1.00" per UAT script
	FeeCurrency        string      `json:"feeCurrency,omitempty"`        // Fee currency, defaults to same as Currency
	CustomerNumber     string      `json:"customerNumber" binding:"required"`
	Notes              string      `json:"notes,omitempty"`
}

// TransferToDanaResponse response dari transfer ke DANA balance
type TransferToDanaResponse struct {
	ResponseCode       string                 `json:"responseCode"`
	ResponseMessage    string                 `json:"responseMessage"`
	PartnerReferenceNo string                 `json:"partnerReferenceNo,omitempty"`
	ReferenceNo        string                 `json:"referenceNo,omitempty"`
	CustomerNumber     string                 `json:"customerNumber,omitempty"`
	CustomerName       string                 `json:"customerName,omitempty"`
	Amount             string                 `json:"amount,omitempty"`
	Currency           string                 `json:"currency,omitempty"`
	FeeAmount          string                 `json:"feeAmount,omitempty"`
	FeeCurrency        string                 `json:"feeCurrency,omitempty"`
	AdditionalInfo     map[string]interface{} `json:"additionalInfo,omitempty"`
	RawDana            interface{}            `json:"rawDana,omitempty"`
}

// TransferToDanaInquiryStatusRequest request untuk cek status transfer
type TransferToDanaInquiryStatusRequest struct {
	OriginalPartnerReferenceNo string                 `json:"originalPartnerReferenceNo" binding:"required"`
	OriginalReferenceNo        string                 `json:"originalReferenceNo,omitempty"`
	OriginalExternalId         string                 `json:"originalExternalId,omitempty"`
	ServiceCode                string                 `json:"serviceCode,omitempty"` // Default: "38"
	AdditionalInfo             map[string]interface{} `json:"additionalInfo,omitempty"`
}

// TransferToDanaInquiryStatusResponse response dari cek status transfer
type TransferToDanaInquiryStatusResponse struct {
	ResponseCode               string      `json:"responseCode"`
	ResponseMessage            string      `json:"responseMessage"`
	OriginalPartnerReferenceNo string      `json:"originalPartnerReferenceNo,omitempty"`
	OriginalReferenceNo        string      `json:"originalReferenceNo,omitempty"`
	OriginalExternalId         string      `json:"originalExternalId,omitempty"`
	ServiceCode                string      `json:"serviceCode,omitempty"`
	Amount                     string      `json:"amount,omitempty"`
	Currency                   string      `json:"currency,omitempty"`
	LatestTransactionStatus    string      `json:"latestTransactionStatus,omitempty"`
	TransactionStatusDesc      string      `json:"transactionStatusDesc,omitempty"`
	RawDana                    interface{} `json:"rawDana,omitempty"`
}

// ============================================================
// PAYMENT GATEWAY TYPES (Hosted Checkout & API Checkout)
// ============================================================

// PaymentOrderGoods item barang dalam order
type PaymentOrderGoods struct {
	GoodsID     string `json:"goodsId,omitempty"`
	GoodsName   string `json:"goodsName,omitempty"`
	GoodsAmount string `json:"goodsAmount,omitempty"`
	GoodsQty    string `json:"goodsQty,omitempty"`
}

// PaymentOrderAdditionalInfo info tambahan untuk Payment Gateway
type PaymentOrderAdditionalInfo struct {
	EnvInfo *PaymentEnvInfo   `json:"envInfo,omitempty"`
	MCC     string            `json:"mcc,omitempty"`
	Order   *PaymentOrderInfo `json:"order,omitempty"`
}

// PaymentEnvInfo info environment
type PaymentEnvInfo struct {
	SourcePlatform string `json:"sourcePlatform,omitempty"` // IPG, APP, WEB
	TerminalType   string `json:"terminalType,omitempty"`   // WEB, MOBILE, TABLET
}

// PaymentOrderInfo info order
type PaymentOrderInfo struct {
	Goods      []PaymentOrderGoods `json:"goods,omitempty"`
	OrderTitle string              `json:"orderTitle,omitempty"`
	Scenario   string              `json:"scenario,omitempty"` // API, HOSTED
}

// CreatePaymentOrderRequest untuk membuat order Payment Gateway
type CreatePaymentOrderRequest struct {
	PartnerReferenceNo string                      `json:"partnerReferenceNo" binding:"required"`
	MerchantID         string                      `json:"merchantId" binding:"required"`
	Amount             Amount                      `json:"amount" binding:"required"`
	PayOptionDetails   []PaymentOptionDetail       `json:"payOptionDetails,omitempty"`
	URLParams          []URLParam                  `json:"urlParams,omitempty"`
	AdditionalInfo     *PaymentOrderAdditionalInfo `json:"additionalInfo,omitempty"`
	ValidUpTo          string                      `json:"validUpTo,omitempty"` // Format: 2006-01-02T15:04:05+07:00
	Notes              string                      `json:"notes,omitempty"`
}

// PaymentOptionDetail detail opsi pembayaran
type PaymentOptionDetail struct {
	PayMethod      string  `json:"payMethod,omitempty"` // NETWORK_PAY, BALANCE_PAY, etc
	PayOption      string  `json:"payOption,omitempty"` // NETWORK_PAY_PG_QRIS, etc
	TransAmount    *Amount `json:"transAmount,omitempty"`
	PromotedAmount *Amount `json:"promotedAmount,omitempty"`
}

// URLParam parameter URL untuk redirect
type URLParam struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// CreatePaymentOrderResponse response dari create payment order
type CreatePaymentOrderResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	ReferenceNo        string      `json:"referenceNo"`
	CheckoutURL        string      `json:"checkoutUrl,omitempty"` // Untuk Hosted Checkout
	PaymentStatus      string      `json:"paymentStatus,omitempty"`
	PaidTime           string      `json:"paidTime,omitempty"`
	RawDana            interface{} `json:"rawDana,omitempty"`
}

// QueryPaymentRequest untuk query status pembayaran
type QueryPaymentRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" binding:"required"`
	MerchantID         string `json:"merchantId,omitempty"`
}

// QueryPaymentResponse response query pembayaran
type QueryPaymentResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	ReferenceNo        string      `json:"referenceNo"`
	PaymentStatus      string      `json:"paymentStatus"` // 00=SUCCESS, 01=INITIATED, 02=PAYING, 05=CANCELLED
	PaymentAmount      string      `json:"paymentAmount,omitempty"`
	Currency           string      `json:"currency,omitempty"`
	PaidTime           string      `json:"paidTime,omitempty"`
	RawDana            interface{} `json:"rawDana,omitempty"`
}

// CancelPaymentRequest untuk membatalkan pembayaran
type CancelPaymentRequest struct {
	PartnerReferenceNo  string `json:"partnerReferenceNo" binding:"required"`
	OriginalReferenceNo string `json:"originalReferenceNo,omitempty"`
	MerchantID          string `json:"merchantId,omitempty"`
	Reason              string `json:"reason,omitempty"`
}

// CancelPaymentResponse response cancel pembayaran
type CancelPaymentResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	ReferenceNo        string      `json:"referenceNo"`
	CancelStatus       string      `json:"cancelStatus,omitempty"`
	RawDana            interface{} `json:"rawDana,omitempty"`
}

// RefundPaymentRequest untuk refund pembayaran
type RefundPaymentRequest struct {
	PartnerReferenceNo  string `json:"partnerReferenceNo" binding:"required"`
	OriginalReferenceNo string `json:"originalReferenceNo" binding:"required"`
	MerchantID          string `json:"merchantId,omitempty"`
	RefundAmount        Amount `json:"refundAmount" binding:"required"`
	Reason              string `json:"reason" binding:"required"`
}

// RefundPaymentResponse response refund pembayaran
type RefundPaymentResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	ReferenceNo        string      `json:"referenceNo"`
	RefundReferenceNo  string      `json:"refundReferenceNo,omitempty"`
	RefundStatus       string      `json:"refundStatus,omitempty"`
	RawDana            interface{} `json:"rawDana,omitempty"`
}

// ConsultPayRequest untuk consult pay (jika diperlukan)
type ConsultPayRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" binding:"required"`
	MerchantID         string `json:"merchantId,omitempty"`
	PayOption          string `json:"payOption,omitempty"`
}

// ConsultPayResponse response consult pay
type ConsultPayResponse struct {
	ResponseCode       string      `json:"responseCode"`
	ResponseMessage    string      `json:"responseMessage"`
	PartnerReferenceNo string      `json:"partnerReferenceNo"`
	ReferenceNo        string      `json:"referenceNo"`
	PayOption          string      `json:"payOption,omitempty"`
	PayURL             string      `json:"payUrl,omitempty"`
	RawDana            interface{} `json:"rawDana,omitempty"`
}

// PaymentGatewayNotification payload notifikasi dari DANA
type PaymentGatewayNotification struct {
	PartnerReferenceNo string `json:"partnerReferenceNo"`
	ReferenceNo        string `json:"referenceNo"`
	MerchantID         string `json:"merchantId"`
	TransactionStatus  string `json:"transactionStatus"`
	Amount             Amount `json:"amount"`
	PaidAmount         Amount `json:"paidAmount,omitempty"`
	PaidTime           string `json:"paidTime,omitempty"`
	Signature          string `json:"signature,omitempty"`
}

// Payment Status Codes
const (
	PaymentStatusSuccess   = "00" // Payment completed successfully
	PaymentStatusInitiated = "01" // Order created, waiting for payment
	PaymentStatusPaying    = "02" // Payment is being processed
	PaymentStatusCancelled = "05" // Order was cancelled
	PaymentStatusNotFound  = "07" // Order not found
)
