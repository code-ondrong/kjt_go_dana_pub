package dana

import (
	"bytes"
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	danaSDK "github.com/dana-id/dana-go/v2"
	configSDK "github.com/dana-id/dana-go/v2/config"
	"github.com/dana-id/dana-go/v2/disbursement/v1"
	merchant_management "github.com/dana-id/dana-go/v2/merchant_management/v1"

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

	// Set API key — ENV and DANA_ENV both required; SDK checks apiKey.ENV for debug mode logic
	configuration.APIKey = &configSDK.APIKey{
		ENV:           string(env),
		DANA_ENV:      string(env),
		X_PARTNER_ID:  cfg.PartnerID,
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
	log.Printf("[DANA SDK] ===== CONFIG DEBUG =====")
	log.Printf("[DANA SDK] X_PARTNER_ID (PartnerID): %s", cfg.PartnerID)
	log.Printf("[DANA SDK] CLIENT_ID:                %s", cfg.ClientID)
	log.Printf("[DANA SDK] MERCHANT_ID:              %s", cfg.MerchantID)
	log.Printf("[DANA SDK] SHOP_ID:                  %s", cfg.ShopID)
	log.Printf("[DANA SDK] DIVISION_ID:              %s", cfg.DivisionID)
	log.Printf("[DANA SDK] CHARGE_TARGET:            %s", cfg.ChargeTarget)
	log.Printf("[DANA SDK] CLIENT_SECRET:             %s...%s", cfg.ClientSecret[:8], cfg.ClientSecret[len(cfg.ClientSecret)-4:])
	log.Printf("[DANA SDK] PRIVATE_KEY length:        %d chars", len(cfg.PrivateKey))
	log.Printf("[DANA SDK] ORIGIN:                   %s", cfg.Origin)
	log.Printf("[DANA SDK] =========================")

	return &SDKClient{
		danaClient:      danaClient,
		cfg:             cfg,
		merchantAPI:     danaClient.MerchantManagementAPI,
		disbursementAPI: danaClient.DisbursementAPI,
	}, nil
}

// GetConfig returns the DANA configuration
func (c *SDKClient) GetConfig() *config.DANAConfig {
	return c.cfg
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
		MerchantId:     req.MerchantId,
		ShopParentType: req.ShopParentType,
		MainName:       req.ShopName,
		ExternalShopId: req.ShopAlias,
		SizeType:       mappedSizeType,
		// MCC codes is required - use default if not provided
		MccCodes: []string{"5734"}, // 5734 = Computer Software Stores (default)
	}

	// Handle parent type: when DIVISION, shopParentId goes to ParentDivisionId
	// When MERCHANT, shopParentId goes to MerchantId (already set above)
	shopParentType := strings.ToUpper(req.ShopParentType)
	if shopParentType == "DIVISION" || shopParentType == "EXTERNAL_DIVISION" {
		// When parent is a division, shopParentId is the division ID
		sdkReq.SetParentDivisionId(req.ShopParentId)
		log.Printf("[DANA SDK] CreateShop - ParentType is DIVISION, setting ParentDivisionId: %s, MerchantId: %s", req.ShopParentId, req.MerchantId)
	} else {
		// When parent is MERCHANT, shopParentId is the merchant ID
		sdkReq.MerchantId = req.ShopParentId
		log.Printf("[DANA SDK] CreateShop - ParentType is MERCHANT, setting MerchantId: %s", req.ShopParentId)
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
	log.Printf("[DANA SDK] CreateDivision called for merchant %s, name %s",
		req.MerchantId, req.DivisionName)

	sdkReq := merchant_management.NewCreateDivisionRequestWithDefaults()
	sdkReq.SetMerchantId(req.MerchantId)

	// Set ParentRoleType from request or default to HEAD_OFFICE
	parentRoleType := req.ParentRoleType
	if parentRoleType == "" {
		parentRoleType = "HEAD_OFFICE"
	}
	sdkReq.SetParentRoleType(parentRoleType)

	sdkReq.SetDivisionName(req.DivisionName)
	sdkReq.SetExternalDivisionId(req.ExternalDivisionId)

	// Set DivisionDesc if provided
	if req.DivisionDesc != "" {
		sdkReq.SetDivisionDescription(req.DivisionDesc)
	}

	sdkReq.SetDivisionType("DIVISION")

	// Set SizeType if provided (map to SDK enum)
	if req.SizeType != "" {
		sdkReq.SetSizeType(req.SizeType)
	}

	// Set required fields for CreateDivisionRequest
	// ApiVersion must be > 2 for full functionality
	sdkReq.SetApiVersion("3")

	// Set ParentDivisionId if provided (only for DIVISION or EXTERNAL_DIVISION parent types)
	if req.ParentDivisionId != "" {
		sdkReq.SetParentDivisionId(req.ParentDivisionId)
	}

	// Set DivisionAddress - required field
	divisionAddress := merchant_management.NewAddressInfo()
	if req.DivisionAddress != nil {
		if req.DivisionAddress.Country != "" {
			divisionAddress.SetCountry(req.DivisionAddress.Country)
		}
		if req.DivisionAddress.Province != "" {
			divisionAddress.SetProvince(req.DivisionAddress.Province)
		}
		if req.DivisionAddress.City != "" {
			divisionAddress.SetCity(req.DivisionAddress.City)
		}
		if req.DivisionAddress.Area != "" {
			divisionAddress.SetArea(req.DivisionAddress.Area)
		}
		if req.DivisionAddress.Address1 != "" {
			divisionAddress.SetAddress1(req.DivisionAddress.Address1)
		}
		if req.DivisionAddress.Address2 != "" {
			divisionAddress.SetAddress2(req.DivisionAddress.Address2)
		}
		if req.DivisionAddress.Postcode != "" {
			divisionAddress.SetPostcode(req.DivisionAddress.Postcode)
		}
		if req.DivisionAddress.SubDistrict != "" {
			divisionAddress.SetSubDistrict(req.DivisionAddress.SubDistrict)
		}
	}
	sdkReq.SetDivisionAddress(*divisionAddress)

	// Set DivisionType - required field
	if req.DivisionType != "" {
		sdkReq.SetDivisionType(req.DivisionType)
	} else {
		sdkReq.SetDivisionType("DIVISION") // Default
	}

	// Set ExternalDivisionId - required field
	sdkReq.SetExternalDivisionId(req.ExternalDivisionId)

	// Set DivisionName - required field
	if req.DivisionName != "" {
		sdkReq.SetDivisionName(req.DivisionName)
	} else {
		return nil, fmt.Errorf("division name is required for creation")
	}

	// Set SizeType if provided (map to SDK enum)
	if req.SizeType != "" {
		sdkReq.SetSizeType(req.SizeType)
	} else {
		sdkReq.SetSizeType("UKE") // Default to Small Business
	}

	// Set MCC codes - required field
	if len(req.MccCodes) > 0 {
		sdkReq.SetMccCodes(req.MccCodes)
	} else {
		sdkReq.SetMccCodes([]string{"5812"}) // Default MCC
	}

	// Set ExtInfo - required field
	extInfo := merchant_management.NewCreateDivisionRequestExtInfo()
	if req.ExtInfo != nil {
		if req.ExtInfo.PIC_EMAIL != "" {
			extInfo.SetPIC_EMAIL(req.ExtInfo.PIC_EMAIL)
		}
		if req.ExtInfo.PIC_PHONENUMBER != "" {
			extInfo.SetPIC_PHONENUMBER(req.ExtInfo.PIC_PHONENUMBER)
		}
		if req.ExtInfo.SUBMITTER_EMAIL != "" {
			extInfo.SetSUBMITTER_EMAIL(req.ExtInfo.SUBMITTER_EMAIL)
		}
		if req.ExtInfo.GOODS_SOLD_TYPE != "" {
			extInfo.SetGOODS_SOLD_TYPE(req.ExtInfo.GOODS_SOLD_TYPE)
		}
		if req.ExtInfo.USECASE != "" {
			extInfo.SetUSECASE(req.ExtInfo.USECASE)
		}
		if req.ExtInfo.USER_PROFILING != "" {
			extInfo.SetUSER_PROFILING(req.ExtInfo.USER_PROFILING)
		}
		if req.ExtInfo.AVG_TICKET != "" {
			extInfo.SetAVG_TICKET(req.ExtInfo.AVG_TICKET)
		}
		if req.ExtInfo.OMZET != "" {
			extInfo.SetOMZET(req.ExtInfo.OMZET)
		}
		if req.ExtInfo.EXT_URLS != "" {
			extInfo.SetEXT_URLS(req.ExtInfo.EXT_URLS)
		}
		if req.ExtInfo.BRAND_NAME != "" {
			extInfo.SetBRAND_NAME(req.ExtInfo.BRAND_NAME)
		}
	}
	sdkReq.SetExtInfo(*extInfo)

	// Set BusinessEntity - required when apiVersion > 2
	if req.BusinessEntity != "" {
		sdkReq.SetBusinessEntity(req.BusinessEntity)
	} else {
		sdkReq.SetBusinessEntity("INDIVIDU") // Default
	}

	// Set BusinessDocs - required when apiVersion > 2
	if len(req.BusinessDocs) > 0 {
		var businessDocs []merchant_management.BusinessDocs
		for _, doc := range req.BusinessDocs {
			businessDoc := merchant_management.NewBusinessDocs()
			if doc.DocType != "" {
				businessDoc.SetDocType(doc.DocType)
			}
			if doc.DocId != "" {
				businessDoc.SetDocId(doc.DocId)
			}
			if doc.DocFile != "" {
				businessDoc.SetDocFile(doc.DocFile)
			}
			businessDocs = append(businessDocs, *businessDoc)
		}
		sdkReq.SetBusinessDocs(businessDocs)
	} else {
		// Provide empty slice to satisfy API requirement
		sdkReq.SetBusinessDocs([]merchant_management.BusinessDocs{})
	}

	// Set Owner information - required when apiVersion > 2
	ownerName := merchant_management.NewUserName()
	if req.OwnerName != nil {
		if req.OwnerName.FirstName != "" {
			ownerName.SetFirstName(req.OwnerName.FirstName)
		}
		if req.OwnerName.LastName != "" {
			ownerName.SetLastName(req.OwnerName.LastName)
		}
	}
	sdkReq.SetOwnerName(*ownerName)

	ownerPhone := merchant_management.NewMobileNoInfo()
	if req.OwnerPhoneNumber != nil {
		if req.OwnerPhoneNumber.MobileNo != "" {
			ownerPhone.SetMobileNo(req.OwnerPhoneNumber.MobileNo)
		}
		if req.OwnerPhoneNumber.Verified != "" {
			ownerPhone.SetVerified(req.OwnerPhoneNumber.Verified)
		}
	}
	sdkReq.SetOwnerPhoneNumber(*ownerPhone)

	if req.OwnerIdType != "" {
		sdkReq.SetOwnerIdType(req.OwnerIdType)
	} else {
		sdkReq.SetOwnerIdType("KTP") // Default
	}

	if req.OwnerIdNo != "" {
		sdkReq.SetOwnerIdNo(req.OwnerIdNo)
	}

	ownerAddress := merchant_management.NewAddressInfo()
	if req.OwnerAddress != nil {
		if req.OwnerAddress.Country != "" {
			ownerAddress.SetCountry(req.OwnerAddress.Country)
		}
		if req.OwnerAddress.Province != "" {
			ownerAddress.SetProvince(req.OwnerAddress.Province)
		}
		if req.OwnerAddress.City != "" {
			ownerAddress.SetCity(req.OwnerAddress.City)
		}
		if req.OwnerAddress.Area != "" {
			ownerAddress.SetArea(req.OwnerAddress.Area)
		}
		if req.OwnerAddress.Address1 != "" {
			ownerAddress.SetAddress1(req.OwnerAddress.Address1)
		}
		if req.OwnerAddress.Address2 != "" {
			ownerAddress.SetAddress2(req.OwnerAddress.Address2)
		}
		if req.OwnerAddress.Postcode != "" {
			ownerAddress.SetPostcode(req.OwnerAddress.Postcode)
		}
		if req.OwnerAddress.SubDistrict != "" {
			ownerAddress.SetSubDistrict(req.OwnerAddress.SubDistrict)
		}
	}
	sdkReq.SetOwnerAddress(*ownerAddress)

	// Set Director PICs - required when apiVersion > 2
	if len(req.DirectorPics) > 0 {
		var directorPics []merchant_management.PicInfo
		for _, pic := range req.DirectorPics {
			picInfo := merchant_management.NewPicInfo()
			if pic.PicName != "" {
				picInfo.SetPicName(pic.PicName)
			}
			if pic.PicPosition != "" {
				picInfo.SetPicPosition(pic.PicPosition)
			}
			directorPics = append(directorPics, *picInfo)
		}
		sdkReq.SetDirectorPics(directorPics)
	} else {
		// Provide empty slice to satisfy API requirement
		sdkReq.SetDirectorPics([]merchant_management.PicInfo{})
	}

	// Set NonDirector PICs - required when apiVersion > 2
	if len(req.NonDirectorPics) > 0 {
		var nonDirectorPics []merchant_management.PicInfo
		for _, pic := range req.NonDirectorPics {
			picInfo := merchant_management.NewPicInfo()
			if pic.PicName != "" {
				picInfo.SetPicName(pic.PicName)
			}
			if pic.PicPosition != "" {
				picInfo.SetPicPosition(pic.PicPosition)
			}
			nonDirectorPics = append(nonDirectorPics, *picInfo)
		}
		sdkReq.SetNonDirectorPics(nonDirectorPics)
	} else {
		// Provide empty slice to satisfy API requirement
		sdkReq.SetNonDirectorPics([]merchant_management.PicInfo{})
	}

	// Set PgDivisionFlag
	if req.PgDivisionFlag != "" {
		sdkReq.SetPgDivisionFlag(req.PgDivisionFlag)
	}

	// Log request JSON for debugging
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] CreateDivision request JSON:\n%s", string(reqJSON))

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
	resp.MainName = req.DivisionName

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
		if info.DivisionType != nil {
			resp.DivisionDetail.DivisionType = *info.DivisionType
		}
		if info.ParentRoleType != nil {
			resp.DivisionDetail.ParentRoleType = *info.ParentRoleType
		}
		if info.PgDivisionFlag != nil {
			resp.DivisionDetail.PgDivisionFlag = *info.PgDivisionFlag
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

	// Set required fields for UpdateDivisionRequest
	// NewExternalDivisionId is required
	if req.NewExternalDivisionId != "" {
		sdkReq.SetNewExternalDivisionId(req.NewExternalDivisionId)
	} else {
		sdkReq.SetNewExternalDivisionId(req.DivisionId)
	}

	// DivisionName is required
	if req.MainName != nil && *req.MainName != "" {
		sdkReq.SetDivisionName(*req.MainName)
	} else {
		return nil, fmt.Errorf("division name is required for update")
	}

	// DivisionAddress is required
	divisionAddress := merchant_management.NewAddressInfo()
	if req.DivisionAddress != nil {
		if req.DivisionAddress.Country != "" {
			divisionAddress.SetCountry(req.DivisionAddress.Country)
		}
		if req.DivisionAddress.Province != "" {
			divisionAddress.SetProvince(req.DivisionAddress.Province)
		}
		if req.DivisionAddress.City != "" {
			divisionAddress.SetCity(req.DivisionAddress.City)
		}
		if req.DivisionAddress.Area != "" {
			divisionAddress.SetArea(req.DivisionAddress.Area)
		}
		if req.DivisionAddress.Address1 != "" {
			divisionAddress.SetAddress1(req.DivisionAddress.Address1)
		}
		if req.DivisionAddress.Address2 != "" {
			divisionAddress.SetAddress2(req.DivisionAddress.Address2)
		}
		if req.DivisionAddress.Postcode != "" {
			divisionAddress.SetPostcode(req.DivisionAddress.Postcode)
		}
		if req.DivisionAddress.SubDistrict != "" {
			divisionAddress.SetSubDistrict(req.DivisionAddress.SubDistrict)
		}
	}
	sdkReq.SetDivisionAddress(*divisionAddress)

	// DivisionType is required
	if req.DivisionType != "" {
		sdkReq.SetDivisionType(req.DivisionType)
	} else {
		sdkReq.SetDivisionType("DIVISION") // Default
	}

	// MccCodes is required
	if len(req.MccCodes) > 0 {
		sdkReq.SetMccCodes(req.MccCodes)
	} else {
		sdkReq.SetMccCodes([]string{"5812"}) // Default MCC
	}

	// ExtInfo is required
	extInfo := make(map[string]interface{})
	if req.ExtInfo != nil && len(req.ExtInfo) > 0 {
		extInfo = req.ExtInfo
	}
	sdkReq.SetExtInfo(extInfo)

	// Optional fields
	if req.DivisionDesc != nil {
		sdkReq.SetDivisionDescription(*req.DivisionDesc)
	}

	// ApiVersion is optional but recommended
	if req.ApiVersion != nil && *req.ApiVersion != "" {
		sdkReq.SetApiVersion(*req.ApiVersion)
	} else {
		sdkReq.SetApiVersion("3")
	}

	// LogoUrlMap is optional
	if req.LogoUrlMap != nil && len(req.LogoUrlMap) > 0 {
		sdkReq.SetLogoUrlMap(req.LogoUrlMap)
	}

	// BusinessEntity is optional but recommended for apiVersion > 2
	if req.BusinessEntity != nil && *req.BusinessEntity != "" {
		sdkReq.SetBusinessEntity(*req.BusinessEntity)
	}

	// BusinessDocs is optional for apiVersion > 2
	if len(req.BusinessDocs) > 0 {
		var businessDocs []merchant_management.BusinessDocs
		for _, doc := range req.BusinessDocs {
			businessDoc := merchant_management.NewBusinessDocs()
			if doc.DocType != "" {
				businessDoc.SetDocType(doc.DocType)
			}
			if doc.DocId != "" {
				businessDoc.SetDocId(doc.DocId)
			}
			if doc.DocFile != "" {
				businessDoc.SetDocFile(doc.DocFile)
			}
			businessDocs = append(businessDocs, *businessDoc)
		}
		sdkReq.SetBusinessDocs(businessDocs)
	}

	// BusinessEndDate is optional for apiVersion > 2
	if req.BusinessEndDate != nil && *req.BusinessEndDate != "" {
		sdkReq.SetBusinessEndDate(*req.BusinessEndDate)
	}

	// Owner information is optional for apiVersion > 2
	if req.OwnerName != nil {
		ownerName := merchant_management.NewUserName()
		if req.OwnerName.FirstName != "" {
			ownerName.SetFirstName(req.OwnerName.FirstName)
		}
		if req.OwnerName.LastName != "" {
			ownerName.SetLastName(req.OwnerName.LastName)
		}
		sdkReq.SetOwnerName(*ownerName)
	}

	if req.OwnerPhoneNumber != nil {
		ownerPhone := merchant_management.NewMobileNoInfo()
		if req.OwnerPhoneNumber.MobileNo != "" {
			ownerPhone.SetMobileNo(req.OwnerPhoneNumber.MobileNo)
		}
		if req.OwnerPhoneNumber.Verified != "" {
			ownerPhone.SetVerified(req.OwnerPhoneNumber.Verified)
		}
		sdkReq.SetOwnerPhoneNumber(*ownerPhone)
	}

	if req.OwnerIdType != nil && *req.OwnerIdType != "" {
		sdkReq.SetOwnerIdType(*req.OwnerIdType)
	}

	if req.OwnerIdNo != nil && *req.OwnerIdNo != "" {
		sdkReq.SetOwnerIdNo(*req.OwnerIdNo)
	}

	if req.OwnerAddress != nil {
		ownerAddress := merchant_management.NewAddressInfo()
		if req.OwnerAddress.Country != "" {
			ownerAddress.SetCountry(req.OwnerAddress.Country)
		}
		if req.OwnerAddress.Province != "" {
			ownerAddress.SetProvince(req.OwnerAddress.Province)
		}
		if req.OwnerAddress.City != "" {
			ownerAddress.SetCity(req.OwnerAddress.City)
		}
		if req.OwnerAddress.Area != "" {
			ownerAddress.SetArea(req.OwnerAddress.Area)
		}
		if req.OwnerAddress.Address1 != "" {
			ownerAddress.SetAddress1(req.OwnerAddress.Address1)
		}
		if req.OwnerAddress.Address2 != "" {
			ownerAddress.SetAddress2(req.OwnerAddress.Address2)
		}
		if req.OwnerAddress.Postcode != "" {
			ownerAddress.SetPostcode(req.OwnerAddress.Postcode)
		}
		if req.OwnerAddress.SubDistrict != "" {
			ownerAddress.SetSubDistrict(req.OwnerAddress.SubDistrict)
		}
		sdkReq.SetOwnerAddress(*ownerAddress)
	}

	// DirectorPics is optional for apiVersion > 2
	if len(req.DirectorPics) > 0 {
		var directorPics []merchant_management.PicInfo
		for _, pic := range req.DirectorPics {
			picInfo := merchant_management.NewPicInfo()
			if pic.PicName != "" {
				picInfo.SetPicName(pic.PicName)
			}
			if pic.PicPosition != "" {
				picInfo.SetPicPosition(pic.PicPosition)
			}
			directorPics = append(directorPics, *picInfo)
		}
		sdkReq.SetDirectorPics(directorPics)
	}

	// NonDirectorPics is optional for apiVersion > 2
	if len(req.NonDirectorPics) > 0 {
		var nonDirectorPics []merchant_management.PicInfo
		for _, pic := range req.NonDirectorPics {
			picInfo := merchant_management.NewPicInfo()
			if pic.PicName != "" {
				picInfo.SetPicName(pic.PicName)
			}
			if pic.PicPosition != "" {
				picInfo.SetPicPosition(pic.PicPosition)
			}
			nonDirectorPics = append(nonDirectorPics, *picInfo)
		}
		sdkReq.SetNonDirectorPics(nonDirectorPics)
	}

	// SizeType is optional for apiVersion > 2
	if req.SizeType != nil && *req.SizeType != "" {
		sdkReq.SetSizeType(*req.SizeType)
	}

	// PgDivisionFlag is optional
	if req.PgDivisionFlag != nil && *req.PgDivisionFlag != "" {
		sdkReq.SetPgDivisionFlag(*req.PgDivisionFlag)
	}

	// Log request JSON for debugging
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] UpdateDivision request JSON:\n%s", string(reqJSON))

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

// AccountInquiry queries DANA account information using the official SDK
func (c *SDKClient) AccountInquiry(ctx context.Context, req *AccountInquiryRequest) (*AccountInquiryResponse, error) {
	log.Printf("[DANA SDK] AccountInquiry called for customer %s", req.CustomerNumber)

	// Format amount correctly if provided (ISO-4217: 2 decimal places)
	var amount *disbursement.Money
	if string(req.Amount) != "" {
		formattedAmount := formatAmount(req.Amount.String())
		curr := req.Currency
		if curr == "" {
			curr = "IDR"
		}
		amount = disbursement.NewMoney(formattedAmount, curr)
	} else {
		amount = disbursement.NewMoney("0.00", "IDR")
	}

	additionalInfo := disbursement.NewDanaAccountInquiryRequestAdditionalInfo("AGENT_TOPUP_FOR_USER_SETTLE")
	// Per DANA UAT script: default is MERCHANT mode (chargeTarget=null, externalDivisionId=null)
	// Only use DIVISION mode if CHARGE_TARGET env is explicitly set to "DIVISION"
	if c.cfg.ChargeTarget == "DIVISION" && c.cfg.DivisionID != "" {
		additionalInfo.SetExternalDivisionId(c.cfg.DivisionID)
		additionalInfo.SetChargeTarget("DIVISION")
		log.Printf("[DANA SDK] AccountInquiry using DIVISION mode with ExternalDivisionId=%s", c.cfg.DivisionID)
	} else {
		// Default: MERCHANT mode (no externalDivisionId, no chargeTarget) - matches UAT script
		log.Printf("[DANA SDK] AccountInquiry using MERCHANT mode")
	}

	// Using NewDanaAccountInquiryRequest per docs
	sdkReq := disbursement.NewDanaAccountInquiryRequest(*amount, *additionalInfo)
	sdkReq.SetCustomerNumber(req.CustomerNumber)

	if req.PartnerReferenceNo != "" {
		sdkReq.SetPartnerReferenceNo(req.PartnerReferenceNo)
	}

	loc := time.FixedZone("WIB", 7*3600)
	sdkReq.SetTransactionDate(time.Now().In(loc).Format("2006-01-02T15:04:05+07:00"))

	// ===== DEBUG: Dump full request payload =====
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] ===== ACCOUNT INQUIRY REQUEST PAYLOAD =====")
	log.Printf("[DANA SDK] %s", string(reqJSON))
	log.Printf("[DANA SDK] ============================================")

	sdkResp, httpResp, err := c.disbursementAPI.DanaAccountInquiry(ctx).
		DanaAccountInquiryRequest(*sdkReq).
		Execute()

	if err != nil {
		// DANA API may return non-200 HTTP status codes (e.g., 403 for "Inactive Account")
		// but the response body contains valid DANA error information that should be returned
		var errorBody map[string]interface{}
		if rawErr, ok := err.(interface{ Body() []byte }); ok {
			if jsonErr := json.Unmarshal(rawErr.Body(), &errorBody); jsonErr == nil {
				log.Printf("[DANA SDK] AccountInquiry DANA error response: %s", string(rawErr.Body()))
				respCode, _ := errorBody["responseCode"].(string)
				respMsg, _ := errorBody["responseMessage"].(string)
				result := &AccountInquiryResponse{
					ResponseCode:    respCode,
					ResponseMessage: respMsg,
					RawDana:         errorBody,
				}
				if customerName, ok := errorBody["customerName"].(string); ok {
					result.CustomerName = customerName
				}
				if customerNum, ok := errorBody["customerNumber"].(string); ok {
					result.CustomerNumber = customerNum
				}
				if partnerRef, ok := errorBody["partnerReferenceNo"].(string); ok {
					result.PartnerReferenceNo = partnerRef
				}
				if refNo, ok := errorBody["referenceNo"].(string); ok {
					result.ReferenceNo = refNo
				}
				if amt, ok := errorBody["amount"].(map[string]interface{}); ok {
					if val, ok := amt["value"].(string); ok {
						result.Amount = val
					}
					if cur, ok := amt["currency"].(string); ok {
						result.Currency = cur
					}
				}
				if feeAmt, ok := errorBody["feeAmount"].(map[string]interface{}); ok {
					if val, ok := feeAmt["value"].(string); ok {
						result.FeeAmount = val
					}
					if cur, ok := feeAmt["currency"].(string); ok {
						result.FeeCurrency = cur
					}
				}
				if addInfo, ok := errorBody["additionalInfo"].(map[string]interface{}); ok {
					result.AdditionalInfo = addInfo
				}
				log.Printf("[DANA SDK] AccountInquiry DANA error. ResponseCode: %s, Message: %s", respCode, respMsg)
				return result, nil
			}
		}
		// If we can't parse the error body, return the original error
		log.Printf("[DANA SDK] AccountInquiry error: %v", err)
		return nil, fmt.Errorf("account inquiry failed: %w", err)
	}
	defer httpResp.Body.Close()

	var rawBody interface{}
	// Since sdkResp is populated, we can serialize that or rely on reading from httpResp
	// DANA SDK might not close body but probably exhaust it. We will decode from httpResp just in case it's still available, otherwise it's fine
	json.NewDecoder(httpResp.Body).Decode(&rawBody)

	resp := &AccountInquiryResponse{
		ResponseCode:    sdkResp.ResponseCode,
		ResponseMessage: sdkResp.ResponseMessage,
		CustomerName:    sdkResp.CustomerName,
		RawDana:         rawBody,
	}
	if sdkResp.PartnerReferenceNo != nil {
		resp.PartnerReferenceNo = *sdkResp.PartnerReferenceNo
	}
	if sdkResp.ReferenceNo != nil {
		resp.ReferenceNo = *sdkResp.ReferenceNo
	}
	if sdkResp.CustomerNumber != nil {
		resp.CustomerNumber = *sdkResp.CustomerNumber
	}
	if sdkResp.CustomerMonthlyInLimit != nil {
		resp.CustomerMonthlyInLimit = *sdkResp.CustomerMonthlyInLimit
	}
	if sdkResp.Amount.Value != "" {
		resp.Amount = sdkResp.Amount.Value
		resp.Currency = sdkResp.Amount.Currency
	}
	if sdkResp.FeeAmount.Value != "" {
		resp.FeeAmount = sdkResp.FeeAmount.Value
		resp.FeeCurrency = sdkResp.FeeAmount.Currency
	}
	if sdkResp.MinAmount.Value != "" {
		resp.MinAmount = sdkResp.MinAmount.Value
		resp.MinCurrency = sdkResp.MinAmount.Currency
	}
	if sdkResp.MaxAmount.Value != "" {
		resp.MaxAmount = sdkResp.MaxAmount.Value
		resp.MaxCurrency = sdkResp.MaxAmount.Currency
	}
	if sdkResp.AdditionalInfo != nil {
		resp.AdditionalInfo = sdkResp.AdditionalInfo
	}

	log.Printf("[DANA SDK] AccountInquiry success. ResponseCode: %s, Customer: %s", resp.ResponseCode, resp.CustomerName)
	return resp, nil
}

