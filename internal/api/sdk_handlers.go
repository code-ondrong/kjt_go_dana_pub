package api

import (
	"net/http"

	"log"

	"github.com/gin-gonic/gin"

	"kjt_go_dana/internal/dana"
	"kjt_go_dana/internal/sse"
)

// SDKAPIHandler handles HTTP requests using the Official DANA SDK
type SDKAPIHandler struct {
	danaClient *dana.SDKClient
	sseBroker  *sse.Broker
}

// NewSDKAPIHandler creates a new SDK-based API handler
func NewSDKAPIHandler(danaClient *dana.SDKClient, sseBroker *sse.Broker) *SDKAPIHandler {
	return &SDKAPIHandler{
		danaClient: danaClient,
		sseBroker:  sseBroker,
	}
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ============================================================
// REQUEST/RESPONSE TYPES FOR HTTP API
// ============================================================

// HTTPCreateShopRequest represents the HTTP request for creating a shop
type HTTPCreateShopRequest struct {
	ShopParentId    string `json:"shopParentId" binding:"required"`
	ExternalShopID  string `json:"externalShopId" binding:"required"`
	ShopName        string `json:"shopName" binding:"required"`
	ShopDesc        string `json:"shopDesc,omitempty"`
	ShopParentType  string `json:"shopParentType" binding:"required"`
	SizeType        string `json:"sizeType" binding:"required"`
	ShopAddress     string `json:"shopAddress,omitempty"`
	ShopCity        string `json:"shopCity,omitempty"`
	ShopProvince    string `json:"shopProvince,omitempty"`
	ShopCountryCode string `json:"shopCountryCode,omitempty"`
	ShopPostalCode  string `json:"shopPostalCode,omitempty"`
}

// HTTPQueryShopRequest represents the HTTP request for querying a shop
type HTTPQueryShopRequest struct {
	ShopParentId string `form:"shopParentId" binding:"required"` // MerchantId
	ShopID       string `form:"shopId" binding:"required"`       // Mandatory
	ShopIdType   string `form:"shopIdType"`                      // Optional
}

// HTTPUpdateShopRequest represents the HTTP request for updating a shop
type HTTPUpdateShopRequest struct {
	ShopID       string  `json:"shopId" binding:"required"`
	ShopIdType   string  `json:"shopIdType" binding:"required"`
	ShopParentId string  `json:"shopParentId" binding:"required"`
	ShopName     *string `json:"shopName,omitempty"`
	ShopDesc     *string `json:"shopDesc,omitempty"`
}

// HTTPCreateDivisionRequest represents the HTTP request for creating a division
type HTTPCreateDivisionRequest struct {
	MerchantId         string `json:"merchantId" binding:"required"`
	ExternalDivisionId string `json:"externalDivisionId" binding:"required"`
	MainName           string `json:"mainName" binding:"required"`
	DivisionDesc       string `json:"divisionDesc,omitempty"`
}

// HTTPQueryDivisionRequest represents the HTTP request for querying a division
type HTTPQueryDivisionRequest struct {
	MerchantId     string `form:"merchantId" binding:"required"`
	DivisionId     string `form:"divisionId" binding:"required"`
	DivisionIdType string `form:"divisionIdType"` // Optional: auto-detected if empty
}

// HTTPUpdateDivisionRequest represents the HTTP request for updating a division
type HTTPUpdateDivisionRequest struct {
	DivisionId     string  `json:"divisionId" binding:"required"`
	DivisionIdType string  `json:"divisionIdType"` // Optional: auto-detected if empty
	MerchantId     string  `json:"merchantId" binding:"required"`
	MainName       *string `json:"mainName,omitempty"`
	DivisionDesc   *string `json:"divisionDesc,omitempty"`
}

// ============================================================
// SHOP MANAGEMENT HANDLERS - SDK Implementation
// ============================================================

// CreateShop handles shop creation requests using SDK
func (h *SDKAPIHandler) CreateShop(c *gin.Context) {
	var req HTTPCreateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SDK Handler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	log.Printf("[SDK Handler] CreateShop - ShopParentId: %s, ExternalShopID: %s, ShopName: %s",
		req.ShopParentId, req.ExternalShopID, req.ShopName)

	// Convert HTTP request to DANA request
	danaReq := &dana.CreateShopRequest{
		ShopParentId:    req.ShopParentId,
		ShopAlias:       req.ExternalShopID,
		ShopName:        req.ShopName,
		ShopParentType:  req.ShopParentType,
		SizeType:        req.SizeType,
		ShopAddress:     req.ShopDesc,
		ShopCity:        req.ShopCity,
		ShopProvince:    req.ShopProvince,
		ShopCountryCode: req.ShopCountryCode,
		ShopPostalCode:  req.ShopPostalCode,
	}

	// Call DANA API using SDK
	resp, err := h.danaClient.CreateShop(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[SDK Handler] CreateShop failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to create shop: " + err.Error(),
		})
		return
	}

	log.Printf("[SDK Handler] CreateShop success - ShopID: %s", resp.ShopID)

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// QueryShop handles shop query requests using SDK
func (h *SDKAPIHandler) QueryShop(c *gin.Context) {
	var req HTTPQueryShopRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		log.Printf("[SDK Handler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "merchantId (shopParentId) and shopId are required for EXTERNAL_ID query",
		})
		return
	}

	// Smart Detection: If shopId is numeric and long, it's likely an INNER_ID
	// Use user-provided type if available, otherwise detect
	shopIdType := req.ShopIdType
	if shopIdType == "" {
		shopIdType = "EXTERNAL_ID"
		if isNumeric(req.ShopID) && len(req.ShopID) >= 16 {
			shopIdType = "INNER_ID"
		}
	}

	log.Printf("[SDK Handler] QueryShop - ShopID: %s, Using Type: %s", req.ShopID, shopIdType)

	// Convert HTTP request to DANA request
	danaReq := &dana.QueryShopRequest{
		ShopParentId: req.ShopParentId,
		ShopID:       req.ShopID,
		ShopIdType:   shopIdType,
	}

	// Call DANA API using SDK
	resp, err := h.danaClient.QueryShop(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[SDK Handler] QueryShop failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to query shop: " + err.Error(),
		})
		return
	}

	log.Printf("[SDK Handler] QueryShop success - Found shops")

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// UpdateShop handles shop update requests using SDK
func (h *SDKAPIHandler) UpdateShop(c *gin.Context) {
	var req HTTPUpdateShopRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("[SDK Handler] Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   err.Error(),
		})
		return
	}

	log.Printf("[SDK Handler] UpdateShop - ShopID: %s", req.ShopID)

	// Convert HTTP request to DANA request
	danaReq := &dana.UpdateShopRequest{
		ShopID:       req.ShopID,
		ShopIdType:   req.ShopIdType,
		ShopParentId: req.ShopParentId,
	}

	// Handle optional fields
	if req.ShopName != nil {
		danaReq.ShopName = *req.ShopName
	}
	if req.ShopDesc != nil {
		danaReq.ShopAddress = *req.ShopDesc
	}

	// Call DANA API using SDK
	resp, err := h.danaClient.UpdateShop(c.Request.Context(), danaReq)
	if err != nil {
		log.Printf("[SDK Handler] UpdateShop failed: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{
			Success: false,
			Error:   "Failed to update shop: " + err.Error(),
		})
		return
	}

	log.Printf("[SDK Handler] UpdateShop success")

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// ============================================================
// DIVISION MANAGEMENT HANDLERS - SDK Implementation
// ============================================================

