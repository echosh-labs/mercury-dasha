"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  Video,
  Eye,
  Clock,
  Users,
  CheckCircle2,
  AlertCircle,
  RefreshCw,
  LogOut,
  Upload,
  BarChart3,
  Layers,
  Sparkles,
  ExternalLink,
  ShieldCheck,
  Tag,
  Radio,
  FileVideo,
  Play,
  Calendar,
  Compass,
  ArrowUpRight,
  BookOpen,
} from "lucide-react";
import { useYouTubeStudio, UploadDraft } from "@/lib/youtube-context";
import { useToast } from "@/lib/toast-context";
import { AnalyticsReport, FinancialReport } from "@/lib/types/treasury";
import LifesWorkSeedsModal from "@/components/foundations/lifes-work-seeds-modal";

export default function FoundationsCreatorHub() {
  const {
    status: ytStatus,
    loading: loadingStatus,
    activeJobs,
    openUploader,
    startUpload,
    cancelJob,
    connectChannel,
    disconnectChannel,
    refreshStatus,
  } = useYouTubeStudio();

  const { success: toastSuccess, error: toastError } = useToast();

  const [activeTab, setActiveTab] = useState<"overview" | "upload" | "generation" | "jobs" | "analytics" | "seeds">("overview");
  const [isSeedsModalOpen, setIsSeedsModalOpen] = useState<boolean>(false);
  const [analytics, setAnalytics] = useState<AnalyticsReport | null>(null);
  const [loadingAnalytics, setLoadingAnalytics] = useState<boolean>(false);

  // Video generation draft parameters
  const [genTitle, setGenTitle] = useState<string>("Mercury Dasha Ephemeris Chronicle");
  const [genOrientation, setGenOrientation] = useState<"16:9" | "9:16" | "1:1">("16:9");
  const [genAttachChrono, setGenAttachChrono] = useState<boolean>(true);
  const [genTags, setGenTags] = useState<string>("MercuryDasha, echosh-labs, VedicAstrology, DashaEngine");

  // Direct upload form parameters
  const [uploadFilePath, setUploadFilePath] = useState<string>("");
  const [uploadTitle, setUploadTitle] = useState<string>("");
  const [uploadDescription, setUploadDescription] = useState<string>("");
  const [uploadPrivacy, setUploadPrivacy] = useState<"private" | "unlisted" | "public">("private");
  const [uploadingDirect, setUploadingDirect] = useState<boolean>(false);

  // Load YouTube Analytics Feed
  const loadAnalytics = useCallback(async () => {
    if (!ytStatus?.authenticated) return;
    setLoadingAnalytics(true);
    try {
      const res = await fetch("/api/v1/youtube/analytics");
      if (res.ok) {
        setAnalytics(await res.json());
      }
    } catch (err) {
      console.error("Failed to load YouTube analytics:", err);
    } finally {
      setLoadingAnalytics(false);
    }
  }, [ytStatus?.authenticated]);

  useEffect(() => {
    loadAnalytics();
  }, [loadAnalytics]);

  const handleDirectUploadSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!uploadFilePath.trim()) {
      toastError("Please enter a valid file path on the host.");
      return;
    }
    setUploadingDirect(true);
    try {
      const draft: UploadDraft = {
        filePath: uploadFilePath.trim(),
        title: uploadTitle.trim() || "Mercury Dasha Video Recording",
        description: uploadDescription.trim(),
        tags: genTags.split(",").map((t) => t.trim()).filter((t) => t.length > 0),
        categoryID: "22",
        privacyStatus: uploadPrivacy,
        madeForKids: false,
        embeddable: true,
        attachChronoContext: genAttachChrono,
      };

      const job = await startUpload(draft);
      if (job) {
        toastSuccess(`Upload initiated: Job ${job.id}`);
        setUploadFilePath("");
        setUploadTitle("");
        setUploadDescription("");
        setActiveTab("jobs");
      }
    } catch (err: any) {
      toastError(`Upload failed: ${err.message}`);
    } finally {
      setUploadingDirect(false);
    }
  };

  return (
    <div className="space-y-6 animate-in fade-in duration-300">
      {/* 1. Channel Presence & Production Studio HUD */}
      <div className="bg-gradient-to-r from-red-950/60 via-slate-900 to-slate-950 p-6 rounded-2xl border border-red-500/20 shadow-2xl">
        <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4">
          <div>
            <div className="flex items-center space-x-3">
              <span className="p-2.5 rounded-xl bg-red-500/10 border border-red-500/30 text-red-400">
                <Video className="w-6 h-6" />
              </span>
              <div>
                <div className="flex flex-wrap items-center gap-2">
                  <h2 className="text-xl font-bold text-white tracking-tight font-serif">
                    Foundations Creator Studio
                  </h2>
                  <span className="bg-red-500/10 text-red-300 border border-red-500/20 text-xs font-mono px-2.5 py-0.5 rounded-full flex items-center space-x-1">
                    <Radio className="w-3 h-3 text-red-400 animate-pulse" />
                    <span>BROADCASTING ENGINE</span>
                  </span>
                  <span className={`text-xs font-mono px-2 py-0.5 rounded border ${
                    ytStatus?.authenticated
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/30"
                      : "bg-amber-500/10 text-amber-300 border-amber-500/30"
                  }`}>
                    {ytStatus?.authenticated ? "CHANNEL CONNECTED" : "OAUTH PENDING"}
                  </span>
                </div>
                <p className="text-xs text-slate-400 mt-1 max-w-2xl">
                  Unified Video Pipeline, Sonic Chronicle Rendering, Channel Management, and Multi-Format Content Production for <span className="text-cyan-300 font-mono">echosh-labs</span>.
                </p>
              </div>
            </div>
          </div>

          <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
            <button
              onClick={() => {
                refreshStatus();
                loadAnalytics();
              }}
              disabled={loadingStatus}
              className="px-3 py-1.5 rounded-lg bg-slate-800 hover:bg-slate-700 text-slate-300 border border-slate-700 transition flex items-center space-x-1.5"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loadingStatus ? "animate-spin text-red-400" : ""}`} />
              <span>Refresh Status</span>
            </button>

            {ytStatus?.authenticated ? (
              <button
                onClick={disconnectChannel}
                className="px-3 py-1.5 rounded-lg bg-slate-800/80 hover:bg-rose-950/40 text-slate-400 hover:text-rose-300 border border-slate-700 hover:border-rose-800/50 transition flex items-center space-x-1.5"
                title="Disconnect Channel Credentials"
              >
                <LogOut className="w-3.5 h-3.5" />
                <span>Disconnect</span>
              </button>
            ) : (
              <button
                onClick={connectChannel}
                className="px-4 py-1.5 rounded-lg bg-red-600 hover:bg-red-500 text-white font-semibold transition flex items-center space-x-1.5 shadow-lg shadow-red-900/30"
              >
                <span>Connect Channel</span>
                <ExternalLink className="w-3.5 h-3.5" />
              </button>
            )}
          </div>
        </div>

        {/* Channel Identity & Quota Status Row */}
        {ytStatus?.authenticated && ytStatus.channel ? (
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mt-6 pt-5 border-t border-slate-800/80">
            {/* Channel Profile */}
            <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 flex items-center space-x-3.5">
              {ytStatus.channel.thumbnail_url ? (
                <img
                  src={ytStatus.channel.thumbnail_url}
                  alt={ytStatus.channel.title}
                  className="w-12 h-12 rounded-full border border-red-500/40 shrink-0"
                />
              ) : (
                <div className="w-12 h-12 rounded-full bg-red-950/40 flex items-center justify-center text-red-400 shrink-0">
                  <Video className="w-6 h-6" />
                </div>
              )}
              <div className="min-w-0">
                <h4 className="text-sm font-bold text-white truncate flex items-center space-x-1.5">
                  <span>{ytStatus.channel.title}</span>
                  <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400 shrink-0" />
                </h4>
                <div className="text-xs text-slate-400 font-mono mt-0.5">
                  {Number(ytStatus.channel.subscriber_count || 0).toLocaleString()} subscribers
                </div>
                <div className="text-[10px] text-slate-500 font-mono truncate">
                  {Number(ytStatus.channel.video_count || 0).toLocaleString()} videos • ID: {ytStatus.channel.channel_id}
                </div>
              </div>
            </div>

            {/* Daily API Quota Gauge */}
            <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 space-y-2">
              <div className="flex justify-between items-center text-xs font-mono">
                <span className="text-slate-400 uppercase text-[10px]">YouTube API Daily Quota</span>
                <span className="text-cyan-300 font-semibold">
                  {ytStatus.quota?.used_today.toLocaleString()} / {ytStatus.quota?.daily_limit.toLocaleString()}
                </span>
              </div>
              <div className="h-2 w-full bg-slate-800 rounded-full overflow-hidden">
                <div
                  className="h-full bg-gradient-to-r from-cyan-500 to-red-500 rounded-full transition-all"
                  style={{
                    width: `${Math.min(
                      ((ytStatus.quota?.used_today || 0) / (ytStatus.quota?.daily_limit || 10000)) * 100,
                      100
                    )}%`,
                  }}
                />
              </div>
              <div className="text-[10px] text-slate-500 font-mono flex justify-between">
                <span>Reset: {ytStatus.quota?.reset_time_utc ? new Date(ytStatus.quota.reset_time_utc).toLocaleTimeString() : "Midnight UTC"}</span>
                <span>{((((ytStatus.quota?.used_today || 0) / (ytStatus.quota?.daily_limit || 10000)) * 100)).toFixed(1)}% used</span>
              </div>
            </div>

            {/* Active Pipeline Jobs */}
            <div className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 flex flex-col justify-between">
              <div className="flex items-center justify-between">
                <span className="text-xs font-mono uppercase text-slate-400">Pipeline Queue</span>
                <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-purple-500/10 text-purple-300 border border-purple-500/20">
                  {activeJobs.length} active
                </span>
              </div>
              <div className="text-xl font-bold text-white mt-1">
                {activeJobs.length > 0 ? (
                  <span className="text-amber-400 flex items-center space-x-2">
                    <RefreshCw className="w-4 h-4 animate-spin" />
                    <span>Uploading...</span>
                  </span>
                ) : (
                  <span className="text-emerald-400 text-sm font-mono flex items-center space-x-1.5">
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Queue Idle</span>
                  </span>
                )}
              </div>
              <button
                onClick={() => openUploader()}
                className="mt-2 text-xs text-red-400 hover:text-red-300 font-mono flex items-center space-x-1"
              >
                <span>Open Docked Uploader →</span>
              </button>
            </div>
          </div>
        ) : (
          <div className="mt-6 p-4 rounded-xl bg-slate-950/60 border border-amber-500/20 flex flex-col sm:flex-row items-center justify-between gap-4 text-xs">
            <div className="flex items-center space-x-3">
              <AlertCircle className="w-5 h-5 text-amber-400 shrink-0" />
              <div className="text-slate-300">
                Connect your YouTube creator channel with Google OAuth to enable direct video uploads, live quota telemetry, and audience analytics feeds.
              </div>
            </div>
            <button
              onClick={connectChannel}
              className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 text-white font-semibold whitespace-nowrap transition"
            >
              Authorize YouTube Channel
            </button>
          </div>
        )}
      </div>

      {/* 2. Sub-Navigation Tabs */}
      <div className="flex border-b border-slate-800 space-x-2 text-xs font-mono font-medium overflow-x-auto pb-1">
        <button
          onClick={() => setActiveTab("overview")}
          className={`px-4 py-2 rounded-lg transition ${
            activeTab === "overview"
              ? "bg-slate-800 text-red-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          Studio Overview
        </button>

        <button
          onClick={() => setActiveTab("upload")}
          className={`px-4 py-2 rounded-lg transition flex items-center space-x-1.5 ${
            activeTab === "upload"
              ? "bg-slate-800 text-cyan-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          <Upload className="w-3.5 h-3.5 text-cyan-400" />
          <span>Video Uploader</span>
        </button>

        <button
          onClick={() => setActiveTab("generation")}
          className={`px-4 py-2 rounded-lg transition flex items-center space-x-1.5 ${
            activeTab === "generation"
              ? "bg-slate-800 text-amber-300 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          <Sparkles className="w-3.5 h-3.5 text-amber-400" />
          <span>Timeline Manifest Generator</span>
        </button>

        <button
          onClick={() => setActiveTab("seeds")}
          className={`px-4 py-2 rounded-lg transition flex items-center space-x-1.5 ${
            activeTab === "seeds"
              ? "bg-slate-800 text-amber-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          <BookOpen className="w-3.5 h-3.5 text-amber-400" />
          <span>Story Seeds &amp; Life&apos;s Work</span>
        </button>

        <button
          onClick={() => setActiveTab("jobs")}
          className={`px-4 py-2 rounded-lg transition flex items-center space-x-1.5 ${
            activeTab === "jobs"
              ? "bg-slate-800 text-purple-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          <Layers className="w-3.5 h-3.5 text-purple-400" />
          <span>Upload Jobs ({activeJobs.length})</span>
        </button>

        <button
          onClick={() => setActiveTab("analytics")}
          className={`px-4 py-2 rounded-lg transition flex items-center space-x-1.5 ${
            activeTab === "analytics"
              ? "bg-slate-800 text-emerald-400 border border-slate-700 font-semibold"
              : "text-slate-400 hover:text-slate-200"
          }`}
        >
          <BarChart3 className="w-3.5 h-3.5 text-emerald-400" />
          <span>Audience &amp; Retention</span>
        </button>
      </div>

      {/* 3. Sub-View: Studio Overview */}
      {activeTab === "overview" && (
        <div className="space-y-6">
          {/* Top Audience KPI Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
              <div className="text-[11px] font-mono uppercase text-slate-400 flex items-center justify-between">
                <span>Views (28 Days)</span>
                <Eye className="w-4 h-4 text-cyan-400" />
              </div>
              <div className="text-2xl font-bold text-white mt-1">
                {(analytics?.total_views || 0).toLocaleString()}
              </div>
              <div className="text-[10px] font-mono text-slate-500 mt-1">
                Avg Duration: {(analytics?.average_view_duration_secs || 0).toFixed(0)}s
              </div>
            </div>

            <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
              <div className="text-[11px] font-mono uppercase text-slate-400 flex items-center justify-between">
                <span>Watch Time</span>
                <Clock className="w-4 h-4 text-amber-400" />
              </div>
              <div className="text-2xl font-bold text-white mt-1">
                {((analytics?.total_minutes_watched || 0) / 60).toFixed(1)} hrs
              </div>
              <div className="text-[10px] font-mono text-slate-500 mt-1">
                {(analytics?.total_minutes_watched || 0).toFixed(0)} total minutes
              </div>
            </div>

            <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
              <div className="text-[11px] font-mono uppercase text-slate-400 flex items-center justify-between">
                <span>Net Subscriber Gain</span>
                <Users className="w-4 h-4 text-emerald-400" />
              </div>
              <div className="text-2xl font-bold text-emerald-400 mt-1">
                +{(analytics?.net_subscribers || 0).toLocaleString()}
              </div>
              <div className="text-[10px] font-mono text-slate-500 mt-1">
                +{analytics?.subscribers_gained || 0} / -{analytics?.subscribers_lost || 0}
              </div>
            </div>

            <div className="p-4 rounded-xl bg-slate-900/90 border border-slate-800">
              <div className="text-[11px] font-mono uppercase text-slate-400 flex items-center justify-between">
                <span>Engagement Actions</span>
                <Sparkles className="w-4 h-4 text-purple-400" />
              </div>
              <div className="text-2xl font-bold text-purple-300 mt-1">
                {((analytics?.total_likes || 0) + (analytics?.total_comments || 0)).toLocaleString()}
              </div>
              <div className="text-[10px] font-mono text-slate-500 mt-1">
                {analytics?.total_likes || 0} likes • {analytics?.total_comments || 0} comments
              </div>
            </div>
          </div>

          {/* Cross-Route Synergy Link */}
          <div className="p-5 rounded-2xl bg-gradient-to-r from-emerald-950/30 via-slate-900 to-slate-950 border border-emerald-500/20 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="flex items-center space-x-3">
              <span className="p-2 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/30">
                <ShieldCheck className="w-5 h-5" />
              </span>
              <div>
                <h4 className="text-sm font-bold text-white font-mono flex items-center space-x-2">
                  <span>AMRA Sovereign Financial Ledger</span>
                  <span className="text-[10px] bg-emerald-500/10 text-emerald-300 border border-emerald-500/30 px-2 py-0.5 rounded-full">
                    IMMUTABLE YIELD
                  </span>
                </h4>
                <p className="text-xs text-slate-400 mt-0.5">
                  YouTube Partner ad accruals and CPM revenue are audited and reconciled in the AMRA Treasury.
                </p>
              </div>
            </div>
            <a
              href="/treasury/"
              className="px-4 py-2 rounded-xl bg-emerald-600/90 hover:bg-emerald-500 text-white font-mono text-xs font-semibold flex items-center space-x-1.5 transition shadow"
            >
              <span>View AMRA Treasury &amp; Ledger</span>
              <ArrowUpRight className="w-3.5 h-3.5" />
            </a>
          </div>
        </div>
      )}

      {/* 4. Sub-View: Video Uploader Form */}
      {activeTab === "upload" && (
        <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
          <div className="lg:col-span-8 bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <h3 className="text-sm font-semibold text-white font-mono flex items-center space-x-2">
                <Upload className="w-4 h-4 text-red-400" />
                <span>Upload Video to YouTube Channel</span>
              </h3>
              <div className="flex items-center space-x-2">
                <button
                  type="button"
                  onClick={() => setIsSeedsModalOpen(true)}
                  className="px-2.5 py-1 rounded-lg bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/30 text-[11px] font-mono flex items-center space-x-1 transition"
                >
                  <Sparkles className="w-3 h-3 text-amber-400" />
                  <span>Import from Life&apos;s Work</span>
                </button>
                <span className="text-xs font-mono text-slate-400">Resumable Chunked Stream</span>
              </div>
            </div>

            <form onSubmit={handleDirectUploadSubmit} className="space-y-4 text-xs font-mono">
              <div className="space-y-1.5">
                <label className="text-slate-300 font-semibold block">File Path on Disk</label>
                <input
                  type="text"
                  placeholder="/home/justin/Dropbox/recordings/dnd_ephemeris.mp4"
                  value={uploadFilePath}
                  onChange={(e) => setUploadFilePath(e.target.value)}
                  className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3.5 py-2.5 text-slate-200 focus:outline-none focus:border-red-500/50"
                  required
                />
                <p className="text-[11px] text-slate-500">Must be an existing video file on the POSIX or Windows substrate.</p>
              </div>

              <div className="space-y-1.5">
                <label className="text-slate-300 font-semibold block">Video Title</label>
                <input
                  type="text"
                  placeholder="Mercury Dasha • Sidereal Moon Ingress Swati Pada 4"
                  value={uploadTitle}
                  onChange={(e) => setUploadTitle(e.target.value)}
                  className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3.5 py-2.5 text-slate-200 focus:outline-none focus:border-red-500/50"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-slate-300 font-semibold block">Description</label>
                <textarea
                  rows={3}
                  placeholder="Automated session chronicle generated by Mercury Dasha with live Chrono-Pulse telemetry."
                  value={uploadDescription}
                  onChange={(e) => setUploadDescription(e.target.value)}
                  className="w-full bg-slate-950/80 border border-slate-800 rounded-lg p-3 text-slate-200 focus:outline-none focus:border-red-500/50"
                />
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-slate-300 font-semibold block">Privacy State</label>
                  <select
                    value={uploadPrivacy}
                    onChange={(e) => setUploadPrivacy(e.target.value as any)}
                    className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3 py-2 text-slate-300 focus:outline-none focus:border-red-500/50"
                  >
                    <option value="private">Private (Recommended for verification)</option>
                    <option value="unlisted">Unlisted (Shareable link)</option>
                    <option value="public">Public (Immediate broadcast)</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-slate-300 font-semibold block">Tags (CSV)</label>
                  <input
                    type="text"
                    value={genTags}
                    onChange={(e) => setGenTags(e.target.value)}
                    className="w-full bg-slate-950/80 border border-slate-800 rounded-lg px-3.5 py-2 text-slate-200 focus:outline-none focus:border-red-500/50"
                  />
                </div>
              </div>

              <div className="pt-2 flex items-center space-x-2">
                <input
                  type="checkbox"
                  id="attachChronoCheck"
                  checked={genAttachChrono}
                  onChange={(e) => setGenAttachChrono(e.target.checked)}
                  className="rounded bg-slate-800 border-slate-700 text-red-500 focus:ring-0"
                />
                <label htmlFor="attachChronoCheck" className="text-slate-300 cursor-pointer text-xs">
                  Attach Live ChronoContext (active Hora, sacred metal, day lord) to metadata description
                </label>
              </div>

              <button
                type="submit"
                disabled={uploadingDirect || !ytStatus?.authenticated}
                className="w-full py-2.5 rounded-xl bg-red-600 hover:bg-red-500 text-white font-bold text-xs transition flex items-center justify-center space-x-2 shadow-lg shadow-red-900/30 disabled:opacity-50"
              >
                <Upload className={`w-4 h-4 ${uploadingDirect ? "animate-bounce" : ""}`} />
                <span>{uploadingDirect ? "Queuing Pipeline Job..." : "Queue Sovereign Upload Job"}</span>
              </button>
            </form>
          </div>

          <div className="lg:col-span-4 space-y-4 text-xs font-mono">
            <div className="p-5 rounded-2xl bg-slate-900/90 border border-slate-800 space-y-3">
              <h4 className="text-slate-200 font-bold flex items-center space-x-2">
                <Radio className="w-4 h-4 text-red-400" />
                <span>Upload Engine Details</span>
              </h4>
              <p className="text-slate-400 leading-relaxed font-sans">
                Uploads stream through the Go backend resumable upload pipeline. Files larger than 50MB are chunked and trackable through live background progress callbacks.
              </p>
              <div className="p-3 rounded-lg bg-slate-950/80 border border-slate-800/80 space-y-1 text-[11px]">
                <div className="text-cyan-300 font-semibold">Resumable OAuth2 Endpoint</div>
                <div className="text-slate-400">Quota cost: 1,600 units / video upload</div>
                <div className="text-slate-400">Daily headroom: {((ytStatus?.quota?.daily_limit || 10000) - (ytStatus?.quota?.used_today || 0)).toLocaleString()} units</div>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* 5. Sub-View: Timeline Manifest Generator */}
      {activeTab === "generation" && (
        <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-5">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <div>
              <h3 className="text-sm font-semibold text-white font-mono flex items-center space-x-2">
                <Sparkles className="w-4 h-4 text-amber-400" />
                <span>Timeline Manifest Video Compiler (`internal/timeline`)</span>
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Assemble Sonic Chronicle audio markers into video presentation manifests with automated ducking and orientation layout.
              </p>
            </div>
            <div className="flex items-center space-x-2">
              <button
                type="button"
                onClick={() => setIsSeedsModalOpen(true)}
                className="px-3 py-1.5 rounded-lg bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/30 text-xs font-mono flex items-center space-x-1.5 transition"
              >
                <Sparkles className="w-3.5 h-3.5 text-amber-400" />
                <span>Load Story Seed</span>
              </button>
              <span className="text-xs font-mono text-amber-300 bg-amber-500/10 border border-amber-500/20 px-2.5 py-0.5 rounded-full">
                CANVAS RENDERING
              </span>
            </div>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <button
              onClick={() => setGenOrientation("16:9")}
              className={`p-4 rounded-xl border text-left transition ${
                genOrientation === "16:9"
                  ? "bg-amber-500/10 border-amber-500/50 text-amber-300"
                  : "bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700"
              }`}
            >
              <div className="font-bold text-sm font-mono">16:9 Landscape</div>
              <div className="text-xs mt-1 font-sans">Standard YouTube Longform (1920x1080)</div>
            </button>

            <button
              onClick={() => setGenOrientation("9:16")}
              className={`p-4 rounded-xl border text-left transition ${
                genOrientation === "9:16"
                  ? "bg-amber-500/10 border-amber-500/50 text-amber-300"
                  : "bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700"
              }`}
            >
              <div className="font-bold text-sm font-mono">9:16 Vertical</div>
              <div className="text-xs mt-1 font-sans">YouTube Shorts &amp; Mobile Reels (1080x1920)</div>
            </button>

            <button
              onClick={() => setGenOrientation("1:1")}
              className={`p-4 rounded-xl border text-left transition ${
                genOrientation === "1:1"
                  ? "bg-amber-500/10 border-amber-500/50 text-amber-300"
                  : "bg-slate-950/60 border-slate-800 text-slate-400 hover:border-slate-700"
              }`}
            >
              <div className="font-bold text-sm font-mono">1:1 Square</div>
              <div className="text-xs mt-1 font-sans">Social Feeds &amp; Square Compendiums (1080x1080)</div>
            </button>
          </div>

          <div className="p-4 rounded-xl bg-slate-950/80 border border-slate-800/80 text-xs font-mono space-y-2">
            <div className="text-slate-300 font-semibold">Compiler Manifest Structure:</div>
            <div className="text-slate-400 leading-relaxed">
              • Primary Voice Audio: Ingested from selected Sonic Chronicle WAV session.<br />
              • Ambient Music Bed: Auto-ducked by -15dB during speaker activity.<br />
              • ChronoContext Metadata: Encodes Lahiri Ayanamsha moon degree and governing alchemical metal directly into video description.
            </div>
          </div>
        </div>
      )}

      {/* 6. Sub-View: Upload Jobs Queue */}
      {activeTab === "jobs" && (
        <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <h3 className="text-sm font-semibold text-white font-mono flex items-center space-x-2">
              <Layers className="w-4 h-4 text-purple-400" />
              <span>Sovereign Upload Jobs &amp; Video Queue</span>
            </h3>
            <span className="text-xs font-mono text-purple-300">
              {activeJobs.length} active tasks
            </span>
          </div>

          <div className="space-y-3">
            {activeJobs.length > 0 ? (
              activeJobs.map((job) => (
                <div
                  key={job.id}
                  className="p-4 rounded-xl bg-slate-950/70 border border-slate-800 space-y-3"
                >
                  <div className="flex items-center justify-between text-xs font-mono">
                    <span className="font-bold text-slate-200 truncate">{job.title || job.file_name}</span>
                    <span className={`px-2 py-0.5 rounded uppercase text-[10px] ${
                      job.status === "completed"
                        ? "bg-emerald-500/10 text-emerald-400 border border-emerald-500/20"
                        : job.status === "uploading"
                        ? "bg-amber-500/10 text-amber-300 border border-amber-500/20"
                        : "bg-purple-500/10 text-purple-300 border border-purple-500/20"
                    }`}>
                      {job.status}
                    </span>
                  </div>

                  <div className="space-y-1">
                    <div className="flex justify-between text-[10px] font-mono text-slate-400">
                      <span>Progress: {job.progress_pct}%</span>
                      <span>{(job.bytes_uploaded / (1024 * 1024)).toFixed(1)} MB / {(job.total_bytes / (1024 * 1024)).toFixed(1)} MB</span>
                    </div>
                    <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-gradient-to-r from-red-500 to-purple-500 transition-all duration-300"
                        style={{ width: `${job.progress_pct}%` }}
                      />
                    </div>
                  </div>

                  <div className="flex justify-between items-center text-[10px] font-mono text-slate-500 pt-1">
                    <span>Job ID: {job.id}</span>
                    {job.status === "uploading" && (
                      <button
                        onClick={() => cancelJob(job.id)}
                        className="text-rose-400 hover:text-rose-300"
                      >
                        Cancel Job
                      </button>
                    )}
                  </div>
                </div>
              ))
            ) : (
              <div className="py-12 text-center text-xs font-mono text-slate-500">
                No active upload jobs in queue. Use the Video Uploader or Dock to queue new content.
              </div>
            )}
          </div>
        </div>
      )}

      {/* 7. Sub-View: Audience & Retention Feed */}
      {activeTab === "analytics" && (
        <div className="bg-slate-900/90 border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <div>
              <h3 className="text-sm font-semibold text-white font-mono flex items-center space-x-2">
                <BarChart3 className="w-4 h-4 text-emerald-400" />
                <span>Audience Engagement &amp; Retention Telemetry</span>
              </h3>
              <p className="text-xs text-slate-400 mt-0.5">
                Time-series viewer interaction and watch duration queried via YouTube Analytics v2.
              </p>
            </div>
            <button
              onClick={loadAnalytics}
              className="p-1.5 rounded-lg bg-slate-800 text-slate-400 hover:text-white transition"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loadingAnalytics ? "animate-spin" : ""}`} />
            </button>
          </div>

          <div className="overflow-x-auto">
            <table className="w-full text-left text-xs font-mono">
              <thead className="border-b border-slate-800 text-slate-400 uppercase text-[10px] bg-slate-950/40">
                <tr>
                  <th className="py-2.5 px-3">Date</th>
                  <th className="py-2.5 px-3">Views</th>
                  <th className="py-2.5 px-3">Watch Time (min)</th>
                  <th className="py-2.5 px-3">Avg Duration</th>
                  <th className="py-2.5 px-3">Net Subs</th>
                  <th className="py-2.5 px-3">Likes</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-slate-800/60 text-slate-300">
                {analytics?.daily_rows && analytics.daily_rows.length > 0 ? (
                  analytics.daily_rows.map((row) => (
                    <tr key={row.day} className="hover:bg-slate-800/30 transition">
                      <td className="py-2.5 px-3 text-white font-medium">{row.day}</td>
                      <td className="py-2.5 px-3">{row.views.toLocaleString()}</td>
                      <td className="py-2.5 px-3">{row.minutes_watched.toFixed(1)}</td>
                      <td className="py-2.5 px-3">{row.average_view_duration_secs.toFixed(0)}s</td>
                      <td className="py-2.5 px-3">
                        <span className={row.subscribers_gained >= row.subscribers_lost ? "text-emerald-400" : "text-rose-400"}>
                          {row.subscribers_gained - row.subscribers_lost >= 0 ? "+" : ""}
                          {row.subscribers_gained - row.subscribers_lost}
                        </span>
                      </td>
                      <td className="py-2.5 px-3 text-purple-300">{row.likes}</td>
                    </tr>
                  ))
                ) : (
                  <tr>
                    <td colSpan={6} className="py-8 text-center text-slate-500">
                      No daily audience rows available. Ensure your YouTube channel is paired and has viewer activity.
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* 8. Sub-View: Story Seeds & Life's Work */}
      {activeTab === "seeds" && (
        <div className="space-y-6">
          <div className="bg-gradient-to-r from-amber-950/40 via-slate-900 to-slate-950 border border-amber-500/20 rounded-2xl p-6 shadow-xl flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <div className="flex items-center space-x-2">
                <Sparkles className="w-5 h-5 text-amber-400" />
                <h3 className="text-base font-bold text-white font-serif">
                  Foundations Story Seeds &amp; Personal Lore Matrix
                </h3>
              </div>
              <p className="text-xs text-slate-400 mt-1 max-w-2xl">
                Personal writings, blessings, Kybalion notes, and Google Recorder transcripts transformed into archetypal narrative hooks, character dialogue, and 3-act alchemical story progressions.
              </p>
            </div>

            <button
              onClick={() => setIsSeedsModalOpen(true)}
              className="py-2.5 px-5 rounded-xl bg-gradient-to-r from-amber-600 to-amber-500 hover:from-amber-500 hover:to-amber-400 text-white font-bold text-xs font-mono flex items-center justify-center space-x-2 shadow-lg shadow-amber-900/30 whitespace-nowrap transition"
            >
              <Sparkles className="w-4 h-4" />
              <span>Synthesize New Seed</span>
            </button>
          </div>

          <div className="p-8 rounded-2xl bg-slate-900/80 border border-slate-800 text-center space-y-4">
            <div className="w-12 h-12 rounded-full bg-amber-500/10 border border-amber-500/20 text-amber-400 flex items-center justify-center mx-auto">
              <BookOpen className="w-6 h-6" />
            </div>
            <div className="max-w-md mx-auto space-y-1">
              <h4 className="text-sm font-bold text-white font-serif">Life&apos;s Work Lore Sanctuary</h4>
              <p className="text-xs text-slate-400">
                Explore all 926 personal notes across 14 sanctuaries or query documents resonant with your active Vimshottari Dasha transit.
              </p>
            </div>
            <button
              onClick={() => setIsSeedsModalOpen(true)}
              className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-amber-300 border border-amber-500/30 text-xs font-mono inline-flex items-center space-x-2 transition"
            >
              <span>Launch Lore Matrix &amp; Synthesizer →</span>
            </button>
          </div>
        </div>
      )}

      {/* 9. Interactive Story Seeds & Lore Drawer/Modal */}
      <LifesWorkSeedsModal
        isOpen={isSeedsModalOpen}
        onClose={() => setIsSeedsModalOpen(false)}
        onLoadDraft={(draft) => {
          setUploadTitle(draft.title);
          setUploadDescription(draft.description);
          setGenTags(draft.tags.join(", "));
          if (draft.filePath) {
            setUploadFilePath(draft.filePath);
          }
          setActiveTab("upload");
        }}
      />
    </div>
  );
}