// TransferToDana transfers funds to a DANA balance using the official SDK
// formatAmount formats an amount string to have exactly 2 decimal places (ISO-4217)
// Example: "10000" -> "10000.00", "10" -> "10.00", "1.5" -> "1.50"
func formatAmount(amountStr string) string {
	if strings.Contains(amountStr, ".") {
		parts := strings.Split(amountStr, ".")
		intPart := parts[0]
		decPart := parts[1]
		if len(decPart) >= 2 {
			decPart = decPart[:2]
		} else {
			decPart = decPart + strings.Repeat("0", 2-len(decPart))
		}
		return intPart + "." + decPart
	}
	return amountStr + ".00"
}

func (c *SDKClient) TransferToDana(ctx context.Context, req *TransferToDanaRequest) (*TransferToDanaResponse, error) {
	log.Printf("[DANA SDK] TransferToDana called for customer %s, amount %s %s",
		req.CustomerNumber, req.Amount, req.Currency)

	// DANA requires amount format with 2 decimal places (ISO-4217)
	formattedAmount := formatAmount(req.Amount.String())
	log.Printf("[DANA SDK] Original amount: %s, Formatted amount: %s", req.Amount.String(), formattedAmount)

	amount := disbursement.NewMoney(formattedAmount, req.Currency)

	// FeeAmount is MANDATORY for TransferToDana API
	// Per DANA UAT script: default feeAmount is "1.00" with currency "IDR"
	// If caller specifies feeAmount, use it; otherwise default to "1.00"
	feeAmountStr := "1.00"
	feeCurrency := req.Currency
	if string(req.FeeAmount) != "" {
		feeAmountStr = formatAmount(req.FeeAmount.String())
	}
	if req.FeeCurrency != "" {
		feeCurrency = req.FeeCurrency
	}
	feeAmount := disbursement.NewMoney(feeAmountStr, feeCurrency)

	// AdditionalInfo is MANDATORY - FundType is required
	// "AGENT_TOPUP_FOR_USER_SETTLE" is the standard value for agent-to-user topup
	additionalInfo := disbursement.NewTransferToDanaRequestAdditionalInfo("AGENT_TOPUP_FOR_USER_SETTLE")

	// Per DANA UAT script: default is MERCHANT mode (chargeTarget=null, externalDivisionId=null)
	// Only use DIVISION mode if CHARGE_TARGET env is explicitly set to "DIVISION"
	if c.cfg.ChargeTarget == "DIVISION" && c.cfg.DivisionID != "" {
		additionalInfo.SetExternalDivisionId(c.cfg.DivisionID)
		additionalInfo.SetChargeTarget("DIVISION")
		log.Printf("[DANA SDK] TransferToDana using DIVISION mode with ExternalDivisionId=%s", c.cfg.DivisionID)
	} else {
		// Default: MERCHANT mode (no externalDivisionId, no chargeTarget) - matches UAT script
		log.Printf("[DANA SDK] TransferToDana using MERCHANT mode")
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

	// ===== DEBUG: Dump full request payload =====
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] ===== TRANSFER TO DANA REQUEST PAYLOAD =====")
	log.Printf("[DANA SDK] %s", string(reqJSON))
	log.Printf("[DANA SDK] ============================================")

	// Call SDK
	sdkResp, httpResp, err := c.disbursementAPI.TransferToDana(ctx).
		TransferToDanaRequest(*sdkReq).
		Execute()

	if err != nil {
		// DANA API may return non-200 HTTP status codes (e.g., 403 for "Inactive Account")
		// but the response body contains valid DANA error information that should be returned
		var errorBody map[string]interface{}
		if rawErr, ok := err.(interface{ Body() []byte }); ok {
			if jsonErr := json.Unmarshal(rawErr.Body(), &errorBody); jsonErr == nil {
				log.Printf("[DANA SDK] TransferToDana DANA error response: %s", string(rawErr.Body()))
				respCode, _ := errorBody["responseCode"].(string)
				respMsg, _ := errorBody["responseMessage"].(string)
				result := &TransferToDanaResponse{
					ResponseCode:    respCode,
					ResponseMessage: respMsg,
					RawDana:         errorBody,
				}
				if partnerRef, ok := errorBody["partnerReferenceNo"].(string); ok {
					result.PartnerReferenceNo = partnerRef
				}
				if refNo, ok := errorBody["referenceNo"].(string); ok {
					result.ReferenceNo = refNo
				}
				if custNum, ok := errorBody["customerNumber"].(string); ok {
					result.CustomerNumber = custNum
				}
				if custName, ok := errorBody["customerName"].(string); ok {
					result.CustomerName = custName
				}
				if amt, ok := errorBody["amount"].(map[string]interface{}); ok {
					if val, ok := amt["value"].(string); ok {
						result.Amount = val
					}
					if cur, ok := amt["currency"].(string); ok {
						result.Currency = cur
					}
				}
				if feeAmt, ok := errorBody["feeAmount"].(map[string]interface{}); ok {
					if val, ok := feeAmt["value"].(string); ok {
						result.FeeAmount = val
					}
					if cur, ok := feeAmt["currency"].(string); ok {
						result.FeeCurrency = cur
					}
				}
				if addInfo, ok := errorBody["additionalInfo"].(map[string]interface{}); ok {
					result.AdditionalInfo = addInfo
				}
				log.Printf("[DANA SDK] TransferToDana DANA error. ResponseCode: %s, Message: %s", respCode, respMsg)
				return result, nil
			}
		}
		// If we can't parse the error body, return the original error
		log.Printf("[DANA SDK] TransferToDana error: %v", err)
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
	if sdkResp.CustomerNumber != nil {
		resp.CustomerNumber = *sdkResp.CustomerNumber
	}
	if sdkResp.CustomerName != nil {
		resp.CustomerName = *sdkResp.CustomerName
	}
	if sdkResp.Amount.Value != "" {
		resp.Amount = sdkResp.Amount.Value
		resp.Currency = sdkResp.Amount.Currency
	}
	if sdkResp.AdditionalInfo != nil {
		resp.AdditionalInfo = sdkResp.AdditionalInfo
	}

	log.Printf("[DANA SDK] TransferToDana success. ResponseCode: %s", resp.ResponseCode)
	return resp, nil
}

// TransferToDanaInquiryStatus checks the status of a transfer using the official DANA SDK
// Implements POST /v1.0/emoney/topup-status.htm for UAT
// Expected UAT response: responseCode 2003900, responseMessage "Successful", latestTransactionStatus "00"
func (c *SDKClient) TransferToDanaInquiryStatus(ctx context.Context, req *TransferToDanaInquiryStatusRequest) (*TransferToDanaInquiryStatusResponse, error) {
	log.Printf("[DANA SDK] TransferToDanaInquiryStatus called for originalPartnerRef %s", req.OriginalPartnerReferenceNo)

	// serviceCode is required by DANA SDK, default to "38" (disbursement service code)
	serviceCode := req.ServiceCode
	if serviceCode == "" {
		serviceCode = "38"
	}

	// Create SDK request using the official DANA SDK
	sdkReq := disbursement.NewTransferToDanaInquiryStatusRequest(
		req.OriginalPartnerReferenceNo,
		serviceCode,
	)

	// Set optional fields
	if req.OriginalReferenceNo != "" {
		sdkReq.SetOriginalReferenceNo(req.OriginalReferenceNo)
	}
	if req.OriginalExternalId != "" {
		sdkReq.SetOriginalExternalId(req.OriginalExternalId)
	}
	if req.AdditionalInfo != nil && len(req.AdditionalInfo) > 0 {
		sdkReq.SetAdditionalInfo(req.AdditionalInfo)
	}

	// Log request for debugging
	reqJSON, _ := json.MarshalIndent(sdkReq, "", "  ")
	log.Printf("[DANA SDK] TransferToDanaInquiryStatus request JSON:\n%s", string(reqJSON))

	// Call SDK
	sdkResp, httpResp, err := c.disbursementAPI.TransferToDanaInquiryStatus(ctx).
		TransferToDanaInquiryStatusRequest(*sdkReq).
		Execute()

	if err != nil {
		// DANA API may return non-200 HTTP status codes (e.g., 404 for "Transaction Not Found")
		// but the response body contains valid DANA error information that should be returned
		// Parse the error body to extract DANA response code and message
		var errorBody map[string]interface{}
		if rawErr, ok := err.(interface{ Body() []byte }); ok {
			if jsonErr := json.Unmarshal(rawErr.Body(), &errorBody); jsonErr == nil {
				log.Printf("[DANA SDK] TransferToDanaInquiryStatus DANA error response: %s", string(rawErr.Body()))
				respCode, _ := errorBody["responseCode"].(string)
				respMsg, _ := errorBody["responseMessage"].(string)
				result := &TransferToDanaInquiryStatusResponse{
					ResponseCode:    respCode,
					ResponseMessage: respMsg,
					RawDana:         errorBody,
				}
				if origPartnerRef, ok := errorBody["originalPartnerReferenceNo"].(string); ok {
					result.OriginalPartnerReferenceNo = origPartnerRef
				}
				if origRef, ok := errorBody["originalReferenceNo"].(string); ok {
					result.OriginalReferenceNo = origRef
				}
				if svcCode, ok := errorBody["serviceCode"].(string); ok {
					result.ServiceCode = svcCode
				}
				if status, ok := errorBody["latestTransactionStatus"].(string); ok {
					result.LatestTransactionStatus = status
				}
				if desc, ok := errorBody["transactionStatusDesc"].(string); ok {
					result.TransactionStatusDesc = desc
				}
				log.Printf("[DANA SDK] TransferToDanaInquiryStatus DANA error. ResponseCode: %s, Message: %s", respCode, respMsg)
				return result, nil
			}
		}
		// If we can't parse the error body, return the original error
		log.Printf("[DANA SDK] TransferToDanaInquiryStatus error: %v", err)
		return nil, fmt.Errorf("transfer to dana inquiry status failed: %w", err)
	}
	defer httpResp.Body.Close()

	log.Printf("[DANA SDK] TransferToDanaInquiryStatus HTTP status: %d", httpResp.StatusCode)

	// Log response body for debugging
	bodyBytes, _ := json.MarshalIndent(sdkResp, "", "  ")
	log.Printf("[DANA SDK] TransferToDanaInquiryStatus response JSON:\n%s", string(bodyBytes))

	// Convert SDK response to our response type
	resp := &TransferToDanaInquiryStatusResponse{
		ResponseCode:    sdkResp.ResponseCode,
		ResponseMessage: sdkResp.ResponseMessage,
	}

	if sdkResp.OriginalPartnerReferenceNo != "" {
		resp.OriginalPartnerReferenceNo = sdkResp.OriginalPartnerReferenceNo
	}
	if sdkResp.OriginalReferenceNo != nil {
		resp.OriginalReferenceNo = *sdkResp.OriginalReferenceNo
	}
	if sdkResp.OriginalExternalId != nil {
		resp.OriginalExternalId = *sdkResp.OriginalExternalId
	}
	if sdkResp.ServiceCode != "" {
		resp.ServiceCode = sdkResp.ServiceCode
	}
	if sdkResp.LatestTransactionStatus != "" {
		resp.LatestTransactionStatus = sdkResp.LatestTransactionStatus
	}
	if sdkResp.TransactionStatusDesc != "" {
		resp.TransactionStatusDesc = sdkResp.TransactionStatusDesc
	}
	// Amount
	if sdkResp.Amount.Value != "" {
		resp.Amount = sdkResp.Amount.Value
		resp.Currency = sdkResp.Amount.Currency
	}

	log.Printf("[DANA SDK] TransferToDanaInquiryStatus success. ResponseCode: %s, Status: %s", resp.ResponseCode, resp.LatestTransactionStatus)
	return resp, nil
}

// ============================================================
// PAYMENT GATEWAY API - SNAP Implementation
// ============================================================

// CreatePaymentOrder creates a payment order using DANA Payment Gateway API (SNAP)
func (c *SDKClient) CreatePaymentOrder(ctx context.Context, req *CreatePaymentOrderRequest) (*CreatePaymentOrderResponse, error) {
	log.Printf("[DANA SDK] CreatePaymentOrder called for partnerRef %s, merchant %s, amount %s",
		req.PartnerReferenceNo, req.MerchantID, req.Amount.Value)

	// Build request body
	body := map[string]interface{}{
		"partnerReferenceNo": req.PartnerReferenceNo,
		"merchantId":         req.MerchantID,
		"amount": map[string]string{
			"value":    req.Amount.Value,
			"currency": req.Amount.Currency,
		},
	}

	// Add optional fields
	if req.ValidUpTo != "" {
		body["validUpTo"] = req.ValidUpTo
	}
	if req.Notes != "" {
		body["notes"] = req.Notes
	}
	if req.AdditionalInfo != nil {
		body["additionalInfo"] = req.AdditionalInfo
	}
	if len(req.PayOptionDetails) > 0 {
		body["payOptionDetails"] = req.PayOptionDetails
	}
	if len(req.URLParams) > 0 {
		body["urlParams"] = req.URLParams
	}

	// Marshal to JSON (minified for consistent hashing)
	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("create payment order marshal failed: %w", err)
	}

	// Build headers
	timestamp := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02T15:04:05+07:00")
	signature, err := c.generateSignature(bodyJSON, timestamp)
	if err != nil {
		return nil, fmt.Errorf("signature generation failed: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   signature,
		"X-PARTNER-ID":  c.cfg.PartnerID,
		"X-EXTERNAL-ID": req.PartnerReferenceNo,
		"CHANNEL-ID":    c.cfg.ChannelID,
		"ORIGIN":        c.cfg.Origin,
	}

	// Determine endpoint
	endpoint := "/payment-gateway/v1.0/debit/payment-host-to-host/createOrder.htm"
	if c.cfg.Environment == "production" {
		endpoint = "/payment-gateway/v1.0/debit/payment-host-to-host/createOrder.htm"
	}

	url := c.cfg.BaseURL() + endpoint

	log.Printf("[DANA SDK] CreatePaymentOrder sending to %s", url)
	log.Printf("[DANA SDK] Headers: CHANNEL-ID=%s, ORIGIN=%s, X-PARTNER-ID=%s", c.cfg.ChannelID, c.cfg.Origin, c.cfg.PartnerID)
	log.Printf("[DANA SDK] Request body: %s", string(bodyJSON))

	// Make HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("create payment order request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	log.Printf("[DANA SDK] CreatePaymentOrder HTTP status: %d", resp.StatusCode)
	log.Printf("[DANA SDK] Response body: %s", string(respBody))

	// Handle empty response body
	if len(respBody) == 0 {
		if resp.StatusCode == 200 || resp.StatusCode == 201 {
			// Empty response with success status - maybe DANA returns empty on success?
			log.Printf("[DANA SDK] Warning: Empty response body with status %d", resp.StatusCode)
			// Return a minimal response
			return &CreatePaymentOrderResponse{
				ResponseCode:       "2005400",
				ResponseMessage:    "Success (empty response)",
				PartnerReferenceNo: req.PartnerReferenceNo,
				RawDana:            map[string]interface{}{},
			}, nil
		}
		return nil, fmt.Errorf("empty response body with status %d", resp.StatusCode)
	}

	// Parse response
	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w (body: %s)", err, string(respBody))
	}

	result := &CreatePaymentOrderResponse{
		ResponseCode:       getString(apiResp, "responseCode"),
		ResponseMessage:    getString(apiResp, "responseMessage"),
		PartnerReferenceNo: req.PartnerReferenceNo,
		RawDana:            apiResp,
	}

	// Extract checkout URL for Hosted Checkout
	if checkoutURL, ok := apiResp["checkoutUrl"].(string); ok {
		result.CheckoutURL = checkoutURL
	}
	if referenceNo, ok := apiResp["referenceNo"].(string); ok {
		result.ReferenceNo = referenceNo
	}
	if paymentStatus, ok := apiResp["paymentStatus"].(string); ok {
		result.PaymentStatus = paymentStatus
	}
	if paidTime, ok := apiResp["paidTime"].(string); ok {
		result.PaidTime = paidTime
	}

	log.Printf("[DANA SDK] CreatePaymentOrder success. ResponseCode: %s, CheckoutURL: %s", result.ResponseCode, result.CheckoutURL)
	return result, nil
}

