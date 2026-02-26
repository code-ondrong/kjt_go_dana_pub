package dana

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	danaSDK "github.com/dana-id/dana-go"
	configSDK "github.com/dana-id/dana-go/config"
	"github.com/dana-id/dana-go/disbursement/v1"
	merchant_management "github.com/dana-id/dana-go/merchant_management/v1"

	"kjt_go_dana/internal/config"
)

// containsPEMHeaders checks if the key already has PEM headers
func containsPEMHeaders(key string) bool {
	return strings.HasPrefix(key, "-----BEGIN") && strings.Contains(key, "-----END")
}

// formatPEMKey adds PEM headers to a base64-encoded key
func formatPEMKey(key, keyType string) string {
	// Remove any existing whitespace
	key = strings.TrimSpace(key)
	// Remove existing PEM headers if present
	if strings.HasPrefix(key, "-----BEGIN") {
		parts := strings.Split(key, "-----")
		if len(parts) >= 3 {
			key = strings.TrimSpace(parts[2])
		}
	}
	// Add proper PEM headers
	return fmt.Sprintf("-----BEGIN %s-----\n%s\n-----END %s-----", keyType, key, keyType)
}

// SDKClient wraps the official DANA SDK
type SDKClient struct {
	danaClient      *danaSDK.APIClient
	cfg             *config.DANAConfig
	merchantAPI     *merchant_management.MerchantManagementAPIService
	disbursementAPI *disbursement.DisbursementAPIService
}

// NewSDKClient creates a new SDK-based DANA client
func NewSDKClient(cfg *config.DANAConfig) (*SDKClient, error) {
	// Get environment from config
	env := configSDK.ENV_SANDBOX
	if cfg.Environment == "production" || cfg.Environment == "PRODUCTION" {
		env = configSDK.ENV_PRODUCTION
	}

	// Use NewConfiguration like the official SDK documentation
	configuration := configSDK.NewConfiguration()

	// Set API key for OPEN_API authentication
	configuration.APIKey = &configSDK.APIKey{
		DANA_ENV:      string(env),
		X_PARTNER_ID:  cfg.ClientID,
		CLIENT_ID:     cfg.ClientID,
		CLIENT_SECRET: cfg.ClientSecret,
		PRIVATE_KEY:   cfg.PrivateKey,
		ORIGIN:        cfg.Origin,
	}

	// Don't override servers - let SDK determine based on DANA_ENV
	if env == configSDK.ENV_SANDBOX {
		log.Printf("[DANA SDK] Using Sandbox environment")
	} else {
		log.Printf("[DANA SDK] Using Production environment")
	}

	// Create SDK client
	danaClient := danaSDK.NewAPIClient(configuration)

	log.Printf("[DANA SDK] Client initialized successfully")
	log.Printf("[DANA SDK] Environment: %s", env)
	log.Printf("[DANA SDK] Client ID: %s", cfg.ClientID)
	if cfg.Origin != "" {
		log.Printf("[DANA SDK] ORIGIN header set: %s", cfg.Origin)
	}

	return &SDKClient{
		danaClient:      danaClient,
		cfg:             cfg,
		merchantAPI:     danaClient.MerchantManagementAPI,
		disbursementAPI: danaClient.DisbursementAPI,
	}, nil
}

// ============================================================
// SHOP MANAGEMENT API - SDK Implementation
// ============================================================