// CreateDivision handles division creation requests using SDK
func (h *SDKAPIHandler) CreateDivision(c *gin.Context) {
	var req HTTPCreateDivisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return
	}

	danaReq := &dana.CreateDivisionRequest{
		MerchantId:         req.MerchantId,
		ExternalDivisionId: req.ExternalDivisionId,
		MainName:           req.MainName,
		DivisionDesc:       req.DivisionDesc,
	}

	resp, err := h.danaClient.CreateDivision(c.Request.Context(), danaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: resp})
}

// QueryDivision handles division query requests using SDK
func (h *SDKAPIHandler) QueryDivision(c *gin.Context) {
	var req HTTPQueryDivisionRequest
	if err := h.checkBindQuery(c, &req); err != nil {
		return
	}

	// Smart Detection for DivisionIdType
	divisionIdType := req.DivisionIdType
	if divisionIdType == "" {
		divisionIdType = "EXTERNAL_ID"
		if isNumeric(req.DivisionId) && len(req.DivisionId) >= 16 {
			divisionIdType = "INNER_ID"
		}
	}

	log.Printf("[SDK Handler] QueryDivision - DivisionId: %s, Using Type: %s", req.DivisionId, divisionIdType)

	danaReq := &dana.QueryDivisionRequest{
		MerchantId:     req.MerchantId,
		DivisionId:     req.DivisionId,
		DivisionIdType: divisionIdType,
	}

	resp, err := h.danaClient.QueryDivision(c.Request.Context(), danaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: resp})
}

