#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Sovereign Cross-Substrate Port Forwarding & Connectivity Verification
# ==============================================================================
# Validates end-to-end connectivity between Windows Host and WSL2 Engine,
# detecting Mirrored vs NAT networking modes, Windows netsh portproxy status,
# and executing bidirectional HTTP smoke probes.
#
# Usage:
#   ./scripts/verify-port-forwarding.sh [PORT]
#   PORT=8080 ./scripts/verify-port-forwarding.sh
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
echo -e "${BOLD}${CYAN}     CROSS-SUBSTRATE PORT FORWARDING & CONNECTIVITY VERIFIER     ${NC}"
echo -e "${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}\n"

# 1. Substrate & Networking Mode Detection
echo -e "${BOLD}[Stage 1/4] Detecting WSL2 & Windows Networking Architecture...${NC}"

IS_WSL=false
if [ -f /proc/sys/fs/binfmt_misc/WSLInterop ] || grep -qi microsoft /proc/version 2>/dev/null; then
  IS_WSL=true
fi

if [ "$IS_WSL" = false ]; then
  echo -e "  • Environment:       ${YELLOW}Native Linux / Non-WSL substrate${NC}"
  echo -e "  • Port Forwarding:   Not applicable (Direct POSIX socket binding)"
  exit 0
fi

WSL_IP=$(ip -4 addr show eth0 2>/dev/null | grep -oP '(?<=inet\s)\d+(\.\d+){3}' | head -n 1 || hostname -I | awk '{print $1}')
echo -e "  • Substrate:         ${GREEN}WSL2 Ubuntu (Canonical Host)${NC}"
echo -e "  • WSL2 IP:           ${CYAN}${WSL_IP}${NC}"

WSLCONFIG="/mnt/c/Users/justi/.wslconfig"
NETWORKING_MODE="NAT (Default)"
IS_MIRRORED=false
if [ -f "$WSLCONFIG" ] && grep -qi "networkingMode=mirrored" "$WSLCONFIG"; then
  NETWORKING_MODE="Mirrored (Shared Windows-WSL Network Stack)"
  IS_MIRRORED=true
fi
echo -e "  • Networking Mode:   ${GREEN}${NETWORKING_MODE}${NC}"

# 2. Inspect Windows Host Portproxy & Firewall Configuration
echo -e "\n${BOLD}[Stage 2/4] Inspecting Windows Host Port Configuration for :${TARGET_PORT}...${NC}"

PORTPROXY_RULES=$(cmd.exe /c "netsh interface portproxy show all" < /dev/null 2>/dev/null | tr -d '\r' || true)
HAS_PORTPROXY_RULE=false
PORTPROXY_TARGET=""

if echo "$PORTPROXY_RULES" | grep -E "(0\.0\.0\.0|\*)\s+${TARGET_PORT}" >/dev/null 2>&1; then
  HAS_PORTPROXY_RULE=true
  PORTPROXY_TARGET=$(echo "$PORTPROXY_RULES" | grep -E "(0\.0\.0\.0|\*)\s+${TARGET_PORT}" | awk '{print $3":"$4}')
fi

if [ "$HAS_PORTPROXY_RULE" = true ]; then
  echo -e "  • Windows PortProxy: ${YELLOW}Active rule detected${NC} -> ${CYAN}${PORTPROXY_TARGET}${NC}"
else
  echo -e "  • Windows PortProxy: ${GREEN}No portproxy rule active for :${TARGET_PORT}${NC}"
fi

# Assessment based on mode:
if [ "$IS_MIRRORED" = true ]; then
  echo -e "  • Architecture Note: In ${GREEN}Mirrored Mode${NC}, Windows and WSL share the network stack."
  if [ "$HAS_PORTPROXY_RULE" = true ]; then
    echo -e "  ${YELLOW}⚠️ WARNING: An active netsh portproxy rule on :${TARGET_PORT} will conflict with mirrored networking!${NC}"
    echo -e "     Portproxy creates a loop on Windows host (iphlpsvc) preventing WSL from binding :${TARGET_PORT}."
    echo -e "     Remedy: Delete portproxy rule using: ${CYAN}make fix-portproxy${NC} (or run scripts/remedy-portproxy.bat)"
  else
    echo -e "  ${GREEN}✔ Status: Mirrored networking enables direct localhost and LAN connectivity to Windows with zero portproxy overhead.${NC}"
  fi
else
  echo -e "  • Architecture Note: In ${YELLOW}NAT Mode${NC}, Windows requires netsh portproxy forwarding to ${WSL_IP}:${TARGET_PORT}."
  if [ "$HAS_PORTPROXY_RULE" = true ]; then
    if [[ "$PORTPROXY_TARGET" == *"$WSL_IP"* ]]; then
      echo -e "  ${GREEN}✔ Status: Portproxy rule correctly forwards to current WSL IP (${WSL_IP}).${NC}"
    else
      echo -e "  ${RED}❌ Status: Portproxy rule forwards to stale IP (${PORTPROXY_TARGET}) instead of ${WSL_IP}!${NC}"
      echo -e "     Remedy: Update portproxy using: ${CYAN}make fix-portproxy${NC}"
    fi
  else
    echo -e "  ${YELLOW}⚠️ Status: Portproxy rule missing for NAT mode. Run 'make fix-portproxy' to configure.${NC}"
  fi
