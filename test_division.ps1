# Test Division API Endpoints - v2 (with fixes)
$baseUrl = "http://localhost:8888/api/v1/division"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  DANA Division API - Live Test v2" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 1. Test CREATE Division (with auto-generated externalDivisionId)
Write-Host "`n[1] POST /api/v1/division/create" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor DarkGray

$timestamp = [int](Get-Date -UFormat %s)
$createBody = @{
    merchantId = "216620060009037857198"
    externalDivisionId = "KJT-DIV-$timestamp"
    mainName = "KJT Division Test v2"
    divisionType = "REGION"
    parentRoleType = "MERCHANT"
    sizeType = "UKE"
    businessEntity = "INDIVIDU"
    ownerIdType = "KTP"
    ownerIdNo = "3172010000000001"
    mccCodes = @("5812")
    divisionAddress = @{
        country = "ID"
        province = "DKI Jakarta"
        city = "Jakarta Selatan"
        address1 = "Jl. Sudirman No. 1"
        postcode = "12190"
    }
    ownerName = @{
        firstName = "Budi"
        lastName = "Santoso"
    }
    ownerPhoneNumber = @{
        mobileNo = "081234567890"
        verified = "true"
    }
    ownerAddress = @{
        country = "ID"
        province = "DKI Jakarta"
        city = "Jakarta Selatan"
        address1 = "Jl. Sudirman No. 1"
        postcode = "12190"
    }
    extInfo = @{
        BRAND_NAME = "KJT Test Brand"
        PIC_EMAIL = "test@kjt.co.id"
    }
} | ConvertTo-Json -Depth 5

Write-Host "Request:" -ForegroundColor DarkGray
Write-Host $createBody

try {
    $createResp = Invoke-RestMethod -Uri "$baseUrl/create" -Method POST -ContentType "application/json" -Body $createBody -TimeoutSec 30
    Write-Host "`nResponse (200 OK):" -ForegroundColor Green
    $createResp | ConvertTo-Json -Depth 10
} catch {
    Write-Host "`nError:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $responseBody = $reader.ReadToEnd()
        Write-Host $responseBody
    }
}

# 2. Test QUERY Division (should now include divisionType)
Write-Host "`n[2] GET /api/v1/division/query" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor DarkGray

$divisionId = "216650000003116824728"
$merchantId = "216620060009037857198"

Write-Host "Request: divisionId=$divisionId, merchantId=$merchantId"

try {
    $queryResp = Invoke-RestMethod -Uri "$baseUrl/query?merchantId=$merchantId&divisionId=$divisionId" -Method GET -TimeoutSec 30
    Write-Host "`nResponse (200 OK):" -ForegroundColor Green
    $queryResp | ConvertTo-Json -Depth 10
    
    # Save divisionType for update test
    $detectedType = $queryResp.data.divisionType
    if ($detectedType) {
        Write-Host "`n  >>> Auto-detected divisionType: $detectedType" -ForegroundColor Magenta
    }
} catch {
    Write-Host "`nError:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $responseBody = $reader.ReadToEnd()
        Write-Host $responseBody
    }
}

# 3. Test UPDATE Division (without divisionType - should auto-detect)
Write-Host "`n[3] POST /api/v1/division/update (auto-detect divisionType)" -ForegroundColor Yellow
Write-Host "----------------------------------------" -ForegroundColor DarkGray

$updateBody = @{
    divisionId = $divisionId
    divisionIdType = "INNER_ID"
    merchantId = $merchantId
    mainName = "KJT Division Updated v2"
    divisionDesc = "Updated via live test v2"
} | ConvertTo-Json

Write-Host "Request (NO divisionType - should auto-detect):" -ForegroundColor DarkGray
Write-Host $updateBody

try {
    $updateResp = Invoke-RestMethod -Uri "$baseUrl/update" -Method POST -ContentType "application/json" -Body $updateBody -TimeoutSec 30
    Write-Host "`nResponse (200 OK):" -ForegroundColor Green
    $updateResp | ConvertTo-Json -Depth 10
} catch {
    Write-Host "`nError:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $reader.BaseStream.Position = 0
        $responseBody = $reader.ReadToEnd()
        Write-Host $responseBody
    }
}

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "  Test Complete" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan