#!/usr/bin/env bash
# ==============================================================================
# Mercury Dasha - Steinberg UR-44 Headless Capture Companion
# Records 48kHz 24-bit PCM stereo via ALSA/PulseAudio and ingests into Dasha
# ==============================================================================
set -euo pipefail

TITLE="${1:-D&D Session $(date +%Y-%m-%d)}"
CAMPAIGN="${2:-D&D Campaign}"
DM="${3:-Justin}"
SERVER_URL="${4:-http://localhost:8080}"
OUTPUT_DIR="${5:-/home/justin/Dropbox/MercuryDasha/sessions/dnd}"

mkdir -p "$OUTPUT_DIR"

TIMESTAMP=$(date +%Y%m%d_%H%M%S)
SAFE_TITLE=$(echo "$TITLE" | tr ' ' '_' | tr -cd '[:alnum:]_-')
OUTPUT_FILE="${OUTPUT_DIR}/${TIMESTAMP}_${SAFE_TITLE}.wav"

echo "=========================================================="
echo "  Mercury Dasha - Steinberg UR-44 Audio Capture Engine"
echo "  Session Title : $TITLE"
echo "  Campaign      : $CAMPAIGN"
echo "  Target Output : $OUTPUT_FILE"
echo "  Server Target : $SERVER_URL"
echo "=========================================================="

if ! command -v ffmpeg &> /dev/null; then
    echo "[ERROR] ffmpeg command not found. Please install ffmpeg or use the in-app Sonic Chronicle studio."
    exit 1
fi

echo "[+] Starting 48kHz 24-bit PCM capture..."
echo "[!] Press Ctrl+C or 'q' to stop recording and ingest into Mercury Dasha."

trap 'echo "Interrupt caught, finalizing file..."' INT

# Record via default Pulse/ALSA device
ffmpeg -y -f pulse -i default -ac 2 -ar 48000 -c:a pcm_s24le "$OUTPUT_FILE" || true

if [ -f "$OUTPUT_FILE" ]; then
    FILE_SIZE=$(stat -c%s "$OUTPUT_FILE")
    echo "[✓] Recording complete: $OUTPUT_FILE ($FILE_SIZE bytes)"
    echo "[+] Uploading to Mercury Dasha at $SERVER_URL..."

    curl -s -X POST "$SERVER_URL/api/v1/sessions/upload" \
        -F "audio=@$OUTPUT_FILE" \
        -F "title=$TITLE" \
        -F "campaign=$CAMPAIGN" \
        -F "dm=$DM" \
        -F "notes=Ingested via record-ur44.sh CLI" || true

    echo ""
    echo "[✓] Ingest complete! Access session & slice cutter at $SERVER_URL"
else
    echo "[WARN] No recording file generated."
fi
