#!/usr/bin/env bash
# ==============================================================================
# echosh-labs / mercury-dasha
# Sovereign Port Contention Identification & Remediation Script
# ==============================================================================
# Usage:
#   ./scripts/clear-port.sh [PORT]
#   PORT=8080 ./scripts/clear-port.sh
# ==============================================================================

set -euo pipefail

TARGET_PORT="${1:-${PORT:-8080}}"
TARGET_PORT="${TARGET_PORT#:}"

BOLD='\033[1m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
RED='\033[0;31m'
NC='\033[0m'

is_port_in_use() {
  local p="$1"
  if command -v python3 >/dev/null 2>&1; then
    if python3 -c "import socket; s=socket.socket(socket.AF_INET, socket.SOCK_STREAM); s.bind(('0.0.0.0', $p)); s.close()" >/dev/null 2>&1; then
      return 1 # Available
    else
      return 0 # In use
    fi
  fi

  if command -v lsof >/dev/null 2>&1; then
    lsof -ti :"$p" >/dev/null 2>&1 && return 0
  fi
  if command -v fuser >/dev/null 2>&1; then
    fuser "$p/tcp" >/dev/null 2>&1 && return 0
  fi
  if command -v ss >/dev/null 2>&1; then
    ss -tulpn "( sport = :$p )" 2>/dev/null | grep -q ":$p" && return 0
  fi
  return 1
}

get_linux_pids() {
  local p="$1"
  local pids=""

  if command -v lsof >/dev/null 2>&1; then
    pids=$(lsof -ti :"$p" 2>/dev/null || true)
  fi

  if [ -z "$pids" ] && command -v fuser >/dev/null 2>&1; then
    pids=$(fuser "$p/tcp" 2>/dev/null | tr -s ' ' '\n' | grep -E '^[0-9]+$' || true)
  fi

  if [ -z "$pids" ] && command -v ss >/dev/null 2>&1; then
    pids=$(ss -tulpn "( sport = :$p )" 2>/dev/null | grep -o 'pid=[0-9]*' | cut -d= -f2 || true)
  fi

  local filtered=""
  local my_pid=$$
  for pid in $pids; do
    if [ "$pid" != "$my_pid" ] && [ -d "/proc/$pid" ]; then
      filtered="$filtered $pid"
    fi
  done
  echo "$filtered" | xargs
}

is_wsl() {
  [ -f /proc/sys/fs/binfmt_misc/WSLInterop ] || grep -qi microsoft /proc/version 2>/dev/null
}

if ! is_port_in_use "$TARGET_PORT"; then
  echo -e "  ${GREEN}✔ Port :${TARGET_PORT} is free and ready.${NC}"
  exit 0
fi

echo -e "  ${YELLOW}⚠️ Port :${TARGET_PORT} is currently in use. Resolving port contention...${NC}"

# 1. Check and terminate Linux processes
LINUX_PIDS=$(get_linux_pids "$TARGET_PORT")
if [ -n "$LINUX_PIDS" ]; then
  for pid in $LINUX_PIDS; do
    CMD="unknown"
    if [ -f "/proc/$pid/comm" ]; then
      CMD=$(cat "/proc/$pid/comm" 2>/dev/null || echo "unknown")
    fi
    echo -e "  • Found contentious Linux process: PID ${CYAN}${pid}${NC} (${YELLOW}${CMD}${NC}) on :${TARGET_PORT}"
    kill -15 "$pid" 2>/dev/null || true
  done

  # Allow graceful exit
  for i in $(seq 1 10); do
    sleep 0.1
    if ! is_port_in_use "$TARGET_PORT"; then
      break
    fi
  done

  # Force kill lingering Linux PIDs
  STILL_PIDS=$(get_linux_pids "$TARGET_PORT")
  if [ -n "$STILL_PIDS" ]; then
    for pid in $STILL_PIDS; do
      echo -e "  • Force killing lingering Linux PID ${RED}${pid}${NC}..."
      kill -9 "$pid" 2>/dev/null || true
    done
    sleep 0.2
  fi
fi

# 2. Check WSL2 Windows Host Contention if still occupied
if is_port_in_use "$TARGET_PORT" && is_wsl; then
  echo -e "  • Checking Windows host for port :${TARGET_PORT} contention..."
  
  # Check if a Windows process is listening
  WIN_LISTENER=$(cmd.exe /c "cd /d C:\ && netstat -ano | findstr LISTENING | findstr :${TARGET_PORT}" < /dev/null 2>/dev/null || true)
  if [ -n "$WIN_LISTENER" ]; then
    WIN_PID=$(echo "$WIN_LISTENER" | awk '{print $NF}' | tr -d '\r' | head -n 1)
    echo -e "  • Windows host listener detected: PID ${CYAN}${WIN_PID}${NC}"
    
    # Check if this PID is svchost / iphlpsvc (netsh portproxy)
    WIN_SVC=$(cmd.exe /c "tasklist /svc" < /dev/null 2>/dev/null | grep "$WIN_PID" || true)
    if echo "$WIN_SVC" | grep -qi "iphlpsvc"; then
      echo -e "  ${YELLOW}⚠️ Contentious Windows netsh portproxy detected forwarding port :${TARGET_PORT}.${NC}"
      echo -e "  • Attempting to trigger elevated removal via Windows..."
      cmd.exe /c "powershell -WindowStyle Hidden -Command \"Start-Process cmd -ArgumentList '/c netsh interface portproxy delete v4tov4 listenport=${TARGET_PORT} listenaddress=0.0.0.0' -Verb RunAs\"" < /dev/null 2>/dev/null || true
      
      # Wait briefly to see if user accepted or rule vanished
      for i in $(seq 1 15); do
        sleep 0.2
        if ! is_port_in_use "$TARGET_PORT"; then
          break
        fi
      done
    else
      echo -e "  • Attempting to terminate contentious Windows PID ${WIN_PID}..."
      cmd.exe /c "taskkill /F /PID ${WIN_PID}" < /dev/null 2>/dev/null || true
      sleep 0.3
    fi
  fi
fi

if ! is_port_in_use "$TARGET_PORT"; then
  echo -e "  ${GREEN}✔ Port :${TARGET_PORT} successfully cleared and reclaimed!${NC}"
  exit 0
fi

echo -e "  ${RED}❌ Port :${TARGET_PORT} remains in use.${NC}"
if is_wsl; then
  echo -e "  ${YELLOW}Tip for WSL2:${NC} If held by Windows netsh portproxy, run in an elevated Windows PowerShell:"
  echo -e "    ${CYAN}netsh interface portproxy delete v4tov4 listenport=${TARGET_PORT} listenaddress=0.0.0.0${NC}"
  echo -e "  Alternatively, choose an alternative port: ${CYAN}PORT=8081 make dev${NC}"
fi
exit 1
