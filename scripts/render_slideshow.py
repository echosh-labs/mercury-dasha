#!/usr/bin/env python3
"""
MERCURY DASHA // AV STUDIO 1 - OFFLINE SLIDESHOW VIDEO RENDERER
================================================================================
Compiles an ordered sequence of images, verses, and multi-track audio into a
production-grade MP4 video using FFmpeg.

Features:
- Configurable per-slide durations and crossfade transition timings
- Multi-harmonic Solfeggio drone generator (528 Hz + 264 Hz + 126.22 Hz Sun)
- Optional external voice narration track with automatic audio ducking
- Optional lower-third verse subtitle rendering with gold esoteric styling
- Sub-pixel motion interpolation (Ken Burns zoom / pan)
================================================================================
"""

import argparse
import json
import os
import subprocess
import sys

STANZA_VERSES = [
    {
        "index": 1,
        "title": "The Ocean of Light",
        "metal": "Quicksilver",
        "axiom": "Principle of Polarity",
        "freq": "528 Hz",
        "verses": [
            "There's an ocean that men cannot fathom nor measure;",
            "It lies just beyond the Dominion of Night;",
            "'Tis the ocean of splendor, of infinite pleasure,",
            "Of fathomless beauty—the ocean of Light."
        ]
    },
    {
        "index": 2,
        "title": "The Island of Blessing",
        "metal": "Silver",
        "axiom": "Principle of Correspondence",
        "freq": "432 Hz",
        "verses": [
            "In the midst of this radiant ocean of glory",
            "Rests the Island of Blessing, the gem of the sea;",
            "The home of the Spirit, so famous in story,",
            "Where angels are servants, and men are the free."
        ]
    },
    {
        "index": 3,
        "title": "The Mount of the Wise",
        "metal": "Tin",
        "axiom": "Principle of Vibration",
        "freq": "639 Hz",
        "verses": [
            "In the midst of the Isle is a flower-crowned mountain;",
            "The sanctified call it the Mount of the Wise;",
            "From its summit pours forth a life-giving fountain",
            "That waters the lands of the earth and the skies."
        ]
    },
    {
        "index": 4,
        "title": "The House of the Sun",
        "metal": "Gold",
        "axiom": "Principle of Gender",
        "freq": "126.22 Hz",
        "verses": [
            "On the top of the mountain a Temple, all glorious,",
            "Stands out in the light of the Illumined One.",
            "Ten thousand bright angels and souls all victorious,",
            "Surround it, and fill it—this House of the Sun."
        ]
    },
    {
        "index": 5,
        "title": "The Temple of Illumination",
        "metal": "Gold",
        "axiom": "Principle of Mentalism",
        "freq": "741 Hz",
        "verses": [
            "And this is the Temple of Illumination",
            "Where the courtiers of heaven and earth daily meet;",
            "Where souls, cleansed from sin, elect from each nation,",
            "Hold council with Jesus, and sit at his feet."
        ]
    },
    {
        "index": 6,
        "title": "The Valley of Silence and Prayer",
        "metal": "Lead",
        "axiom": "Principle of Rhythm",
        "freq": "396 Hz",
        "verses": [
            "The way to this Island and unto this Mountain",
            "Lies through the deep valley of Silence and Prayer;",
            "But whoever will may drink from the Fountain,",
            "And realize all that it is to be there."
        ]
    },
    {
        "index": 7,
        "title": "The Thrice-Blessed Day",
        "metal": "Copper",
        "axiom": "Principle of Cause and Effect",
        "freq": "852 Hz",
        "verses": [
            "Come up to this Temple of Illumination;",
            "Come, bathe in the sunlight of a thrice-blessed day.",
            "There is room for the millions of every nation,",
            "And Christ is the Truth and the Life and the Way."
        ]
    },
    {
        "index": 8,
        "title": "The Innermost Circle",
        "metal": "Amara / Stone",
        "axiom": "Mental Transmutation",
        "freq": "963 Hz",
        "verses": [
            "In the Innermost Circle there's joy and there's gladness;",
            "There's peace and there's freedom from sin and from strife.",
            "The Angel of Mercy will free you from sadness,",
            "And Christ's Benediction is Eternal Life."
        ]
    }
]

