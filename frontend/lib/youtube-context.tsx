"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";

export interface ChannelProfile {
  channel_id: string;
  title: string;
  description: string;
  custom_url: string;
  thumbnail_url: string;
  subscriber_count: number;
  video_count: number;
  uploads_playlist_id: string;
  last_synced_at: string;
}

export interface UploadJob {
  id: string;
  trigger_source: string;
  file_path: string;
  file_name: string;
  title?: string;
  total_bytes: number;
  bytes_uploaded: number;
  progress_pct: number;
  status: "queued" | "uploading" | "completed" | "failed" | "cancelled";
  video_id?: string;
  video_url?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
  completed_at?: string;
}

export interface QuotaTracker {
  daily_limit: number;
  used_today: number;
  reset_time_utc: string;
}

export interface YouTubeStatus {
  configured: boolean;
  authenticated: boolean;
  token_expiry?: string;
  channel?: ChannelProfile | null;
  recent_jobs?: UploadJob[];
  quota?: QuotaTracker;
}

export interface UploadDraft {
  filePath: string;
  title: string;
  description: string;
  tags: string[];
  categoryID: string;
  privacyStatus: "private" | "unlisted" | "public";
  madeForKids: boolean;
  embeddable: boolean;
  attachChronoContext: boolean;
}

interface YouTubeStudioContextType {
  status: YouTubeStatus | null;
  loading: boolean;
  isDrawerOpen: boolean;
  draft: UploadDraft | null;
  activeJobs: UploadJob[];
  openUploader: (initial?: Partial<UploadDraft>) => void;
  closeDrawer: () => void;
  refreshStatus: () => Promise<void>;
  startUpload: (draft: UploadDraft) => Promise<UploadJob | null>;
  cancelJob: (id: string) => Promise<void>;
  connectChannel: () => Promise<void>;
  disconnectChannel: () => Promise<void>;
}

const defaultDraft: UploadDraft = {
  filePath: "",
  title: "",
  description: "",
  tags: ["MercuryDasha", "echosh-labs"],
  categoryID: "22",
  privacyStatus: "private",
  madeForKids: false,
  embeddable: true,
  attachChronoContext: true,
};

const YouTubeStudioContext = createContext<YouTubeStudioContextType | undefined>(undefined);

export function YouTubeStudioProvider({ children }: { children: ReactNode }) {
  const [status, setStatus] = useState<YouTubeStatus | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [isDrawerOpen, setIsDrawerOpen] = useState<boolean>(false);
  const [draft, setDraft] = useState<UploadDraft | null>(null);
  const [activeJobs, setActiveJobs] = useState<UploadJob[]>([]);

  const refreshStatus = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/youtube/status");
      if (res.ok) {
        const data: YouTubeStatus = await res.json();
        setStatus(data);
        if (data.recent_jobs) {
          const ongoing = data.recent_jobs.filter(
            (j) => j.status === "uploading" || j.status === "queued"
          );
          setActiveJobs(ongoing);
        }
      }
    } catch (err) {
      console.error("Failed to fetch YouTube status:", err);
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    refreshStatus();
  }, [refreshStatus]);

  // Polling for active jobs
  useEffect(() => {
    if (activeJobs.length === 0) return;

    const interval = setInterval(async () => {
      try {
        const res = await fetch("/api/v1/youtube/jobs?limit=5");
        if (res.ok) {
          const data = await res.json();
          const jobs: UploadJob[] = data.jobs || [];
          const ongoing = jobs.filter(
            (j) => j.status === "uploading" || j.status === "queued"
          );
          setActiveJobs(ongoing);

          // If all completed, refresh overall status once
          if (ongoing.length === 0 && activeJobs.length > 0) {
            refreshStatus();
          }
        }
      } catch (err) {
        console.error("Error polling YouTube jobs:", err);
      }
    }, 2500);

    return () => clearInterval(interval);
  }, [activeJobs.length, refreshStatus]);

  const openUploader = useCallback((initial?: Partial<UploadDraft>) => {
    setDraft({
      ...defaultDraft,
      ...initial,
    });
    setIsDrawerOpen(true);
  }, []);

  const closeDrawer = useCallback(() => {
    setIsDrawerOpen(false);
  }, []);

  const connectChannel = useCallback(async () => {
    try {
      const res = await fetch("/api/v1/youtube/auth/url");
      if (!res.ok) {
        const err = await res.json();
        alert(`OAuth initialization failed: ${err.error || "Check credentials"}`);
        return;
      }
      const data = await res.json();
      if (data.auth_url) {
        window.location.href = data.auth_url;
      }
    } catch (err) {
      console.error("Connect channel error:", err);
      alert("Failed to contact backend for OAuth URL");
    }
  }, []);

  const disconnectChannel = useCallback(async () => {
    if (!confirm("Are you sure you want to disconnect your YouTube channel?")) return;
    try {
      const res = await fetch("/api/v1/youtube/auth/disconnect", { method: "POST" });
      if (res.ok) {
        await refreshStatus();
      }
    } catch (err) {
      console.error("Disconnect channel error:", err);
    }
  }, [refreshStatus]);

  const startUpload = useCallback(
    async (uploadDraft: UploadDraft): Promise<UploadJob | null> => {
      try {
        const payload = {
          file_path: uploadDraft.filePath,
          title: uploadDraft.title,
          description: uploadDraft.description,
          tags: uploadDraft.tags,
          category_id: uploadDraft.categoryID,
          privacy_status: uploadDraft.privacyStatus,
          made_for_kids: uploadDraft.madeForKids,
          embeddable: uploadDraft.embeddable,
          attach_chrono_context: uploadDraft.attachChronoContext,
        };

        const res = await fetch("/api/v1/youtube/upload?async=true", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(payload),
        });

        if (!res.ok) {
          const err = await res.json();
          throw new Error(err.error || "Upload initiation failed");
        }

        const job: UploadJob = await res.json();
        setActiveJobs((prev) => [job, ...prev.filter((j) => j.id !== job.id)]);
        await refreshStatus();
        return job;
      } catch (err: any) {
        console.error("Upload error:", err);
        alert(`Upload error: ${err.message}`);
        return null;
      }
    },
    [refreshStatus]
  );

  const cancelJob = useCallback(async (id: string) => {
    try {
      await fetch(`/api/v1/youtube/jobs/${id}/cancel`, { method: "POST" });
      setActiveJobs((prev) => prev.filter((j) => j.id !== id));
      await refreshStatus();
    } catch (err) {
      console.error("Cancel job error:", err);
    }
  }, [refreshStatus]);

  return (
    <YouTubeStudioContext.Provider
      value={{
        status,
        loading,
        isDrawerOpen,
        draft,
        activeJobs,
        openUploader,
        closeDrawer,
        refreshStatus,
        startUpload,
        cancelJob,
        connectChannel,
        disconnectChannel,
      }}
    >
      {children}
    </YouTubeStudioContext.Provider>
  );
}

export function useYouTubeStudio() {
  const ctx = useContext(YouTubeStudioContext);
  if (!ctx) {
    throw new Error("useYouTubeStudio must be used within a YouTubeStudioProvider");
  }
  return ctx;
}
