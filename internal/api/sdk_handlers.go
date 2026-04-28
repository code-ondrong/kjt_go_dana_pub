package api

import (
	"fmt"
	"net/http"
	"time"

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
	MerchantId      string `json:"merchantId,omitempty"`            // Auto-filled from config if empty; required by DANA
	ShopParentId    string `json:"shopParentId" binding:"required"` // MerchantId or DivisionId depending on shopParentType
	ExternalShopID  string `json:"externalShopId" binding:"required"`
	ShopName        string `json:"shopName" binding:"required"`
	ShopDesc        string `json:"shopDesc,omitempty"`
	ShopParentType  string `json:"shopParentType" binding:"required"` // MERCHANT, DIVISION, or EXTERNAL_DIVISION
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
	MerchantId         string                             `json:"merchantId" binding:"required"`
	ExternalDivisionId string                             `json:"externalDivisionId,omitempty"` // Optional - auto-generated if empty
	MainName           string                             `json:"mainName" binding:"required"`
	DivisionDesc       string                             `json:"divisionDesc,omitempty"`
	ParentRoleType     string                             `json:"parentRoleType,omitempty"` // MERCHANT, HEAD_OFFICE, BRANCH_OFFICE
	ParentDivisionId   string                             `json:"parentDivisionId,omitempty"`
	DivisionType       string                             `json:"divisionType,omitempty"`
	SizeType           string                             `json:"sizeType,omitempty"`
	MccCodes           []string                           `json:"mccCodes,omitempty"`
	DivisionAddress    *dana.AddressInfo                  `json:"divisionAddress,omitempty"`
	ExtInfo            *dana.CreateDivisionRequestExtInfo `json:"extInfo,omitempty"`
	BusinessEntity     string                             `json:"businessEntity,omitempty"`
	BusinessDocs       []dana.BusinessDocs                `json:"businessDocs,omitempty"`
	OwnerName          *dana.UserName                     `json:"ownerName,omitempty"`
	OwnerPhoneNumber   *dana.MobileNoInfo                 `json:"ownerPhoneNumber,omitempty"`
	OwnerIdType        string                             `json:"ownerIdType,omitempty"`
	OwnerIdNo          string                             `json:"ownerIdNo,omitempty"`
	OwnerAddress       *dana.AddressInfo                  `json:"ownerAddress,omitempty"`
	DirectorPics       []dana.PicInfo                     `json:"directorPics,omitempty"`
	NonDirectorPics    []dana.PicInfo                     `json:"nonDirectorPics,omitempty"`
	PgDivisionFlag     string                             `json:"pgDivisionFlag,omitempty"`
}

// HTTPQueryDivisionRequest represents the HTTP request for querying a division
type HTTPQueryDivisionRequest struct {
	MerchantId     string `form:"merchantId" binding:"required"`
	DivisionId     string `form:"divisionId" binding:"required"`
	DivisionIdType string `form:"divisionIdType"` // Optional: auto-detected if empty
}