// UpdateDivision handles division update requests using SDK
func (h *SDKAPIHandler) UpdateDivision(c *gin.Context) {
	var req HTTPUpdateDivisionRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	// Smart Detection for DivisionIdType
	divisionIdType := req.DivisionIdType
	if divisionIdType == "" {
		divisionIdType = "EXTERNAL_ID"
		if isNumeric(req.DivisionId) && len(req.DivisionId) >= 16 {
			divisionIdType = "INNER_ID"
		}
	}

	log.Printf("[SDK Handler] UpdateDivision - DivisionId: %s, Using Type: %s", req.DivisionId, divisionIdType)

	danaReq := &dana.UpdateDivisionRequest{
		DivisionId:     req.DivisionId,
		DivisionIdType: divisionIdType,
		MerchantId:     req.MerchantId,
		MainName:       req.MainName,
		DivisionDesc:   req.DivisionDesc,
	}

	resp, err := h.danaClient.UpdateDivision(c.Request.Context(), danaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: resp})
}

// Helper to check query binding
func (h *SDKAPIHandler) checkBindQuery(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindQuery(req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return err
	}
	return nil
}

// Helper to check JSON binding
func (h *SDKAPIHandler) checkBindJSON(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: err.Error()})
		return err
	}
	return nil
}

// ============================================================
// DISBURSEMENT HANDLERS - SDK Implementation
// ============================================================

// TransferToDana handles disbursement to DANA balance requests using SDK
func (h *SDKAPIHandler) TransferToDana(c *gin.Context) {
	var req dana.TransferToDanaRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] TransferToDana - Customer: %s, Amount: %s",
		req.CustomerNumber, req.Amount)

	resp, err := h.danaClient.TransferToDana(c.Request.Context(), &req)
	if err != nil {
		log.Printf("[SDK Handler] TransferToDana error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{Success: true, Data: resp})
}

// TransferToDanaInquiryStatus handles transfer status inquiry using SDK
func (h *SDKAPIHandler) TransferToDanaInquiryStatus(c *gin.Context) {
	var req dana.TransferToDanaInquiryStatusRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] TransferToDanaInquiryStatus - OriginalPartnerRef: %s",
		req.OriginalPartnerReferenceNo)

	resp, err := h.danaClient.TransferToDanaInquiryStatus(c.Request.Context(), &req)
	if err != nil {
		log.Printf("[SDK Handler] TransferToDanaInquiryStatus error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// ============================================================
// HEALTH CHECK
// ============================================================

// HealthCheck returns SDK health status
func (h *SDKAPIHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data: map[string]interface{}{
			"status":  "ok",
			"backend": "official_dana_sdk",
			"version": "v1.2.11",
		},
	})
}

// Helper to check if string is numeric
func isNumeric(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