// CreateShop creates a new shop using the official SDK
func (c *SDKClient) CreateShop(ctx context.Context, req *CreateShopRequest) (*CreateShopResponse, error) {
	log.Printf("[DANA SDK] CreateShop called with shopParentId: %s, shopName: %s",
		req.ShopParentId, req.ShopName)

	// Map sizeType from common names to DANA SDK enum values
	// MICRO -> UMI, SMALL -> UKE, MEDIUM -> UME, LARGE -> UBE
	mappedSizeType := mapSizeType(req.SizeType)

	// Create request directly with all required fields
	sdkReq := &merchant_management.CreateShopRequest{
		MerchantId:     req.ShopParentId,
		ShopParentType: req.ShopParentType,
		MainName:       req.ShopName,
		ExternalShopId: req.ShopAlias,
		SizeType:       mappedSizeType,
		// MCC codes is required - use default if not provided
		MccCodes: []string{"5734"}, // 5734 = Computer Software Stores (default)
	}

	// Set optional description
	if req.ShopAddress != "" {
		sdkReq.ShopDesc = &req.ShopAddress
	}

	// Set shop address only if we have meaningful address data
	if req.ShopCity != "" || req.ShopProvince != "" || req.ShopCountryCode != "" {
		address := merchant_management.NewAddressInfo()
		if req.ShopCountryCode != "" {
			address.SetCountry(req.ShopCountryCode)
		}
		if req.ShopProvince != "" {
			address.SetProvince(req.ShopProvince)
		}
		if req.ShopCity != "" {
			address.SetCity(req.ShopCity)
		}
		if req.ShopAddress != "" {
			address.SetAddress1(req.ShopAddress)
		}
		if req.ShopPostalCode != "" {
			address.SetPostcode(req.ShopPostalCode)
		}
		sdkReq.ShopAddress = address
	}

	// Log request JSON for debugging
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] CreateShop request JSON:\n%s", string(reqJSON))

	// Call SDK
	sdkResp, httpResp, err := c.merchantAPI.CreateShop(ctx).
		CreateShopRequest(*sdkReq).
		Execute()

	if err != nil {
		log.Printf("[DANA SDK] CreateShop error: %v", err)
		return nil, fmt.Errorf("create shop failed: %w", err)
	}
	defer httpResp.Body.Close()

	log.Printf("[DANA SDK] CreateShop HTTP status: %d", httpResp.StatusCode)

	// Log response body for debugging
	bodyBytes, _ := json.MarshalIndent(sdkResp.Response.Body, "", "  ")
	log.Printf("[DANA SDK] CreateShop response JSON:\n%s", string(bodyBytes))

	// Convert SDK response to our response type
	resp := &CreateShopResponse{
		ResponseCode:    sdkResp.Response.Body.ResultInfo.ResultCodeId,
		ResponseMessage: sdkResp.Response.Body.ResultInfo.ResultMsg,
	}

	if sdkResp.Response.Body.ShopId != nil {
		resp.ShopID = *sdkResp.Response.Body.ShopId
	}
	resp.MerchantID = req.ShopParentId
	resp.ShopName = req.ShopName
	resp.ShopStatus = "ACTIVE"

	log.Printf("[DANA SDK] CreateShop success. ResponseCode: %s, ShopID: %s", resp.ResponseCode, resp.ShopID)
	return resp, nil
}

// mapSizeType maps common size type names to DANA SDK enum values
func mapSizeType(sizeType string) string {
	// Normalize to uppercase
	sizeType = strings.ToUpper(sizeType)

	switch sizeType {
	case "MICRO", "UMI":
		return "UMI" // Usaha Mikro
	case "SMALL", "UKE":
		return "UKE" // Usaha Kecil
	case "MEDIUM", "UME":
		return "UME" // Usaha Menengah
	case "LARGE", "UBE":
		return "UBE" // Usaha Besar
	default:
		// Default to UKE if unknown
		log.Printf("[DANA SDK] Unknown sizeType '%s', defaulting to UKE (Small)", sizeType)
		return "UKE"
	}
}

