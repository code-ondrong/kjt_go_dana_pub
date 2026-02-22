package api

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"

	"kjt_go_dana/internal/dana"
	"kjt_go_dana/internal/sse"
)

type APIHandler struct {
	danaClient *dana.Client
	sseBroker  *sse.Broker

	mu      sync.RWMutex
	qrStore map[string]*dana.QRData
}

func NewAPIHandler(danaClient *dana.Client, sseBroker *sse.Broker) *APIHandler {
	return &APIHandler{
		danaClient: danaClient,
		sseBroker:  sseBroker,
		qrStore:    make(map[string]*dana.QRData),
	}
}

type CreateQRRequest struct {
	PartnerReferenceNo string  `json:"partnerReferenceNo"`
	Amount             float64 `json:"amount" binding:"required"`
	Description        string  `json:"description"`
	ExpiryMinutes      int     `json:"expiryMinutes"`
}

func (h *APIHandler) CreateQR(c *gin.Context) {
	var req CreateQRRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.PartnerReferenceNo == "" {
		req.PartnerReferenceNo = uuid.New().String()
	}
	if req.ExpiryMinutes <= 0 {
		req.ExpiryMinutes = 15
	}

	expiredTime := time.Now().Add(time.Duration(req.ExpiryMinutes) * time.Minute).
		Format("2006-01-02T15:04:05+07:00")

	danaReq := &dana.CreateQRRequest{
		PartnerReferenceNo: req.PartnerReferenceNo,
		Amount: dana.Amount{
			Value:    fmt.Sprintf("%.2f", req.Amount),
			Currency: "IDR",
		},
		ExpiredTime: expiredTime,
		AdditionalInfo: map[string]interface{}{
			"description": req.Description,
		},
	}

	resp, err := h.danaClient.CreateQR(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[API] CreateQR Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Generate QR Image
	var qrImageBase64 string
	if resp.QRContent != "" {
		png, err := qrcode.Encode(resp.QRContent, qrcode.Medium, 256)
		if err == nil {
			qrImageBase64 = "data:image/png;base64," + base64.StdEncoding.EncodeToString(png)
		}
	}

	// Save to store
	qrData := &dana.QRData{
		PartnerReferenceNo: resp.PartnerReferenceNo,
		ReferenceNo:        resp.ReferenceNo,
		QRContent:          resp.QRContent,
		QRImageBase64:      qrImageBase64,
		Status:             dana.TransactionStatusInitiated,
		Amount:             danaReq.Amount.Value,
		Currency:           "IDR",
		CreatedAt:          time.Now(),
	}

	h.mu.Lock()
	h.qrStore[req.PartnerReferenceNo] = qrData
	h.mu.Unlock()

	// Notify SSE
	h.sseBroker.PublishPaymentUpdate(req.PartnerReferenceNo, sse.PaymentEvent{
		PartnerReferenceNo: resp.PartnerReferenceNo,
		ReferenceNo:        resp.ReferenceNo,
		Status:             dana.TransactionStatusInitiated,
		Amount:             danaReq.Amount.Value,
		Currency:           "IDR",
		Message:            "QR berhasil dibuat, menunggu pembayaran",
	})

	c.JSON(http.StatusOK, gin.H{
		"partnerReferenceNo": resp.PartnerReferenceNo,
		"referenceNo":        resp.ReferenceNo,
		"qrContent":          resp.QRContent,
		"qrImageBase64":      qrImageBase64,
		"expiredTime":        expiredTime,
		"status":             dana.TransactionStatusInitiated,
	})
}

func (h *APIHandler) QueryQR(c *gin.Context) {
	partnerRef := c.Param("partnerReferenceNo")
	if partnerRef == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "partnerReferenceNo is required"})
		return
	}

	resp, err := h.danaClient.QueryQR(c.Request.Context(), &dana.QueryQRRequest{
		PartnerReferenceNo: partnerRef,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update store
	h.mu.Lock()
	if qr, ok := h.qrStore[partnerRef]; ok {
		qr.Status = resp.TransactionStatus
	}
	h.mu.Unlock()

	// Notify SSE
	h.sseBroker.PublishPaymentUpdate(partnerRef, sse.PaymentEvent{
		PartnerReferenceNo: resp.PartnerReferenceNo,
		ReferenceNo:        resp.ReferenceNo,
		Status:             resp.TransactionStatus,
		Amount:             resp.Amount.Value,
		Currency:           resp.Amount.Currency,
		PaidAt:             resp.PaidTime,
	})

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) HandleWebhook(c *gin.Context) {
	var payload dana.NotificationPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[Webhook] Payment update: %s -> %s", payload.PartnerReferenceNo, payload.TransactionStatus)

	// Update store
	h.mu.Lock()
	if qr, ok := h.qrStore[payload.PartnerReferenceNo]; ok {
		qr.Status = payload.TransactionStatus
	}
	h.mu.Unlock()

	// Notify SSE
	h.sseBroker.PublishPaymentUpdate(payload.PartnerReferenceNo, sse.PaymentEvent{
		PartnerReferenceNo: payload.PartnerReferenceNo,
		ReferenceNo:        payload.ReferenceNo,
		Status:             payload.TransactionStatus,
		Amount:             payload.PaidAmount.Value,
		Currency:           payload.PaidAmount.Currency,
		PaidAt:             payload.PaidTime,
		Message:            "Update dari webhook DANA",
	})

	c.JSON(http.StatusOK, gin.H{
		"responseCode":    "2005400",
		"responseMessage": "SUCCESS",
	})
}

func (h *APIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":     "ok",
		"time":       time.Now().Format(time.RFC3339),
		"sseClients": h.sseBroker.ClientCount(),
	})
}

