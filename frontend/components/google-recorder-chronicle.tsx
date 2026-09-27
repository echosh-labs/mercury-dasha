"use client";

import React, { useState, useEffect, useRef, useCallback } from "react";
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
  ExternalLink,
  Copy,
  Check,
  Sparkles,
  Bookmark,
  Share2,
  Sliders,
  ChevronRight,
  AlertCircle,
  FolderTree,
  Headphones,
  FileAudio,
  Hash,
  Filter
} from "lucide-react";

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

interface Recording {
  id: string;
  title: string;
  recorded_at: string;
  duration: string;
  duration_ms: number;
  location?: string;
  has_transcript: boolean;
  is_synced?: boolean;
  audio_path?: string;
  transcript_txt?: string;
  transcript_json?: string;
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

export default function GoogleRecorderChronicle() {
  const [status, setStatus] = useState<StatusOverview | null>(null);
  const [syncStatus, setSyncStatus] = useState<SyncStatus | null>(null);
  const [recordings, setRecordings] = useState<Recording[]>([]);
  const [selectedRec, setSelectedRec] = useState<Recording | null>(null);
  const [transcript, setTranscript] = useState<TranscriptPayload | null>(null);
  const [transcriptSearch, setTranscriptSearch] = useState<string>("");

  // Loading States
  const [loadingList, setLoadingList] = useState<boolean>(false);
  const [loadingTranscript, setLoadingTranscript] = useState<boolean>(false);
  const [isSyncing, setIsSyncing] = useState<boolean>(false);

  // Search & Filter
  const [searchQuery, setSearchQuery] = useState<string>("");
  const [filterMode, setFilterMode] = useState<"all" | "synced" | "pending">("all");
  const [syncLimit, setSyncLimit] = useState<number>(10);
  const [forceSync, setForceSync] = useState<boolean>(false);

  // Audio Playback State
  const audioRef = useRef<HTMLAudioElement | null>(null);
  const transcriptContainerRef = useRef<HTMLDivElement | null>(null);
  const [isPlaying, setIsPlaying] = useState<boolean>(false);
  const [currentTime, setCurrentTime] = useState<number>(0);
  const [duration, setDuration] = useState<number>(0);
  const [playbackRate, setPlaybackRate] = useState<number>(1.0);
  const [volume, setVolume] = useState<number>(0.85);
  const [muted, setMuted] = useState<boolean>(false);
  const [autoScroll, setAutoScroll] = useState<boolean>(true);

  // Feedback notifications
  const [feedback, setFeedback] = useState<{ type: "success" | "error"; msg: string } | null>(null);
  const [copiedId, setCopiedId] = useState<string | null>(null);

  // 1. Fetch MCP Server Health & Sync Status
  const fetchStatus = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/recorder/status");
      if (res.ok) {
        const data = await res.json();
        setStatus(data);
        if (data.is_syncing) {
          setIsSyncing(true);
        }
      }
    } catch (err) {
      console.warn("Failed to fetch recorder status:", err);
    }
  }, []);

  // 2. Fetch Recordings Catalog
  const fetchRecordings = useCallback(async () => {
    setLoadingList(true);
    try {
      const params = new URLSearchParams();
      if (searchQuery.trim()) {
        params.set("q", searchQuery.trim());
      }
      const res = await fetch(`/api/v1/recorder/recordings?${params.toString()}`);
      if (res.ok) {
        const data = await res.json();
        const list: Recording[] = data.recordings || [];
        setRecordings(list);
        if (!selectedRec && list.length > 0) {
          setSelectedRec(list[0]);
        }
      }
    } catch (err) {
      console.error("Failed to load recordings:", err);
    } finally {
      setLoadingList(false);
    }
  }, [searchQuery, selectedRec]);

  // Initial Boot
  useEffect(() => {
    fetchStatus();
    fetchRecordings();
    const interval = setInterval(fetchStatus, 6000);
    return () => clearInterval(interval);
  }, [fetchStatus, fetchRecordings]);

  // Load Transcript whenever active recording changes
  useEffect(() => {
    if (!selectedRec) return;

    let isMounted = true;
    setLoadingTranscript(true);
    setTranscript(null);

    fetch(`/api/v1/recorder/recordings/${encodeURIComponent(selectedRec.id)}/transcript`)
      .then((res) => {
        if (!res.ok) throw new Error("Transcript not found");
        return res.json();
      })
      .then((data: TranscriptPayload) => {
        if (isMounted) {
          setTranscript(data);
        }
      })
      .catch(() => {
        if (isMounted) {
          setTranscript(null);
        }
      })
      .finally(() => {
        if (isMounted) {
          setLoadingTranscript(false);
        }
      });

    return () => {
      isMounted = false;
    };
  }, [selectedRec?.id]);

  // Reset audio playback on recording selection
  useEffect(() => {
    if (audioRef.current) {
      audioRef.current.pause();
      setIsPlaying(false);
      setCurrentTime(0);
      audioRef.current.currentTime = 0;
      audioRef.current.load();
    }
  }, [selectedRec?.id]);

  // Trigger On-Demand Sync
  const triggerSync = async (all = false) => {
    setIsSyncing(true);
    setFeedback(null);
    try {
      const payload = {
        all: all,
        limit: all ? 0 : syncLimit,
        force: forceSync,
        download_audio: true,
        download_transcript: true,
        transcript_format: "both",
        download_metadata: true,
      };

      const res = await fetch("/api/v1/recorder/sync", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload),
      });

      if (res.ok) {
        setFeedback({
          type: "success",
          msg: all ? "Full library synchronization triggered successfully!" : `Synchronized latest ${syncLimit} recordings to Dropbox!`,
        });
        await fetchRecordings();
        await fetchStatus();
      } else {
        const errData = await res.json();
        setFeedback({
          type: "error",
          msg: errData.error || "Failed to trigger sync",
        });
      }
    } catch (err: any) {
      setFeedback({
        type: "error",
        msg: err.message || "Network error syncing with MCP server",
      });
    } finally {
      setIsSyncing(false);
    }
  };

  // Audio Playback Controls
  const togglePlayPause = () => {
    if (!audioRef.current) return;
    if (isPlaying) {
      audioRef.current.pause();
      setIsPlaying(false);
    } else {
      audioRef.current
        .play()
        .then(() => setIsPlaying(true))
        .catch((err) => console.warn("Playback error:", err));
    }
  };

  const seekRelative = (seconds: number) => {
    if (!audioRef.current) return;
    audioRef.current.currentTime = Math.max(0, Math.min(duration, audioRef.current.currentTime + seconds));
  };

  const seekToMs = (ms: number) => {
    if (!audioRef.current) return;
    audioRef.current.currentTime = ms / 1000;
    if (!isPlaying) {
      audioRef.current
        .play()
        .then(() => setIsPlaying(true))
        .catch(() => {});
    }
  };

  const handleRateChange = (rate: number) => {
    setPlaybackRate(rate);
    if (audioRef.current) {
      audioRef.current.playbackRate = rate;
    }
  };

  const handleVolumeChange = (vol: number) => {
    const clamped = Math.max(0, Math.min(1, vol));
    setVolume(clamped);
    if (clamped > 0 && muted) setMuted(false);
    if (audioRef.current) {
      audioRef.current.volume = clamped;
      audioRef.current.muted = false;
    }
  };

  const toggleMute = () => {
    const next = !muted;
    setMuted(next);
    if (audioRef.current) {
      audioRef.current.muted = next;
    }
  };

  // Format Helper: Milliseconds or Seconds -> "MM:SS" / "HH:MM:SS"
  const formatTime = (secs: number) => {
    if (isNaN(secs) || secs < 0) return "00:00";
    const h = Math.floor(secs / 3600);
    const m = Math.floor((secs % 3600) / 60);
    const s = Math.floor(secs % 60);
    if (h > 0) {
      return `${h}:${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
    }
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  const formatDate = (isoStr: string) => {
    if (!isoStr) return "";
    try {
      const d = new Date(isoStr);
      return d.toLocaleDateString("en-US", {
        month: "short",
        day: "numeric",
        year: "numeric",
        hour: "numeric",
        minute: "2-digit",
      });
    } catch {
      return isoStr;
    }
  };

  // Find Active Transcript Paragraph for Karaoke Highlighting
  const currentMs = currentTime * 1000;
  const activeParagraphIndex = transcript?.paragraphs?.findIndex((p, idx, arr) => {
    const next = arr[idx + 1];
    const end = p.end_ms && p.end_ms > p.start_ms ? p.end_ms : (next ? next.start_ms : p.start_ms + 15000);
    return currentMs >= p.start_ms && currentMs < end;
  }) ?? -1;

  // Filtered recordings
  const filteredRecordings = recordings.filter((r) => {
    if (filterMode === "synced" && !r.is_synced) return false;
    if (filterMode === "pending" && r.is_synced) return false;
    if (searchQuery.trim()) {
      const q = searchQuery.toLowerCase();
      return r.title.toLowerCase().includes(q) || r.location?.toLowerCase().includes(q);
    }
    return true;
  });

  // Filtered transcript paragraphs
  const filteredParagraphs = transcript?.paragraphs?.filter((p) => {
    if (!transcriptSearch.trim()) return true;
    const q = transcriptSearch.toLowerCase();
    return p.text.toLowerCase().includes(q) || p.speaker.toLowerCase().includes(q);
  }) || [];

  // Copy Quote Helper
  const copyQuote = (p: DiarizedParagraph) => {
    if (!selectedRec) return;
    const quote = `"${p.text}"\n— ${p.speaker} [${formatTime(p.start_ms / 1000)}] in "${selectedRec.title}"`;
    navigator.clipboard.writeText(quote);
    setCopiedId(`quote-${p.start_ms}`);
    setTimeout(() => setCopiedId(null), 2500);
  };

  const getAudioUrl = (rec: Recording) => {
    return `/api/v1/recorder/recordings/${encodeURIComponent(rec.id)}/audio`;
  };

  return (
    <div className="space-y-6">
      {/* 1. Header Banner & MCP Synchronizer Controls */}
      <div className="bg-gradient-to-r from-slate-900 via-[#101726] to-slate-900 p-6 rounded-2xl border border-slate-800 shadow-2xl">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-6">
          <div className="space-y-2">
            <div className="flex items-center gap-3">
              <div className="p-2.5 rounded-xl bg-rose-500/10 text-rose-400 border border-rose-500/20 shadow-lg shadow-rose-500/10">
                <Mic className="w-6 h-6" />
              </div>
              <div>
                <div className="flex items-center gap-2.5 flex-wrap">
                  <h2 className="text-xl font-bold text-white tracking-tight">
                    Google Recorder Sovereign Chronicle
                  </h2>
                  <span
                    className={`text-xs px-2.5 py-0.5 rounded-full font-mono flex items-center gap-1.5 border ${
                      status?.mcp_connected
                        ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                        : "bg-red-500/10 text-red-400 border-red-500/30"
                    }`}
                  >
                    <span className={`w-2 h-2 rounded-full ${status?.mcp_connected ? "bg-emerald-400 animate-pulse" : "bg-red-400"}`} />
                    <span>{status?.mcp_connected ? "MCP CONNECTED (Port 8091)" : "MCP DISCONNECTED"}</span>
                  </span>

                  {status?.auth_valid && (
                    <span className="text-xs px-2 py-0.5 rounded-full font-mono bg-cyan-500/10 text-cyan-300 border border-cyan-500/30 flex items-center gap-1">
                      <ShieldCheck className="w-3 h-3 text-cyan-400" />
                      <span>AUTH VALID</span>
                    </span>
                  )}
                </div>
                <p className="text-xs text-slate-400 mt-1">
                  Synchronizing Google recordings &amp; diarized transcripts directly into{" "}
                  <code className="text-rose-300 font-mono bg-rose-950/40 px-1.5 py-0.5 rounded border border-rose-800/40">
                    {status?.output_dir || "/home/justin/Dropbox/audio/recorder"}
                  </code>
                </p>
              </div>
            </div>
          </div>

          {/* Sync Actions & Auto-Cadence */}
          <div className="flex flex-wrap items-center gap-3">
            <div className="flex items-center gap-1.5 bg-slate-950/70 border border-slate-800 px-3 py-1.5 rounded-xl text-xs font-mono text-slate-400">
              <Clock className="w-3.5 h-3.5 text-amber-400" />
              <span>Auto-Sync: <span className="text-amber-300 font-semibold">{status?.auto_sync_mins || 30}m cycle</span></span>
            </div>

            <div className="flex items-center gap-1 bg-slate-950/70 border border-slate-800 px-2.5 py-1.5 rounded-xl text-xs font-mono text-slate-300">
              <span className="text-slate-500">Batch:</span>
              {[5, 10, 25].map((n) => (
                <button
                  key={n}
                  onClick={() => setSyncLimit(n)}
                  className={`px-1.5 py-0.5 rounded text-[11px] transition ${
                    syncLimit === n ? "bg-rose-600 text-white font-semibold" : "hover:text-white"
                  }`}
                >
                  {n}
                </button>
              ))}
            </div>

            <button
              onClick={() => triggerSync(false)}
              disabled={isSyncing || !status?.mcp_connected}
              className={`flex items-center gap-2 px-4 py-2 rounded-xl text-xs font-semibold font-mono transition shadow-lg ${
                isSyncing
                  ? "bg-rose-500/20 text-rose-300 border border-rose-500/30 cursor-not-allowed animate-pulse"
                  : "bg-rose-600 hover:bg-rose-500 text-white shadow-rose-900/30 active:scale-95"
              }`}
            >
              <RefreshCw className={`w-3.5 h-3.5 ${isSyncing ? "animate-spin" : ""}`} />
              <span>{isSyncing ? "Syncing MCP Archive..." : `Sync Recent (${syncLimit})`}</span>
            </button>

            <button
              onClick={() => triggerSync(true)}
              disabled={isSyncing || !status?.mcp_connected}
              className="flex items-center gap-1.5 px-3 py-2 rounded-xl text-xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition"
              title="Synchronize entire historical recording library"
            >
              <Download className="w-3.5 h-3.5 text-slate-400" />
              <span>Sync All</span>
            </button>
          </div>
        </div>

        {/* Feedback Alert */}
        {feedback && (
          <div
            className={`mt-4 p-3 rounded-xl text-xs font-mono flex items-center justify-between border ${
              feedback.type === "success"
                ? "bg-emerald-950/40 text-emerald-300 border-emerald-800/60"
                : "bg-red-950/40 text-red-300 border-red-800/60"
            }`}
          >
            <div className="flex items-center gap-2">
              {feedback.type === "success" ? <CheckCircle className="w-4 h-4 text-emerald-400" /> : <AlertCircle className="w-4 h-4 text-red-400" />}
              <span>{feedback.msg}</span>
            </div>
            <button onClick={() => setFeedback(null)} className="text-slate-400 hover:text-white">
              ✕
            </button>
          </div>
        )}
      </div>

      {/* 2. Main Studio Split Grid: Recordings Explorer (Left) & Player / Diarized Transcript (Right) */}
      <div className="grid grid-cols-1 lg:grid-cols-12 gap-6 items-start">
        {/* Left Column: Recordings List (5 cols) */}
        <div className="lg:col-span-5 space-y-4">
          <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-4 space-y-3 shadow-xl">
            {/* Search and Filters */}
            <div className="flex items-center gap-2">
              <div className="relative flex-1">
                <Search className="w-3.5 h-3.5 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                <input
                  type="text"
                  placeholder="Search recordings, locations..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full bg-slate-950 border border-slate-800 rounded-xl pl-8 pr-3 py-2 text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:border-rose-500/50 transition font-mono"
                />
              </div>

              <div className="flex items-center gap-1 bg-slate-950 border border-slate-800 p-1 rounded-xl text-[11px] font-mono">
                <button
                  onClick={() => setFilterMode("all")}
                  className={`px-2 py-1 rounded-lg transition ${
                    filterMode === "all" ? "bg-rose-600 text-white font-semibold" : "text-slate-400 hover:text-white"
                  }`}
                >
                  All ({recordings.length})
                </button>
                <button
                  onClick={() => setFilterMode("synced")}
                  className={`px-2 py-1 rounded-lg transition ${
                    filterMode === "synced" ? "bg-emerald-600 text-white font-semibold" : "text-slate-400 hover:text-white"
                  }`}
                >
                  Local
                </button>
              </div>
            </div>

            {/* Recordings List */}
            <div className="space-y-2 max-h-[620px] overflow-y-auto pr-1">
              {loadingList ? (
                <div className="p-8 text-center text-xs font-mono text-slate-500 flex items-center justify-center gap-2">
                  <RefreshCw className="w-4 h-4 animate-spin text-rose-400" />
                  <span>Loading recording catalog...</span>
                </div>
              ) : filteredRecordings.length === 0 ? (
                <div className="p-8 text-center text-xs font-mono text-slate-500 space-y-2">
                  <Mic className="w-6 h-6 text-slate-600 mx-auto" />
                  <p>No recordings found matching filter.</p>
                  <p className="text-[11px] text-slate-600">Click &ldquo;Sync Recent&rdquo; to pull files from Google Recorder.</p>
                </div>
              ) : (
                filteredRecordings.map((rec) => {
                  const isSelected = selectedRec?.id === rec.id;
                  return (
                    <div
                      key={rec.id}
                      onClick={() => setSelectedRec(rec)}
                      className={`p-3 rounded-xl border transition-all cursor-pointer group ${
                        isSelected
                          ? "bg-gradient-to-r from-rose-950/40 via-slate-900 to-slate-900 border-rose-500/40 shadow-lg shadow-rose-950/20"
                          : "bg-slate-950/60 border-slate-800/80 hover:border-slate-700 hover:bg-slate-900/60"
                      }`}
                    >
                      <div className="flex items-start justify-between gap-2">
                        <div className="space-y-1 min-w-0">
                          <h4 className="text-xs font-semibold text-slate-200 truncate group-hover:text-rose-300 transition">
                            {rec.title}
                          </h4>
                          <div className="flex items-center gap-2 text-[11px] font-mono text-slate-400">
                            <span className="flex items-center gap-1">
                              <Calendar className="w-3 h-3 text-slate-500" />
                              <span>{formatDate(rec.recorded_at)}</span>
                            </span>
                          </div>
                        </div>

                        <div className="flex flex-col items-end gap-1.5 flex-shrink-0">
                          <span className="text-[11px] font-mono font-semibold px-2 py-0.5 rounded-md bg-slate-800 text-slate-300 border border-slate-700">
                            {rec.duration || "00:00"}
                          </span>

                          <div className="flex items-center gap-1">
                            {rec.is_synced ? (
                              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-emerald-500/10 text-emerald-400 border border-emerald-500/30 flex items-center gap-0.5">
                                <Check className="w-2.5 h-2.5" />
                                <span>M4A</span>
                              </span>
                            ) : (
                              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/30">
                                Cloud
                              </span>
                            )}

                            {rec.has_transcript && (
                              <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-cyan-500/10 text-cyan-300 border border-cyan-500/30 flex items-center gap-0.5">
                                <FileText className="w-2.5 h-2.5" />
                                <span>Text</span>
                              </span>
                            )}
                          </div>
                        </div>
                      </div>
                    </div>
                  );
                })
              )}
            </div>
          </div>
        </div>

        {/* Right Column: Audio Player & Synchronized Diarized Transcript (7 cols) */}
        <div className="lg:col-span-7 space-y-4">
          {selectedRec ? (
            <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-5 space-y-5 shadow-2xl">
              {/* Recording Metadata Header */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800/80 pb-4">
                <div className="space-y-1">
                  <div className="flex items-center gap-2">
                    <h3 className="text-base font-bold text-white tracking-tight">
                      {selectedRec.title}
                    </h3>
                    <span className="text-[11px] font-mono px-2 py-0.5 rounded-full bg-rose-500/10 text-rose-400 border border-rose-500/20">
                      {selectedRec.duration}
                    </span>
                  </div>
                  <div className="flex items-center gap-3 text-xs font-mono text-slate-400">
                    <span>Recorded: {formatDate(selectedRec.recorded_at)}</span>
                    {selectedRec.location && <span>• {selectedRec.location}</span>}
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <a
                    href={getAudioUrl(selectedRec)}
                    download={`${selectedRec.title}.m4a`}
                    className="flex items-center gap-1.5 px-3 py-1.5 rounded-xl text-xs font-mono bg-slate-800 hover:bg-slate-700 text-slate-200 border border-slate-700 transition"
                    title="Download audio stream"
                  >
                    <Download className="w-3.5 h-3.5 text-slate-400" />
                    <span>Download</span>
                  </a>
                </div>
              </div>

              {/* High-Performance Audio Player Bar */}
              <div className="p-4 rounded-xl bg-slate-950/80 border border-slate-800/80 space-y-3">
                <audio
                  ref={audioRef}
                  src={getAudioUrl(selectedRec)}
                  preload="metadata"
                  onTimeUpdate={(e) => setCurrentTime(e.currentTarget.currentTime)}
                  onDurationChange={(e) => setDuration(e.currentTarget.duration)}
                  onEnded={() => setIsPlaying(false)}
                  onPlay={() => setIsPlaying(true)}
                  onPause={() => setIsPlaying(false)}
                />

                {/* Scrubber Progress Bar */}
                <div className="space-y-1">
                  <input
                    type="range"
                    min={0}
                    max={duration || 100}
                    step={0.1}
                    value={currentTime}
                    onChange={(e) => {
                      const val = parseFloat(e.target.value);
                      setCurrentTime(val);
                      if (audioRef.current) audioRef.current.currentTime = val;
                    }}
                    className="w-full h-2 bg-slate-800 rounded-lg appearance-none cursor-pointer accent-rose-500 hover:accent-rose-400 transition"
                  />
                  <div className="flex items-center justify-between text-[11px] font-mono text-slate-400">
                    <span>{formatTime(currentTime)}</span>
                    <span>{formatTime(duration)}</span>
                  </div>
                </div>

                {/* Player Transport Controls */}
                <div className="flex flex-wrap items-center justify-between gap-4 pt-1">
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => seekRelative(-5)}
                      className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
                      title="Skip back 5s"
                    >
                      <RotateCcw className="w-3.5 h-3.5" />
                    </button>

                    <button
                      onClick={togglePlayPause}
                      className="p-3 rounded-xl bg-rose-600 hover:bg-rose-500 text-white font-semibold transition shadow-lg shadow-rose-950/50 active:scale-95"
                      title={isPlaying ? "Pause" : "Play"}
                    >
                      {isPlaying ? <Pause className="w-4 h-4 fill-white" /> : <Play className="w-4 h-4 fill-white ml-0.5" />}
                    </button>

                    <button
                      onClick={() => seekRelative(5)}
                      className="p-2 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
                      title="Skip forward 5s"
                    >
                      <RotateCw className="w-3.5 h-3.5" />
                    </button>
                  </div>

                  {/* Playback Rate Multipliers */}
                  <div className="flex items-center gap-1 bg-slate-900 border border-slate-800 p-1 rounded-xl text-[11px] font-mono">
                    {[0.75, 1.0, 1.25, 1.5, 2.0].map((rate) => (
                      <button
                        key={rate}
                        onClick={() => handleRateChange(rate)}
                        className={`px-1.5 py-0.5 rounded transition ${
                          playbackRate === rate ? "bg-rose-600 text-white font-semibold" : "text-slate-400 hover:text-white"
                        }`}
                      >
                        {rate}x
                      </button>
                    ))}
                  </div>

                  {/* Volume Slider */}
                  <div className="flex items-center gap-2">
                    <button onClick={toggleMute} className="text-slate-400 hover:text-white transition">
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
                      min={0}
                      max={1}
                      step={0.05}
                      value={muted ? 0 : volume}
                      onChange={(e) => handleVolumeChange(parseFloat(e.target.value))}
                      className="w-20 h-1.5 bg-slate-800 rounded appearance-none cursor-pointer accent-rose-500"
                    />
                  </div>
                </div>
              </div>

              {/* Synchronized Diarized Transcript Panel */}
              <div className="space-y-3">
                <div className="flex items-center justify-between gap-3">
                  <div className="flex items-center gap-2">
                    <FileText className="w-4 h-4 text-cyan-400" />
                    <h4 className="text-xs font-bold text-slate-200 uppercase tracking-wider font-mono">
                      Synchronized Diarized Transcript
                    </h4>
                    {transcript?.paragraphs && (
                      <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-cyan-950/60 text-cyan-300 border border-cyan-800/60">
                        {transcript.paragraphs.length} blocks
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => setAutoScroll(!autoScroll)}
                      className={`text-[10px] font-mono px-2 py-1 rounded-lg border transition ${
                        autoScroll
                          ? "bg-rose-500/10 text-rose-300 border-rose-500/30"
                          : "bg-slate-800 text-slate-400 border-slate-700"
                      }`}
                    >
                      Karaoke Scroll: {autoScroll ? "ON" : "OFF"}
                    </button>
                  </div>
                </div>

                {/* Transcript Search Filter */}
                <div className="relative">
                  <Search className="w-3 h-3 absolute left-3 top-1/2 -translate-y-1/2 text-slate-500" />
                  <input
                    type="text"
                    placeholder="Search inside transcript text..."
                    value={transcriptSearch}
                    onChange={(e) => setTranscriptSearch(e.target.value)}
                    className="w-full bg-slate-950 border border-slate-800 rounded-xl pl-8 pr-3 py-1.5 text-xs text-slate-300 placeholder:text-slate-500 focus:outline-none focus:border-cyan-500/50 transition font-mono"
                  />
                </div>

                {/* Paragraphs Display */}
                <div
                  ref={transcriptContainerRef}
                  className="space-y-2.5 max-h-[380px] overflow-y-auto pr-1 rounded-xl p-1"
                >
                  {loadingTranscript ? (
                    <div className="p-8 text-center text-xs font-mono text-slate-500 flex items-center justify-center gap-2">
                      <RefreshCw className="w-4 h-4 animate-spin text-cyan-400" />
                      <span>Loading synchronized transcript...</span>
                    </div>
                  ) : !transcript || filteredParagraphs.length === 0 ? (
                    <div className="p-8 text-center text-xs font-mono text-slate-500 space-y-1">
                      <FileText className="w-5 h-5 text-slate-600 mx-auto" />
                      <p>No transcript available for this recording.</p>
                      <p className="text-[11px] text-slate-600">Run sync to pull latest transcripts from Google Recorder.</p>
                    </div>
                  ) : (
                    filteredParagraphs.map((para, idx) => {
                      const isActive = activeParagraphIndex === idx;
                      const isCopied = copiedId === `quote-${para.start_ms}`;

                      return (
                        <div
                          key={idx}
                          id={`para-${idx}`}
                          className={`p-3 rounded-xl border transition-all ${
                            isActive
                              ? "bg-gradient-to-r from-amber-500/10 via-rose-500/5 to-slate-900 border-amber-500/40 shadow-lg shadow-amber-500/5 ring-1 ring-amber-500/20"
                              : "bg-slate-950/60 border-slate-800/80 hover:border-slate-700 hover:bg-slate-900/40"
                          }`}
                        >
                          <div className="flex items-center justify-between gap-2 mb-1.5">
                            <div className="flex items-center gap-2">
                              {/* Click-to-Seek Timestamp Badge */}
                              <button
                                onClick={() => seekToMs(para.start_ms)}
                                className="flex items-center gap-1 text-[11px] font-mono px-2 py-0.5 rounded-md bg-rose-500/10 text-rose-300 hover:bg-rose-500/20 border border-rose-500/20 transition group"
                                title="Jump audio playback to this moment"
                              >
                                <Play className="w-2.5 h-2.5 fill-current text-rose-400 group-hover:scale-110 transition-transform" />
                                <span>{formatTime(para.start_ms / 1000)}</span>
                              </button>

                              {/* Speaker Badge */}
                              <span className="text-[10px] font-mono font-semibold px-2 py-0.5 rounded bg-cyan-950/80 text-cyan-300 border border-cyan-800/60">
                                {para.speaker}
                              </span>
                            </div>

                            {/* Copy Quote Button */}
                            <button
                              onClick={() => copyQuote(para)}
                              className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 hover:bg-slate-700 text-slate-300 flex items-center gap-1 transition"
                              title="Copy quote with attribution"
                            >
                              {isCopied ? <Check className="w-3 h-3 text-emerald-400" /> : <Copy className="w-3 h-3 text-slate-400" />}
                              <span>{isCopied ? "Copied!" : "Quote"}</span>
                            </button>
                          </div>

                          <p
                            onClick={() => seekToMs(para.start_ms)}
                            className={`text-xs leading-relaxed transition cursor-pointer ${
                              isActive ? "text-slate-100 font-medium" : "text-slate-300 hover:text-slate-100"
                            }`}
                          >
                            {para.text}
                          </p>
                        </div>
                      );
                    })
                  )}
                </div>
              </div>
            </div>
          ) : (
            <div className="bg-slate-900/80 border border-slate-800 rounded-2xl p-12 text-center text-xs font-mono text-slate-500 space-y-3">
              <Headphones className="w-10 h-10 text-slate-700 mx-auto" />
              <p className="text-sm font-semibold text-slate-300">No Recording Selected</p>
              <p>Select a recording from the left panel to begin listening and inspect the synchronized transcript.</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