// QueryShop queries shop information using the official SDK
func (c *SDKClient) QueryShop(ctx context.Context, req *QueryShopRequest) (*QueryShopResponse, error) {
	log.Printf("[DANA SDK] QueryShop called with shopID: %s, shopIdType: %s",
		req.ShopID, req.ShopIdType)

	// Create SDK request
	sdkReq := merchant_management.NewQueryShopRequest(req.ShopID, req.ShopIdType)

	// In DANA:
	// 1. For EXTERNAL_ID: MerchantId is MANDATORY.
	// 2. For INNER_ID: MerchantId is NOT needed (globally unique).
	//    Providing it can cause empty results if the MerchantId doesn't match perfectly.
	if req.ShopParentId != "" {
		if req.ShopIdType == "EXTERNAL_ID" {
			sdkReq.SetMerchantId(req.ShopParentId)
			log.Printf("[DANA SDK] EXTERNAL_ID query: Setting MerchantId: %s", req.ShopParentId)
		} else {
			log.Printf("[DANA SDK] INNER_ID query: Skipping MerchantId to avoid filter mismatch")
		}
	}

	// Log the actual request for debugging
	log.Printf("[DANA SDK] Sending QueryShop request:")
	log.Printf("  - ShopID: %s", sdkReq.ShopId)
	log.Printf("  - ShopIdType: %s", sdkReq.ShopIdType)
	if sdkReq.MerchantId != nil {
		log.Printf("  - MerchantId: %s", *sdkReq.MerchantId)
	} else {
		if req.ShopIdType == "EXTERNAL_ID" {
			log.Printf("  - MerchantId: <nil> (WARNING: Required for EXTERNAL_ID)")
		} else {
			log.Printf("  - MerchantId: <nil> (Not Required for INNER_ID)")
		}
	}

	// Call SDK
	sdkResp, httpResp, err := c.merchantAPI.QueryShop(ctx).
		QueryShopRequest(*sdkReq).
		Execute()

	if err != nil {
		log.Printf("[DANA SDK] QueryShop error: %v", err)
		return nil, fmt.Errorf("query shop failed: %w", err)
	}
	defer httpResp.Body.Close()

	log.Printf("[DANA SDK] QueryShop HTTP status: %d", httpResp.StatusCode)

	// Log the actual response body for debugging
	bodyJSON, _ := json.MarshalIndent(sdkResp.Response.Body, "", "  ")
	log.Printf("[DANA SDK] QueryShop response Body:\n%s", string(bodyJSON))

	// Convert SDK response to our response type
	resp := &QueryShopResponse{
		ResponseCode:    sdkResp.Response.Body.ResultInfo.ResultCodeId,
		ResponseMessage: sdkResp.Response.Body.ResultInfo.ResultMsg,
	}

	// Parse shop resource info
	if sdkResp.Response.Body.ShopResourceInfo != nil {
		shop := sdkResp.Response.Body.ShopResourceInfo

		// Only map if we actually have some data (check a few key fields)
		if shop.MainName != nil || shop.ExternalShopId != nil || shop.MerchantId != nil {
			shopInfo := ShopInfo{
				ShopID:     req.ShopID, // Fallback to requested ID
				ShopStatus: "ACTIVE",   // Default status
			}

			// Map basic info
			if shop.MainName != nil {
				shopInfo.ShopName = *shop.MainName
			}
			if shop.ExternalShopId != nil {
				shopInfo.ShopAlias = *shop.ExternalShopId
				// If we queried by INNER_ID, the response EXTERNAL_ID is useful
				if req.ShopIdType == "INNER_ID" {
					shopInfo.ShopAlias = *shop.ExternalShopId
				} else {
					shopInfo.ShopID = *shop.ExternalShopId
				}
			}
			if shop.MerchantId != nil {
				shopInfo.ShopParentId = *shop.MerchantId
			}
			if shop.ParentRoleType != nil {
				shopInfo.ShopParentType = *shop.ParentRoleType
			}
			if shop.SizeType != nil {
				shopInfo.SizeType = *shop.SizeType
			}
			if shop.ParentDivisionId != nil {
				shopInfo.ParentDivisionId = *shop.ParentDivisionId
			}
			if shop.Nmid != nil {
				shopInfo.Nmid = *shop.Nmid
			}
			if shop.Lat != nil {
				shopInfo.ShopLat = *shop.Lat
			}
			if shop.Ln != nil {
				shopInfo.ShopLong = *shop.Ln
			}

			// Map maps
			if shop.LogoUrlMap != nil {
				shopInfo.LogoUrlMap = shop.LogoUrlMap
			}
			if shop.ExtInfo != nil {
				shopInfo.ExtInfo = shop.ExtInfo
			}

			if shop.ShopAddress != nil {
				addr := shop.ShopAddress
				if addr.Address1 != nil {
					shopInfo.ShopAddress = *addr.Address1
				}
				if addr.Address2 != nil {
					shopInfo.ShopAddress2 = *addr.Address2
				}
				if addr.City != nil {
					shopInfo.ShopCity = *addr.City
				}
				if addr.Province != nil {
					shopInfo.ShopProvince = *addr.Province
				}
				if addr.Country != nil {
					shopInfo.ShopCountryCode = *addr.Country
				}
				if addr.Postcode != nil {
					shopInfo.ShopPostalCode = *addr.Postcode
				}
				if addr.SubDistrict != nil {
					shopInfo.ShopSubDistrict = *addr.SubDistrict
				}
				if addr.Area != nil {
					shopInfo.ShopArea = *addr.Area
				}
			}

			resp.ShopDetailInfoList = []ShopInfo{shopInfo}
		}
	}

	// Always provide the raw body for "results yang lengkap"
	resp.RawDANA = sdkResp.Response.Body

	log.Printf("[DANA SDK] QueryShop success. ResponseCode: %s", resp.ResponseCode)
	return resp, nil
}

// UpdateShop updates shop information using the official SDK
func (c *SDKClient) UpdateShop(ctx context.Context, req *UpdateShopRequest) (*UpdateShopResponse, error) {
	log.Printf("[DANA SDK] UpdateShop called with shopID: %s", req.ShopID)

	// Map sizeType if provided - default to UKE (Small Business)
	mappedSizeType := "UKE"
	if req.SizeType != "" {
		mappedSizeType = mapSizeType(req.SizeType)
	}

	// Convert request to SDK type - need to provide required fields
	sdkReq := &merchant_management.UpdateShopRequest{
		ShopId:      req.ShopID,
		MerchantId:  req.ShopParentId, // Required field
		ShopIdType:  req.ShopIdType,
		ShopAddress: *merchant_management.NewAddressInfo(), // Required field
		SizeType:    &mappedSizeType,                       // SizeType is required by DANA API
	}

	// Set optional fields
	if req.ShopName != "" {
		sdkReq.MainName = &req.ShopName
	}
	if req.ShopAlias != "" {
		sdkReq.NewExternalShopId = &req.ShopAlias
	}
	if req.ShopAddress != "" {
		sdkReq.ShopDesc = &req.ShopAddress
	}

	// Log request JSON for debugging
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] UpdateShop request JSON:\n%s", string(reqJSON))

	// Call SDK
	sdkResp, httpResp, err := c.merchantAPI.UpdateShop(ctx).
		UpdateShopRequest(*sdkReq).
		Execute()

	if err != nil {
		log.Printf("[DANA SDK] UpdateShop error: %v", err)
		return nil, fmt.Errorf("update shop failed: %w", err)
	}
	defer httpResp.Body.Close()

	log.Printf("[DANA SDK] UpdateShop HTTP status: %d", httpResp.StatusCode)

	// Log response body for debugging
	bodyBytes, _ := json.MarshalIndent(sdkResp.Response.Body, "", "  ")
	log.Printf("[DANA SDK] UpdateShop response JSON:\n%s", string(bodyBytes))

	// Convert SDK response to our response type
	resp := &UpdateShopResponse{
		ResponseCode:    sdkResp.Response.Body.ResultInfo.ResultCodeId,
		ResponseMessage: sdkResp.Response.Body.ResultInfo.ResultMsg,
		ShopID:          req.ShopID,
	}

	if req.ShopName != "" {
		resp.ShopName = req.ShopName
	}
	resp.ShopStatus = "ACTIVE"

	log.Printf("[DANA SDK] UpdateShop success. ResponseCode: %s", resp.ResponseCode)
	return resp, nil
}

// ============================================================
// DIVISION MANAGEMENT API - SDK Implementation
// ============================================================