fi

# 3. Active Cross-Substrate HTTP Smoke Probe
echo -e "\n${BOLD}[Stage 3/4] Testing Cross-Substrate Port :${TARGET_PORT} Accessibility...${NC}"

EPHEMERAL_PROBE=false
PROBE_PID=""

# Check if an engine instance is already listening on TARGET_PORT
ALREADY_LISTENING=false
if python3 -c "import socket; s=socket.socket(); s.bind(('0.0.0.0', $TARGET_PORT)); s.close()" >/dev/null 2>&1; then
  ALREADY_LISTENING=false
else
  ALREADY_LISTENING=true
fi

if [ "$ALREADY_LISTENING" = false ]; then
  echo -e "  • Engine not currently running. Launching ephemeral probe on :${TARGET_PORT}..."
  python3 -m http.server "$TARGET_PORT" >/dev/null 2>&1 &
  PROBE_PID=$!
  EPHEMERAL_PROBE=true
  sleep 0.3
fi

cleanup_probe() {
  if [ "$EPHEMERAL_PROBE" = true ] && [ -n "$PROBE_PID" ]; then
    kill -9 "$PROBE_PID" >/dev/null 2>&1 || true
  fi
}
trap cleanup_probe EXIT INT TERM

# Probe 1: Direct WSL Localhost
echo -n "  • [WSL2 -> WSL2]  GET http://127.0.0.1:${TARGET_PORT}/: "
WSL_STATUS=$(curl -s -I --connect-timeout 2 --max-time 3 "http://127.0.0.1:${TARGET_PORT}/" 2>/dev/null | head -n 1 || echo "FAIL")
if echo "$WSL_STATUS" | grep -qE "200|301|302|404"; then
  echo -e "${GREEN}PASS (${WSL_STATUS})${NC}"
else
  echo -e "${RED}FAIL (${WSL_STATUS})${NC}"
fi

# Probe 2: Windows Host Localhost
echo -n "  • [Win11 -> WSL2] GET http://localhost:${TARGET_PORT}/: "
WIN_CURL_OUT=$(cmd.exe /c "curl.exe -s -I http://localhost:${TARGET_PORT}/" < /dev/null 2>/dev/null | tr -d '\r' || echo "ERR")
if echo "$WIN_CURL_OUT" | grep -qE "200|301|302|404"; then
  WIN_HEADER=$(echo "$WIN_CURL_OUT" | head -n 1)
  echo -e "${GREEN}PASS (${WIN_HEADER} - Windows Browser Access Verified!)${NC}"
  WIN_SUCCESS=true
else
  echo -e "${RED}FAIL (Could not connect from Windows host)${NC}"
  WIN_SUCCESS=false
fi

# Probe 3: Windows Host via LAN IP
echo -n "  • [Win11 -> LAN]  GET http://${WSL_IP}:${TARGET_PORT}/: "
WIN_LAN_OUT=$(cmd.exe /c "curl.exe -s -I http://${WSL_IP}:${TARGET_PORT}/" < /dev/null 2>/dev/null | tr -d '\r' || echo "ERR")
if echo "$WIN_LAN_OUT" | grep -qE "200|301|302|404"; then
  LAN_HEADER=$(echo "$WIN_LAN_OUT" | head -n 1)
  echo -e "${GREEN}PASS (${LAN_HEADER} - Network Adapter Forwarding Verified!)${NC}"
else
  echo -e "${YELLOW}NOTE (LAN binding filtered or pending firewall allow)${NC}"
fi

# 4. Summary & Operational Directives
echo -e "\n${BOLD}[Stage 4/4] Verification Summary...${NC}"
if [ "$WIN_SUCCESS" = true ]; then
  echo -e "  ${BOLD}${GREEN}🎉 SUCCESS: Port :${TARGET_PORT} is fully accessible from Windows host!${NC}"
  echo -e "  • You can open your browser on Windows and navigate directly to:"
  echo -e "    ${CYAN}http://localhost:${TARGET_PORT}${NC}"
else
  echo -e "  ${BOLD}${YELLOW}⚠️ Windows host cannot connect to port :${TARGET_PORT}.${NC}"
  if [ "$HAS_PORTPROXY_RULE" = true ] && [ "$IS_MIRRORED" = true ]; then
    echo -e "  • Cause: Contentious Windows netsh portproxy rule is trapping port :${TARGET_PORT}."
    echo -e "  • Remedy: Run ${CYAN}make fix-portproxy${NC} or double-click ${CYAN}scripts/remedy-portproxy.bat${NC}"
  fi
fi

cleanup_probe
trap - EXIT INT TERM
echo -e "\n${BOLD}${CYAN}══════════════════════════════════════════════════════════════════${NC}\n"
