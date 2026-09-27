# ==============================================================================
# echosh-labs / mercury-dasha
# Sovereign Local Test & Execution Script (PowerShell)
# ==============================================================================

param (
    [int]$Port = 8080,
    [string]$DbPath = ".data/mercury-dasha-dev.db",
    [switch]$RunOnly,
    [switch]$TestOnly,
    [switch]$BuildOnly
)

$ErrorActionPreference = "Stop"

Write-Host "===============================================================" -ForegroundColor Cyan
Write-Host " [MERCURY STACK] Sovereign Local Test & Launch Environment" -ForegroundColor Cyan
Write-Host "===============================================================" -ForegroundColor Cyan

# 1. Run Unit & Integration Tests
if (-not $RunOnly) {
    Write-Host "`n[1/4] Running Go Unit & Dasha Engine Tests..." -ForegroundColor Yellow
    Push-Location backend
    try {
        go test -v ./...
    } finally {
        Pop-Location
    }
    Write-Host "[OK] All Go tests passed successfully." -ForegroundColor Green

    if ($TestOnly) {
        Write-Host "`n[OK] Test run complete." -ForegroundColor Green
        exit 0
    }

    # 2. Build Frontend Static Export
    Write-Host "`n[2/4] Building Next.js Static Export Bundle..." -ForegroundColor Yellow
    Push-Location frontend
    try {
        npm run build
    } finally {
        Pop-Location
    }
    Write-Host "[OK] Next.js export generated." -ForegroundColor Green

    # 3. Synchronize bundle to backend embed.FS
    Write-Host "`n[3/4] Synchronizing Static Export to Backend Embed FS..." -ForegroundColor Yellow
    if (-not (Test-Path "backend\cmd\server\frontend_out")) {
        New-Item -ItemType Directory -Path "backend\cmd\server\frontend_out" -Force | Out-Null
    }
    Copy-Item -Path "frontend\out\*" -Destination "backend\cmd\server\frontend_out" -Recurse -Force
    
    # 4. Compile Go Single Binary
    Write-Host "`n[4/4] Compiling Single Binary Engine..." -ForegroundColor Yellow
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" -Force | Out-Null
    }
    Push-Location backend
    try {
        go build -o ../bin/mercury-dasha.exe ./cmd/server/main.go
    } finally {
        Pop-Location
    }
    Write-Host "[OK] Compiled: bin/mercury-dasha.exe" -ForegroundColor Green

    if ($BuildOnly) {
        if (-not (Test-Path ".data")) {
            New-Item -ItemType Directory -Path ".data" -Force | Out-Null
        }
        Write-Host "`n[OK] Environment and single binary ready in .\bin\mercury-dasha.exe" -ForegroundColor Green
        Write-Host "     Launch anytime with: .\scripts\test-local.ps1 -RunOnly" -ForegroundColor Cyan
        exit 0
    }
}

# Ensure .data directory exists
if (-not (Test-Path ".data")) {
    New-Item -ItemType Directory -Path ".data" -Force | Out-Null
}

# 5. Resolve Port Contention if port is occupied
$activeConn = Get-NetTCPConnection -LocalPort $Port -State Listen -ErrorAction SilentlyContinue
if ($activeConn) {
    $contentiousPid = $activeConn.OwningProcess | Select-Object -First 1
    Write-Host "`n[WARN] Port $Port is currently in use by PID $contentiousPid. Clearing port contention..." -ForegroundColor Yellow
    try {
        Stop-Process -Id $contentiousPid -Force -ErrorAction Stop
        Start-Sleep -Milliseconds 300
        Write-Host "[OK] Contentious process $contentiousPid terminated. Port $Port cleared." -ForegroundColor Green
    } catch {
        Write-Host "[WARN] Could not terminate PID $contentiousPid. Please run as Administrator or free port manually." -ForegroundColor Red
    }
}

# Launch the server
Write-Host "`n===============================================================" -ForegroundColor Green
Write-Host " Serving Sovereign Mercury Stack on http://localhost:$Port" -ForegroundColor Green
Write-Host "    * Web Dashboard & Dasha Studio: http://localhost:$Port" -ForegroundColor Cyan
Write-Host "    * Health & Telemetry Endpoint:  http://localhost:$Port/healthz" -ForegroundColor Cyan
Write-Host "    * Chrono-Pulse Stream (SSE):    http://localhost:$Port/api/v1/stream/pulse" -ForegroundColor Cyan
Write-Host "    * Planetary Hora Endpoint:      http://localhost:$Port/api/dasha/hora" -ForegroundColor Cyan
Write-Host "    * Database File (BoltDB):       $DbPath" -ForegroundColor Cyan
Write-Host "===============================================================`n" -ForegroundColor Green

$env:PORT = "$Port"
$env:BOLT_DB_PATH = "$DbPath"
$env:ENV = "development"

& ".\bin\mercury-dasha.exe"