// CreateDivision creates a new division using the official SDK
func (c *SDKClient) CreateDivision(ctx context.Context, req *CreateDivisionRequest) (*CreateDivisionResponse, error) {
	log.Printf("[DANA SDK] CreateDivision called for merchant %s, name %s", req.MerchantId, req.MainName)

	sdkReq := merchant_management.NewCreateDivisionRequestWithDefaults()
	sdkReq.SetMerchantId(req.MerchantId)
	sdkReq.SetParentRoleType("MERCHANT")
	sdkReq.SetDivisionName(req.MainName)
	sdkReq.SetExternalDivisionId(req.ExternalDivisionId)
	sdkReq.SetDivisionType("DIVISION")

	if req.DivisionDesc != "" {
		sdkReq.SetDivisionDescription(req.DivisionDesc)
	}

	if len(req.MccCodes) > 0 {
		sdkReq.SetMccCodes(req.MccCodes)
	} else {
		sdkReq.SetMccCodes([]string{"5812"}) // Default MCC
	}

	sdkResp, httpResp, err := c.merchantAPI.CreateDivision(ctx).
		CreateDivisionRequest(*sdkReq).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("create division failed: %w", err)
	}
	defer httpResp.Body.Close()

	resp := &CreateDivisionResponse{
		ResponseCode:    sdkResp.Response.Body.ResultInfo.ResultCodeId,
		ResponseMessage: sdkResp.Response.Body.ResultInfo.ResultMsg,
	}

	if sdkResp.Response.Body.DivisionId != nil {
		resp.DivisionID = *sdkResp.Response.Body.DivisionId
	}
	resp.MerchantID = req.MerchantId
	resp.MainName = req.MainName

	return resp, nil
}

// QueryDivision queries division information using the official SDK
func (c *SDKClient) QueryDivision(ctx context.Context, req *QueryDivisionRequest) (*QueryDivisionResponse, error) {
	log.Printf("[DANA SDK] QueryDivision called for merchant %s, division %s", req.MerchantId, req.DivisionId)

	sdkReq := merchant_management.NewQueryDivisionRequestWithDefaults()
	sdkReq.SetDivisionId(req.DivisionId)
	sdkReq.SetDivisionIdType(req.DivisionIdType)

	// Skip MerchantId for INNER_ID to avoid filter mismatch
	if req.MerchantId != "" {
		if req.DivisionIdType == "EXTERNAL_ID" {
			sdkReq.SetMerchantId(req.MerchantId)
			log.Printf("[DANA SDK] Division EXTERNAL_ID query: Setting MerchantId: %s", req.MerchantId)
		} else {
			log.Printf("[DANA SDK] Division INNER_ID query: Skipping MerchantId to avoid filter mismatch")
		}
	}

	sdkResp, httpResp, err := c.merchantAPI.QueryDivision(ctx).
		QueryDivisionRequest(*sdkReq).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("query division failed: %w", err)
	}
	defer httpResp.Body.Close()

	resp := &QueryDivisionResponse{
		ResponseCode:    sdkResp.Response.Body.ResultInfo.ResultCodeId,
		ResponseMessage: sdkResp.Response.Body.ResultInfo.ResultMsg,
	}

	if sdkResp.Response.Body.DivisionResourceInfo != nil {
		info := sdkResp.Response.Body.DivisionResourceInfo
		resp.DivisionDetail = &DivisionInfo{
			DivisionID: req.DivisionId,
		}
		if info.MerchantId != nil {
			resp.DivisionDetail.MerchantID = *info.MerchantId
		}
		if info.ExternalDivisionId != nil {
			resp.DivisionDetail.ExternalDivisionId = *info.ExternalDivisionId
		}
		if info.DivisionName != nil {
			resp.DivisionDetail.MainName = *info.DivisionName
		}
		if info.DivisionDescription != nil {
			resp.DivisionDetail.DivisionDesc = *info.DivisionDescription
		}
		// Status is usually active if found
		resp.DivisionDetail.Status = "ACTIVE"
	}

	return resp, nil
}

