"use client";

import React, { useState, useEffect, useRef, useCallback, useMemo } from "react";
import {
  Mic,
  Radio,
  Play,
  Pause,
  RotateCcw,
  RotateCw,
  Volume2,
  VolumeX,
  Volume1,
  Search,
  FileText,
  RefreshCw,
  Clock,
  Calendar,
  ShieldCheck,
  CheckCircle,
  Download,
  Copy,
  Check,
  Sparkles,
  Bookmark,
  Sliders,
  ChevronRight,
  AlertCircle,
  FolderTree,
  Headphones,
  FileAudio,
  Hash,
  Filter,
  Scissors,
  ChevronDown,
  ChevronUp,
  Activity,
  Layers,
  HardDrive,
  Info,
  Wand2,
  Square,
  Circle,
  Zap,
  Swords,
  Scroll,
  Tag
} from "lucide-react";

export interface UnifiedAudioItem {
  id: string;
  title: string;
  source: "recorder" | "vault" | "session" | string;
  category: string;
  format: string;
  duration?: string;
  duration_ms?: number;
  size_bytes: number;
  recorded_at?: string;
  mod_time?: string;
  has_transcript: boolean;
  is_synced: boolean;
  audio_url: string;
  transcript_url?: string;
  transcript_snippet?: string;
  metadata?: Record<string, any>;
  tags?: string[];
  path: string;
}

interface DiarizedParagraph {
  speaker: string;
  start_ms: number;
  end_ms?: number;
  text: string;
}

interface TranscriptPayload {
  recording_id: string;
  raw_text: string;
  paragraphs: DiarizedParagraph[];
  json?: any;
}

interface CatalogResponse {
  items: UnifiedAudioItem[];
  total: number;
  counts: {
    total: number;
    with_transcript: number;
    recorder: number;
    vault: number;
    sessions: number;
    formats: Record<string, number>;
  };
}

interface SyncStatus {
  is_syncing: boolean;
  current_item?: string;
  total_seen: number;
  total_downloaded: number;
  total_skipped: number;
  total_errors: number;
  last_error?: string;
  manifest_total: number;
}

interface StatusOverview {
  mcp_connected: boolean;
  mcp_url: string;
  auth_valid: boolean;
  output_dir: string;
  total_local_recordings: number;
  last_sync_time: string;
  auto_sync_mins: number;
  is_syncing: boolean;
}