// ============================================================
// SHOP MANAGEMENT HANDLERS
// ============================================================

// CreateShopRequest handler request untuk create shop
type CreateShopRequest struct {
	// Parent info
	ShopParentType string `json:"shopParentType" binding:"required"` // MERCHANT atau DIVISION
	ShopParentId   string `json:"shopParentId" binding:"required"`   // Merchant ID atau Division ID

	// Info dasar
	ShopName         string `json:"shopName" binding:"required"`
	ShopAlias        string `json:"shopAlias,omitempty"`
	ShopType         string `json:"shopType,omitempty"`
	ShopCategoryCode string `json:"shopCategoryCode,omitempty"`
	SizeType         string `json:"sizeType,omitempty"`

	// Alamat
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

	// Jam operasional
	ShopOpenTime  string `json:"shopOpenTime,omitempty"`
	ShopCloseTime string `json:"shopCloseTime,omitempty"`

	// Loyalty dan Business Entity
	Loyalty         string `json:"loyalty,omitempty"`
	BusinessEntity  string `json:"businessEntity,omitempty"`

	// Identitas pemilik
	OwnerIdType    string `json:"ownerIdType,omitempty"`
	OwnerId        string `json:"ownerId,omitempty"`
	OwnerName      string `json:"ownerName,omitempty"`
	OwnerBirthDate string `json:"ownerBirthDate,omitempty"`

	// Kepemilikan
	ShopOwning string `json:"shopOwning,omitempty"`

	// Dokumen dan resource
	BusinessDocs                 []dana.BusinessDoc          `json:"businessDocs,omitempty"`
	MobileNoInfo                 []dana.MobileNoInfo          `json:"mobileNoInfo,omitempty"`
	MerchantResourceInformation  []dana.MerchantResourceInfo  `json:"merchantResourceInformation,omitempty"`

	// Bank
	BankAccountNo   string `json:"bankAccountNo,omitempty"`
	BankAccountName string `json:"bankAccountName,omitempty"`
	BankCode        string `json:"bankCode,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// UpdateShopRequest handler request untuk update shop
type UpdateShopRequest struct {
	ShopID string `json:"shopId" binding:"required"`

	// Field yang bisa diupdate (opsional - hanya isi yang berubah)
	ShopName         string `json:"shopName,omitempty"`
	ShopAlias        string `json:"shopAlias,omitempty"`
	ShopType         string `json:"shopType,omitempty"`
	ShopCategoryCode string `json:"shopCategoryCode,omitempty"`
	SizeType         string `json:"sizeType,omitempty"`

	// Alamat
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

	// Jam operasional
	ShopOpenTime  string `json:"shopOpenTime,omitempty"`
	ShopCloseTime string `json:"shopCloseTime,omitempty"`

	// Status dan loyalty
	ShopStatus string `json:"shopStatus,omitempty"`
	Loyalty    string `json:"loyalty,omitempty"`

	// Resource
	MerchantResourceInformation []dana.MerchantResourceInfo `json:"merchantResourceInformation,omitempty"`

	AdditionalInfo map[string]interface{} `json:"additionalInfo,omitempty"`
}

// QueryShopRequest handler request untuk query shop
type QueryShopRequest struct {
	// Untuk query shop spesifik
	ShopID string `json:"shopId,omitempty"`

	// Untuk query list shop
	ShopParentType string `json:"shopParentType,omitempty"` // MERCHANT atau DIVISION
	ShopParentId   string `json:"shopParentId,omitempty"`   // Merchant ID atau Division ID

	// Filter
	ShopStatus string `json:"shopStatus,omitempty"`

	// Pagination
	PageNo   int32 `json:"pageNo,omitempty"`
	PageSize int32 `json:"pageSize,omitempty"`
}

// CreateShop membuat shop baru
func (h *APIHandler) CreateShop(c *gin.Context) {
	var req CreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[API] CreateShop request: shopParentType=%s, shopParentId=%s, shopName=%s",
		req.ShopParentType, req.ShopParentId, req.ShopName)

	danaReq := &dana.CreateShopRequest{
		ShopParentType:              req.ShopParentType,
		ShopParentId:                req.ShopParentId,
		ShopName:                    req.ShopName,
		ShopAlias:                   req.ShopAlias,
		ShopType:                    req.ShopType,
		ShopCategoryCode:            req.ShopCategoryCode,
		SizeType:                    req.SizeType,
		ShopAddress:                 req.ShopAddress,
		ShopCity:                    req.ShopCity,
		ShopProvince:                req.ShopProvince,
		ShopPostalCode:              req.ShopPostalCode,
		ShopCountryCode:             req.ShopCountryCode,
		ShopLat:                     req.ShopLat,
		ShopLong:                    req.ShopLong,
		ShopPhoneNo:                 req.ShopPhoneNo,
		ShopMobileNo:                req.ShopMobileNo,
		ShopEmail:                   req.ShopEmail,
		ShopOpenTime:                req.ShopOpenTime,
		ShopCloseTime:               req.ShopCloseTime,
		Loyalty:                     req.Loyalty,
		BusinessEntity:              req.BusinessEntity,
		OwnerIdType:                 req.OwnerIdType,
		OwnerId:                     req.OwnerId,
		OwnerName:                   req.OwnerName,
		OwnerBirthDate:              req.OwnerBirthDate,
		ShopOwning:                  req.ShopOwning,
		BusinessDocs:                req.BusinessDocs,
		MobileNoInfo:                req.MobileNoInfo,
		MerchantResourceInformation: req.MerchantResourceInformation,
		BankAccountNo:               req.BankAccountNo,
		BankAccountName:             req.BankAccountName,
		BankCode:                    req.BankCode,
		AdditionalInfo:              req.AdditionalInfo,
	}

	resp, err := h.danaClient.CreateShop(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[API] CreateShop Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to create shop",
			"details":    err.Error(),
			"request":    gin.H{"shopParentType": req.ShopParentType, "shopParentId": req.ShopParentId, "shopName": req.ShopName},
			"suggestion": "Please check your DANA credentials and API endpoint path",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"responseCode":    resp.ResponseCode,
		"responseMessage": resp.ResponseMessage,
		"shopId":          resp.ShopID,
		"shopName":        resp.ShopName,
		"shopStatus":      resp.ShopStatus,
		"createdAt":       resp.CreatedAt,
	})
}

// UpdateShop mengupdate informasi shop
func (h *APIHandler) UpdateShop(c *gin.Context) {
	var req UpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"details": err.Error(),
		})
		return
	}

	log.Printf("[API] UpdateShop request: shopId=%s, shopName=%s", req.ShopID, req.ShopName)

	danaReq := &dana.UpdateShopRequest{
		ShopID:                      req.ShopID,
		ShopName:                    req.ShopName,
		ShopAlias:                   req.ShopAlias,
		ShopType:                    req.ShopType,
		ShopCategoryCode:            req.ShopCategoryCode,
		SizeType:                    req.SizeType,
		ShopAddress:                 req.ShopAddress,
		ShopCity:                    req.ShopCity,
		ShopProvince:                req.ShopProvince,
		ShopPostalCode:              req.ShopPostalCode,
		ShopCountryCode:             req.ShopCountryCode,
		ShopLat:                     req.ShopLat,
		ShopLong:                    req.ShopLong,
		ShopPhoneNo:                 req.ShopPhoneNo,
		ShopMobileNo:                req.ShopMobileNo,
		ShopEmail:                   req.ShopEmail,
		ShopOpenTime:                req.ShopOpenTime,
		ShopCloseTime:               req.ShopCloseTime,
		ShopStatus:                  req.ShopStatus,
		Loyalty:                     req.Loyalty,
		MerchantResourceInformation: req.MerchantResourceInformation,
		AdditionalInfo:              req.AdditionalInfo,
	}

	resp, err := h.danaClient.UpdateShop(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[API] UpdateShop Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "Failed to update shop",
			"details":    err.Error(),
			"request":    gin.H{"shopId": req.ShopID, "shopName": req.ShopName},
			"suggestion": "Please check your DANA credentials and API endpoint path",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"responseCode":    resp.ResponseCode,
		"responseMessage": resp.ResponseMessage,
		"shopId":          resp.ShopID,
		"shopName":        resp.ShopName,
		"shopStatus":      resp.ShopStatus,
		"updatedAt":       resp.UpdatedAt,
	})
}

// QueryShop mendapatkan informasi shop
func (h *APIHandler) QueryShop(c *gin.Context) {
	var req QueryShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Jika request body kosong, coba ambil dari query params
		req.ShopID = c.Query("shopId")
		req.ShopParentType = c.Query("shopParentType")
		req.ShopParentId = c.Query("shopParentId")
		req.ShopStatus = c.Query("shopStatus")

		// Parse pagination params from query string
		if pageNoStr := c.Query("pageNo"); pageNoStr != "" {
			if pageNo, err := strconv.ParseInt(pageNoStr, 10, 32); err == nil {
				pageNo := int32(pageNo)
				req.PageNo = pageNo
			}
		}
		if pageSizeStr := c.Query("pageSize"); pageSizeStr != "" {
			if pageSize, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil {
				pageSz := int32(pageSize)
				req.PageSize = pageSz
			}
		}
	}

	log.Printf("[API] QueryShop request: shopId=%s, shopParentType=%s, shopParentId=%s",
		req.ShopID, req.ShopParentType, req.ShopParentId)

	danaReq := &dana.QueryShopRequest{
		ShopID:         req.ShopID,
		ShopIdType:     dana.ShopIdTypeExternalID, // BUGFIX #3: Default to EXTERNAL_SHOP_ID
		ShopParentType: req.ShopParentType,
		ShopParentId:   req.ShopParentId,
		ShopStatus:     req.ShopStatus,
		PageNo:         req.PageNo,
		PageSize:       req.PageSize,
	}

	// Set default pagination
	if danaReq.PageNo <= 0 {
		danaReq.PageNo = 1
	}
	if danaReq.PageSize <= 0 {
		danaReq.PageSize = 20
	}

	resp, err := h.danaClient.QueryShop(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[API] QueryShop Error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":         "Failed to query shop",
			"details":       err.Error(),
			"request":       gin.H{"shopId": req.ShopID, "shopParentType": req.ShopParentType, "shopParentId": req.ShopParentId},
			"suggestion":    "Please check your DANA credentials and API endpoint path",
		})
		return
	}

	// Calculate total count
	totalCount := int64(0)
	if resp.TotalCount != nil {
		totalCount = *resp.TotalCount
	} else if len(resp.ShopDetailInfoList) > 0 {
		totalCount = int64(len(resp.ShopDetailInfoList))
	}

	c.JSON(http.StatusOK, gin.H{
		"responseCode":       resp.ResponseCode,
		"responseMessage":    resp.ResponseMessage,
		"totalCount":         totalCount,
		"shops":              resp.ShopDetailInfoList,
		"pageNo":             resp.PageNo,
		"pageSize":           resp.PageSize,
	})
}