// UpdateDivision updates division information using the official SDK
func (c *SDKClient) UpdateDivision(ctx context.Context, req *UpdateDivisionRequest) (*UpdateDivisionResponse, error) {
	log.Printf("[DANA SDK] UpdateDivision called for division %s", req.DivisionId)

	sdkReq := merchant_management.NewUpdateDivisionRequestWithDefaults()
	sdkReq.SetDivisionId(req.DivisionId)
	sdkReq.SetDivisionIdType(req.DivisionIdType)

	// Skip MerchantId for INNER_ID to avoid filter mismatch
	if req.MerchantId != "" {
		if req.DivisionIdType == "EXTERNAL_ID" {
			sdkReq.SetMerchantId(req.MerchantId)
			log.Printf("[DANA SDK] Division EXTERNAL_ID update: Setting MerchantId: %s", req.MerchantId)
		} else {
			log.Printf("[DANA SDK] Division INNER_ID update: Skipping MerchantId to avoid filter mismatch")
		}
	}

	if req.MainName != nil {
		sdkReq.SetDivisionName(*req.MainName)
	}
	if req.DivisionDesc != nil {
		sdkReq.SetDivisionDescription(*req.DivisionDesc)
	}

	sdkResp, httpResp, err := c.merchantAPI.UpdateDivision(ctx).
		UpdateDivisionRequest(*sdkReq).
		Execute()

	if err != nil {
		return nil, fmt.Errorf("update division failed: %w", err)
	}
	defer httpResp.Body.Close()

	resp := &UpdateDivisionResponse{
		ResponseCode:    sdkResp.Response.Body.ResultInfo.ResultCodeId,
		ResponseMessage: sdkResp.Response.Body.ResultInfo.ResultMsg,
		DivisionID:      req.DivisionId,
	}

	return resp, nil
}

// ============================================================
// DISBURSEMENT API - SDK Implementation
// ============================================================

// TransferToDana transfers funds to a DANA balance using the official SDK
func (c *SDKClient) TransferToDana(ctx context.Context, req *TransferToDanaRequest) (*TransferToDanaResponse, error) {
	log.Printf("[DANA SDK] TransferToDana called for customer %s, amount %s %s",
		req.CustomerNumber, req.Amount, req.Currency)

	// DANA requires amount format with 2 decimal places (ISO-4217)
	// Example: 10000 -> "10000.00", 10 -> "10.00"
	amountStr := req.Amount.String()

	// Format amount to have exactly 2 decimal places
	var formattedAmount string
	if strings.Contains(amountStr, ".") {
		parts := strings.Split(amountStr, ".")
		intPart := parts[0]
		decPart := parts[1]
		if len(decPart) >= 2 {
			decPart = decPart[:2]
		} else {
			decPart = decPart + strings.Repeat("0", 2-len(decPart))
		}
		formattedAmount = intPart + "." + decPart
	} else {
		formattedAmount = amountStr + ".00"
	}

	log.Printf("[DANA SDK] Original amount: %s, Formatted amount: %s", amountStr, formattedAmount)

	amount := disbursement.NewMoney(formattedAmount, req.Currency)

	// FeeAmount is MANDATORY for TransferToDana API
	// Set to 0.00 as default
	feeAmount := disbursement.NewMoney("0.00", req.Currency)

	// AdditionalInfo is MANDATORY - FundType is required
	// "AGENT_TOPUP_FOR_USER_SETTLE" is the standard value for agent-to-user topup
	additionalInfo := disbursement.NewTransferToDanaRequestAdditionalInfo("AGENT_TOPUP_FOR_USER_SETTLE")

	// If ClientID is available, use it as ExternalDivisionId as requested
	// This helps resolve MERCHANT_ACCOUNT_NOT_EXIST in many Sandbox setups
	if c.cfg.ClientID != "" {
		additionalInfo.SetExternalDivisionId(c.cfg.ClientID)
		additionalInfo.SetChargeTarget("DIVISION")
		log.Printf("[DANA SDK] Using ClientID %s as ExternalDivisionId", c.cfg.ClientID)
	}
	// Create SDK request with required parameters
	sdkReq := disbursement.NewTransferToDanaRequest(
		req.PartnerReferenceNo,
		*amount,
		*feeAmount,
		*additionalInfo,
	)

	sdkReq.SetCustomerNumber(req.CustomerNumber)

	if req.Notes != "" {
		sdkReq.SetNotes(req.Notes)
	}

	// TransactionDate is mandatory and MUST be in GMT+7 (+07:00) format for DANA
	loc := time.FixedZone("WIB", 7*3600)
	sdkReq.SetTransactionDate(time.Now().In(loc).Format("2006-01-02T15:04:05+07:00"))

	// Call SDK
	sdkResp, httpResp, err := c.disbursementAPI.TransferToDana(ctx).
		TransferToDanaRequest(*sdkReq).
		Execute()

	if err != nil {
		if httpResp != nil {
			body, _ := io.ReadAll(httpResp.Body)
			log.Printf("[DANA SDK] TransferToDana error body: %s", string(body))
		}
		return nil, fmt.Errorf("transfer to dana failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Capture raw response body for debugging/complete data
	var rawBody interface{}
	json.NewDecoder(httpResp.Body).Decode(&rawBody)

	resp := &TransferToDanaResponse{
		ResponseCode:    sdkResp.ResponseCode,
		ResponseMessage: sdkResp.ResponseMessage,
		RawDana:         rawBody,
	}

	if sdkResp.ReferenceNo != nil {
		resp.ReferenceNo = *sdkResp.ReferenceNo
	}

	log.Printf("[DANA SDK] TransferToDana success. ResponseCode: %s", resp.ResponseCode)
	return resp, nil
}

// TransferToDanaInquiryStatus checks the status of a transfer using the official SDK
func (c *SDKClient) TransferToDanaInquiryStatus(ctx context.Context, req *TransferToDanaInquiryStatusRequest) (*TransferToDanaInquiryStatusResponse, error) {
	log.Printf("[DANA SDK] TransferToDanaInquiryStatus called for originalPartnerRef %s", req.OriginalPartnerReferenceNo)

	// Create SDK request
	// ServiceCode "38" is mandatory for TransferToDana Inquiry
	sdkReq := disbursement.NewTransferToDanaInquiryStatusRequest(req.OriginalPartnerReferenceNo, "38")
	if req.OriginalReferenceNo != "" {
		sdkReq.SetOriginalReferenceNo(req.OriginalReferenceNo)
	}

	// Call SDK
	sdkResp, httpResp, err := c.disbursementAPI.TransferToDanaInquiryStatus(ctx).
		TransferToDanaInquiryStatusRequest(*sdkReq).
		Execute()

	if err != nil {
		if httpResp != nil {
			body, _ := io.ReadAll(httpResp.Body)
			log.Printf("[DANA SDK] TransferToDanaInquiryStatus error body: %s", string(body))
		}
		return nil, fmt.Errorf("transfer to dana inquiry status failed: %w", err)
	}
	defer httpResp.Body.Close()

	var rawBody interface{}
	json.NewDecoder(httpResp.Body).Decode(&rawBody)

	resp := &TransferToDanaInquiryStatusResponse{
		ResponseCode:            sdkResp.ResponseCode,
		ResponseMessage:         sdkResp.ResponseMessage,
		LatestTransactionStatus: sdkResp.LatestTransactionStatus,
		TransactionStatusDesc:   sdkResp.TransactionStatusDesc,
		RawDana:                 rawBody,
	}

	if sdkResp.OriginalReferenceNo != nil {
		resp.OriginalReferenceNo = *sdkResp.OriginalReferenceNo
	}

	log.Printf("[DANA SDK] TransferToDanaInquiryStatus success. ResponseCode: %s, Status: %s", resp.ResponseCode, resp.LatestTransactionStatus)
	return resp, nil
}