// QueryPayment queries payment status using DANA Payment Gateway API
func (c *SDKClient) QueryPayment(ctx context.Context, req *QueryPaymentRequest) (*QueryPaymentResponse, error) {
	log.Printf("[DANA SDK] QueryPayment called for partnerRef %s", req.PartnerReferenceNo)

	// Build request body
	body := map[string]interface{}{
		"partnerReferenceNo": req.PartnerReferenceNo,
	}
	if req.MerchantID != "" {
		body["merchantId"] = req.MerchantID
	}

	bodyJSON, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("query payment marshal failed: %w", err)
	}

	// Build headers
	timestamp := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02T15:04:05+07:00")
	signature, err := c.generateSignature(bodyJSON, timestamp)
	if err != nil {
		return nil, fmt.Errorf("signature generation failed: %w", err)
	}

	headers := map[string]string{
		"Content-Type":  "application/json",
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   signature,
		"X-PARTNER-ID":  c.cfg.PartnerID,
		"X-EXTERNAL-ID": req.PartnerReferenceNo,
		"CHANNEL-ID":    c.cfg.ChannelID,
		"ORIGIN":        c.cfg.Origin,
	}

	url := c.cfg.BaseURL() + "/payment-gateway/v1.0/debit/payment-host-to-host/query.htm"

	log.Printf("[DANA SDK] QueryPayment sending to %s", url)

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	if err != nil {
		return nil, fmt.Errorf("create http request failed: %w", err)
	}

	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("query payment request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response failed: %w", err)
	}

	log.Printf("[DANA SDK] QueryPayment HTTP status: %d", resp.StatusCode)

	var apiResp map[string]interface{}
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return nil, fmt.Errorf("parse response failed: %w", err)
	}

	result := &QueryPaymentResponse{
		ResponseCode:       getString(apiResp, "responseCode"),
		ResponseMessage:    getString(apiResp, "responseMessage"),
		PartnerReferenceNo: req.PartnerReferenceNo,
		RawDana:            apiResp,
	}

	if referenceNo, ok := apiResp["referenceNo"].(string); ok {
		result.ReferenceNo = referenceNo
	}
	if paymentStatus, ok := apiResp["paymentStatus"].(string); ok {
		result.PaymentStatus = paymentStatus
	}
	if paymentAmount, ok := apiResp["paymentAmount"].(string); ok {
		result.PaymentAmount = paymentAmount
	}
	if currency, ok := apiResp["currency"].(string); ok {
		result.Currency = currency
	}
	if paidTime, ok := apiResp["paidTime"].(string); ok {
		result.PaidTime = paidTime
	}

	log.Printf("[DANA SDK] QueryPayment success. ResponseCode: %s, Status: %s", result.ResponseCode, result.PaymentStatus)
	return result, nil
}

