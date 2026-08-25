param (
    [string]$TargetUrl = "http://localhost:8080"
)

Write-Host "==================================================================" -ForegroundColor Cyan
Write-Host "🎯 Testing mercury-dasha Engine at: $TargetUrl" -ForegroundColor Cyan
Write-Host "==================================================================" -ForegroundColor Cyan

# 1. Health Telemetry
Write-Host "`n1. [GET] /healthz -> Checking Service Telemetry & BoltDB Stats" -ForegroundColor Yellow
$health = Invoke-RestMethod -Uri "$TargetUrl/healthz" -Method Get
$health | ConvertTo-Json -Depth 5 | Write-Host

# 2. Write Document
Write-Host "`n2. [POST] /api/v1/meta/agent:dasha:config -> Writing JSON Document" -ForegroundColor Yellow
$body = @{
    agent_id = "dasha-01"
    version = "1.0.0"
    framework = "echosh-labs"
    model_settings = @{
        temperature = 0.2
        max_tokens = 4096
    }
    tags = @("production", "mercury", "engine")
} | ConvertTo-Json -Depth 5

$res = Invoke-RestMethod -Uri "$TargetUrl/api/v1/meta/agent:dasha:config" -Method Post -Body $body -ContentType "application/json"
$res | ConvertTo-Json | Write-Host

# 3. List Keys
Write-Host "`n3. [GET] /api/v1/meta -> Listing All Keys" -ForegroundColor Yellow
$keys = Invoke-RestMethod -Uri "$TargetUrl/api/v1/meta" -Method Get
$keys | ConvertTo-Json | Write-Host

# 4. Fetch Document
Write-Host "`n4. [GET] /api/v1/meta/agent:dasha:config -> Fetching Document" -ForegroundColor Yellow
$doc = Invoke-RestMethod -Uri "$TargetUrl/api/v1/meta/agent:dasha:config" -Method Get
$doc | ConvertTo-Json -Depth 5 | Write-Host

Write-Host "`n==================================================================" -ForegroundColor Green
Write-Host "🎉 Verification Complete! Open in browser: $TargetUrl" -ForegroundColor Green
Write-Host "==================================================================" -ForegroundColor Green