export default function UnifiedAudioPortal() {
  // Catalog State
  const [items, setItems] = useState<UnifiedAudioItem[]>([]);
  const [totalCount, setTotalCount] = useState<number>(0);
  const [counts, setCounts] = useState<CatalogResponse["counts"]>({
    total: 0,
    with_transcript: 0,
    recorder: 0,
    vault: 0,
    sessions: 0,
    formats: {},
  });
  const [selectedItem, setSelectedItem] = useState<UnifiedAudioItem | null>(null);
  const [loadingList, setLoadingList] = useState<boolean>(true);

  // Filters & Search
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [sourceFilter, setSourceFilter] = useState<"all" | "transcript" | "vault" | "session">("all");
  const [formatFilter, setFormatFilter] = useState<string>("all");

  // Transcript & Right Panel State
  const [transcript, setTranscript] = useState<TranscriptPayload | null>(null);
  const [loadingTranscript, setLoadingTranscript] = useState<boolean>(false);
  const [transcriptSearch, setTranscriptSearch] = useState<string>("");
  const [copiedId, setCopiedId] = useState<string | null>(null);
  const [autoScroll, setAutoScroll] = useState<boolean>(true);

  // MCP & Sync State
  const [status, setStatus] = useState<StatusOverview | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [isSyncing, setIsSyncing] = useState<boolean>(false);

  // Audio Playback Engine
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const transcriptContainerRef = useRef<HTMLDivElement | null>(null);
  const [isPlaying, setIsPlaying] = useState<boolean>(false);
  const [currentTime, setCurrentTime] = useState<number>(0);
  const [duration, setDuration] = useState<number>(0);
  const [playbackRate, setPlaybackRate] = useState<number>(1.0);
  const [volume, setVolume] = useState<number>(0.85);
  const [muted, setMuted] = useState<boolean>(false);

  // Studio Drawer State
  const [showStudioDrawer, setShowStudioDrawer] = useState<boolean>(false);
  const [devices, setDevices] = useState<MediaDeviceInfo[]>([]);
  const [selectedDeviceId, setSelectedDeviceId] = useState<string>("");
  const [studioFidelity, setStudioFidelity] = useState<boolean>(true);
  const [isMonitoring, setIsMonitoring] = useState<boolean>(false);
  const [peakDb, setPeakDb] = useState<number>(-60);
  const [isClipping, setIsClipping] = useState<boolean>(false);

  // Live Recording State
  const [isRecording, setIsRecording] = useState<boolean>(false);
  const [isPaused, setIsPaused] = useState<boolean>(false);
  const [recordDuration, setRecordDuration] = useState<number>(0);
  const [recordingSessionId, setRecordingSessionId] = useState<string | null>(null);
  const [sessionTitle, setSessionTitle] = useState<string>(`Studio Session ${new Date().toLocaleDateString()}`);
  const [sessionCampaign, setSessionCampaign] = useState<string>("Chronicle / Studio");
  const [sessionDm, setSessionDm] = useState<string>("Justin");
  const [isSavingRecord, setIsSavingRecord] = useState<boolean>(false);
  const [placedMarkers, setPlacedMarkers] = useState<Array<{ id: string; timestamp_ms: number; label: string; category: string }>>([]);

  // Slicer & Marker Tools (For Transcriptless / Studio Audio)
  const [sliceStartMs, setSliceStartMs] = useState<number>(0);
  const [sliceEndMs, setSliceEndMs] = useState<number>(5000);
  const [sliceLabel, setSliceLabel] = useState<string>("Encounter Excerpt");
  const [isExtractingSlice, setIsExtractingSlice] = useState<boolean>(false);
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);
  const [transcribeQueued, setTranscribeQueued] = useState<boolean>(false);

  // Studio Audio Graph & Recording Refs
  const audioContextRef = useRef<AudioContext | null>(null);
  const analyserRef = useRef<AnalyserNode | null>(null);
  const mediaStreamRef = useRef<MediaStream | null>(null);
  const animationFrameRef = useRef<number | null>(null);
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const miniCanvasRef = useRef<HTMLCanvasElement | null>(null);
  const mediaRecorderRef = useRef<MediaRecorder | null>(null);
  const recordedChunksRef = useRef<Blob[]>([]);
  const recordTimerRef = useRef<NodeJS.Timeout | null>(null);
  const recordingSessionIdRef = useRef<string | null>(null);
  const chunkUploadPromiseRef = useRef<Promise<any>>(Promise.resolve());
  const peakCapsRef = useRef<number[]>(new Array(48).fill(0));

  // Direct DOM Refs for 60 FPS zero-overhead VU meter rendering
  const peakTextRef = useRef<HTMLSpanElement | null>(null);
  const peakBarRef = useRef<HTMLDivElement | null>(null);
  const peakBadgeRef = useRef<HTMLDivElement | null>(null);
  const floatingPeakTextRef = useRef<HTMLSpanElement | null>(null);
  const floatingPeakBarRef = useRef<HTMLDivElement | null>(null);

  // Helper notification
  const showFeedback = (type: "success" | "error", msg: string) => {
    setFeedback({ type, msg });
    setTimeout(() => setFeedback(null), 4500);
  };

  // 1. Fetch Unified Catalog
  const fetchCatalog = useCallback(async () => {
    setLoadingList(true);
    try {
      const res = await fetch("/api/v1/audio/catalog?limit=500");
      if (res.ok) {
        const data: CatalogResponse = await res.json();
        setItems(data.items || []);
        setTotalCount(data.total || 0);
        if (data.counts) {
          setCounts(data.counts);
        }
        if (!selectedItem && data.items && data.items.length > 0) {
          setSelectedItem(data.items[0]);
        }
      } else {
        showFeedback("error", `Catalog error: ${res.statusText}`);
      }
    } catch (err: any) {
      console.error("Failed to fetch audio catalog:", err);
      showFeedback("error", "Could not connect to unified audio catalog API.");
    } finally {
      setLoadingList(false);
    }
  }, [selectedItem]);

  // 2. Fetch Recorder Status
  const fetchStatus = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/recorder/status");
      if (res.ok) {
        const data: StatusOverview = await res.json();
        setStatus(data);
        setIsSyncing(data.is_syncing);
      }
    } catch (err) {
      console.warn("Could not fetch recorder status:", err);
    }
  }, []);

  useEffect(() => {
    fetchCatalog();
    fetchStatus();
    refreshDevices();
  }, []);

  // Poll sync status if active
  useEffect(() => {
    let timer: NodeJS.Timeout | null = null;
    if (isSyncing) {
      timer = setInterval(async () => {
        try {
          const res = await fetch("/api/v1/recorder/sync/status");
          if (res.ok) {
            const data: SyncStatus = await res.json();
            setSyncStatus(data);
            if (!data.is_syncing) {
              setIsSyncing(false);
              showFeedback("success", `Sync complete: ${data.total_downloaded} downloaded, ${data.total_skipped} skipped.`);
              fetchCatalog();
              fetchStatus();
            }
          }
        } catch {
          // ignore transient errors
        }
      }, 2000);
    }
    return () => {
      if (timer) clearInterval(timer);
    };
  }, [isSyncing, fetchCatalog, fetchStatus]);

  // 3. Load Transcript when selectedItem changes
  useEffect(() => {
    if (!selectedItem) {
      setTranscript(null);
      return;
    }

    setTranscribeQueued(false);

    // If item has transcript and is from recorder
    if (selectedItem.has_transcript && selectedItem.source === "recorder") {
      setLoadingTranscript(true);
      fetch(`/api/v1/recorder/recordings/${selectedItem.id}/transcript`)
        .then((res) => {
          if (res.ok) return res.json();
          throw new Error(`Failed to load transcript (${res.status})`);
        })
        .then((data: TranscriptPayload) => {
          setTranscript(data);
        })
        .catch((err) => {
          console.warn("Transcript fetch failed:", err);
          setTranscript(null);
        })
        .finally(() => {
          setLoadingTranscript(false);
        });
    } else {
      setTranscript(null);
      setLoadingTranscript(false);
    }

    // Reset slice cutter range to track start/end
    setSliceStartMs(0);
    setSliceEndMs(selectedItem.duration_ms ? Math.min(selectedItem.duration_ms, 15000) : 10000);
    setSliceLabel(`Excerpt - ${selectedItem.title.slice(0, 24)}`);
  }, [selectedItem?.id]);

  // 4. Audio Playback Event Handlers
  useEffect(() => {
    const audio = audioRef.current;
    if (!audio) return;

    const onPlay = () => setIsPlaying(true);
    const onPause = () => setIsPlaying(false);
    const onTimeUpdate = () => {
      setCurrentTime(audio.currentTime);
      if (audio.duration && !isNaN(audio.duration)) {
        setDuration(audio.duration);
      }
    };
    const onLoadedMetadata = () => {
      if (audio.duration && !isNaN(audio.duration)) {
        setDuration(audio.duration);
      }
    };
    const onEnded = () => {
      setIsPlaying(false);
      setCurrentTime(0);
    };

    audio.addEventListener("play", onPlay);
    audio.addEventListener("pause", onPause);
    audio.addEventListener("timeupdate", onTimeUpdate);
    audio.addEventListener("loadedmetadata", onLoadedMetadata);
    audio.addEventListener("ended", onEnded);

    return () => {
      audio.removeEventListener("play", onPlay);
      audio.removeEventListener("pause", onPause);
      audio.removeEventListener("timeupdate", onTimeUpdate);
      audio.removeEventListener("loadedmetadata", onLoadedMetadata);
      audio.removeEventListener("ended", onEnded);
    };
  }, [selectedItem]);

  const togglePlay = () => {
    const audio = audioRef.current;
    if (!audio) return;
    if (isPlaying) {
      audio.pause();
    } else {
      audio.play().catch((e) => console.warn("Playback prevented:", e));
    }
  };

  const seekTo = (seconds: number) => {
    const audio = audioRef.current;
    if (!audio) return;
    audio.currentTime = Math.max(0, Math.min(seconds, duration || audio.duration || 999999));
    if (!isPlaying) {
      audio.play().catch(() => {});
    }
  };

  const skipSeconds = (delta: number) => {
    const audio = audioRef.current;
    if (!audio) return;
    seekTo(audio.currentTime + delta);
  };

  const handleVolumeChange = (newVol: number) => {
    setVolume(newVol);
    if (audioRef.current) {
      audioRef.current.volume = newVol;
      if (newVol > 0 && muted) {
        setMuted(false);
        audioRef.current.muted = false;
      }
    }
  };

  const toggleMute = () => {
    if (audioRef.current) {
      const nextMute = !muted;
      setMuted(nextMute);
      audioRef.current.muted = nextMute;
    }
  };

  const handlePlaybackRateChange = (rate: number) => {
    setPlaybackRate(rate);
    if (audioRef.current) {
      audioRef.current.playbackRate = rate;
    }
  };

  // Active Paragraph Tracking & Auto-Scroll
  const currentMs = currentTime * 1000;
  const activeParagraphIndex = useMemo(() => {
    if (!transcript || !transcript.paragraphs || transcript.paragraphs.length === 0) return -1;
    for (let i = transcript.paragraphs.length - 1; i >= 0; i--) {
      const p = transcript.paragraphs[i];
      if (currentMs >= p.start_ms) {
        if (p.end_ms && currentMs > p.end_ms + 1500) continue;
        return i;
      }
    }
    return -1;
  }, [transcript, currentMs]);

  useEffect(() => {
    if (!autoScroll || activeParagraphIndex < 0) return;
    const el = document.getElementById(`speaker-para-${activeParagraphIndex}`);
    if (el && transcriptContainerRef.current) {
      el.scrollIntoView({ behavior: "smooth", block: "nearest" });
    }
  }, [activeParagraphIndex, autoScroll]);

  // Trigger Google Recorder Sync
  const triggerSync = async () => {
    if (isSyncing) return;
    try {
      setIsSyncing(true);
      const res = await fetch("/api/v1/recorder/sync", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ limit: 50, force: false }),
      });
      if (res.ok) {
        showFeedback("success", "Google Recorder sync started...");
      } else {
        const data = await res.json();
        showFeedback("error", `Sync error: ${data.error || res.statusText}`);
        setIsSyncing(false);
      }
    } catch (err: any) {
      showFeedback("error", `Sync failed: ${err.message}`);
      setIsSyncing(false);
    }
  };

  // Queue Future Transcription Placeholder
  const queueTranscription = async () => {
    if (!selectedItem) return;
    try {
      const res = await fetch("/api/v1/audio/transcribe", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ id: selectedItem.id, path: selectedItem.path }),
      });
      if (res.ok) {
        setTranscribeQueued(true);
        showFeedback("success", `Registered "${selectedItem.title}" for future Voice AI transcription!`);
      }
    } catch {
      showFeedback("error", "Could not queue transcription.");
    }
  };

  // Copy Quote with Attribution
  const copyQuote = (speaker: string, timecode: string, text: string, idx: number) => {
    const quote = `> "${text}"\n— **${speaker}** (${timecode}), *${selectedItem?.title || "Sonic Chronicle"}*`;
    navigator.clipboard.writeText(quote);
    setCopiedId(`para-${idx}`);
    showFeedback("success", "Quote copied with attribution.");
    setTimeout(() => setCopiedId(null), 2500);
  };

  // Extract Audio Slice (for session audio)
  const extractSlice = async () => {
    if (!selectedItem) return;
    setIsExtractingSlice(true);
    try {
      const res = await fetch(`/api/v1/sessions/${selectedItem.id}/slice`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          start_ms: sliceStartMs,
          end_ms: sliceEndMs,
          label: sliceLabel,
        }),
      });
      if (res.ok) {
        showFeedback("success", `Precision slice extracted: ${sliceLabel}`);
        fetchCatalog();
      } else {
        showFeedback("error", "Slice extraction is only supported on active session files.");
      }
    } catch {
      showFeedback("error", "Failed to extract slice.");
    } finally {
      setIsExtractingSlice(false);
    }
  };

  // Steinberg Audio Hardware Monitoring
  const refreshDevices = async () => {
    try {
      if (!navigator.mediaDevices?.enumerateDevices) return;
      const allDevices = await navigator.mediaDevices.enumerateDevices();
      const audioInputs = allDevices.filter((d) => d.kind === "audioinput");
      setDevices(audioInputs);
      const steinberg = audioInputs.find((d) =>
        d.label.toLowerCase().includes("steinberg") || d.label.toLowerCase().includes("ur44") || d.label.toLowerCase().includes("line")
      );
      if (steinberg) {
        setSelectedDeviceId(steinberg.deviceId);
      } else if (audioInputs.length > 0 && !selectedDeviceId) {
        setSelectedDeviceId(audioInputs[0].deviceId);
      }
    } catch (err) {
      console.warn("Could not enumerate audio devices:", err);
    }
  };

  // Robust Audio Stream Opener (with fallback if exact device constraints fail)
  const acquireAudioStream = async (devId?: string): Promise<MediaStream> => {
    const targetId = devId || selectedDeviceId;
    if (targetId) {
      try {
        return await navigator.mediaDevices.getUserMedia({
          audio: {
            deviceId: { ideal: targetId },
            echoCancellation: !studioFidelity,
            noiseSuppression: !studioFidelity,
            autoGainControl: !studioFidelity,
          },
        });
      } catch (err1) {
        console.warn("Targeted audio constraints failed, trying generic constraint:", err1);
      }
    }
    return await navigator.mediaDevices.getUserMedia({ audio: true });
  };

  const startMonitoring = async () => {
    try {
      if (isMonitoring && mediaStreamRef.current) return;
      
      const stream = await acquireAudioStream(selectedDeviceId);
      mediaStreamRef.current = stream;

      const AudioCtxClass = window.AudioContext || (window as any).webkitAudioContext;
      const audioCtx = new AudioCtxClass();
      if (audioCtx.state === "suspended") {
        await audioCtx.resume();
      }
      audioContextRef.current = audioCtx;

      const source = audioCtx.createMediaStreamSource(stream);
      const analyser = audioCtx.createAnalyser();
      analyser.fftSize = 256;
      analyser.smoothingTimeConstant = 0.8;
      source.connect(analyser);
      analyserRef.current = analyser;

      setIsMonitoring(true);
      showFeedback("success", "Line monitor opened: live signal active.");
    } catch (err: any) {
      console.error("Failed to access audio input:", err);
      showFeedback("error", `Could not open audio input: ${err.message || "Permission denied"}`);
    }
  };

  const stopMonitoring = () => {
    if (isRecording) {
      showFeedback("error", "Please stop recording before closing the monitor.");
      return;
    }
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = null;
    }
    if (mediaStreamRef.current) {
      mediaStreamRef.current.getTracks().forEach((t) => t.stop());
      mediaStreamRef.current = null;
    }
    if (audioContextRef.current) {
      audioContextRef.current.close();
      audioContextRef.current = null;
    }
    analyserRef.current = null;
    setIsMonitoring(false);
    setPeakDb(-60);
    setIsClipping(false);

    // Reset DOM VU meter indicators directly
    if (peakTextRef.current) peakTextRef.current.textContent = "-∞ dB";
    if (peakBarRef.current) peakBarRef.current.style.width = "0%";
    if (peakBadgeRef.current) {
      peakBadgeRef.current.textContent = "STANDBY";
      peakBadgeRef.current.className = "text-[10px] text-slate-500 text-right";
    }
  };

  const drawSpectralMonitor = () => {
    if (!analyserRef.current) return;
    if (animationFrameRef.current) {
      cancelAnimationFrame(animationFrameRef.current);
      animationFrameRef.current = null;
    }

    const analyser = analyserRef.current;
    const bufferLength = analyser.frequencyBinCount; // 128 bins with fftSize = 256
    const freqData = new Uint8Array(bufferLength);
    const timeData = new Uint8Array(bufferLength);
    const numBars = 40;

    const render = () => {
      animationFrameRef.current = requestAnimationFrame(render);
      if (!analyserRef.current) return;

      analyser.getByteFrequencyData(freqData);
      analyser.getByteTimeDomainData(timeData);

      // 1. Calculate Peak dB from time-domain audio samples
      let peak = 0;
      for (let i = 0; i < bufferLength; i++) {
        const val = Math.abs((timeData[i] - 128) / 128);
        if (val > peak) peak = val;
      }
      const rawDb = peak > 0 ? 20 * Math.log10(peak) : -60;
      const db = Math.max(-60, Math.min(0, Math.round(rawDb)));
      const pct = Math.min(100, Math.max(0, ((db + 60) / 60) * 100));
      const clipping = db >= -0.5;

      // 2. Direct DOM update for VU Meters (0 React re-renders = butter-smooth 60 FPS)
      if (peakTextRef.current) {
        peakTextRef.current.textContent = `${db} dB`;
        peakTextRef.current.className = clipping ? "text-rose-400 font-bold" : "text-cyan-300";
      }
      if (peakBarRef.current) {
        peakBarRef.current.style.width = `${pct}%`;
        peakBarRef.current.className = `h-full transition-all duration-75 ${
          clipping ? "bg-rose-500 shadow-[0_0_8px_#f43f5e]" : db > -12 ? "bg-amber-400" : "bg-cyan-400"
        }`;
      }
      if (peakBadgeRef.current) {
        peakBadgeRef.current.textContent = clipping ? "CLIPPING WARN" : isRecording ? "LIVE BROADCAST PCM" : "LIVE MONITOR";
        peakBadgeRef.current.className = `text-[10px] text-right font-mono ${clipping ? "text-rose-400 font-bold animate-pulse" : "text-slate-500"}`;
      }
      if (floatingPeakTextRef.current) {
        floatingPeakTextRef.current.textContent = `${db} dB`;
      }
      if (floatingPeakBarRef.current) {
        floatingPeakBarRef.current.style.width = `${pct}%`;
        floatingPeakBarRef.current.className = `h-full ${clipping ? "bg-rose-500" : db > -12 ? "bg-amber-400" : "bg-cyan-400"}`;
      }

      // 3. Render Main Canvas: Multi-Band FFT Equalizer Bars + Overlaid Waveform
      const canvas = canvasRef.current;
      if (canvas) {
        const ctx = canvas.getContext("2d");
        if (ctx) {
          ctx.fillStyle = "rgb(10, 15, 29)";
          ctx.fillRect(0, 0, canvas.width, canvas.height);

          // Render FFT Equalizer Bars with logarithmic bin distribution
          const barWidth = (canvas.width / numBars) - 2;
          for (let i = 0; i < numBars; i++) {
            const binIdx = Math.min(bufferLength - 1, Math.floor(Math.pow(i / numBars, 1.35) * bufferLength));
            const val = freqData[binIdx] / 255.0;
            const barHeight = Math.max(2, val * (canvas.height - 8));
            const x = i * (barWidth + 2) + 2;
            const y = canvas.height - barHeight;

            // Multi-stop energetic gradient from cyan through emerald to amber/rose
            const grad = ctx.createLinearGradient(0, canvas.height, 0, 0);
            grad.addColorStop(0, "#06b6d4");
            grad.addColorStop(0.5, "#10b981");
            grad.addColorStop(0.85, "#f59e0b");
            grad.addColorStop(1, "#f43f5e");
            ctx.fillStyle = grad;
            ctx.fillRect(x, y, barWidth, barHeight);

            // Peak Hold cap
            if (barHeight > (peakCapsRef.current[i] || 0)) {
              peakCapsRef.current[i] = barHeight;
            } else {
              peakCapsRef.current[i] = Math.max(0, (peakCapsRef.current[i] || 0) - 0.7);
            }
            const capY = canvas.height - (peakCapsRef.current[i] || 0) - 2;
            ctx.fillStyle = (peakCapsRef.current[i] || 0) > canvas.height - 12 ? "#f43f5e" : "#38bdf8";
            ctx.fillRect(x, capY, barWidth, 1.5);
          }

          // Luminous center waveform overlay
          ctx.beginPath();
          ctx.lineWidth = 1.2;
          ctx.strokeStyle = clipping ? "rgba(244, 63, 94, 0.7)" : "rgba(56, 189, 248, 0.45)";
          const sliceWidth = canvas.width / bufferLength;
          let wx = 0;
          for (let j = 0; j < bufferLength; j++) {
            const v = timeData[j] / 128.0;
            const wy = (v * canvas.height) / 2;
            if (j === 0) ctx.moveTo(wx, wy);
            else ctx.lineTo(wx, wy);
            wx += sliceWidth;
          }
          ctx.stroke();
        }
      }

      // 4. Render Mini Floating Canvas (if mounted)
      const miniCanvas = miniCanvasRef.current;
      if (miniCanvas) {
        const mctx = miniCanvas.getContext("2d");
        if (mctx) {
          mctx.fillStyle = "rgb(15, 23, 42)";
          mctx.fillRect(0, 0, miniCanvas.width, miniCanvas.height);
          const miniBars = 24;
          const mBarWidth = (miniCanvas.width / miniBars) - 1.5;
          for (let k = 0; k < miniBars; k++) {
            const binIdx = Math.min(bufferLength - 1, Math.floor(Math.pow(k / miniBars, 1.35) * bufferLength));
            const val = freqData[binIdx] / 255.0;
            const barH = Math.max(2, val * (miniCanvas.height - 4));
            const mx = k * (mBarWidth + 1.5) + 1;
            const my = miniCanvas.height - barH;

            mctx.fillStyle = val > 0.8 ? "#f43f5e" : val > 0.5 ? "#f59e0b" : "#06b6d4";
            mctx.fillRect(mx, my, mBarWidth, barH);
          }
        }
      }
    };
    render();
  };

  // Keep visualizer loop synchronized with monitoring / recording state
  useEffect(() => {
    if (!canvasRef.current && !miniCanvasRef.current) return;
    if (isMonitoring || isRecording) {
      drawSpectralMonitor();
    } else {
      const canvas = canvasRef.current;
      const ctx = canvas?.getContext("2d");
      if (ctx && canvas) {
        ctx.fillStyle = "rgb(10, 15, 29)";
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.lineWidth = 1.5;
        ctx.strokeStyle = "rgba(6, 182, 212, 0.4)";
        ctx.beginPath();
        ctx.moveTo(0, canvas.height / 2);
        ctx.lineTo(canvas.width, canvas.height / 2);
        ctx.stroke();
      }
    }
    return () => {
      if (animationFrameRef.current) {
        cancelAnimationFrame(animationFrameRef.current);
        animationFrameRef.current = null;
      }
    };
  }, [isMonitoring, isRecording, showStudioDrawer]);

  // Start Live Steinberg / Studio Recording
  const startRecording = async () => {
    try {
      setShowStudioDrawer(true);
      // 1. Ensure audio stream is active
      let stream = mediaStreamRef.current;
      if (!stream) {
        stream = await acquireAudioStream(selectedDeviceId);
        mediaStreamRef.current = stream;

        if (!audioContextRef.current) {
          const AudioCtxClass = window.AudioContext || (window as any).webkitAudioContext;
          const audioCtx = new AudioCtxClass();
          if (audioCtx.state === "suspended") await audioCtx.resume();
          audioContextRef.current = audioCtx;
          const source = audioCtx.createMediaStreamSource(stream);
          const analyser = audioCtx.createAnalyser();
          analyser.fftSize = 256;
          analyser.smoothingTimeConstant = 0.8;
          source.connect(analyser);
          analyserRef.current = analyser;
          setIsMonitoring(true);
        }
      }

      // 2. Initialize live session in backend
      const title = sessionTitle.trim() || `Studio Session ${new Date().toLocaleDateString()}`;
      const campaign = sessionCampaign.trim() || "Chronicle / Studio";
      const selectedDevice = devices.find((d) => d.deviceId === selectedDeviceId);
      const inputDeviceLabel = selectedDevice?.label || "Steinberg UR-44 (Line In)";

      let activeId = `session-${Date.now()}`;
      try {
        const initRes = await fetch("/api/v1/sessions", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            title,
            campaign,
            dm: sessionDm.trim() || "Justin",
            input_device: inputDeviceLabel,
            format: "audio/webm",
            sample_rate: 48000,
            channels: 2,
            bit_depth: 16,
          }),
        });
        if (initRes.ok) {
          const initData = await initRes.json();
          if (initData.id) {
            activeId = initData.id;
          }
        }
      } catch (e) {
        console.warn("Could not pre-init session on backend, will finalize on stop:", e);
      }

      setRecordingSessionId(activeId);
      recordingSessionIdRef.current = activeId;
      setPlacedMarkers([]);
      recordedChunksRef.current = [];
      chunkUploadPromiseRef.current = Promise.resolve();

      // 3. Setup MediaRecorder
      const mimeType = typeof MediaRecorder !== "undefined" && MediaRecorder.isTypeSupported("audio/webm;codecs=opus")
        ? "audio/webm;codecs=opus"
        : typeof MediaRecorder !== "undefined" && MediaRecorder.isTypeSupported("audio/webm")
        ? "audio/webm"
        : "";

      const mr = mimeType ? new MediaRecorder(stream, { mimeType }) : new MediaRecorder(stream);
      mediaRecorderRef.current = mr;

      mr.ondataavailable = async (e) => {
        if (e.data && e.data.size > 0) {
          recordedChunksRef.current.push(e.data);
          const sid = recordingSessionIdRef.current;
          if (sid) {
            const p = fetch(`/api/v1/sessions/${sid}/chunk`, {
              method: "POST",
              headers: { "Content-Type": "application/octet-stream" },
              body: e.data,
            }).catch((err) => {
              console.warn("Streaming chunk append failed:", err);
            });
            chunkUploadPromiseRef.current = chunkUploadPromiseRef.current.then(() => p);
          }
        }
      };

      mr.start(1000); // 1 second timeslices for steady streaming
      setIsRecording(true);
      setIsPaused(false);
      setRecordDuration(0);

      if (recordTimerRef.current) clearInterval(recordTimerRef.current);
      recordTimerRef.current = setInterval(() => {
        setRecordDuration((prev) => prev + 1);
      }, 1000);

      showFeedback("success", `Recording started: "${title}"`);
    } catch (err: any) {
      console.error("Failed to start recording:", err);
      showFeedback("error", `Could not start recording: ${err.message || "Microphone access denied"}`);
    }
  };

  // Toggle Pause / Resume Live Recording
  const togglePauseRecording = () => {
    const mr = mediaRecorderRef.current;
    if (!mr || !isRecording) return;
    if (isPaused) {
      mr.resume();
      setIsPaused(false);
      recordTimerRef.current = setInterval(() => {
        setRecordDuration((prev) => prev + 1);
      }, 1000);
      showFeedback("success", "Recording resumed");
    } else {
      mr.pause();
      setIsPaused(true);
      if (recordTimerRef.current) clearInterval(recordTimerRef.current);
      showFeedback("success", "Recording paused");
    }
  };

  // Stop & Ingest Live Recording Session
  const stopRecording = async () => {
    const mr = mediaRecorderRef.current;
    if (!mr || !isRecording) return;

    setIsSavingRecord(true);
    if (recordTimerRef.current) {
      clearInterval(recordTimerRef.current);
      recordTimerRef.current = null;
    }

    mr.onstop = async () => {
      try {
        const title = sessionTitle.trim() || `Studio Session ${new Date().toLocaleDateString()}`;
        const campaign = sessionCampaign.trim() || "Chronicle / Studio";
        const dm = sessionDm.trim() || "Justin";
        const sid = recordingSessionIdRef.current;

        // Ensure all in-flight streaming chunks are flushed and completed on server
        await chunkUploadPromiseRef.current;

        let completedItem: any = null;
        if (sid) {
          try {
            const compRes = await fetch(`/api/v1/sessions/${sid}/complete`, { method: "POST" });
            if (compRes.ok) {
              completedItem = await compRes.json();
            }
          } catch (e) {
            console.warn("Session completion API error:", e);
          }
        }

        // ONLY fall back to multipart upload if session was NEVER pre-initialized on backend
        if (!sid && recordedChunksRef.current.length > 0) {
          const mime = mr.mimeType || "audio/webm";
          const blob = new Blob(recordedChunksRef.current, { type: mime });
          const ext = mime.includes("wav") ? "wav" : "webm";
          const formData = new FormData();
          formData.append("audio", blob, `${title.replace(/[^a-zA-Z0-9_-]/g, "_")}.${ext}`);
          formData.append("title", title);
          formData.append("campaign", campaign);
          formData.append("dm", dm);
          formData.append("notes", `Recorded via Live Studio Capture (${formatTimecode(recordDuration)})`);

          const uploadRes = await fetch("/api/v1/sessions/upload", {
            method: "POST",
            body: formData,
          });
          if (uploadRes.ok) {
            completedItem = await uploadRes.json();
          }
        }

        showFeedback("success", `Session saved & ingested: "${title}"!`);
        setIsRecording(false);
        setIsPaused(false);
        setRecordingSessionId(null);
        recordingSessionIdRef.current = null;

        await fetchCatalog();
        if (completedItem && completedItem.id) {
          const createdItem: UnifiedAudioItem = {
            id: completedItem.id,
            title: completedItem.title || title,
            source: "session",
            category: "studio_session",
            format: completedItem.format ? completedItem.format.replace("audio/", "") : "webm",
            duration: formatSec(completedItem.duration_sec || recordDuration),
            duration_ms: Math.round((completedItem.duration_sec || recordDuration) * 1000),
            size_bytes: completedItem.size_bytes || 0,
            recorded_at: completedItem.start_time || new Date().toISOString(),
            mod_time: new Date().toISOString(),
            has_transcript: false,
            is_synced: true,
            audio_url: `/api/v1/sessions/${completedItem.id}/stream`,
            path: completedItem.file_path || `MercuryDasha/sessions/dnd/${title}`,
          };
          setSelectedItem(createdItem);
        }
      } catch (err: any) {
        console.error("Failed to finalize recording:", err);
        showFeedback("error", `Failed to save recording: ${err.message}`);
      } finally {
        setIsSavingRecord(false);
        setIsRecording(false);
        setIsPaused(false);
      }
    };

    mr.stop();
  };

  // Drop Millisecond-Accurate Encounter Marker
  const dropMarker = async (category: string, label: string) => {
    const sid = recordingSessionIdRef.current;
    const timestampMs = Math.round(recordDuration * 1000);
    const newMarker = {
      id: `marker-${Date.now()}`,
      timestamp_ms: timestampMs,
      label,
      category,
    };
    setPlacedMarkers((prev) => [...prev, newMarker]);

    if (sid) {
      try {
        await fetch(`/api/v1/sessions/${sid}/markers`, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({
            timestamp_ms: timestampMs,
            label,
            category,
            notes: `Encounter marker dropped at ${formatTimecode(recordDuration)}`,
          }),
        });
        showFeedback("success", `Marked [${label}] at ${formatTimecode(recordDuration)}`);
      } catch {
        // maintained locally
      }
    }
  };

  // Format Timecode for live recording HUD (HH:MM:SS)
  const formatTimecode = (seconds: number): string => {
    const h = Math.floor(seconds / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = seconds % 60;
    if (h > 0) {
      return `${h.toString().padStart(2, "0")}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
    }
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  // Filtered Items Calculation
  const filteredItems = useMemo(() => {
    return items.filter((item) => {
      // Source filter
      if (sourceFilter === "transcript" && !item.has_transcript) return false;
      if (sourceFilter === "vault" && item.source !== "vault") return false;
      if (sourceFilter === "session" && item.source !== "session") return false;

      // Format filter
      if (formatFilter !== "all" && item.format !== formatFilter) return false;

      // Search Query
      if (searchQuery.trim()) {
        const q = searchQuery.toLowerCase();
        const matchTitle = item.title.toLowerCase().includes(q);
        const matchPath = item.path.toLowerCase().includes(q);
        const matchFormat = item.format.toLowerCase().includes(q);
        if (!matchTitle && !matchPath && !matchFormat) return false;
      }

      return true;
    });
  }, [items, sourceFilter, formatFilter, searchQuery]);

  // Format Milliseconds helper
  const formatMs = (ms: number): string => {
    const totalSec = Math.floor(ms / 1000);
    const m = Math.floor(totalSec / 60);
    const s = totalSec % 60;
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  const formatSec = (sec: number): string => {
    if (isNaN(sec)) return "00:00";
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  const formatBytes = (bytes: number): string => {
    if (!bytes || bytes === 0) return "0 B";
    if (bytes < 1024) return `${bytes} B`;
    if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
    return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
  };

  return (
    <div className="space-y-4">
      {/* ── TOP BANNER & FEEDBACK ────────────────────────────────────────── */}
      {feedback && (
        <div
          className={`flex items-center justify-between px-4 py-3 rounded-xl border text-sm animate-in fade-in slide-in-from-top-2 transition shadow-lg ${
            feedback.type === "success"
              ? "bg-emerald-950/80 border-emerald-500/40 text-emerald-200"
              : "bg-rose-950/80 border-rose-500/40 text-rose-200"
          }`}
        >
          <div className="flex items-center space-x-2">
            {feedback.type === "success" ? <CheckCircle className="w-4 h-4 text-emerald-400" /> : <AlertCircle className="w-4 h-4 text-rose-400" />}
            <span>{feedback.msg}</span>
          </div>
          <button onClick={() => setFeedback(null)} className="text-xs opacity-70 hover:opacity-100">
            Dismiss
          </button>
        </div>
      )}

      {/* ── UNIFIED AUDIO HEADER & STATS BAR ──────────────────────────────── */}
      <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-5 shadow-xl relative overflow-hidden backdrop-blur-md">
        <div className="absolute top-0 right-0 w-96 h-96 bg-cyan-500/5 rounded-full blur-3xl pointer-events-none" />

        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div className="flex items-center space-x-3.5">
            <span className="p-3 rounded-xl bg-cyan-500/10 border border-cyan-500/20 text-cyan-400 shadow-inner">
              <Headphones className="w-6 h-6" />
            </span>
            <div>
              <h2 className="text-xl font-bold text-white tracking-tight flex items-center space-x-2.5">
                <span>Sonic Chronicle Portal</span>
                <span className="text-xs px-2.5 py-0.5 rounded-full bg-cyan-500/20 text-cyan-300 font-mono border border-cyan-500/30">
                  UNIFIED STOREHOUSE
                </span>
                {status?.mcp_connected && (
                  <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-300 font-mono flex items-center space-x-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    <span>RECORDER MCP LINKED</span>
                  </span>
                )}
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                Sovereign Dropbox Audio Vault • Diarized Google Transcripts • Lossless Steinberg PCM • Universal Multi-Format Playback
              </p>
            </div>
          </div>

          {/* Quick Metrics & Studio Drawer Toggle */}
          <div className="flex flex-wrap items-center gap-2.5">
            <div className="flex items-center space-x-2 bg-slate-800/80 border border-slate-700/80 px-3 py-1.5 rounded-xl text-xs font-mono">
              <HardDrive className="w-3.5 h-3.5 text-cyan-400" />
              <span className="text-slate-400">Total Assets:</span>
              <span className="text-cyan-300 font-semibold">{totalCount}</span>
              <span className="text-slate-600">•</span>
              <span className="text-emerald-400 font-semibold">{counts.with_transcript} with Diarization</span>
            </div>

            <button
              onClick={triggerSync}
              disabled={isSyncing}
              className={`px-3 py-1.5 rounded-xl border text-xs font-medium transition flex items-center space-x-1.5 ${
                isSyncing
                  ? "bg-cyan-950/60 border-cyan-500/40 text-cyan-300 cursor-not-allowed"
                  : "bg-slate-800 hover:bg-slate-700 border-slate-700 text-slate-200"
              }`}
              title="Sync Google Recorder recordings & transcripts"
            >
              <RefreshCw className={`w-3.5 h-3.5 text-cyan-400 ${isSyncing ? "animate-spin" : ""}`} />
              <span>{isSyncing ? "Syncing..." : "Sync Recorder"}</span>
            </button>

            <button
              onClick={() => setShowStudioDrawer(!showStudioDrawer)}
              className={`px-3 py-1.5 rounded-xl border text-xs font-medium transition flex items-center space-x-1.5 shadow-sm ${
                showStudioDrawer
                  ? "bg-cyan-600 border-cyan-500 text-white"
                  : "bg-slate-800/90 hover:bg-slate-700 border-slate-700 text-slate-300"
              }`}
            >
              <Mic className="w-3.5 h-3.5" />
              <span>Live Studio Capture</span>
              {showStudioDrawer ? <ChevronUp className="w-3.5 h-3.5 ml-0.5" /> : <ChevronDown className="w-3.5 h-3.5 ml-0.5" />}
            </button>
          </div>
        </div>

        {/* Sync Progress Bar (Active during background sync) */}
        {isSyncing && (
          <div className="mt-4 pt-3 border-t border-slate-800/80 flex items-center justify-between text-xs text-cyan-300">
            <div className="flex items-center space-x-2">
              <span className="w-2 h-2 rounded-full bg-cyan-400 animate-ping" />
              <span>Synchronizing Google Recorder into <code className="text-slate-300">/Dropbox/audio/recorder/</code></span>
              {syncStatus?.current_item && (
                <span className="text-slate-400 truncate max-w-xs font-mono">[{syncStatus.current_item}]</span>
              )}
            </div>
            <div className="font-mono text-slate-400">
              Downloaded: {syncStatus?.total_downloaded || 0} • Skipped: {syncStatus?.total_skipped || 0}
            </div>
          </div>
        )}

        {/* ── COLLAPSIBLE STEINBERG STUDIO CAPTURE DRAWER ──────────────────── */}
        {showStudioDrawer && (
          <div className="mt-4 pt-4 border-t border-slate-800 space-y-4 animate-in fade-in slide-in-from-top-2">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 bg-slate-950/60 p-3.5 rounded-xl border border-slate-800">
              <div className="flex items-center space-x-3">
                <Sliders className="w-4 h-4 text-cyan-400" />
                <span className="text-xs font-semibold text-white">Line Input Source:</span>
                <select
                  value={selectedDeviceId}
                  onChange={(e) => setSelectedDeviceId(e.target.value)}
                  className="bg-slate-900 border border-slate-700 rounded-lg text-xs px-2.5 py-1 text-slate-200 focus:outline-none focus:border-cyan-500"
                >
                  {devices.length === 0 && <option value="">Default Audio Input</option>}
                  {devices.map((d) => (
                    <option key={d.deviceId} value={d.deviceId}>
                      {d.label || `Line Input ${d.deviceId.slice(0, 8)}`}
                    </option>
                  ))}
                </select>
                <button onClick={refreshDevices} title="Refresh Devices" className="text-slate-400 hover:text-cyan-300">
                  <RefreshCw className="w-3.5 h-3.5" />
                </button>
              </div>

              <div className="flex items-center space-x-2.5">
                <button
                  onClick={() => setStudioFidelity(!studioFidelity)}
                  className={`px-2.5 py-1 rounded-lg border text-xs font-mono transition flex items-center space-x-1 ${
                    studioFidelity
                      ? "bg-emerald-950/50 border-emerald-500/40 text-emerald-300"
                      : "bg-slate-800 border-slate-700 text-slate-400"
                  }`}
                >
                  <ShieldCheck className="w-3 h-3" />
                  <span>{studioFidelity ? "RAW PCM (NO AGC)" : "SPEECH FILTERED"}</span>
                </button>

                {!isMonitoring ? (
                  <button
                    onClick={startMonitoring}
                    className="px-3 py-1 rounded-lg bg-cyan-600 hover:bg-cyan-500 text-white text-xs font-medium transition flex items-center space-x-1 shadow"
                  >
                    <Activity className="w-3.5 h-3.5" />
                    <span>Open Line Monitor</span>
                  </button>
                ) : (
                  <button
                    onClick={stopMonitoring}
                    className="px-3 py-1 rounded-lg bg-rose-600/80 hover:bg-rose-600 text-white text-xs font-medium transition flex items-center space-x-1 shadow"
                  >
                    <Activity className="w-3.5 h-3.5" />
                    <span>Close Monitor</span>
                  </button>
                )}
              </div>
            </div>

            {/* FFT Spectral Frequency Monitor & Real-Time Waveform */}
            <div className={`p-3.5 rounded-xl border transition-all ${
              isRecording
                ? "bg-slate-950 border-rose-500/50 shadow-lg shadow-rose-950/30 ring-1 ring-rose-500/30"
                : isMonitoring
                ? "bg-slate-950 border-cyan-500/40 shadow-md ring-1 ring-cyan-500/20"
                : "bg-slate-950/80 border-slate-800 opacity-85"
            }`}>
              <div className="flex items-center justify-between text-[11px] font-mono text-slate-400 mb-2 px-1">
                <span className="flex items-center gap-1.5 font-semibold text-cyan-300">
                  <Activity className="w-3.5 h-3.5 text-cyan-400 animate-pulse" />
                  FFT Spectral Monitor & Live Oscilloscope
                </span>
                <span className="text-[11px] text-slate-400">48 kHz • 256 Bins • Multi-Band EQ</span>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-4 gap-3">
                <div className="md:col-span-3 relative">
                  <canvas ref={canvasRef} width={640} height={75} className="w-full h-18 rounded-lg border border-slate-800 bg-slate-950 shadow-inner" />
                  {!isMonitoring && !isRecording && (
                    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                      <span className="text-[11px] font-mono text-slate-400 bg-slate-950/90 px-3 py-1 rounded border border-slate-800 shadow">
                        Spectral Monitor Standby • Click "Open Line Monitor" or "Record Live Session"
                      </span>
                    </div>
                  )}
                </div>
                <div className="flex flex-col justify-center space-y-1.5 p-2.5 rounded-lg bg-slate-900 border border-slate-800 text-xs font-mono">
                  <div className="flex justify-between">
                    <span className="text-slate-400">Peak Level:</span>
                    <span ref={peakTextRef} className={isClipping ? "text-rose-400 font-bold" : isMonitoring || isRecording ? "text-cyan-300" : "text-slate-500"}>
                      {isMonitoring || isRecording ? `${peakDb} dB` : "-∞ dB"}
                    </span>
                  </div>
                  <div className="w-full bg-slate-800 h-2.5 rounded-full overflow-hidden">
                    <div
                      ref={peakBarRef}
                      className={`h-full transition-all duration-75 ${
                        isClipping ? "bg-rose-500" : peakDb > -12 ? "bg-amber-400" : "bg-cyan-400"
                      }`}
                      style={{ width: `${isMonitoring || isRecording ? Math.min(100, Math.max(0, ((peakDb + 60) / 60) * 100)) : 0}%` }}
                    />
                  </div>
                  <div ref={peakBadgeRef} className="text-[10px] text-slate-500 text-right">
                    {isClipping ? "CLIPPING WARN" : isRecording ? "LIVE BROADCAST PCM" : isMonitoring ? "LIVE MONITOR" : "STANDBY"}
                  </div>
                </div>
              </div>
            </div>

            {/* ── RECORDING CONTROLS & SESSION METADATA BAR ──────────────── */}
            <div className={`p-4 rounded-xl border transition-all ${
              isRecording
                ? "bg-rose-950/20 border-rose-500/40 shadow-lg shadow-rose-950/30 ring-1 ring-rose-500/30"
                : "bg-slate-950/60 border-slate-800"
            } space-y-3.5`}>
              {/* Row 1: Session Metadata Inputs */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3 text-xs">
                <div>
                  <label className="text-[10px] font-mono text-slate-400 block mb-1">
                    SESSION TITLE
                  </label>
                  <input
                    type="text"
                    value={sessionTitle}
                    onChange={(e) => setSessionTitle(e.target.value)}
                    disabled={isRecording}
                    placeholder="e.g. Amber Temple Delve"
                    className="w-full bg-slate-900 border border-slate-700/80 rounded-lg px-2.5 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 disabled:opacity-60"
                  />
                </div>

                <div>
                  <label className="text-[10px] font-mono text-slate-400 block mb-1">
                    CAMPAIGN / REALM
                  </label>
                  <input
                    type="text"
                    value={sessionCampaign}
                    onChange={(e) => setSessionCampaign(e.target.value)}
                    disabled={isRecording}
                    placeholder="e.g. Curse of Strahd"
                    className="w-full bg-slate-900 border border-slate-700/80 rounded-lg px-2.5 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 disabled:opacity-60"
                  />
                </div>

                <div>
                  <label className="text-[10px] font-mono text-slate-400 block mb-1">
                    DM / SPEAKER
                  </label>
                  <input
                    type="text"
                    value={sessionDm}
                    onChange={(e) => setSessionDm(e.target.value)}
                    disabled={isRecording}
                    placeholder="e.g. Justin"
                    className="w-full bg-slate-900 border border-slate-700/80 rounded-lg px-2.5 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500 disabled:opacity-60"
                  />
                </div>
              </div>

              {/* Row 2: Primary Recording Transport & Live Status Bar */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 pt-2 border-t border-slate-800/80">
                <div className="flex items-center space-x-3">
                  {!isRecording ? (
                    <button
                      onClick={startRecording}
                      disabled={isSavingRecord}
                      className="px-4 py-2 rounded-xl bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold transition flex items-center space-x-2 shadow-lg shadow-rose-900/40 hover:scale-[1.02] active:scale-[0.98]"
                    >
                      <Circle className="w-4 h-4 fill-current text-white animate-pulse" />
                      <span>Record Live Session</span>
                    </button>
                  ) : (
                    <div className="flex items-center space-x-2">
                      <button
                        onClick={togglePauseRecording}
                        className={`px-3 py-1.5 rounded-xl border text-xs font-semibold transition flex items-center space-x-1.5 shadow ${
                          isPaused
                            ? "bg-amber-600 hover:bg-amber-500 border-amber-500 text-white"
                            : "bg-slate-800 hover:bg-slate-700 border-slate-700 text-slate-200"
                        }`}
                      >
                        {isPaused ? (
                          <>
                            <Play className="w-3.5 h-3.5 fill-current" />
                            <span>Resume</span>
                          </>
                        ) : (
                          <>
                            <Pause className="w-3.5 h-3.5 fill-current" />
                            <span>Pause</span>
                          </>
                        )}
                      </button>

                      <button
                        onClick={stopRecording}
                        disabled={isSavingRecord}
                        className="px-3.5 py-1.5 rounded-xl bg-slate-800 hover:bg-rose-600 border border-slate-700 hover:border-rose-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow"
                      >
                        {isSavingRecord ? (
                          <>
                            <RefreshCw className="w-3.5 h-3.5 animate-spin text-cyan-400" />
                            <span>Ingesting to Storehouse...</span>
                          </>
                        ) : (
                          <>
                            <Square className="w-3.5 h-3.5 fill-current text-rose-400" />
                            <span>Stop & Ingest</span>
                          </>
                        )}
                      </button>
                    </div>
                  )}

                  {/* Live Recording HUD Badge */}
                  {isRecording && (
                    <div className="flex items-center space-x-2 bg-slate-900/90 border border-rose-500/40 px-3 py-1.5 rounded-xl">
                      <span className={`w-2.5 h-2.5 rounded-full ${isPaused ? "bg-amber-400" : "bg-rose-500 animate-ping"}`} />
                      <span className="text-xs font-mono font-bold text-rose-300">
                        {isPaused ? "PAUSED" : "REC"} {formatTimecode(recordDuration)}
                      </span>
                    </div>
                  )}
                </div>

                <div className="flex items-center space-x-2 text-[11px] font-mono text-slate-400">
                  <span className="px-2 py-0.5 rounded bg-slate-900 border border-slate-800 text-slate-300">
                    48kHz Linear PCM / Opus Stereo
                  </span>
                  <span className="text-slate-500">•</span>
                  <span className="text-slate-400">
                    {isRecording ? `${placedMarkers.length} Markers Stamped` : "Live Line-In Standby"}
                  </span>
                </div>
              </div>

              {/* Row 3: Live Encounter Marker Toolbar (Active during recording) */}
              {isRecording && (
                <div className="pt-3 border-t border-slate-800/80 space-y-2 animate-in fade-in slide-in-from-top-1">
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] font-mono uppercase tracking-wider text-cyan-400 flex items-center space-x-1">
                      <Bookmark className="w-3 h-3" />
                      <span>Live Encounter Bookmarks (Stamp Millisecond Timecode)</span>
                    </span>
                    <span className="text-[10px] font-mono text-slate-500">
                      Auto-ingested into BoltDB session ledger
                    </span>
                  </div>

                  <div className="flex flex-wrap items-center gap-1.5">
                    <button
                      onClick={() => dropMarker("combat", "Combat Encounter")}
                      className="px-2.5 py-1 rounded-lg bg-rose-950/60 hover:bg-rose-900 border border-rose-500/40 text-rose-300 text-xs font-medium transition flex items-center space-x-1"
                    >
                      <Swords className="w-3 h-3 text-rose-400" />
                      <span>Combat</span>
                    </button>

                    <button
                      onClick={() => dropMarker("roleplay", "Roleplay Dialogue")}
                      className="px-2.5 py-1 rounded-lg bg-purple-950/60 hover:bg-purple-900 border border-purple-500/40 text-purple-300 text-xs font-medium transition flex items-center space-x-1"
                    >
                      <FileText className="w-3 h-3 text-purple-400" />
                      <span>Roleplay</span>
                    </button>

                    <button
                      onClick={() => dropMarker("lore", "Lore & Exposition")}
                      className="px-2.5 py-1 rounded-lg bg-cyan-950/60 hover:bg-cyan-900 border border-cyan-500/40 text-cyan-300 text-xs font-medium transition flex items-center space-x-1"
                    >
                      <Scroll className="w-3 h-3 text-cyan-400" />
                      <span>Lore</span>
                    </button>

                    <button
                      onClick={() => dropMarker("crit", "Critical Roll / Triumph")}
                      className="px-2.5 py-1 rounded-lg bg-amber-950/60 hover:bg-amber-900 border border-amber-500/40 text-amber-300 text-xs font-medium transition flex items-center space-x-1"
                    >
                      <Zap className="w-3 h-3 text-amber-400" />
                      <span>Crit</span>
                    </button>

                    <button
                      onClick={() => dropMarker("loot", "Loot & Discovery")}
                      className="px-2.5 py-1 rounded-lg bg-emerald-950/60 hover:bg-emerald-900 border border-emerald-500/40 text-emerald-300 text-xs font-medium transition flex items-center space-x-1"
                    >
                      <Sparkles className="w-3 h-3 text-emerald-400" />
                      <span>Loot</span>
                    </button>

                    <button
                      onClick={() => dropMarker("bookmark", "General Bookmark")}
                      className="px-2.5 py-1 rounded-lg bg-slate-800 hover:bg-slate-700 border border-slate-700 text-slate-300 text-xs font-medium transition flex items-center space-x-1"
                    >
                      <Tag className="w-3 h-3 text-slate-400" />
                      <span>Custom Marker</span>
                    </button>
                  </div>

                  {/* Placed Markers Strip */}
                  {placedMarkers.length > 0 && (
                    <div className="flex items-center space-x-2 pt-1.5 overflow-x-auto scrollbar-thin">
                      <span className="text-[10px] font-mono text-slate-500 shrink-0">STAMPED:</span>
                      {placedMarkers.map((m) => (
                        <span
                          key={m.id}
                          className="px-2 py-0.5 rounded-full bg-slate-900 border border-slate-800 text-[10px] font-mono text-cyan-300 whitespace-nowrap shrink-0 flex items-center space-x-1"
                        >
                          <span className="text-slate-500">{formatMs(m.timestamp_ms)}</span>
                          <span className="font-semibold text-slate-200">{m.label}</span>
                        </span>
                      ))}
                    </div>
                  )}
                </div>
              )}
            </div>
          </div>
        )}
      </div>

      {/* ── MAIN UNIFIED WORKSPACE (EXPLORER + ADAPTIVE RIGHT WORKSPACE) ────── */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-5">
        {/* ── LEFT COLUMN: UNIFIED AUDIO EXPLORER (5 COLS) ────────────────── */}
        <div className="lg:col-span-5 space-y-3 flex flex-col h-[740px]">
          {/* Search & Filter Header */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-3.5 shadow space-y-3">
            {/* Global Search Bar */}
            <div className="relative">
              <Search className="w-4 h-4 absolute left-3 top-2.5 text-slate-400 pointer-events-none" />
              <input
                type="text"
                placeholder="Search titles, filenames, formats..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="w-full bg-slate-950 border border-slate-700/80 rounded-lg pl-9 pr-3 py-1.5 text-xs text-white placeholder-slate-500 focus:outline-none focus:border-cyan-500"
              />
            </div>

            {/* Quick Source Filters */}
            <div className="flex items-center space-x-1 text-xs overflow-x-auto pb-1 scrollbar-none">
              <button
                onClick={() => setSourceFilter("all")}
                className={`px-2.5 py-1 rounded-lg font-medium transition whitespace-nowrap ${
                  sourceFilter === "all"
                    ? "bg-cyan-600 text-white shadow-sm"
                    : "bg-slate-800 hover:bg-slate-750 text-slate-300"
                }`}
              >
                All Audio ({counts.total})
              </button>

              <button
                onClick={() => setSourceFilter("transcript")}
                className={`px-2.5 py-1 rounded-lg font-medium transition whitespace-nowrap flex items-center space-x-1 ${
                  sourceFilter === "transcript"
                    ? "bg-emerald-600 text-white shadow-sm"
                    : "bg-slate-800 hover:bg-slate-750 text-slate-300"
                }`}
              >
                <FileText className="w-3 h-3" />
                <span>With Transcripts ({counts.with_transcript})</span>
              </button>

              <button
                onClick={() => setSourceFilter("vault")}
                className={`px-2.5 py-1 rounded-lg font-medium transition whitespace-nowrap ${
                  sourceFilter === "vault"
                    ? "bg-amber-600 text-white shadow-sm"
                    : "bg-slate-800 hover:bg-slate-750 text-slate-300"
                }`}
              >
                Vault ({counts.vault})
              </button>

              <button
                onClick={() => setSourceFilter("session")}
                className={`px-2.5 py-1 rounded-lg font-medium transition whitespace-nowrap ${
                  sourceFilter === "session"
                    ? "bg-purple-600 text-white shadow-sm"
                    : "bg-slate-800 hover:bg-slate-750 text-slate-300"
                }`}
              >
                Studio Slices ({counts.sessions})
              </button>
            </div>

            {/* Format Pills */}
            <div className="flex items-center space-x-1.5 pt-1 border-t border-slate-800/80 text-[11px] font-mono overflow-x-auto scrollbar-none">
              <span className="text-slate-500 uppercase tracking-wider text-[10px]">Format:</span>
              {["all", "m4a", "wav", "mp3", "amr", "flac"].map((fmt) => (
                <button
                  key={fmt}
                  onClick={() => setFormatFilter(fmt)}
                  className={`px-2 py-0.5 rounded uppercase transition ${
                    formatFilter === fmt
                      ? "bg-slate-700 text-cyan-300 font-bold border border-cyan-500/30"
                      : "text-slate-400 hover:text-slate-200"
                  }`}
                >
                  {fmt === "amr" ? "AMR (WAV)" : fmt}
                </button>
              ))}
            </div>
          </div>

          {/* Unified Track List */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-2 flex-1 overflow-y-auto space-y-1.5 shadow">
            {loadingList ? (
              <div className="flex flex-col items-center justify-center h-48 space-y-2 text-slate-400 text-xs">
                <RefreshCw className="w-5 h-5 animate-spin text-cyan-400" />
                <span>Loading unified audio assets...</span>
              </div>
            ) : filteredItems.length === 0 ? (
              <div className="text-center py-12 text-slate-500 text-xs">
                <FolderTree className="w-8 h-8 mx-auto mb-2 opacity-40 text-slate-400" />
                <p>No audio files match your filters.</p>
              </div>
            ) : (
              filteredItems.map((item) => {
                const isSelected = selectedItem?.id === item.id;
                const isCurrentPlaying = isSelected && isPlaying;

                return (
                  <div
                    key={item.id}
                    onClick={() => setSelectedItem(item)}
                    className={`p-2.5 rounded-xl border transition cursor-pointer group relative ${
                      isSelected
                        ? "bg-cyan-950/40 border-cyan-500/50 shadow-md ring-1 ring-cyan-500/30"
                        : "bg-slate-950/60 hover:bg-slate-800/60 border-slate-800/80"
                    }`}
                  >
                    <div className="flex items-start justify-between gap-2">
                      <div className="flex items-start space-x-2.5 min-w-0">
                        {/* Play / Format Icon Indicator */}
                        <div
                          className={`mt-0.5 p-2 rounded-lg transition shrink-0 ${
                            isSelected
                              ? "bg-cyan-500 text-slate-950"
                              : "bg-slate-800 text-cyan-400 group-hover:bg-slate-700"
                          }`}
                        >
                          {isCurrentPlaying ? (
                            <Activity className="w-3.5 h-3.5 animate-pulse" />
                          ) : (
                            <FileAudio className="w-3.5 h-3.5" />
                          )}
                        </div>

                        <div className="min-w-0">
                          <h4 className="text-xs font-semibold text-slate-100 truncate group-hover:text-cyan-200">
                            {item.title}
                          </h4>

                          <div className="flex flex-wrap items-center gap-1.5 mt-1 text-[10px] font-mono">
                            {/* Source Badge */}
                            <span
                              className={`px-1.5 py-0.2 rounded font-medium ${
                                item.source === "recorder"
                                  ? "bg-emerald-950 text-emerald-300 border border-emerald-500/30"
                                  : item.source === "session"
                                  ? "bg-purple-950 text-purple-300 border border-purple-500/30"
                                  : "bg-amber-950 text-amber-300 border border-amber-500/30"
                              }`}
                            >
                              {item.source === "recorder" ? "RECORDER" : item.source === "session" ? "STUDIO" : "VAULT"}
                            </span>

                            {/* Format Pill */}
                            <span className="px-1.5 py-0.2 rounded bg-slate-800 text-slate-300 uppercase">
                              {item.format === "amr" ? "AMR -> WAV" : item.format}
                            </span>

                            {/* Duration */}
                            {item.duration && (
                              <span className="text-slate-400 flex items-center space-x-0.5">
                                <Clock className="w-2.5 h-2.5" />
                                <span>{item.duration}</span>
                              </span>
                            )}

                            {/* Transcript Status Indicator */}
                            {item.has_transcript ? (
                              <span className="px-1.5 py-0.2 rounded bg-emerald-500/20 text-emerald-300 font-sans font-medium flex items-center space-x-0.5">
                                <FileText className="w-2.5 h-2.5" />
                                <span>Diarized</span>
                              </span>
                            ) : (
                              <span className="text-slate-500">Audio Only</span>
                            )}
                          </div>
                        </div>
                      </div>

                      {/* File Size */}
                      <span className="text-[10px] font-mono text-slate-500 shrink-0">
                        {formatBytes(item.size_bytes)}
                      </span>
                    </div>
                  </div>
                );
              })
            )}
          </div>
        </div>

        {/* ── RIGHT COLUMN: ADAPTIVE PLAYBACK & INSPECTOR WORKSPACE (7 COLS) ─── */}
        <div className="lg:col-span-7 flex flex-col h-[740px] space-y-3">
          {/* Universal Native Audio Element (Hidden) */}
          <audio
            ref={audioRef}
            src={selectedItem ? selectedItem.audio_url : ""}
            preload="metadata"
          />

          {/* ── UNIVERSAL AUDIO PLAYER BAR ─────────────────────────────────── */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-4 shadow-xl space-y-3">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
              {/* Track Title & Metadata */}
              <div className="min-w-0">
                <div className="flex items-center space-x-2">
                  <span
                    className={`w-2 h-2 rounded-full ${
                      isPlaying ? "bg-emerald-400 animate-ping" : "bg-slate-500"
                    }`}
                  />
                  <h3 className="text-sm font-bold text-white truncate max-w-md">
                    {selectedItem ? selectedItem.title : "No Track Selected"}
                  </h3>
                </div>
                <p className="text-[11px] font-mono text-slate-400 mt-0.5 truncate">
                  {selectedItem ? selectedItem.path : "Select an audio asset from the explorer"}
                </p>
              </div>

              {/* Playback Rate & Volume Controls */}
              <div className="flex items-center space-x-3 shrink-0">
                {/* Speed Multiplier */}
                <div className="flex items-center space-x-1 bg-slate-950 px-2 py-1 rounded-lg border border-slate-800 text-xs font-mono">
                  {[0.75, 1.0, 1.25, 1.5, 2.0].map((rate) => (
                    <button
                      key={rate}
                      onClick={() => handlePlaybackRateChange(rate)}
                      className={`px-1.5 py-0.5 rounded transition ${
                        playbackRate === rate
                          ? "bg-cyan-600 text-white font-bold"
                          : "text-slate-400 hover:text-slate-200"
                      }`}
                    >
                      {rate}x
                    </button>
                  ))}
                </div>

                {/* Volume Slider */}
                <div className="flex items-center space-x-1.5">
                  <button onClick={toggleMute} className="text-slate-400 hover:text-cyan-300">
                    {muted || volume === 0 ? (
                      <VolumeX className="w-4 h-4 text-rose-400" />
                    ) : volume < 0.5 ? (
                      <Volume1 className="w-4 h-4" />
                    ) : (
                      <Volume2 className="w-4 h-4" />
                    )}
                  </button>
                  <input
                    type="range"
                    min="0"
                    max="1"
                    step="0.05"
                    value={muted ? 0 : volume}
                    onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
                    className="w-16 h-1.5 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-cyan-500"
                  />
                </div>
              </div>
            </div>

            {/* Scrubber Slider */}
            <div className="space-y-1">
              <input
                type="range"
                min="0"
                max={duration || 100}
                step="0.1"
                value={currentTime}
                onChange={(e) => seekTo(parseFloat(e.target.value))}
                className="w-full h-2 bg-slate-950 rounded-lg appearance-none cursor-pointer accent-cyan-400 hover:accent-cyan-300 transition"
              />
              <div className="flex justify-between text-[11px] font-mono text-slate-400">
                <span>{formatSec(currentTime)}</span>
                <span className="text-slate-500">
                  {selectedItem?.format ? `HTTP 206 [${selectedItem.format.toUpperCase()}]` : "STREAM"}
                </span>
                <span>{formatSec(duration)}</span>
              </div>
            </div>

            {/* Transport Control Buttons */}
            <div className="flex items-center justify-center space-x-4 pt-1">
              <button
                onClick={() => skipSeconds(-5)}
                className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition"
                title="Rewind 5 seconds"
              >
                <RotateCcw className="w-4 h-4" />
              </button>

              <button
                onClick={togglePlay}
                disabled={!selectedItem}
                className={`p-3.5 rounded-full transition shadow-lg ${
                  isPlaying
                    ? "bg-cyan-500 hover:bg-cyan-400 text-slate-950"
                    : "bg-cyan-600 hover:bg-cyan-500 text-white"
                } ${!selectedItem ? "opacity-50 cursor-not-allowed" : ""}`}
              >
                {isPlaying ? <Pause className="w-5 h-5 fill-current" /> : <Play className="w-5 h-5 fill-current ml-0.5" />}
              </button>

              <button
                onClick={() => skipSeconds(5)}
                className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 hover:text-white transition"
                title="Forward 5 seconds"
              >
                <RotateCw className="w-4 h-4" />
              </button>
            </div>
          </div>

          {/* ── ADAPTIVE WORKSPACE CONTENT ─────────────────────────────────── */}
          <div className="bg-slate-900/90 border border-slate-800 rounded-xl p-4 flex-1 flex flex-col overflow-hidden shadow">
            {selectedItem?.has_transcript && transcript ? (
              /* ── SCENARIO A: KARAOKE DIARIZED TRANSCRIPT ─────────────────── */
              <div className="flex flex-col h-full space-y-3">
                {/* Transcript Header & Tools */}
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 pb-3 border-b border-slate-800">
                  <div className="flex items-center space-x-2">
                    <FileText className="w-4 h-4 text-emerald-400" />
                    <span className="text-xs font-bold text-white uppercase tracking-wider">
                      Diarized Speaker Transcript
                    </span>
                    <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-300 border border-emerald-500/20">
                      {transcript.paragraphs?.length || 0} Blocks
                    </span>
                  </div>

                  <div className="flex items-center space-x-2">
                    {/* Transcript Search */}
                    <div className="relative">
                      <Search className="w-3 h-3 absolute left-2 top-2 text-slate-400" />
                      <input
                        type="text"
                        placeholder="Filter text..."
                        value={transcriptSearch}
                        onChange={(e) => setTranscriptSearch(e.target.value)}
                        className="bg-slate-950 border border-slate-700 rounded-md pl-6 pr-2 py-0.5 text-xs text-white focus:outline-none focus:border-cyan-500 w-32"
                      />
                    </div>

                    <button
                      onClick={() => setAutoScroll(!autoScroll)}
                      className={`px-2 py-1 rounded text-[11px] font-mono transition ${
                        autoScroll
                          ? "bg-cyan-950 text-cyan-300 border border-cyan-500/30"
                          : "text-slate-500 hover:text-slate-300"
                      }`}
                      title="Auto-scroll to current speaker block"
                    >
                      {autoScroll ? "Auto-Scroll ON" : "Scroll OFF"}
                    </button>
                  </div>
                </div>

                {/* Diarized Paragraphs Karaoke List */}
                <div
                  ref={transcriptContainerRef}
                  className="flex-1 overflow-y-auto space-y-3 pr-2 scrollbar-thin scrollbar-thumb-slate-700"
                >
                  {transcript.paragraphs && transcript.paragraphs.length > 0 ? (
                    transcript.paragraphs.map((p, idx) => {
                      const isActive = idx === activeParagraphIndex;
                      const timecode = formatMs(p.start_ms);
                      const isCopied = copiedId === `para-${idx}`;

                      if (transcriptSearch.trim()) {
                        const match = p.text.toLowerCase().includes(transcriptSearch.toLowerCase());
                        if (!match) return null;
                      }

                      return (
                        <div
                          key={idx}
                          id={`speaker-para-${idx}`}
                          className={`p-3.5 rounded-xl border transition group ${
                            isActive
                              ? "bg-cyan-950/60 border-cyan-400/80 shadow-lg ring-1 ring-cyan-400/40"
                              : "bg-slate-950/60 hover:bg-slate-800/40 border-slate-800/80"
                          }`}
                        >
                          <div className="flex items-center justify-between mb-1.5">
                            {/* Click-to-Seek Timestamp */}
                            <button
                              onClick={() => seekTo(p.start_ms / 1000)}
                              className="flex items-center space-x-1.5 text-xs font-mono text-cyan-400 hover:text-cyan-200 transition group/btn"
                              title="Jump player to this point"
                            >
                              <Play className="w-3 h-3 group-hover/btn:scale-110 transition" />
                              <span className="font-semibold underline decoration-dotted">{timecode}</span>
                              <span className="text-slate-400 font-sans font-medium text-xs">• {p.speaker}</span>
                            </button>

                            {/* Quote to Story Button */}
                            <button
                              onClick={() => copyQuote(p.speaker, timecode, p.text, idx)}
                              className="text-[11px] text-slate-500 hover:text-cyan-300 transition flex items-center space-x-1 opacity-0 group-hover:opacity-100"
                              title="Copy quote with markdown attribution"
                            >
                              {isCopied ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3" />}
                              <span>{isCopied ? "Copied" : "Quote"}</span>
                            </button>
                          </div>

                          <p
                            onClick={() => seekTo(p.start_ms / 1000)}
                            className={`text-xs leading-relaxed cursor-pointer transition ${
                              isActive ? "text-slate-100 font-medium" : "text-slate-300 hover:text-white"
                            }`}
                          >
                            {p.text}
                          </p>
                        </div>
                      );
                    })
                  ) : (
                    <div className="text-xs text-slate-400 p-4">
                      <p className="font-mono">{transcript.raw_text}</p>
                    </div>
                  )}
                </div>
              </div>
            ) : loadingTranscript ? (
              <div className="flex flex-col items-center justify-center flex-1 space-y-2 text-slate-400 text-xs">
                <RefreshCw className="w-6 h-6 animate-spin text-cyan-400" />
                <span>Loading diarized transcript...</span>
              </div>
            ) : (
              /* ── SCENARIO B: TRANSCRIPTLESS AUDIO INSPECTOR & SLICER ─────── */
              <div className="flex flex-col h-full justify-between space-y-4 overflow-y-auto pr-1">
                <div className="space-y-4">
                  {/* File Metadata Card */}
                  <div className="bg-slate-950/70 border border-slate-800 rounded-xl p-4 space-y-3">
                    <div className="flex items-center space-x-2 text-xs font-bold text-white uppercase tracking-wider">
                      <Info className="w-4 h-4 text-cyan-400" />
                      <span>Audio Asset Specifications</span>
                    </div>

                    <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs font-mono">
                      <div className="p-2.5 rounded-lg bg-slate-900 border border-slate-800/80">
                        <span className="text-slate-500 block text-[10px]">FORMAT</span>
                        <span className="text-cyan-300 font-semibold uppercase">
                          {selectedItem ? selectedItem.format : "N/A"}
                        </span>
                      </div>

                      <div className="p-2.5 rounded-lg bg-slate-900 border border-slate-800/80">
                        <span className="text-slate-500 block text-[10px]">FILE SIZE</span>
                        <span className="text-slate-200">
                          {selectedItem ? formatBytes(selectedItem.size_bytes) : "0 B"}
                        </span>
                      </div>

                      <div className="p-2.5 rounded-lg bg-slate-900 border border-slate-800/80">
                        <span className="text-slate-500 block text-[10px]">SOURCE TYPE</span>
                        <span className="text-amber-300 uppercase">
                          {selectedItem ? selectedItem.source : "VAULT"}
                        </span>
                      </div>

                      <div className="p-2.5 rounded-lg bg-slate-900 border border-slate-800/80">
                        <span className="text-slate-500 block text-[10px]">DURATION</span>
                        <span className="text-slate-200">
                          {selectedItem?.duration || formatSec(duration)}
                        </span>
                      </div>
                    </div>

                    <div className="p-2.5 rounded-lg bg-slate-900 border border-slate-800 text-[11px] font-mono text-slate-400 break-all">
                      <span className="text-slate-500 block text-[10px] mb-0.5">CANONICAL DROPBOX PATH</span>
                      {selectedItem?.path || "N/A"}
                    </div>
                  </div>

                  {/* Future Transcription Affordance Card */}
                  <div className="bg-gradient-to-r from-cyan-950/40 to-slate-950 border border-cyan-500/30 rounded-xl p-4 space-y-2.5 relative overflow-hidden">
                    <div className="flex items-center space-x-2 text-xs font-bold text-cyan-300">
                      <Wand2 className="w-4 h-4 text-cyan-400" />
                      <span>Voice AI Transcription Pipeline (Ready for On-Demand)</span>
                    </div>
                    <p className="text-xs text-slate-300 leading-relaxed">
                      This audio file does not have a synchronized transcript yet. You can listen with full range seeking, or queue this asset for background transcription.
                    </p>
                    <button
                      onClick={queueTranscription}
                      disabled={transcribeQueued}
                      className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold transition flex items-center space-x-1.5 shadow ${
                        transcribeQueued
                          ? "bg-emerald-950/80 border border-emerald-500/50 text-emerald-300"
                          : "bg-cyan-600 hover:bg-cyan-500 text-white"
                      }`}
                    >
                      {transcribeQueued ? (
                        <>
                          <Check className="w-3.5 h-3.5" />
                          <span>Queued for Future Voice AI</span>
                        </>
                      ) : (
                        <>
                          <Sparkles className="w-3.5 h-3.5" />
                          <span>Queue for On-Demand Transcription</span>
                        </>
                      )}
                    </button>
                  </div>

                  {/* Precision Audio Slicer */}
                  <div className="bg-slate-950/70 border border-slate-800 rounded-xl p-4 space-y-3">
                    <div className="flex items-center space-x-2 text-xs font-bold text-white uppercase tracking-wider">
                      <Scissors className="w-4 h-4 text-amber-400" />
                      <span>Precision Audio Slicer</span>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-3 text-xs">
                      <div>
                        <label className="text-[10px] font-mono text-slate-400 block mb-1">
                          START TIME (MS):
                        </label>
                        <div className="flex items-center space-x-1">
                          <input
                            type="number"
                            value={sliceStartMs}
                            onChange={(e) => setSliceStartMs(parseInt(e.target.value) || 0)}
                            className="bg-slate-900 border border-slate-700 rounded-lg px-2 py-1 text-white font-mono w-full"
                          />
                          <button
                            onClick={() => setSliceStartMs(Math.round(currentTime * 1000))}
                            className="px-2 py-1 bg-slate-800 hover:bg-slate-700 text-cyan-300 rounded text-[10px] font-mono whitespace-nowrap"
                            title="Set start to playhead"
                          >
                            Mark
                          </button>
                        </div>
                      </div>

                      <div>
                        <label className="text-[10px] font-mono text-slate-400 block mb-1">
                          END TIME (MS):
                        </label>
                        <div className="flex items-center space-x-1">
                          <input
                            type="number"
                            value={sliceEndMs}
                            onChange={(e) => setSliceEndMs(parseInt(e.target.value) || 0)}
                            className="bg-slate-900 border border-slate-700 rounded-lg px-2 py-1 text-white font-mono w-full"
                          />
                          <button
                            onClick={() => setSliceEndMs(Math.round(currentTime * 1000))}
                            className="px-2 py-1 bg-slate-800 hover:bg-slate-700 text-cyan-300 rounded text-[10px] font-mono whitespace-nowrap"
                            title="Set end to playhead"
                          >
                            Mark
                          </button>
                        </div>
                      </div>

                      <div>
                        <label className="text-[10px] font-mono text-slate-400 block mb-1">
                          EXCERPT LABEL:
                        </label>
                        <input
                          type="text"
                          value={sliceLabel}
                          onChange={(e) => setSliceLabel(e.target.value)}
                          className="bg-slate-900 border border-slate-700 rounded-lg px-2 py-1 text-white text-xs w-full"
                        />
                      </div>
                    </div>

                    <div className="flex justify-end pt-1">
                      <button
                        onClick={extractSlice}
                        disabled={isExtractingSlice || !selectedItem}
                        className="px-3.5 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white text-xs font-semibold transition flex items-center space-x-1.5 shadow"
                      >
                        <Scissors className="w-3.5 h-3.5" />
                        <span>{isExtractingSlice ? "Extracting..." : "Cut & Save Excerpt"}</span>
                      </button>
                    </div>
                  </div>
                </div>

                <div className="p-3 rounded-lg bg-slate-950/40 border border-slate-800 text-[11px] text-slate-500 flex items-center justify-between">
                  <span>AMR on-the-fly transcoding active • 48kHz Linear PCM stream</span>
                  <span className="font-mono">{selectedItem?.id}</span>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* ── FLOATING LIVE RECORDING & SPECTRAL MONITOR WINDOW ──────────────── */}
      {isRecording && (
        <aside aria-label="Live Recording Stream Monitor" className="fixed bottom-6 right-6 z-50 w-84 sm:w-96 bg-slate-950/95 border border-rose-500/80 rounded-2xl p-4 shadow-2xl shadow-rose-950/70 backdrop-blur-xl animate-in slide-in-from-bottom-5">
          <div className="flex items-center justify-between mb-2.5">
            <div className="flex items-center space-x-2">
              <span className="w-2.5 h-2.5 rounded-full bg-rose-500 animate-ping" />
              <span className="text-xs font-bold text-rose-400 tracking-wider font-mono">REC LIVE STREAM</span>
            </div>
            <span className="text-xs font-mono font-bold text-white bg-rose-950/90 px-2.5 py-0.5 rounded border border-rose-500/50">
              {formatTimecode(recordDuration)}
            </span>
          </div>

          {/* Live Mini FFT Spectral Equalizer */}
          <div className="relative mb-2.5">
            <canvas ref={miniCanvasRef} width={350} height={44} className="w-full h-11 rounded-lg bg-slate-900 border border-slate-800 shadow-inner" />
          </div>

          {/* Peak Level VU Bar */}
          <div className="space-y-1 mb-3 bg-slate-900/90 p-2 rounded-lg border border-slate-800">
            <div className="flex justify-between text-[11px] font-mono">
              <span className="text-slate-400 truncate max-w-[180px]">{sessionTitle || "Studio Session"}</span>
              <span ref={floatingPeakTextRef} className="text-cyan-300 font-bold">
                {peakDb} dB
              </span>
            </div>
            <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden">
              <div
                ref={floatingPeakBarRef}
                className="h-full bg-cyan-400 transition-all duration-75"
                style={{ width: `${Math.min(100, Math.max(0, ((peakDb + 60) / 60) * 100))}%` }}
              />
            </div>
          </div>

          {/* Floating Transport Actions */}
          <div className="flex items-center justify-between gap-2 pt-1 border-t border-slate-800/80">
            <button
              onClick={togglePauseRecording}
              className={`px-3 py-1.5 rounded-lg text-xs font-semibold border flex items-center space-x-1.5 transition shadow-sm ${
                isPaused ? "bg-amber-600 border-amber-500 text-white" : "bg-slate-800 border-slate-700 text-slate-200 hover:bg-slate-700"
              }`}
            >
              {isPaused ? <Play className="w-3.5 h-3.5 fill-current" /> : <Pause className="w-3.5 h-3.5 fill-current" />}
              <span>{isPaused ? "Resume" : "Pause"}</span>
            </button>

            <button
              onClick={() => dropMarker("combat", "Combat Bookmark")}
              className="px-2.5 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 border border-slate-700 text-amber-300 text-xs font-semibold flex items-center space-x-1 transition shadow-sm"
              title="Quick Combat Bookmark"
            >
              <Swords className="w-3.5 h-3.5" />
              <span>Mark</span>
            </button>

            <button
              onClick={stopRecording}
              disabled={isSavingRecord}
              className="px-3.5 py-1.5 rounded-lg bg-rose-600 hover:bg-rose-500 text-white text-xs font-bold transition flex items-center space-x-1.5 shadow-md shadow-rose-950 hover:scale-[1.02] active:scale-[0.98]"
            >
              <Square className="w-3.5 h-3.5 fill-current" />
              <span>{isSavingRecord ? "Saving..." : "Stop & Ingest"}</span>
            </button>
          </div>
        </aside>
      )}
    </div>
  );
}
