"use client";

import React, { useState, useEffect, useRef, useCallback } from "react";
import {
  Mic,
  MicOff,
  Play,
  Pause,
  Square,
  RotateCcw,
  Sparkles,
  Film,
  Sliders,
  Download,
  CheckCircle,
  ChevronLeft,
  ChevronRight,
  Volume2,
  Layers,
  Video,
  Flame,
  Activity,
  Radio,
  ExternalLink,
  RefreshCw
} from "lucide-react";

export interface SlideItem {
  index: number;
  title: string;
  subtitle?: string;
  verses: string[];
  image_path: string;
  sacred_metal?: string;
  hermetic_axiom?: string;
  planetary_lord?: string;
  frequency_hz?: number;
  color_hex?: string;
}

export interface SlideshowData {
  id: string;
  title: string;
  subtitle?: string;
  tradition?: string;
  description?: string;
  slides: SlideItem[];
}

interface StationMarker {
  station: number;
  start_ms: number;
  end_ms?: number;
  title: string;
}

interface VocalNarrationStudioProps {
  slideshowId?: string;
  onRenderComplete?: (videoUrl: string) => void;
  className?: string;
}

export default function VocalNarrationStudio({
  slideshowId = "temple-of-illumination",
  onRenderComplete,
  className = ""
}: VocalNarrationStudioProps) {
  // Slideshow & Navigation State
  const [slideshow, setSlideshow] = useState<SlideshowData | null>(null);
  const [availableSlideshows, setAvailableSlideshows] = useState<any[]>([]);
  const [selectedId, setSelectedId] = useState(slideshowId);
  const [currentStationIdx, setCurrentStationIdx] = useState(0);
  const [loading, setLoading] = useState(true);

  // Audio Recording State
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string>("");
  const [isRecording, setIsRecording] = useState(false);
  const [isPaused, setIsPaused] = useState(false);
  const [recordingDurationMs, setRecordingDurationMs] = useState(0);
  const [recordedAudioBlob, setRecordedAudioBlob] = useState<Blob | null>(null);
  const [recordedAudioUrl, setRecordedAudioUrl] = useState<string | null>(null);
  const [stationMarkers, setStationMarkers] = useState<StationMarker[]>([]);

  // Video Rendering State
  const [isRendering, setIsRendering] = useState(false);
  const [renderProgress, setRenderProgress] = useState("");
  const [renderedVideoUrl, setRenderedVideoUrl] = useState<string | null>(null);
  const [renderedVideoStats, setRenderedVideoStats] = useState<{ size_mb: number } | null>(null);

  // Refs
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const audioChunksRef = useRef<Blob[]>([]);
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const streamRef = useRef<MediaStream | null>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const timerIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const recordingStartEpochRef = useRef<number>(0);
  const activeStationStartMsRef = useRef<number>(0);

  // 1. Fetch Slideshow Details & Available Catalog
  const fetchSlideshow = useCallback(async (id: string) => {
    try {
      setLoading(true);
      const res = await fetch(`/api/v1/slideshows/${id}`);
      if (res.ok) {
        const data = await res.json();
        setSlideshow(data);
        setCurrentStationIdx(0);
      }
    } catch (err) {
      console.error("Failed to load slideshow:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchCatalog = useCallback(async () => {
    try {
      const res = await fetch(`/api/v1/slideshows`);
      if (res.ok) {
        const data = await res.json();
        setAvailableSlideshows(data.slideshows || []);
      }
    } catch (err) {
      console.error("Failed to load catalog:", err);
    }
  }, []);

  useEffect(() => {
    fetchCatalog();
    fetchSlideshow(selectedId);
  }, [selectedId, fetchCatalog, fetchSlideshow]);

  // 2. Discover Hardware Audio Interfaces (e.g. Steinberg UR44)
  useEffect(() => {
    async function getDevices() {
      try {
        if (!navigator.mediaDevices?.enumerateDevices) return;
        const allDevices = await navigator.mediaDevices.enumerateDevices();
        const audioInputs = allDevices.filter((d) => d.kind === "audioinput");
        setDevices(audioInputs);

        // Auto-select Steinberg UR44 or first device
        const ur44 = audioInputs.find(
          (d) => d.label.includes("UR44") || d.label.includes("Steinberg")
        );
        if (ur44) {
          setSelectedDeviceId(ur44.deviceId);
        } else if (audioInputs.length > 0) {
          setSelectedDeviceId(audioInputs[0].deviceId);
        }
      } catch (err) {
        console.warn("Audio device enumeration:", err);
      }
    }
    getDevices();
  }, []);

  // 3. Canvas Waveform Visualizer
  const drawWaveform = useCallback(() => {
    if (!canvasRef.current || !analyserRef.current) return;
    const canvas = canvasRef.current;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const analyser = analyserRef.current;
    const bufferLength = analyser.frequencyBinCount;
    const dataArray = new Uint8Array(bufferLength);

    const renderLoop = () => {
      animationFrameRef.current = requestAnimationFrame(renderLoop);
      analyser.getByteFrequencyData(dataArray);

      ctx.clearRect(0, 0, canvas.width, canvas.height);

      // Background subtle gradient
      const bgGrad = ctx.createLinearGradient(0, 0, 0, canvas.height);
      bgGrad.addColorStop(0, "rgba(5, 8, 17, 0.95)");
      bgGrad.addColorStop(1, "rgba(10, 16, 31, 0.95)");
      ctx.fillStyle = bgGrad;
      ctx.fillRect(0, 0, canvas.width, canvas.height);

      // Draw Gold/Cyan Frequency Bars
      const barWidth = (canvas.width / bufferLength) * 2.5;
      let x = 0;

      for (let i = 0; i < bufferLength; i++) {
        const barHeight = (dataArray[i] / 255) * canvas.height * 0.85;

        // Gradient from Imperial Gold to Radiant Cyan
        const barGrad = ctx.createLinearGradient(0, canvas.height, 0, canvas.height - barHeight);
        barGrad.addColorStop(0, "rgba(212, 175, 55, 0.8)");
        barGrad.addColorStop(0.6, "rgba(230, 195, 92, 0.9)");
        barGrad.addColorStop(1, "rgba(92, 219, 149, 1.0)");

        ctx.fillStyle = barGrad;
        ctx.fillRect(x, canvas.height - barHeight, barWidth - 1, barHeight);

        x += barWidth;
        if (x > canvas.width) break;
      }
    };

    renderLoop();
  }, []);

  // 4. Start Live Studio Recording
  const startRecording = async () => {
    try {
      audioChunksRef.current = [];
      setRenderedVideoUrl(null);
      setRecordedAudioBlob(null);
      if (recordedAudioUrl) {
        URL.revokeObjectURL(recordedAudioUrl);
        setRecordedAudioUrl(null);
      }

      const constraints: MediaStreamConstraints = {
        audio: selectedDeviceId
          ? { deviceId: { exact: selectedDeviceId }, channelCount: 1, sampleRate: 48000 }
          : { channelCount: 1, sampleRate: 48000 }
      };

      const stream = await navigator.mediaDevices.getUserMedia(constraints);
      streamRef.current = stream;

      // AudioContext & Analyser setup
      const AudioCtx = window.AudioContext || (window as any).webkitAudioContext;
      const audioCtx = new AudioCtx({ sampleRate: 48000 });
      audioContextRef.current = audioCtx;

      const source = audioCtx.createMediaStreamSource(stream);
      const analyser = audioCtx.createAnalyser();
      analyser.fftSize = 256;
      source.connect(analyser);
      analyserRef.current = analyser;

      drawWaveform();

      // Initialize MediaRecorder
      const mimeTypes = ["audio/webm;codecs=opus", "audio/webm", "audio/wav"];
      let mimeType = "";
      for (const t of mimeTypes) {
        if (MediaRecorder.isTypeSupported(t)) {
          mimeType = t;
          break;
        }
      }

      const recorder = new MediaRecorder(stream, mimeType ? { mimeType } : undefined);
      mediaRecorderRef.current = recorder;

      recorder.ondataavailable = (e) => {
        if (e.data && e.data.size > 0) {
          audioChunksRef.current.push(e.data);
        }
      };

      recorder.onstop = () => {
        const audioBlob = new Blob(audioChunksRef.current, {
          type: recorder.mimeType || "audio/wav"
        });
        setRecordedAudioBlob(audioBlob);
        const url = URL.createObjectURL(audioBlob);
        setRecordedAudioUrl(url);

        // Close visualizer loop
        if (animationFrameRef.current) {
          cancelAnimationFrame(animationFrameRef.current);
        }
        if (audioContextRef.current && audioContextRef.current.state !== "closed") {
          audioContextRef.current.close();
        }
      };

      recorder.start(250); // Emit chunk every 250ms
      setIsRecording(true);
      setIsPaused(false);
      recordingStartEpochRef.current = Date.now();
      activeStationStartMsRef.current = 0;

      // Reset station markers
      if (slideshow && slideshow.slides.length > 0) {
        setStationMarkers([
          {
            station: 1,
            start_ms: 0,
            title: slideshow.slides[0].title
          }
        ]);
      }

      timerIntervalRef.current = setInterval(() => {
        setRecordingDurationMs(Date.now() - recordingStartEpochRef.current);
      }, 50);
    } catch (err) {
      console.error("Failed to start recording:", err);
      alert("Microphone access failed. Please ensure your microphone or audio interface is connected.");
    }
  };

  // 5. Station Navigation & Sentence Marker Sync
  const advanceToStation = (newIdx: number) => {
    if (!slideshow || newIdx < 0 || newIdx >= slideshow.slides.length) return;

    if (isRecording) {
      const nowMs = Date.now() - recordingStartEpochRef.current;
      setStationMarkers((prev) => {
        const updated = [...prev];
        if (updated.length > 0) {
          updated[updated.length - 1].end_ms = nowMs;
        }
        updated.push({
          station: newIdx + 1,
          start_ms: nowMs,
          title: slideshow.slides[newIdx].title
        });
        return updated;
      });
      activeStationStartMsRef.current = nowMs;
    }

    setCurrentStationIdx(newIdx);
  };

  // 6. Stop Recording
  const stopRecording = () => {
    if (mediaRecorderRef.current && mediaRecorderRef.current.state !== "inactive") {
      mediaRecorderRef.current.stop();
    }
    if (streamRef.current) {
      streamRef.current.getTracks().forEach((t) => t.stop());
    }
    if (timerIntervalRef.current) {
      clearInterval(timerIntervalRef.current);
    }
    setIsRecording(false);
    setIsPaused(false);

    // Finalize last station end_ms
    const finalDuration = Date.now() - recordingStartEpochRef.current;
    setStationMarkers((prev) => {
      const updated = [...prev];
      if (updated.length > 0) {
        updated[updated.length - 1].end_ms = finalDuration;
      }
      return updated;
    });
  };

  // 7. Auto-Render Video with Sidechain Ducking
  const triggerAutoRender = async () => {
    if (!recordedAudioBlob && !slideshow) return;

    try {
      setIsRendering(true);
      setRenderProgress("Uploading vocal track and initializing FFmpeg sidechain engine...");

      const formData = new FormData();
      if (recordedAudioBlob) {
        formData.append("audio", recordedAudioBlob, "vocal_narration.webm");
      }
      formData.append("markers", JSON.stringify(stationMarkers));

      setRenderProgress("Compositing 1080p scenes & ducking 528 Hz ambient drone by -14 dB...");

      const res = await fetch(`/api/v1/slideshows/${selectedId}/render`, {
        method: "POST",
        body: formData
      });

      if (!res.ok) {
        const errJson = await res.json().catch(() => ({}));
        throw new Error(errJson.error || `Server responded with ${res.status}`);
      }

      const data = await res.json();
      setRenderedVideoUrl(data.video_url);
      setRenderedVideoStats({ size_mb: data.file_size_mb });
      setRenderProgress("Render complete!");

      if (onRenderComplete) {
        onRenderComplete(data.video_url);
      }
    } catch (err: any) {
      console.error("Render failed:", err);
      alert(`Video compilation failed: ${err.message}`);
    } finally {
      setIsRendering(false);
    }
  };

  // Formatting helpers
  const formatTime = (ms: number) => {
    const totalSec = Math.floor(ms / 1000);
    const m = Math.floor(totalSec / 60);
    const s = totalSec % 60;
    const tenths = Math.floor((ms % 1000) / 100);
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}.${tenths}`;
  };

  const currentSlide = slideshow?.slides[currentStationIdx];

  return (
    <div className={`space-y-6 ${className}`}>
      {/* Studio Header & Configuration Bar */}
      <div className="flex flex-col md:flex-row items-start md:items-center justify-between gap-4 bg-gradient-to-r from-slate-900 via-[#0B1120] to-slate-900 p-5 rounded-2xl border border-slate-800 shadow-xl">
        <div className="flex items-center space-x-3">
          <div className="p-3 bg-amber-500/10 border border-amber-500/30 rounded-xl text-amber-400">
            <Radio className="w-6 h-6 animate-pulse" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white tracking-tight flex items-center gap-2">
              <span>AV Studio 1: Vocal Narration &amp; Teleprompter</span>
              <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                LIVE DUCKING
              </span>
            </h2>
            <p className="text-xs text-slate-400">
              Synchronize live voice narration with the 8 stations of illumination and Solfeggio harmonics.
            </p>
          </div>
        </div>

        {/* Input Selector & Slideshow Dropdown */}
        <div className="flex flex-wrap items-center gap-3">
          {/* Slideshow Selector */}
          <div className="flex items-center space-x-2 bg-slate-950/80 px-3 py-1.5 rounded-lg border border-slate-800 text-xs">
            <Layers className="w-3.5 h-3.5 text-amber-400" />
            <select
              value={selectedId}
              onChange={(e) => setSelectedId(e.target.value)}
              disabled={isRecording || isRendering}
              className="bg-transparent text-slate-200 outline-none cursor-pointer"
            >
              {availableSlideshows.map((s) => (
                <option key={s.id} value={s.id} className="bg-slate-900 text-white">
                  {s.title}
                </option>
              ))}
            </select>
          </div>

          {/* Microphone Selector */}
          <div className="flex items-center space-x-2 bg-slate-950/80 px-3 py-1.5 rounded-lg border border-slate-800 text-xs">
            <Mic className="w-3.5 h-3.5 text-cyan-400" />
            <select
              value={selectedDeviceId}
              onChange={(e) => setSelectedDeviceId(e.target.value)}
              disabled={isRecording || isRendering}
              className="bg-transparent text-slate-200 outline-none cursor-pointer max-w-[200px] truncate"
            >
              {devices.map((d) => (
                <option key={d.deviceId} value={d.deviceId} className="bg-slate-900 text-white">
                  {d.label || `Audio Interface (${d.deviceId.slice(0, 8)})`}
                </option>
              ))}
            </select>
          </div>
        </div>
      </div>

      {/* Main Studio Console Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
        {/* Left Column: Visual Scene & Teleprompter (8 Cols) */}
        <div className="lg:col-span-8 space-y-4">
          {/* Active Station Card */}
          <div className="relative overflow-hidden rounded-2xl border border-slate-800 bg-slate-950 shadow-2xl">
            {/* Visual Slide Background / Preview */}
            <div className="relative aspect-video w-full overflow-hidden bg-black flex items-center justify-center">
              {currentSlide?.image_path ? (
                <img
                  src={
                    currentSlide.image_path.startsWith("assets/")
                      ? `/${currentSlide.image_path}`
                      : `/api/v1/index/content?path=${currentSlide.image_path}`
                  }
                  alt={currentSlide.title}
                  className="w-full h-full object-cover transition-all duration-700 hover:scale-105"
                />
              ) : (
                <div className="text-slate-600 font-mono text-sm">Station Visual Asset Preview</div>
              )}

              {/* Station Badges Overlay */}
              <div className="absolute top-4 left-4 right-4 flex items-center justify-between pointer-events-none">
                <span className="px-3 py-1 rounded-lg bg-black/75 backdrop-blur-md border border-amber-500/40 text-amber-300 font-bold text-xs tracking-wider uppercase">
                  Station {currentStationIdx + 1} of {slideshow?.slides.length || 8}
                </span>

                <div className="flex items-center space-x-2">
                  {currentSlide?.sacred_metal && (
                    <span className="px-2.5 py-1 rounded-md bg-slate-900/80 backdrop-blur-md border border-slate-700 text-cyan-300 text-[11px] font-mono">
                      Metal: {currentSlide.sacred_metal}
                    </span>
                  )}
                  {currentSlide?.frequency_hz && (
                    <span className="px-2.5 py-1 rounded-md bg-slate-900/80 backdrop-blur-md border border-slate-700 text-emerald-300 text-[11px] font-mono">
                      {currentSlide.frequency_hz} Hz
                    </span>
                  )}
                </div>
              </div>

              {/* Lower-Third Subtitle / Verse Overlay */}
              <div className="absolute bottom-0 inset-x-0 p-6 bg-gradient-to-t from-black via-black/85 to-transparent backdrop-blur-[2px]">
                <h3 className="text-lg md:text-xl font-serif text-amber-300 font-bold mb-2">
                  {currentSlide?.title}
                </h3>
                <div className="space-y-1 text-slate-100 font-serif text-sm md:text-base leading-relaxed tracking-wide">
                  {currentSlide?.verses.map((line, i) => (
                    <p key={i} className="drop-shadow-[0_2px_4px_rgba(0,0,0,0.9)]">
                      {line}
                    </p>
                  ))}
                </div>
              </div>
            </div>

            {/* Station Stepper Controls */}
            <div className="flex items-center justify-between p-4 bg-slate-900/90 border-t border-slate-800">
              <button
                onClick={() => advanceToStation(currentStationIdx - 1)}
                disabled={currentStationIdx === 0}
                className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 disabled:opacity-30 disabled:cursor-not-allowed text-xs text-slate-200 transition"
              >
                <ChevronLeft className="w-4 h-4" />
                <span>Prev Station</span>
              </button>

              <div className="text-xs font-mono text-slate-400">
                Axiom: <span className="text-slate-200">{currentSlide?.hermetic_axiom || "Mentalism"}</span>
              </div>

              <button
                onClick={() => advanceToStation(currentStationIdx + 1)}
                disabled={!slideshow || currentStationIdx === slideshow.slides.length - 1}
                className="flex items-center space-x-1.5 px-4 py-1.5 rounded-lg bg-amber-500/20 hover:bg-amber-500/30 border border-amber-500/40 text-amber-300 font-medium text-xs transition"
              >
                <span>Next Station</span>
                <ChevronRight className="w-4 h-4" />
              </button>
            </div>
          </div>
        </div>

        {/* Right Column: Audio Recording Engine & Controls (4 Cols) */}
        <div className="lg:col-span-4 space-y-4">
          {/* Waveform Oscilloscope Card */}
          <div className="rounded-2xl border border-slate-800 bg-slate-950 p-5 space-y-4 shadow-xl">
            <div className="flex items-center justify-between">
              <span className="text-xs font-mono uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                <Activity className="w-3.5 h-3.5 text-cyan-400" />
                <span>Studio Audio Input</span>
              </span>
              <span
                className={`text-xs font-mono px-2 py-0.5 rounded-full font-bold ${
                  isRecording ? "bg-red-500/20 text-red-400 animate-pulse" : "bg-slate-800 text-slate-400"
                }`}
              >
                {isRecording ? "RECORDING" : "STANDBY"}
              </span>
            </div>

            {/* Canvas Frequency Bars */}
            <div className="relative rounded-xl overflow-hidden border border-slate-800 h-28 bg-[#050811]">
              <canvas ref={canvasRef} width={400} height={112} className="w-full h-full block" />
              {!isRecording && (
                <div className="absolute inset-0 flex items-center justify-center text-xs font-mono text-slate-600">
                  Ready to record voice narration
                </div>
              )}
            </div>

            {/* Timer & Level Gauge */}
            <div className="flex items-center justify-between bg-slate-900/60 p-3 rounded-xl border border-slate-800/80">
              <span className="text-xs text-slate-400">Total Duration:</span>
              <span className="text-xl font-mono font-bold text-amber-300 tracking-wider">
                {formatTime(recordingDurationMs)}
              </span>
            </div>

            {/* Primary Action Buttons */}
            <div className="space-y-2">
              {!isRecording ? (
                <button
                  onClick={startRecording}
                  disabled={isRendering}
                  className="w-full py-3.5 rounded-xl bg-gradient-to-r from-red-600 via-rose-600 to-red-600 hover:from-red-500 hover:to-rose-500 text-white font-bold text-sm shadow-lg shadow-red-900/40 flex items-center justify-center space-x-2 transition"
                >
                  <Mic className="w-4 h-4" />
                  <span>Start Live Narration</span>
                </button>
              ) : (
                <button
                  onClick={stopRecording}
                  className="w-full py-3.5 rounded-xl bg-slate-800 hover:bg-slate-700 text-amber-300 border border-amber-500/40 font-bold text-sm shadow-lg flex items-center justify-center space-x-2 transition"
                >
                  <Square className="w-4 h-4 text-red-400 fill-red-400" />
                  <span>Complete &amp; Stop Narration</span>
                </button>
              )}

              {/* Recorded Audio Review Player */}
              {recordedAudioUrl && !isRecording && (
                <div className="p-3 bg-slate-900/80 rounded-xl border border-slate-800 space-y-2">
                  <div className="text-[11px] font-mono text-slate-400 flex items-center justify-between">
                    <span>Recorded Vocal Track:</span>
                    <span className="text-emerald-400 font-bold">Ready</span>
                  </div>
                  <audio src={recordedAudioUrl} controls className="w-full h-8" />
                </div>
              )}

              {/* One-Click Auto-Render Trigger */}
              {recordedAudioBlob && !isRecording && (
                <button
                  onClick={triggerAutoRender}
                  disabled={isRendering}
                  className="w-full py-3.5 rounded-xl bg-gradient-to-r from-amber-600 via-amber-500 to-yellow-500 hover:from-amber-500 hover:to-yellow-400 text-slate-950 font-black text-sm shadow-xl shadow-amber-950/50 flex items-center justify-center space-x-2 transition disabled:opacity-50"
                >
                  {isRendering ? (
                    <>
                      <RefreshCw className="w-4 h-4 animate-spin text-slate-950" />
                      <span>{renderProgress || "Rendering MP4..."}</span>
                    </>
                  ) : (
                    <>
                      <Sparkles className="w-4 h-4 text-slate-950" />
                      <span>Compile Video (Sidechain Ducking)</span>
                    </>
                  )}
                </button>
              )}
            </div>
          </div>

          {/* Rendered MP4 Video Preview Card */}
          {renderedVideoUrl && (
            <div className="rounded-2xl border border-emerald-500/30 bg-emerald-950/20 p-5 space-y-4 shadow-xl">
              <div className="flex items-center justify-between text-xs">
                <span className="text-emerald-400 font-bold flex items-center gap-1.5">
                  <CheckCircle className="w-4 h-4" />
                  <span>Rendered Video Ready</span>
                </span>
                {renderedVideoStats && (
                  <span className="font-mono text-slate-400 text-[11px]">
                    {renderedVideoStats.size_mb.toFixed(2)} MB
                  </span>
                )}
              </div>

              <div className="rounded-xl overflow-hidden border border-emerald-500/30 aspect-video bg-black">
                <video src={renderedVideoUrl} controls className="w-full h-full object-cover" />
              </div>

              <a
                href={`${renderedVideoUrl}&download=1`}
                download="temple_of_illumination_vocal.mp4"
                className="w-full py-2.5 rounded-xl bg-emerald-600 hover:bg-emerald-500 text-white font-bold text-xs flex items-center justify-center space-x-2 transition shadow-lg"
              >
                <Download className="w-3.5 h-3.5" />
                <span>Download Sovereign Video (MP4)</span>
              </a>
            </div>
          )}
        </div>
      </div>

      {/* Station Rail: All 8 Stations Selector */}
      <div className="rounded-2xl border border-slate-800 bg-slate-950/70 p-4">
        <h4 className="text-xs font-mono uppercase tracking-wider text-slate-400 mb-3 flex items-center gap-2">
          <Film className="w-3.5 h-3.5 text-amber-400" />
          <span>Stations of Illumination ({slideshow?.slides.length || 8})</span>
        </h4>

        <div className="grid grid-cols-2 sm:grid-cols-4 md:grid-cols-8 gap-3">
          {slideshow?.slides.map((s, idx) => (
            <button
              key={s.index}
              onClick={() => advanceToStation(idx)}
              className={`p-2 rounded-xl border text-left transition relative overflow-hidden group ${
                idx === currentStationIdx
                  ? "border-amber-500 bg-amber-500/10 shadow-lg shadow-amber-500/10"
                  : "border-slate-800 bg-slate-900/60 hover:border-slate-700"
              }`}
            >
              <div className="aspect-video w-full rounded-lg overflow-hidden bg-black mb-1.5">
                <img
                  src={
                    s.image_path.startsWith("assets/")
                      ? `/${s.image_path}`
                      : `/api/v1/index/content?path=${s.image_path}`
                  }
                  alt={s.title}
                  className="w-full h-full object-cover group-hover:scale-110 transition-transform"
                />
              </div>
              <div className="text-[11px] font-bold text-slate-200 truncate">{s.title}</div>
              <div className="text-[9px] font-mono text-slate-400 truncate">{s.sacred_metal}</div>
            </button>
          ))}
        </div>
      </div>
    </div>
  );
}
