@echo off
:: ==============================================================================
:: echosh-labs / mercury-dasha
:: Windows Port 8080 PortProxy Fix (Run as Administrator)
:: ==============================================================================
echo [MERCURY-DASHA] Aligning Windows Port 8080 Configuration...

:: Check if running elevated
net session >nul 2>&1
if %errorLevel% neq 0 (
    echo Elevating privileges to Administrator...
    powershell -Command "Start-Process '%~0' -Verb RunAs"
    exit /b
)

echo Removing conflicting netsh portproxy rule for port 8080...
netsh interface portproxy delete v4tov4 listenport=8080 listenaddress=0.0.0.0

echo.
echo Current PortProxy Table:
netsh interface portproxy show all

echo.
echo Port 8080 cleared for mirrored WSL2 operation!
timeout /t 3 >nul