// CancelPayment cancels a payment order
func (c *SDKClient) CancelPayment(ctx context.Context, req *CancelPaymentRequest) (*CancelPaymentResponse, error) {
	log.Printf("[DANA SDK] CancelPayment called for partnerRef %s", req.PartnerReferenceNo)

	body := map[string]interface{}{
		"partnerReferenceNo": req.PartnerReferenceNo,
	}
	if req.OriginalReferenceNo != "" {
		body["originalReferenceNo"] = req.OriginalReferenceNo
	}
	if req.MerchantID != "" {
		body["merchantId"] = req.MerchantID
	}
	if req.Reason != "" {
		body["reason"] = req.Reason
	}

	bodyJSON, _ := json.Marshal(body)
	timestamp := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02T15:04:05+07:00")
	signature, _ := c.generateSignature(bodyJSON, timestamp)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   signature,
		"X-PARTNER-ID":  c.cfg.PartnerID,
		"X-EXTERNAL-ID": req.PartnerReferenceNo,
		"CHANNEL-ID":    c.cfg.ChannelID,
		"ORIGIN":        c.cfg.Origin,
	}

	url := c.cfg.BaseURL() + "/payment-gateway/v1.0/debit/payment-host-to-host/cancel.htm"

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("cancel payment request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var apiResp map[string]interface{}
	json.Unmarshal(respBody, &apiResp)

	result := &CancelPaymentResponse{
		ResponseCode:       getString(apiResp, "responseCode"),
		ResponseMessage:    getString(apiResp, "responseMessage"),
		PartnerReferenceNo: req.PartnerReferenceNo,
		RawDana:            apiResp,
	}

	if referenceNo, ok := apiResp["referenceNo"].(string); ok {
		result.ReferenceNo = referenceNo
	}
	if cancelStatus, ok := apiResp["cancelStatus"].(string); ok {
		result.CancelStatus = cancelStatus
	}

	log.Printf("[DANA SDK] CancelPayment success. ResponseCode: %s", result.ResponseCode)
	return result, nil
}

