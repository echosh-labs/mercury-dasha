#!/usr/bin/env python3
"""
MERCURY DASHA // AV STUDIO 1 - LIVE STUDIO VOICE RECORDER & SYNC
================================================================================
Captures studio-grade vocal narration from hardware audio interfaces
(e.g. Steinberg UR44, USB microphones) using Windows DirectShow via FFmpeg,
performs automatic speech pause detection, and triggers sidechain ducked video render.
================================================================================
"""

import argparse
import os
import re
import subprocess
import sys
import time

STANZA_PROMPTS = [
    {
        "station": 1,
        "title": "The Ocean of Light",
        "verses": [
            "There's an ocean that men cannot fathom nor measure;",
            "It lies just beyond the Dominion of Night;",
            "'Tis the ocean of splendor, of infinite pleasure,",
            "Of fathomless beauty—the ocean of Light."
        ]
    },
    {
        "station": 2,
        "title": "The Island of Blessing",
        "verses": [
            "In the midst of this radiant ocean of glory",
            "Rests the Island of Blessing, the gem of the sea;",
            "The home of the Spirit, so famous in story,",
            "Where angels are servants, and men are the free."
        ]
    },
    {
        "station": 3,
        "title": "The Mount of the Wise",
        "verses": [
            "In the midst of the Isle is a flower-crowned mountain;",
            "The sanctified call it the Mount of the Wise;",
            "From its summit pours forth a life-giving fountain",
            "That waters the lands of the earth and the skies."
        ]
    },
    {
        "station": 4,
        "title": "The House of the Sun",
        "verses": [
            "On the top of the mountain a Temple, all glorious,",
            "Stands out in the light of the Illumined One.",
            "Ten thousand bright angels and souls all victorious,",
            "Surround it, and fill it—this House of the Sun."
        ]
    },
    {
        "station": 5,
        "title": "The Temple of Illumination",
        "verses": [
            "And this is the Temple of Illumination",
            "Where the courtiers of heaven and earth daily meet;",
            "Where souls, cleansed from sin, elect from each nation,",
            "Hold council with Jesus, and sit at his feet."
        ]
    },
    {
        "station": 6,
        "title": "The Valley of Silence and Prayer",
        "verses": [
            "The way to this Island and unto this Mountain",
            "Lies through the deep valley of Silence and Prayer;",
            "But whoever will may drink from the Fountain,",
            "And realize all that it is to be there."
        ]
    },
    {
        "station": 7,
        "title": "The Thrice-Blessed Day",
        "verses": [
            "Come up to this Temple of Illumination;",
            "Come, bathe in the sunlight of a thrice-blessed day.",
            "There is room for the millions of every nation,",
            "And Christ is the Truth and the Life and the Way."
        ]
    },
    {
        "station": 8,
        "title": "The Innermost Circle",
        "verses": [
            "In the Innermost Circle there's joy and there's gladness;",
            "There's peace and there's freedom from sin and from strife.",
            "The Angel of Mercy will free you from sadness,",
            "And Christ's Benediction is Eternal Life."
        ]
    }
]

def list_audio_devices():
    """Find DirectShow audio devices available on Windows."""
    try:
        proc = subprocess.run(
            ["ffmpeg", "-list_devices", "true", "-f", "dshow", "-i", "dummy"],
            stderr=subprocess.PIPE,
            stdout=subprocess.PIPE,
            text=True,
            encoding="utf-8",
            errors="ignore"
        )
        output = proc.stderr
        devices = []
        for line in output.splitlines():
            if "(audio)" in line:
                m = re.search(r'"([^"]+)"\s+\(audio\)', line)
                if m:
                    devices.append(m.group(1))
        return devices
    except Exception as e:
        print(f"Error enumerating audio devices: {e}")
        return []

def print_teleprompter():
    print("\n" + "="*70)
    print("      THE TEMPLE OF ILLUMINATION — VOCAL TELEPROMPTER")
    print("="*70)
    for s in STANZA_PROMPTS:
        print(f"\n[Station {s['station']}: {s['title'].upper()}]")
        for line in s['verses']:
            print(f"   {line}")
    print("\n" + "="*70 + "\n")

def main():
    parser = argparse.ArgumentParser(description="Record live vocal narration for AV Studio 1.")
    parser.add_argument("--device", default="", help="DirectShow audio device name (e.g. 'Line (3- Steinberg UR44)')")
    parser.add_argument("--output", default="renders/temple_voice.wav", help="Output audio file path")
    parser.add_argument("--channels", type=int, default=1, help="Audio channels: 1 (mono) or 2 (stereo)")
    parser.add_argument("--sample-rate", type=int, default=48000, help="Sample rate in Hz (default: 48000)")
    parser.add_argument("--auto-render", action="store_true", help="Automatically trigger FFmpeg video compilation upon completion")
    args = parser.parse_args()

    devices = list_audio_devices()
    selected_device = args.device

    if not selected_device:
        # Prioritize Steinberg UR44 if detected
        ur44 = [d for d in devices if "UR44" in d or "Steinberg" in d]
        if ur44:
            selected_device = ur44[0]
            print(f"Auto-selected detected hardware interface: {selected_device}")
        elif devices:
            selected_device = devices[0]
            print(f"Using default audio device: {selected_device}")
        else:
            print("No DirectShow audio input devices detected.")
            sys.exit(1)

    print("\n" + "═"*70)
    print("      AV STUDIO 1: VOCAL NARRATION RECORDER")
    print("═"*70)
    print(f"  • Selected Input:  {selected_device}")
    print(f"  • Format:          48,000 Hz, {'Mono' if args.channels == 1 else 'Stereo'} PCM WAV")
    print(f"  • Output Target:   {args.output}")
    print("═"*70)

    print_teleprompter()

    os.makedirs(os.path.dirname(os.path.abspath(args.output)), exist_ok=True)

    input("Press [ENTER] to start recording studio narration (or Ctrl+C to abort)...")

    cmd = [
        "ffmpeg", "-y",
        "-f", "dshow",
        "-i", f"audio={selected_device}",
        "-ac", str(args.channels),
        "-ar", str(args.sample_rate),
        args.output
    ]

    print("\n🔴 RECORDING IN PROGRESS... Speak clearly into your microphone.")
    print("When finished with all 8 stations, press [Q] or [Ctrl+C] to stop.\n")

    try:
        proc = subprocess.Popen(cmd)
        proc.wait()
    except KeyboardInterrupt:
        print("\nStopping recording...")
        if proc.poll() is None:
            proc.terminate()
            try:
                proc.wait(timeout=3)
            except subprocess.TimeoutExpired:
                proc.kill()

    if os.path.isfile(args.output) and os.path.getsize(args.output) > 1000:
        size_kb = os.path.getsize(args.output) / 1024
        print(f"\n✔ Recording saved successfully: {args.output} ({size_kb:.1f} KB)")
        
        if args.auto_render:
            print("\nTriggering video compilation with sidechain ducking...")
            render_cmd = [
                "python", "scripts/render_slideshow.py",
                "--voice-audio", args.output,
                "--output", "renders/temple_of_illumination_with_voice.mp4"
            ]
            subprocess.run(render_cmd)
    else:
        print("Warning: Recording file was empty or not created.")

if __name__ == "__main__":
    main()