// HTTPUpdateDivisionRequest represents the HTTP request for updating a division
type HTTPUpdateDivisionRequest struct {
	DivisionId            string                 `json:"divisionId" binding:"required"`
	DivisionIdType        string                 `json:"divisionIdType"` // Optional: auto-detected if empty
	MerchantId            string                 `json:"merchantId" binding:"required"`
	NewExternalDivisionId string                 `json:"newExternalDivisionId,omitempty"`
	MainName              *string                `json:"mainName,omitempty"`
	DivisionDesc          *string                `json:"divisionDesc,omitempty"`
	DivisionType          string                 `json:"divisionType,omitempty"`
	DivisionAddress       *dana.AddressInfo      `json:"divisionAddress,omitempty"`
	MccCodes              []string               `json:"mccCodes,omitempty"`
	ExtInfo               map[string]interface{} `json:"extInfo,omitempty"`
	ApiVersion            *string                `json:"apiVersion,omitempty"`
	BusinessEntity        *string                `json:"businessEntity,omitempty"`
	BusinessEndDate       *string                `json:"businessEndDate,omitempty"`
	BusinessDocs          []dana.BusinessDocs    `json:"businessDocs,omitempty"`
	OwnerName             *dana.UserName         `json:"ownerName,omitempty"`
	OwnerPhoneNumber      *dana.MobileNoInfo     `json:"ownerPhoneNumber,omitempty"`
	OwnerIdType           *string                `json:"ownerIdType,omitempty"`
	OwnerIdNo             *string                `json:"ownerIdNo,omitempty"`
	OwnerAddress          *dana.AddressInfo      `json:"ownerAddress,omitempty"`
	DirectorPics          []dana.PicInfo         `json:"directorPics,omitempty"`
	NonDirectorPics       []dana.PicInfo         `json:"nonDirectorPics,omitempty"`
	SizeType              *string                `json:"sizeType,omitempty"`
	PgDivisionFlag        *string                `json:"pgDivisionFlag,omitempty"`
	LogoUrlMap            map[string]string      `json:"logoUrlMap,omitempty"`
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

	log.Printf("[SDK Handler] CreateShop - ShopParentId: %s, ShopParentType: %s, ExternalShopID: %s, ShopName: %s",
		req.ShopParentId, req.ShopParentType, req.ExternalShopID, req.ShopName)

	// Auto-fill merchantId from config if not provided
	if req.MerchantId == "" {
		req.MerchantId = h.danaClient.GetConfig().MerchantID
		log.Printf("[SDK Handler] CreateShop - Auto-filled merchantId from config: %s", req.MerchantId)
	}

	// Convert HTTP request to DANA request
	danaReq := &dana.CreateShopRequest{
		MerchantId:      req.MerchantId,
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

	// Auto-generate externalDivisionId if not provided (timestamp-based unique ID)
	if req.ExternalDivisionId == "" {
		req.ExternalDivisionId = fmt.Sprintf("DIV-%d", time.Now().UnixNano())
		log.Printf("[SDK Handler] CreateDivision - Auto-generated externalDivisionId: %s", req.ExternalDivisionId)
	}

	// Default divisionType to REGION if not provided
	if req.DivisionType == "" {
		req.DivisionType = "REGION"
	}

	// Default parentRoleType to MERCHANT if not provided
	if req.ParentRoleType == "" {
		req.ParentRoleType = "MERCHANT"
	}

	// Default sizeType to UKE if not provided
	if req.SizeType == "" {
		req.SizeType = "UKE"
	}

	// Default businessEntity to INDIVIDU if not provided
	if req.BusinessEntity == "" {
		req.BusinessEntity = "INDIVIDU"
	}

	// Default ownerIdType to KTP if not provided
	if req.OwnerIdType == "" {
		req.OwnerIdType = "KTP"
	}

	// Default mccCodes if not provided
	if len(req.MccCodes) == 0 {
		req.MccCodes = []string{"5812"}
	}

	// Auto-fill divisionAddress if not provided (DANA requires full address)
	if req.DivisionAddress == nil {
		req.DivisionAddress = &dana.AddressInfo{
			Country:  "ID",
			Province: "DKI Jakarta",
			City:     "Jakarta Selatan",
			Area:     "Kebayoran Baru",
			Address1: "Jl. Sudirman No. 1",
			Postcode: "12190",
		}
		log.Printf("[SDK Handler] CreateDivision - Auto-filled default divisionAddress")
	} else {
		// Ensure area is filled (DANA requires it)
		if req.DivisionAddress.Area == "" {
			req.DivisionAddress.Area = "Kebayoran Baru"
		}
	}
	// Auto-fill ownerAddress if not provided (DANA requires full address)
	if req.OwnerAddress == nil {
		req.OwnerAddress = &dana.AddressInfo{
			Country:  "ID",
			Province: "DKI Jakarta",
			City:     "Jakarta Selatan",
			Area:     "Kebayoran Baru",
			Address1: "Jl. Sudirman No. 1",
			Postcode: "12190",
		}
		log.Printf("[SDK Handler] CreateDivision - Auto-filled default ownerAddress")
	} else {
		if req.OwnerAddress.Area == "" {
			req.OwnerAddress.Area = "Kebayoran Baru"
		}
	}
	// Auto-fill extInfo if not provided (DANA requires PIC_PHONENUMBER, SUBMITTER_EMAIL, etc.)
	if req.ExtInfo == nil {
		req.ExtInfo = &dana.CreateDivisionRequestExtInfo{
			PIC_PHONENUMBER: "081234567890",
			PIC_EMAIL:       "pic@kjt.co.id",
			SUBMITTER_EMAIL: "submitter@kjt.co.id",
			BRAND_NAME:      "KJT Brand",
			GOODS_SOLD_TYPE: "SERVICE",
			USECASE:         "PAYMENT",
			USER_PROFILING:  "BUYER",
			AVG_TICKET:      "50000",
			OMZET:           "100000000",
			EXT_URLS:        "https://kjt.co.id",
		}
		log.Printf("[SDK Handler] CreateDivision - Auto-filled default extInfo")
	} else {
		// Fill any missing required extInfo fields with defaults
		if req.ExtInfo.PIC_PHONENUMBER == "" {
			req.ExtInfo.PIC_PHONENUMBER = "081234567890"
		}
		if req.ExtInfo.SUBMITTER_EMAIL == "" {
			req.ExtInfo.SUBMITTER_EMAIL = "submitter@kjt.co.id"
		}
		if req.ExtInfo.PIC_EMAIL == "" {
			req.ExtInfo.PIC_EMAIL = "pic@kjt.co.id"
		}
		if req.ExtInfo.BRAND_NAME == "" {
			req.ExtInfo.BRAND_NAME = "KJT Brand"
		}
		if req.ExtInfo.GOODS_SOLD_TYPE == "" {
			req.ExtInfo.GOODS_SOLD_TYPE = "SERVICE"
		}
		if req.ExtInfo.USECASE == "" {
			req.ExtInfo.USECASE = "PAYMENT"
		}
		if req.ExtInfo.USER_PROFILING == "" {
			req.ExtInfo.USER_PROFILING = "BUYER"
		}
		if req.ExtInfo.AVG_TICKET == "" {
			req.ExtInfo.AVG_TICKET = "50000"
		}
		if req.ExtInfo.OMZET == "" {
			req.ExtInfo.OMZET = "100000000"
		}
		if req.ExtInfo.EXT_URLS == "" {
			req.ExtInfo.EXT_URLS = "https://kjt.co.id"
		}
	}

	// Ensure DirectorPics and NonDirectorPics are non-nil (DANA requires non-empty arrays)
	// DANA rejects empty arrays - must provide at least 1 PIC entry
	if req.DirectorPics == nil || len(req.DirectorPics) == 0 {
		req.DirectorPics = []dana.PicInfo{
			{PicName: "Director", PicPosition: "DIRECTOR"},
		}
	}
	if req.NonDirectorPics == nil || len(req.NonDirectorPics) == 0 {
		req.NonDirectorPics = []dana.PicInfo{
			{PicName: "PIC", PicPosition: "PIC"},
		}
	}

	danaReq := &dana.CreateDivisionRequest{
		MerchantId:         req.MerchantId,
		ExternalDivisionId: req.ExternalDivisionId,
		DivisionName:       req.MainName,
		DivisionDesc:       req.DivisionDesc,
		ParentRoleType:     req.ParentRoleType,
		ParentDivisionId:   req.ParentDivisionId,
		DivisionType:       req.DivisionType,
		SizeType:           req.SizeType,
		MccCodes:           req.MccCodes,
		DivisionAddress:    req.DivisionAddress,
		ExtInfo:            req.ExtInfo,
		BusinessEntity:     req.BusinessEntity,
		BusinessDocs:       req.BusinessDocs,
		OwnerName:          req.OwnerName,
		OwnerPhoneNumber:   req.OwnerPhoneNumber,
		OwnerIdType:        req.OwnerIdType,
		OwnerIdNo:          req.OwnerIdNo,
		OwnerAddress:       req.OwnerAddress,
		DirectorPics:       req.DirectorPics,
		NonDirectorPics:    req.NonDirectorPics,
		PgDivisionFlag:     req.PgDivisionFlag,
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

	// Auto-fetch current division type if not provided (DANA requires divisionType in update)
	divisionType := req.DivisionType
	sizeType := ""
	if req.SizeType != nil {
		sizeType = *req.SizeType
	}
	if divisionType == "" || sizeType == "" {
		log.Printf("[SDK Handler] UpdateDivision - Auto-fetching current division info...")
		queryResp, err := h.danaClient.QueryDivision(c.Request.Context(), &dana.QueryDivisionRequest{
			MerchantId:     req.MerchantId,
			DivisionId:     req.DivisionId,
			DivisionIdType: divisionIdType,
		})
		if err == nil && queryResp.DivisionDetail != nil {
			if divisionType == "" && queryResp.DivisionDetail.DivisionType != "" {
				divisionType = queryResp.DivisionDetail.DivisionType
				log.Printf("[SDK Handler] UpdateDivision - Auto-detected DivisionType: %s", divisionType)
			}
			// DANA requires sizeType in update - default to UKE if not found
			if sizeType == "" {
				sizeType = "UKE"
				log.Printf("[SDK Handler] UpdateDivision - Using default SizeType: %s", sizeType)
			}
		} else {
			if divisionType == "" {
				divisionType = "REGION"
			}
			if sizeType == "" {
				sizeType = "UKE"
			}
			log.Printf("[SDK Handler] UpdateDivision - Query failed, using defaults - DivisionType: %s, SizeType: %s", divisionType, sizeType)
		}
	}

	// Auto-fill divisionAddress if not provided (DANA requires full address in update)
	divisionAddress := req.DivisionAddress
	if divisionAddress == nil {
		divisionAddress = &dana.AddressInfo{
			Country:  "ID",
			Province: "DKI Jakarta",
			City:     "Jakarta Selatan",
			Area:     "Kebayoran Baru",
			Address1: "Jl. Sudirman No. 1",
			Postcode: "12190",
		}
		log.Printf("[SDK Handler] UpdateDivision - Auto-filled default divisionAddress")
	}

	danaReq := &dana.UpdateDivisionRequest{
		DivisionId:            req.DivisionId,
		DivisionIdType:        divisionIdType,
		MerchantId:            req.MerchantId,
		NewExternalDivisionId: req.NewExternalDivisionId,
		MainName:              req.MainName,
		DivisionDesc:          req.DivisionDesc,
		DivisionType:          divisionType,
		DivisionAddress:       divisionAddress,
		MccCodes:              req.MccCodes,
		ExtInfo:               req.ExtInfo,
		ApiVersion:            req.ApiVersion,
		BusinessEntity:        req.BusinessEntity,
		BusinessEndDate:       req.BusinessEndDate,
		BusinessDocs:          req.BusinessDocs,
		OwnerName:             req.OwnerName,
		OwnerPhoneNumber:      req.OwnerPhoneNumber,
		OwnerIdType:           req.OwnerIdType,
		OwnerIdNo:             req.OwnerIdNo,
		OwnerAddress:          req.OwnerAddress,
		DirectorPics:          req.DirectorPics,
		NonDirectorPics:       req.NonDirectorPics,
		SizeType:              &sizeType,
		PgDivisionFlag:        req.PgDivisionFlag,
		LogoUrlMap:            req.LogoUrlMap,
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

// AccountInquiry handles DANA account inquiries using SDK
func (h *SDKAPIHandler) AccountInquiry(c *gin.Context) {
	var req dana.AccountInquiryRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] AccountInquiry - Customer: %s", req.CustomerNumber)

	resp, err := h.danaClient.AccountInquiry(c.Request.Context(), &req)
	if err != nil {
		log.Printf("[SDK Handler] AccountInquiry error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	// Always wrap in APIResponse
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: resp})
}

// TransferToDana handles disbursement to DANA balance requests using SDK
func (h *SDKAPIHandler) TransferToDana(c *gin.Context) {
	var req dana.TransferToDanaRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] TransferToDana - Customer: %s, Amount: %s %s, FeeAmount: %s %s",
		req.CustomerNumber, req.Amount, req.Currency, req.FeeAmount, req.FeeCurrency)

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

// ============================================================
// PAYMENT GATEWAY HANDLERS
// ============================================================

// HTTPCreatePaymentOrderRequest represents the HTTP request for creating a payment order
type HTTPCreatePaymentOrderRequest struct {
	PartnerReferenceNo string `json:"partnerReferenceNo" binding:"required"`
	MerchantID         string `json:"merchantId" binding:"required"`
	Amount             string `json:"amount" binding:"required"`
	Currency           string `json:"currency" binding:"required"`
	OrderTitle         string `json:"orderTitle,omitempty"`
	ValidUpTo          string `json:"validUpTo,omitempty"`
	Notes              string `json:"notes,omitempty"`
	PayMethod          string `json:"payMethod,omitempty"`
	PayOption          string `json:"payOption,omitempty"`
}

// HTTPQueryPaymentRequest represents the HTTP request for querying payment status
type HTTPQueryPaymentRequest struct {
	PartnerReferenceNo string `form:"partnerReferenceNo" binding:"required"`
	MerchantID         string `form:"merchantId"`
}

// HTTPCancelPaymentRequest represents the HTTP request for cancelling payment
type HTTPCancelPaymentRequest struct {
	PartnerReferenceNo  string `json:"partnerReferenceNo" binding:"required"`
	OriginalReferenceNo string `json:"originalReferenceNo,omitempty"`
	Reason              string `json:"reason,omitempty"`
}

// HTTPRefundPaymentRequest represents the HTTP request for refunding payment
type HTTPRefundPaymentRequest struct {
	PartnerReferenceNo  string `json:"partnerReferenceNo" binding:"required"`
	OriginalReferenceNo string `json:"originalReferenceNo" binding:"required"`
	RefundAmount        string `json:"refundAmount" binding:"required"`
	Currency            string `json:"currency" binding:"required"`
	Reason              string `json:"reason" binding:"required"`
}

// CreatePaymentOrder handles payment order creation
func (h *SDKAPIHandler) CreatePaymentOrder(c *gin.Context) {
	var req HTTPCreatePaymentOrderRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] CreatePaymentOrder - PartnerRef: %s, Merchant: %s, Amount: %s",
		req.PartnerReferenceNo, req.MerchantID, req.Amount)

	// Build Payment Gateway request
	pgReq := &dana.CreatePaymentOrderRequest{
		PartnerReferenceNo: req.PartnerReferenceNo,
		MerchantID:         req.MerchantID,
		Amount: dana.Amount{
			Value:    req.Amount,
			Currency: req.Currency,
		},
		Notes:     req.Notes,
		ValidUpTo: req.ValidUpTo,
	}

	// Add additional info for order title
	if req.OrderTitle != "" {
		pgReq.AdditionalInfo = &dana.PaymentOrderAdditionalInfo{
			Order: &dana.PaymentOrderInfo{
				OrderTitle: req.OrderTitle,
			},
		}
	}

	// Add pay option details if specified
	if req.PayMethod != "" || req.PayOption != "" {
		pgReq.PayOptionDetails = []dana.PaymentOptionDetail{
			{
				PayMethod: req.PayMethod,
				PayOption: req.PayOption,
			},
		}
	}

	// Call Payment Gateway
	resp, err := h.danaClient.CreatePaymentOrder(c.Request.Context(), pgReq)
	if err != nil {
		log.Printf("[SDK Handler] CreatePaymentOrder error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// QueryPayment handles payment status query
func (h *SDKAPIHandler) QueryPayment(c *gin.Context) {
	var req HTTPQueryPaymentRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, APIResponse{
			Success: false,
			Error:   "Invalid query parameters: " + err.Error(),
		})
		return
	}

	log.Printf("[SDK Handler] QueryPayment - PartnerRef: %s", req.PartnerReferenceNo)

	pgReq := &dana.QueryPaymentRequest{
		PartnerReferenceNo: req.PartnerReferenceNo,
		MerchantID:         req.MerchantID,
	}

	resp, err := h.danaClient.QueryPayment(c.Request.Context(), pgReq)
	if err != nil {
		log.Printf("[SDK Handler] QueryPayment error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// CancelPayment handles payment cancellation
func (h *SDKAPIHandler) CancelPayment(c *gin.Context) {
	var req HTTPCancelPaymentRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] CancelPayment - PartnerRef: %s", req.PartnerReferenceNo)

	pgReq := &dana.CancelPaymentRequest{
		PartnerReferenceNo:  req.PartnerReferenceNo,
		OriginalReferenceNo: req.OriginalReferenceNo,
		Reason:              req.Reason,
	}

	resp, err := h.danaClient.CancelPayment(c.Request.Context(), pgReq)
	if err != nil {
		log.Printf("[SDK Handler] CancelPayment error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// RefundPayment handles payment refund
func (h *SDKAPIHandler) RefundPayment(c *gin.Context) {
	var req HTTPRefundPaymentRequest
	if err := h.checkBindJSON(c, &req); err != nil {
		return
	}

	log.Printf("[SDK Handler] RefundPayment - PartnerRef: %s, Amount: %s",
		req.PartnerReferenceNo, req.RefundAmount)

	pgReq := &dana.RefundPaymentRequest{
		PartnerReferenceNo:  req.PartnerReferenceNo,
		OriginalReferenceNo: req.OriginalReferenceNo,
		RefundAmount: dana.Amount{
			Value:    req.RefundAmount,
			Currency: req.Currency,
		},
		Reason: req.Reason,
	}

	resp, err := h.danaClient.RefundPayment(c.Request.Context(), pgReq)
	if err != nil {
		log.Printf("[SDK Handler] RefundPayment error: %v", err)
		c.JSON(http.StatusInternalServerError, APIResponse{Success: false, Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    resp,
	})
}

// WebhookPayment handles payment webhook notifications from DANA
// DANA may send GET for URL verification and POST for actual notifications
func (h *SDKAPIHandler) WebhookPayment(c *gin.Context) {
	// Handle GET request (URL verification by DANA)
	if c.Request.Method == "GET" {
		log.Printf("[SDK Handler] WebhookPayment GET verification received")
		// DANA may expect a specific response for verification
		// Return 200 with a simple message or challenge response
		c.JSON(http.StatusOK, map[string]string{
			"status":  "active",
			"message": "Webhook endpoint is verified and active",
		})
		return
	}

	// Handle POST request (actual notification)
	var notification dana.PaymentGatewayNotification
	if err := c.ShouldBindJSON(&notification); err != nil {
		log.Printf("[SDK Handler] WebhookPayment bind error: %v", err)
		c.JSON(http.StatusBadRequest, APIResponse{Success: false, Error: "Invalid webhook payload"})
		return
	}

	log.Printf("[SDK Handler] WebhookPayment received - PartnerRef: %s, Status: %s, Amount: %s",
		notification.PartnerReferenceNo, notification.TransactionStatus, notification.Amount.Value)

	// TODO: Verify webhook signature using DANA_PUBLIC_KEY if required
	// For now, process the notification

	// Publish to SSE broker for real-time updates
	if h.sseBroker != nil {
		eventData := map[string]interface{}{
			"partnerReferenceNo": notification.PartnerReferenceNo,
			"referenceNo":        notification.ReferenceNo,
			"status":             notification.TransactionStatus,
			"amount":             notification.Amount.Value,
			"paidTime":           notification.PaidTime,
		}
		h.sseBroker.Publish(notification.PartnerReferenceNo, "payment_update", eventData)
		log.Printf("[SDK Handler] WebhookPayment published to SSE channel: %s", notification.PartnerReferenceNo)
	}

	// Respond with success to DANA
	// DANA expects HTTP 200 with specific response format
	c.JSON(http.StatusOK, map[string]string{
		"status": "received",
	})
}