// RefundPayment refunds a payment
func (c *SDKClient) RefundPayment(ctx context.Context, req *RefundPaymentRequest) (*RefundPaymentResponse, error) {
	log.Printf("[DANA SDK] RefundPayment called for partnerRef %s, amount %s",
		req.PartnerReferenceNo, req.RefundAmount.Value)

	body := map[string]interface{}{
		"partnerReferenceNo":  req.PartnerReferenceNo,
		"originalReferenceNo": req.OriginalReferenceNo,
		"refundAmount": map[string]string{
			"value":    req.RefundAmount.Value,
			"currency": req.RefundAmount.Currency,
		},
		"reason": req.Reason,
	}
	if req.MerchantID != "" {
		body["merchantId"] = req.MerchantID
	}

	bodyJSON, _ := json.Marshal(body)
	timestamp := time.Now().In(time.FixedZone("WIB", 7*3600)).Format("2006-01-02T15:04:05+07:00")
	signature, _ := c.generateSignature(bodyJSON, timestamp)

	headers := map[string]string{
		"Content-Type":  "application/json",
		"X-TIMESTAMP":   timestamp,
		"X-SIGNATURE":   signature,
		"X-PARTNER-ID":  c.cfg.PartnerID,
		"X-EXTERNAL-ID": req.PartnerReferenceNo,
		"CHANNEL-ID":    c.cfg.ChannelID,
		"ORIGIN":        c.cfg.Origin,
	}

	url := c.cfg.BaseURL() + "/payment-gateway/v1.0/debit/payment-host-to-host/refund.htm"

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyJSON))
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("refund payment request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	var apiResp map[string]interface{}
	json.Unmarshal(respBody, &apiResp)

	result := &RefundPaymentResponse{
		ResponseCode:       getString(apiResp, "responseCode"),
		ResponseMessage:    getString(apiResp, "responseMessage"),
		PartnerReferenceNo: req.PartnerReferenceNo,
		RawDana:            apiResp,
	}

	if refRefNo, ok := apiResp["refundReferenceNo"].(string); ok {
		result.RefundReferenceNo = refRefNo
	}
	if refundStatus, ok := apiResp["refundStatus"].(string); ok {
		result.RefundStatus = refundStatus
	}

	log.Printf("[DANA SDK] RefundPayment success. ResponseCode: %s", result.ResponseCode)
	return result, nil
}

