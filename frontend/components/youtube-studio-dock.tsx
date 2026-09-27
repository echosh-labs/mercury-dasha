"use client";

import React, { useState, useEffect } from "react";
import {
  Video,
  Upload,
  ExternalLink,
  X,
  Play,
  CheckCircle2,
  AlertCircle,
  Clock,
  Sparkles,
  ShieldCheck,
  RefreshCw,
  LogOut,
  Maximize2,
  Minimize2,
  Tag,
  FileVideo,
  Flame,
  Radio
} from "lucide-react";
import { useYouTubeStudio, UploadDraft } from "@/lib/youtube-context";

export default function YouTubeStudioDock() {
  const {
    status,
    loading,
    isDrawerOpen,
    draft,
    activeJobs,
    openUploader,
    closeDrawer,
    startUpload,
    cancelJob,
    connectChannel,
    disconnectChannel,
    refreshStatus,
  } = useYouTubeStudio();

  const [formDraft, setFormDraft] = useState<UploadDraft>({
    filePath: "",
    title: "",
    description: "",
    tags: ["MercuryDasha", "echosh-labs"],
    categoryID: "22",
    privacyStatus: "private",
    madeForKids: false,
    embeddable: true,
    attachChronoContext: true,
  });

  const [tagsInput, setTagsInput] = useState<string>("");
  const [submitting, setSubmitting] = useState<boolean>(false);
  const [isMinimized, setIsMinimized] = useState<boolean>(false);

  // Sync draft from context when drawer opens
  useEffect(() => {
    if (draft) {
      setFormDraft(draft);
      setTagsInput(draft.tags ? draft.tags.join(", ") : "");
    }
  }, [draft]);

  const activeJob = activeJobs.length > 0 ? activeJobs[0] : null;

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formDraft.filePath) {
      alert("Please specify a valid file path on disk.");
      return;
    }

    setSubmitting(true);
    const tags = tagsInput
      .split(",")
      .map((t) => t.trim())
      .filter((t) => t.length > 0);

    const updated = { ...formDraft, tags };
    const job = await startUpload(updated);
    setSubmitting(false);

    if (job) {
      // Keep drawer open or minimize to dock
      setIsMinimized(false);
    }
  };

  return (
    <>
      {/* 1. Floating Persistent Mini-Dock (Bottom-Right) */}
      {activeJob && !isDrawerOpen && (
        <div className="fixed bottom-4 right-4 z-40 bg-slate-900/95 border border-red-500/40 rounded-xl shadow-2xl p-4 w-96 backdrop-blur transition-all duration-300 animate-in fade-in slide-in-from-bottom-5">
          <div className="flex items-center justify-between gap-2 mb-2">
            <div className="flex items-center space-x-2">
              <span className="w-2.5 h-2.5 rounded-full bg-red-500 animate-ping" />
              <span className="text-xs font-bold text-white uppercase tracking-wider">
                YouTube Pipeline
              </span>
            </div>
            <div className="flex items-center space-x-1">
              <button
                onClick={() => openUploader()}
                className="p-1 rounded text-slate-400 hover:text-white hover:bg-slate-800 transition"
                title="Expand Studio Drawer"
              >
                <Maximize2 className="w-3.5 h-3.5" />
              </button>
              <button
                onClick={() => cancelJob(activeJob.id)}
                className="p-1 rounded text-red-400 hover:text-red-300 hover:bg-red-950/40 transition"
                title="Cancel Upload"
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </div>
          </div>

          <div className="text-xs font-mono text-slate-300 truncate mb-2">
            {activeJob.file_name}
          </div>

          {/* Progress Bar */}
          <div className="w-full bg-slate-800 rounded-full h-2 overflow-hidden border border-slate-700">
            <div
              className="bg-gradient-to-r from-red-500 to-amber-500 h-full transition-all duration-500"
              style={{ width: `${Math.max(activeJob.progress_pct, 5)}%` }}
            />
          </div>

          <div className="flex items-center justify-between text-[11px] font-mono text-slate-400 mt-2">
            <span>{activeJob.progress_pct.toFixed(1)}% complete</span>
            <span className="capitalize">{activeJob.status}</span>
          </div>
        </div>
      )}

      {/* 2. Slide-Over Studio Drawer */}
      {isDrawerOpen && (
        <div className="fixed inset-0 z-50 overflow-hidden">
          {/* Backdrop */}
          <div
            className="absolute inset-0 bg-black/60 backdrop-blur-sm transition-opacity"
            onClick={closeDrawer}
          />

          <div className="fixed inset-y-0 right-0 max-w-full flex pl-10">
            <div className="w-screen max-w-xl bg-slate-950 border-l border-slate-800 p-6 flex flex-col shadow-2xl overflow-y-auto">
              {/* Drawer Header */}
              <div className="flex items-center justify-between pb-4 border-b border-slate-800">
                <div className="flex items-center space-x-3">
                  <div className="p-2 rounded-lg bg-red-600/20 border border-red-500/30 text-red-400">
                    <Video className="w-5 h-5" />
                  </div>
                  <div>
                    <h2 className="text-base font-bold text-white tracking-tight flex items-center gap-2">
                      <span>YouTube Sovereign Studio</span>
                      <span className="text-[10px] font-mono uppercase px-2 py-0.5 rounded bg-red-500/10 text-red-400 border border-red-500/20">
                        OAuth 2.0
                      </span>
                    </h2>
                    <p className="text-xs text-slate-400">
                      Sovereign video publishing & automated event-chain pipeline
                    </p>
                  </div>
                </div>
                <button
                  onClick={closeDrawer}
                  className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800 transition"
                >
                  <X className="w-5 h-5" />
                </button>
              </div>

              {/* Channel Account & Quota Status Banner */}
              <div className="my-4 p-4 rounded-xl bg-slate-900/90 border border-slate-800">
                {status?.authenticated && status.channel ? (
                  <div className="flex items-center justify-between">
                    <div className="flex items-center space-x-3">
                      {status.channel.thumbnail_url ? (
                        <img
                          src={status.channel.thumbnail_url}
                          alt={status.channel.title}
                          className="w-10 h-10 rounded-full border border-red-500/40"
                        />
                      ) : (
                        <div className="w-10 h-10 rounded-full bg-red-900/40 flex items-center justify-center text-red-400">
                          <Video className="w-5 h-5" />
                        </div>
                      )}
                      <div>
                        <div className="text-sm font-semibold text-white flex items-center space-x-1">
                          <span>{status.channel.title}</span>
                          <CheckCircle2 className="w-3.5 h-3.5 text-emerald-400" />
                        </div>
                        <div className="text-xs text-slate-400 font-mono">
                          {Number(status.channel.subscriber_count || 0).toLocaleString()} subscribers •{" "}
                          {Number(status.channel.video_count || 0).toLocaleString()} videos
                        </div>
                      </div>
                    </div>
                    <button
                      onClick={disconnectChannel}
                      className="px-2.5 py-1 text-xs font-mono rounded bg-slate-800 hover:bg-red-950/40 text-slate-300 hover:text-red-400 border border-slate-700 hover:border-red-800/40 transition flex items-center space-x-1"
                      title="Revoke and Disconnect"
                    >
                      <LogOut className="w-3 h-3" />
                      <span>Disconnect</span>
                    </button>
                  </div>
                ) : (
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                    <div>
                      <div className="text-sm font-semibold text-white flex items-center space-x-2">
                        <AlertCircle className="w-4 h-4 text-amber-400" />
                        <span>Channel Not Connected</span>
                      </div>
                      <p className="text-xs text-slate-400 mt-0.5">
                        Pair your Google Account to authorize direct video publishing.
                      </p>
                    </div>
                    <button
                      onClick={connectChannel}
                      className="px-4 py-2 rounded-lg bg-red-600 hover:bg-red-500 text-white font-semibold text-xs transition flex items-center justify-center space-x-2 shadow-lg shadow-red-900/20"
                    >
                      <Video className="w-4 h-4" />
                      <span>Connect YouTube</span>
                    </button>
                  </div>
                )}

                {/* Quota Tracker */}
                {status?.quota && (
                  <div className="mt-3 pt-3 border-t border-slate-800 flex items-center justify-between text-xs font-mono text-slate-400">
                    <span className="flex items-center space-x-1.5">
                      <ShieldCheck className="w-3.5 h-3.5 text-cyan-400" />
                      <span>Daily API Quota:</span>
                    </span>
                    <span className="text-cyan-300">
                      {status.quota.used_today.toLocaleString()} / {status.quota.daily_limit.toLocaleString()} units
                    </span>
                  </div>
                )}
              </div>

              {/* Upload Form */}
              <form onSubmit={handleSubmit} className="space-y-4 flex-1">
                {/* File Path Confirmation */}
                <div>
                  <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                    POSIX Source Video Path
                  </label>
                  <div className="flex items-center space-x-2">
                    <div className="relative flex-1">
                      <input
                        type="text"
                        value={formDraft.filePath}
                        onChange={(e) =>
                          setFormDraft({ ...formDraft, filePath: e.target.value })
                        }
                        placeholder="/home/justin/Dropbox/... or local video path"
                        className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-white text-xs font-mono focus:border-red-500 focus:outline-none"
                        required
                      />
                    </div>
                  </div>
                </div>

                {/* Title */}
                <div>
                  <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                    Video Title
                  </label>
                  <input
                    type="text"
                    value={formDraft.title}
                    onChange={(e) =>
                      setFormDraft({ ...formDraft, title: e.target.value })
                    }
                    placeholder="Enter video title"
                    className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-white text-xs focus:border-red-500 focus:outline-none"
                    required
                  />
                </div>

                {/* Description */}
                <div>
                  <div className="flex items-center justify-between mb-1">
                    <label className="text-xs font-mono uppercase text-slate-400">
                      Description
                    </label>
                    <label className="flex items-center space-x-1.5 text-xs font-mono text-amber-400 cursor-pointer">
                      <input
                        type="checkbox"
                        checked={formDraft.attachChronoContext}
                        onChange={(e) =>
                          setFormDraft({
                            ...formDraft,
                            attachChronoContext: e.target.checked,
                          })
                        }
                        className="rounded border-slate-700 bg-slate-900 text-red-500 focus:ring-0"
                      />
                      <Sparkles className="w-3 h-3 text-amber-400" />
                      <span>Attach Dasha & Hora</span>
                    </label>
                  </div>
                  <textarea
                    rows={3}
                    value={formDraft.description}
                    onChange={(e) =>
                      setFormDraft({ ...formDraft, description: e.target.value })
                    }
                    placeholder="Provide a detailed description..."
                    className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-white text-xs focus:border-red-500 focus:outline-none"
                  />
                  {formDraft.attachChronoContext && (
                    <div className="mt-1 text-[11px] font-mono text-slate-500 bg-slate-900/60 p-2 rounded border border-slate-800/80">
                      ⚡ Automatically embeds active planetary Hora ruler, Hermetic Axiom, and Vimshottari cycle signature into description.
                    </div>
                  )}
                </div>

                {/* Privacy & Category */}
                <div className="grid grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                      Privacy Status
                    </label>
                    <select
                      value={formDraft.privacyStatus}
                      onChange={(e: any) =>
                        setFormDraft({
                          ...formDraft,
                          privacyStatus: e.target.value,
                        })
                      }
                      className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-white text-xs font-mono focus:border-red-500 focus:outline-none"
                    >
                      <option value="private">Private (Default)</option>
                      <option value="unlisted">Unlisted</option>
                      <option value="public">Public</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                      Category
                    </label>
                    <select
                      value={formDraft.categoryID}
                      onChange={(e) =>
                        setFormDraft({
                          ...formDraft,
                          categoryID: e.target.value,
                        })
                      }
                      className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-white text-xs font-mono focus:border-red-500 focus:outline-none"
                    >
                      <option value="22">People & Blogs (22)</option>
                      <option value="27">Education (27)</option>
                      <option value="28">Science & Technology (28)</option>
                      <option value="24">Entertainment (24)</option>
                      <option value="10">Music (10)</option>
                      <option value="20">Gaming (20)</option>
                    </select>
                  </div>
                </div>

                {/* Tags */}
                <div>
                  <label className="block text-xs font-mono uppercase text-slate-400 mb-1">
                    Tags (Comma Separated)
                  </label>
                  <input
                    type="text"
                    value={tagsInput}
                    onChange={(e) => setTagsInput(e.target.value)}
                    placeholder="MercuryDasha, astrology, echosh"
                    className="w-full px-3 py-2 rounded-lg bg-slate-900 border border-slate-800 text-white text-xs font-mono focus:border-red-500 focus:outline-none"
                  />
                </div>

                {/* Submit Button */}
                <div className="pt-2">
                  <button
                    type="submit"
                    disabled={submitting || !status?.authenticated}
                    className={`w-full py-2.5 rounded-lg font-semibold text-xs flex items-center justify-center space-x-2 transition ${
                      status?.authenticated
                        ? "bg-red-600 hover:bg-red-500 text-white shadow-lg shadow-red-900/30"
                        : "bg-slate-800 text-slate-500 cursor-not-allowed"
                    }`}
                  >
                    {submitting ? (
                      <>
                        <RefreshCw className="w-4 h-4 animate-spin" />
                        <span>Dispatching Upload Job...</span>
                      </>
                    ) : (
                      <>
                        <Upload className="w-4 h-4" />
                        <span>Upload to YouTube</span>
                      </>
                    )}
                  </button>
                </div>
              </form>

              {/* Past Jobs Section */}
              {status?.recent_jobs && status.recent_jobs.length > 0 && (
                <div className="mt-6 pt-4 border-t border-slate-800">
                  <div className="text-xs font-mono uppercase text-slate-400 mb-3 flex items-center justify-between">
                    <span>Recent Upload History</span>
                    <button
                      onClick={refreshStatus}
                      className="p-1 rounded text-slate-400 hover:text-white"
                      title="Refresh jobs"
                    >
                      <RefreshCw className="w-3 h-3" />
                    </button>
                  </div>

                  <div className="space-y-2 max-h-48 overflow-y-auto pr-1">
                    {status.recent_jobs.map((j) => (
                      <div
                        key={j.id}
                        className="p-2.5 rounded-lg bg-slate-900 border border-slate-800/80 flex items-center justify-between text-xs"
                      >
                        <div className="truncate pr-2">
                          <div className="text-white font-medium truncate">
                            {j.title || j.file_name}
                          </div>
                          <div className="text-[11px] text-slate-400 font-mono">
                            {new Date(j.created_at).toLocaleDateString()} • {j.status}
                            {j.video_id && ` • ID: ${j.video_id}`}
                          </div>
                        </div>

                        <div className="flex items-center space-x-2 shrink-0">
                          {j.status === "completed" && j.video_url && (
                            <a
                              href={j.video_url}
                              target="_blank"
                              rel="noreferrer"
                              className="px-2 py-1 rounded bg-red-950/60 border border-red-800 text-red-300 hover:bg-red-900/60 text-xs font-mono flex items-center space-x-1"
                            >
                              <span>Watch</span>
                              <ExternalLink className="w-3 h-3" />
                            </a>
                          )}
                          {j.status === "uploading" && (
                            <button
                              onClick={() => cancelJob(j.id)}
                              className="px-2 py-1 rounded bg-slate-800 text-red-400 hover:bg-red-950 text-xs font-mono"
                            >
                              Cancel
                            </button>
                          )}
                        </div>
                      </div>
                    ))}
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      )}
    </>
  );
}