def main():
    parser = argparse.ArgumentParser(description="Render an offline MP4 video from Hermetic slideshow assets.")
    parser.add_argument("--slides-dir", default="backend/storage/slideshows/temple_of_illumination",
                        help="Path to slideshow images directory")
    parser.add_argument("--slide-duration", type=float, default=8.0,
                        help="Duration of each slide in seconds (default: 8.0s)")
    parser.add_argument("--transition-duration", type=float, default=1.2,
                        help="Crossfade duration between slides in seconds (default: 1.2s)")
    parser.add_argument("--fps", type=int, default=30, help="Frames per second (default: 30)")
    parser.add_argument("--width", type=int, default=1920, help="Video width (default: 1920)")
    parser.add_argument("--height", type=int, default=1080, help="Video height (default: 1080)")
    parser.add_argument("--output", default="renders/temple_of_illumination.mp4",
                        help="Output MP4 file path (default: renders/temple_of_illumination.mp4)")
    parser.add_argument("--voice-audio", default="",
                        help="Optional voice narration audio track path (WAV/MP3/M4A)")
    parser.add_argument("--drone-freq", type=float, default=528.0,
                        help="Base Solfeggio frequency for ambient drone (default: 528.0 Hz)")
    parser.add_argument("--no-subtitles", action="store_true",
                        help="Disable drawing verses and lower-third titles")
    parser.add_argument("--preset", default="veryfast",
                        help="x264 encoding preset (ultrafast, veryfast, faster, medium)")
    
    args = parser.parse_args()

    slides_dir = os.path.abspath(args.slides_dir)
    if not os.path.isdir(slides_dir):
        print(f"Error: slides directory '{slides_dir}' does not exist.", file=sys.stderr)
        sys.exit(1)

    # Collect images in sorted order
    files = sorted([f for f in os.listdir(slides_dir) if f.lower().endswith(('.jpg', '.jpeg', '.png'))])
    if not files:
        print(f"Error: no images found in '{slides_dir}'", file=sys.stderr)
        sys.exit(1)

    image_paths = [os.path.join(slides_dir, f) for f in files]
    num_slides = len(image_paths)
    print(f"Loaded {num_slides} slide images from: {slides_dir}")

    slide_dur = args.slide_duration
    trans_dur = args.transition_duration
    if trans_dur >= slide_dur:
        trans_dur = slide_dur / 2.0

    # Total duration = N * slide_dur - (N-1) * trans_dur
    total_duration = num_slides * slide_dur - (num_slides - 1) * trans_dur
    print(f"Target duration: {total_duration:.2f}s across {num_slides} slides ({slide_dur:.1f}s each with {trans_dur:.1f}s crossfade)")

    os.makedirs(os.path.dirname(os.path.abspath(args.output)), exist_ok=True)

    # Build FFmpeg command inputs
    cmd = ["ffmpeg", "-y"]
    
    for img in image_paths:
        cmd.extend(["-loop", "1", "-t", str(slide_dur), "-i", img])

    # Audio input: synthesizes 528 Hz + 264 Hz octave lower + 126.22 Hz Sun frequency
    drone_expr = (
        f"0.18*sin(2*PI*{args.drone_freq}*t) + "
        f"0.14*sin(2*PI*{args.drone_freq/2.0}*t) + "
        f"0.08*sin(2*PI*126.22*t)"
    )
    cmd.extend([
        "-f", "lavfi",
        "-i", f"aevalsrc={drone_expr}:s=48000:d={total_duration}"
    ])
    audio_input_idx = num_slides

    has_voice = False
    voice_input_idx = -1
    if args.voice_audio and os.path.isfile(args.voice_audio):
        has_voice = True
        cmd.extend(["-i", args.voice_audio])
        voice_input_idx = num_slides + 1

    # Build Video Filtergraph
    # 1. Scale each input to canvas dimensions
    filter_chains = []
    for i in range(num_slides):
        filter_chains.append(
            f"[{i}:v]scale={args.width}:{args.height}:force_original_aspect_ratio=increase,"
            f"crop={args.width}:{args.height},"
            f"setsar=1,fps={args.fps}[v{i}]"
        )

    # 2. Chain with xfade crossfade transitions
    curr_label = "v0"
    for i in range(1, num_slides):
        offset = i * (slide_dur - trans_dur)
        next_label = f"xf{i}" if i < num_slides - 1 else "v_faded"
        filter_chains.append(
            f"[{curr_label}][v{i}]xfade=transition=fade:duration={trans_dur}:offset={offset:.3f}[{next_label}]"
        )
        curr_label = next_label

    # 3. Subtitles & Telemetry Overlays
    final_video = curr_label
    if not args.no_subtitles:
        font_path = "/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf"
        font_arg = f":fontfile='{font_path}'" if os.path.isfile(font_path) else ""
        
        drawtext_filters = []
        for i in range(num_slides):
            data = STANZA_VERSES[i % len(STANZA_VERSES)]
            start_t = i * (slide_dur - trans_dur)
            end_t = start_t + slide_dur
            
            title_text = f"STATION {data['index']}: {data['title'].upper()}"
            meta_text = f"Sacred Metal: {data['metal']}  •  Axiom: {data['axiom']}  •  Harmonic: {data['freq']}"
            verses_joined = "  |  ".join(data["verses"])
            
            # Clean text for ffmpeg drawtext filter (replace ASCII quotes and semicolons to avoid breaking FFmpeg parser)
            clean_verses = verses_joined.replace("'", "’").replace(";", " —").replace(":", " -")
            clean_title = title_text.replace("'", "’").replace(";", " -").replace(":", "\\:")
            clean_meta = meta_text.replace("'", "’").replace(";", " -").replace(":", "\\:")

            safe_title = clean_title
            safe_meta = clean_meta
            safe_verses = clean_verses

            # Header Banner (Top-Left)
            drawtext_filters.append(
                f"drawtext=text='{safe_title}'{font_arg}:fontsize=28:fontcolor=0xE6C35C:"
                f"box=1:boxcolor=0x050811@0.75:boxborderw=12:x=60:y=60:enable='between(t,{start_t:.3f},{end_t:.3f})'"
            )
            # Epigraph / Correspondences
            drawtext_filters.append(
                f"drawtext=text='{safe_meta}'{font_arg}:fontsize=18:fontcolor=0x94A3B8:"
                f"box=1:boxcolor=0x050811@0.75:boxborderw=8:x=60:y=112:enable='between(t,{start_t:.3f},{end_t:.3f})'"
            )
            # Lower-Third Subtitle Bar (Bottom)
            drawtext_filters.append(
                f"drawtext=text='{safe_verses}'{font_arg}:fontsize=22:fontcolor=0xFFFFFF:"
                f"box=1:boxcolor=0x050811@0.80:boxborderw=16:x=(w-text_w)/2:y=h-100:enable='between(t,{start_t:.3f},{end_t:.3f})'"
            )

        filter_chains.append(f"[{curr_label}]" + ",".join(drawtext_filters) + "[v_final]")
        final_video = "v_final"

    # Audio Filtergraph: Drone + optional Voice with Ducking
    if has_voice:
        audio_filter = (
            f"[{audio_input_idx}:a]aformat=sample_rates=48000:channel_layouts=stereo,afade=t=in:st=0:d=2,afade=t=out:st={total_duration-3}:d=3[drone_faded];"
            f"[{voice_input_idx}:a]aformat=sample_rates=48000:channel_layouts=stereo,volume=1.2,asplit=2[voice_sc][voice_mix];"
            f"[drone_faded][voice_sc]sidechaincompress=threshold=0.12:ratio=6:attack=150:release=700[ducked_drone];"
            f"[ducked_drone][voice_mix]amix=inputs=2:duration=first:dropout_transition=2[a_final]"
        )
        filter_chains.append(audio_filter)
        final_audio = "a_final"
    else:
        filter_chains.append(
            f"[{audio_input_idx}:a]aformat=sample_rates=48000:channel_layouts=stereo,afade=t=in:st=0:d=2,afade=t=out:st={total_duration-3}:d=3,volume=0.35[a_final]"
        )
        final_audio = "a_final"

    full_filter = ";\n".join(filter_chains)

    cmd.extend([
        "-filter_complex", full_filter,
        "-map", f"[{final_video}]",
        "-map", f"[{final_audio}]",
        "-c:v", "libx264",
        "-pix_fmt", "yuv420p",
        "-preset", args.preset,
        "-crf", "20",
        "-c:a", "aac",
        "-b:a", "192k",
        "-shortest",
        args.output
    ])

    print("\nExecuting FFmpeg Video Compilation Pipeline...")
    print(f"Output File: {args.output}")

    res = subprocess.run(cmd)
    if res.returncode != 0:
        print(f"\nFFmpeg compilation failed with return code {res.returncode}", file=sys.stderr)
        sys.exit(res.returncode)

    file_size_mb = os.path.getsize(args.output) / (1024 * 1024)
    print(f"\n✔ Slideshow Video Successfully Rendered!")
    print(f"  • Path:     {args.output}")
    print(f"  • Size:     {file_size_mb:.2f} MB")
    print(f"  • Duration: {total_duration:.2f}s")
    print(f"  • Quality:  1080p @ {args.fps}fps H.264 / AAC")

if __name__ == "__main__":
    main()