// generateSignature generates RSA signature for SNAP requests
func (c *SDKClient) generateSignature(body []byte, timestamp string) (string, error) {
	// Build string to sign: SHA256Hash(Body+Timestamp)
	hash := sha256.Sum256([]byte(string(body) + timestamp))

	// Parse private key - support both PKCS1 and PKCS8 formats
	privateKey := c.cfg.PrivateKey

	// Try to parse the private key
	rsaKey, err := parsePrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("parse private key failed: %w", err)
	}

	// Sign
	signature, err := rsa.SignPKCS1v15(rand.Reader, rsaKey, crypto.SHA256, hash[:])
	if err != nil {
		return "", fmt.Errorf("sign failed: %w", err)
	}

	return base64.StdEncoding.EncodeToString(signature), nil
}

// parsePrivateKey parses a private key in various formats (PEM, DER base64, raw DER)
func parsePrivateKey(keyStr string) (*rsa.PrivateKey, error) {
	keyStr = strings.TrimSpace(keyStr)

	// Check if it's PEM format
	if containsPEMHeaders(keyStr) {
		block, _ := pem.Decode([]byte(keyStr))
		if block == nil {
			return nil, fmt.Errorf("failed to decode PEM block")
		}
		return parseDERBytes(block.Bytes)
	}

	// Try to decode as base64 (might be DER encoded in base64)
	derBytes, err := base64.StdEncoding.DecodeString(keyStr)
	if err == nil {
		// Successfully decoded base64, try to parse as DER
		key, err := parseDERBytes(derBytes)
		if err == nil {
			return key, nil
		}
	}

	// If base64 decode failed or DER parse failed, try to parse as raw DER
	return parseDERBytes([]byte(keyStr))
}

// parseDERBytes tries to parse DER bytes as PKCS8 or PKCS1
func parseDERBytes(derBytes []byte) (*rsa.PrivateKey, error) {
	// Try PKCS8 first
	key, err := x509.ParsePKCS8PrivateKey(derBytes)
	if err == nil {
		if k, ok := key.(*rsa.PrivateKey); ok {
			return k, nil
		}
		return nil, fmt.Errorf("not an RSA key")
	}

	// Try PKCS1
	k, err := x509.ParsePKCS1PrivateKey(derBytes)
	if err != nil {
		return nil, fmt.Errorf("parse private key failed (tried PKCS8 and PKCS1): %w", err)
	}
	return k, nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
