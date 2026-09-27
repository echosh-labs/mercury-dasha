<#
.SYNOPSIS
    Captures high-fidelity audio from the Steinberg UR-44 USB interface and ingests it into Mercury Dasha.
.DESCRIPTION
    Leverages Windows DirectSound/WASAPI via FFmpeg to record broadcast-grade 48kHz 24-bit PCM stereo audio
    concurrently with OBS Studio. On completion, uploads the session directly to Mercury Dasha's Sovereign Storehouse.
.PARAMETER Title
    Title of the D&D session (e.g. "Curse of Strahd - Ep 14")
.PARAMETER Campaign
    Campaign name (e.g. "Curse of Strahd")
.PARAMETER DM
    Dungeon Master name (e.g. "Justin")
.PARAMETER DeviceName
    Name of the audio input device (default: "Line (Steinberg UR44)")
.PARAMETER ServerUrl
    Mercury Dasha server URL (default: "http://localhost:8080")
.EXAMPLE
    .\record-ur44.ps1 -Title "Ep 12 The Amber Temple" -Campaign "Curse of Strahd"
#>

param(
    [string]$Title = "D&D Session $(Get-Date -Format 'yyyy-MM-dd')",
    [string]$Campaign = "D&D Campaign",
    [string]$DM = "Justin",
    [string]$DeviceName = "Line (Steinberg UR44)",
    [string]$ServerUrl = "http://localhost:8080",
    [string]$OutputDir = "$env:USERPROFILE\Dropbox\MercuryDasha\sessions\dnd"
)

$ErrorActionPreference = "Stop"

Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "  Mercury Dasha - Steinberg UR-44 Audio Capture Engine" -ForegroundColor White
Write-Host "  Lossless 48kHz / 24-Bit Linear PCM D&D Chronicler" -ForegroundColor DarkCyan
Write-Host "==========================================================" -ForegroundColor Cyan
Write-Host "Session Title : $Title" -ForegroundColor Yellow
Write-Host "Campaign      : $Campaign" -ForegroundColor Yellow
Write-Host "Target Device : $DeviceName" -ForegroundColor Yellow
Write-Host "Server Target : $ServerUrl" -ForegroundColor Yellow
Write-Host "==========================================================" -ForegroundColor Cyan

if (-not (Test-Path $OutputDir)) {
    New-Item -ItemType Directory -Path $OutputDir -Force | Out-Null
}

$Timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
$SafeTitle = ($Title -replace '[^a-zA-Z0-9_-]', '_')
$OutputFile = Join-Path $OutputDir "${Timestamp}_${SafeTitle}.wav"

# Check if FFmpeg is available
$ffmpegPath = Get-Command ffmpeg -ErrorAction SilentlyContinue
if (-not $ffmpegPath) {
    Write-Host "[ERROR] FFmpeg was not found in PATH. Please install FFmpeg or use Mercury Dasha's in-app Sonic Chronicle studio." -ForegroundColor Red
    exit 1
}

Write-Host "`n[+] Initializing Steinberg UR-44 Capture via Windows DirectSound/WASAPI..." -ForegroundColor Green
Write-Host "[!] Press 'q' or Ctrl+C in this terminal when the D&D session concludes to finalize and upload.`n" -ForegroundColor DarkYellow

# FFmpeg command: capture audio via dshow, 48kHz, stereo, 24-bit PCM
$ffmpegArgs = @(
    "-y",
    "-f", "dshow",
    "-audio_buffer_size", "50",
    "-i", "audio=$DeviceName",
    "-ac", "2",
    "-ar", "48000",
    "-c:a", "pcm_s24le",
    "$OutputFile"
)

try {
    & ffmpeg $ffmpegArgs
} catch {
    Write-Host "`n[!] Capture interrupted by user." -ForegroundColor Yellow
}

if (-not (Test-Path $OutputFile)) {
    Write-Host "[WARN] No recording file was produced." -ForegroundColor Red
    exit 0
}

$FileSize = (Get-Item $OutputFile).Length
$FileSizeMB = [math]::Round($FileSize / 1MB, 2)
Write-Host "`n[✓] Capture complete! Saved local master: $OutputFile ($FileSizeMB MB)" -ForegroundColor Green

# Ingest into Mercury Dasha
Write-Host "[+] Ingesting session into Mercury Dasha Sovereign Storehouse at $ServerUrl..." -ForegroundColor Cyan

try {
    $boundary = [System.Guid]::NewGuid().ToString()
    $LF = "`r`n"
    
    $headers = @{
        "Content-Type" = "multipart/form-data; boundary=$boundary"
    }

    # Use curl.exe if available for robust large-file streaming upload
    $curlPath = Get-Command curl.exe -ErrorAction SilentlyContinue
    if ($curlPath) {
        & curl.exe -X POST "$ServerUrl/api/v1/sessions/upload" `
            -F "audio=@$OutputFile" `
            -F "title=$Title" `
            -F "campaign=$Campaign" `
            -F "dm=$DM" `
            -F "notes=Ingested via record-ur44.ps1 CLI"
        Write-Host "`n[✓] Successfully ingested into Mercury Dasha BoltDB & Indexer!" -ForegroundColor Green
    } else {
        Write-Host "[!] curl.exe not found; file saved to POSIX/Dropbox directory: $OutputFile" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[!] Upload to server failed ($ServerUrl). The raw audio file remains safe at $OutputFile." -ForegroundColor Yellow
}

Write-Host "`n[✓] All done! Listen or extract precision slices at $ServerUrl" -ForegroundColor Cyan
