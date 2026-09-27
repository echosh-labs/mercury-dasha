#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Sovereign Windows PortProxy Auto-Remediation Script
# ==============================================================================
# Automatically aligns Windows host netsh portproxy with WSL2 networking mode:
# - In Mirrored Mode: Removes contentious portproxy rules on :PORT to avoid collision.
# - In NAT Mode: Adds/updates portproxy forwarding to the current WSL2 eth0 IP.
# ==============================================================================

set -euo pipefail

TARGET_PORT="${1:-${PORT:-8080}}"
TARGET_PORT="${TARGET_PORT#:}"

BOLD='\033[1m'
GREEN='\033[0;32m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "\n${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}"
echo -e "${BOLD}${CYAN}        WINDOWS NETSH PORTPROXY SOVEREIGN AUTO-REMEDY             ${NC}"
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}\n"

WSLCONFIG="/mnt/c/Users/justi/.wslconfig"
IS_MIRRORED=false
if [ -f "$WSLCONFIG" ] && grep -qi "networkingMode=mirrored" "$WSLCONFIG"; then
  IS_MIRRORED=true
fi

WSL_IP=$(ip -4 addr show eth0 2>/dev/null | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n 1 || hostname -I | awk '{print $1}')

if [ "$IS_MIRRORED" = true ]; then
  echo -e "  • Mode: ${GREEN}Mirrored Networking Active${NC}"
  echo -e "  • Action: Removing conflicting Windows netsh portproxy rule on port :${TARGET_PORT}..."
  
  WIN_CMD="netsh interface portproxy delete v4tov4 listenport=${TARGET_PORT} listenaddress=0.0.0.0"
  cmd.exe /c "powershell -WindowStyle Hidden -Command \"Start-Process cmd -ArgumentList '/c ${WIN_CMD}' -Verb RunAs\"" < /dev/null 2>/dev/null || true
  
  echo -e "  • Elevation command dispatched to Windows host (UAC prompt requested)."
  echo -e "  • If a Windows UAC prompt appeared on your screen, please click ${CYAN}Yes${NC}."
else
  echo -e "  • Mode: ${YELLOW}NAT Networking Active${NC}"
  echo -e "  • WSL2 IP: ${CYAN}${WSL_IP}${NC}"
  echo -e "  • Action: Forwarding Windows 0.0.0.0:${TARGET_PORT} -> ${WSL_IP}:${TARGET_PORT}..."
  
  WIN_CMD="netsh interface portproxy add v4tov4 listenport=${TARGET_PORT} listenaddress=0.0.0.0 connectport=${TARGET_PORT} connectaddress=${WSL_IP}"
  cmd.exe /c "powershell -WindowStyle Hidden -Command \"Start-Process cmd -ArgumentList '/c ${WIN_CMD}' -Verb RunAs\"" < /dev/null 2>/dev/null || true
  
  echo -e "  • Forwarding command dispatched to Windows host."
fi

# Wait and verify
sleep 1.5
echo -e "\n${BOLD}Current Windows PortProxy Status:${NC}"
cmd.exe /c "netsh interface portproxy show all" < /dev/null 2>/dev/null | tr -d '\r' || true

echo -e "\n${GREEN}✔ Auto-remedy step complete! Run './scripts/verify-port-forwarding.sh' to test.${NC}\n"
