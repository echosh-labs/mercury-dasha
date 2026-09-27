"use client";

import React, { useEffect, useState } from "react";
import {
  Settings,
  Globe,
  MapPin,
  Clock,
  Compass,
  CheckCircle2,
  RefreshCw,
  X,
  Save,
  Sun,
  Moon,
  Sparkles,
  ShieldCheck,
  AlertCircle
} from "lucide-react";
import { DetectClientTemporalEnvironment, IANACentroids, DetectedLocation } from "@/lib/timezone";

interface AdminSettingsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onSettingsSaved?: (settings: DetectedLocation) => void;
}

export default function AdminSettingsModal({ isOpen, onClose, onSettingsSaved }: AdminSettingsModalProps) {
  const [settings, setSettings] = useState<DetectedLocation>({
    timezone: "America/Toronto",
    utc_offset_hours: -4.0,
    latitude: 43.1594,
    longitude: -79.2469,
    auto_detect: true,
    location_name: "St. Catharines, ON",
  });

  const [loading, setLoading] = useState<boolean>(true);
  const [detecting, setDetecting] = useState<boolean>(false);
  const [saving, setSaving] = useState<boolean>(false);
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);
  const [currentTime, setCurrentTime] = useState<Date>(new Date());
  const [solarTimes, setSolarTimes] = useState<any | null>(null);

  // Live seconds ticker
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  // Keyboard Escape dismissal per modern-web-guidance
  useEffect(() => {
    if (!isOpen) return;
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        onClose();
      }
    };
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [isOpen, onClose]);

  // Fetch initial settings from server or auto-detect on open
  useEffect(() => {
    if (!isOpen) return;

    const loadSettings = async () => {
      setLoading(true);
      try {
        const res = await fetch("/api/v1/settings");
        if (res.ok) {
          const serverSettings: DetectedLocation = await res.json();
          setSettings(serverSettings);
        } else {
          // Fallback to client auto-detection
          const detected = await DetectClientTemporalEnvironment();
          setSettings(detected);
        }
      } catch {
        const detected = await DetectClientTemporalEnvironment();
        setSettings(detected);
      } finally {
        setLoading(false);
      }
    };

    loadSettings();
  }, [isOpen]);

  // Fetch solar times whenever coordinates change
  useEffect(() => {
    if (!isOpen) return;

    const fetchSolar = async () => {
      try {
        const res = await fetch(`/api/v1/hora/active?lat=${settings.latitude}&lon=${settings.longitude}`);
        if (res.ok) {
          const data = await res.json();
          setSolarTimes(data);
        }
      } catch (err) {
        console.error("Failed to fetch solar preview:", err);
      }
    };

    fetchSolar();
  }, [isOpen, settings.latitude, settings.longitude]);

  if (!isOpen) return null;

  const handleAutoDetect = async () => {
    setDetecting(true);
    setSaveSuccess(false);
    try {
      const detected = await DetectClientTemporalEnvironment();
      setSettings(detected);
      // Automatically persist auto-detected settings
      await persistSettings(detected);
      setSaveSuccess(true);
      if (onSettingsSaved) onSettingsSaved(detected);
    } catch (err) {
      console.error("Auto-detect failed:", err);
    } finally {
      setDetecting(false);
    }
  };

  const persistSettings = async (toSave: DetectedLocation) => {
    const res = await fetch("/api/v1/settings", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(toSave),
    });
    if (!res.ok) {
      throw new Error("Failed to persist settings");
    }
    // Store in localStorage for fast clientside recovery
    localStorage.setItem("mercury_geo_lat", toSave.latitude.toString());
    localStorage.setItem("mercury_geo_lon", toSave.longitude.toString());
    localStorage.setItem("mercury_timezone", toSave.timezone);
    if (typeof window !== "undefined") {
      window.dispatchEvent(new Event("mercury_settings_updated"));
    }
  };

  const handleManualSave = async () => {
    setSaving(true);
    setSaveSuccess(false);
    try {
      await persistSettings(settings);
      setSaveSuccess(true);
      if (onSettingsSaved) onSettingsSaved(settings);
      setTimeout(() => setSaveSuccess(false), 4000);
    } catch (err) {
      console.error("Save failed:", err);
    } finally {
      setSaving(false);
    }
  };

  const selectQuickCity = (normTz: string) => {
    const centroid = IANACentroids[normTz];
    if (centroid) {
      setSettings((prev) => ({
        ...prev,
        timezone: normTz,
        latitude: centroid.latitude,
        longitude: centroid.longitude,
        location_name: centroid.name,
        auto_detect: false,
      }));
    }
  };

  const formatTimezoneString = (tz: string) => {
    try {
      return currentTime.toLocaleTimeString("en-US", { timeZone: tz, hour: "2-digit", minute: "2-digit", second: "2-digit" });
    } catch {
      return currentTime.toLocaleTimeString();
    }
  };

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="admin-settings-title"
      className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-black/80 backdrop-blur-md animate-in fade-in duration-200"
    >
      <div 
        className="relative w-full max-w-4xl max-h-[90vh] bg-[#0c1322] border border-cyan-900/60 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800/80 bg-slate-900/60">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-xl bg-purple-950/80 border border-purple-700/50 text-purple-400">
              <Settings className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h2 id="admin-settings-title" className="text-lg font-bold text-white tracking-tight">
                  Admin Portal & Temporal Settings Substrate
                </h2>
                <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-purple-950/70 border border-purple-800 text-purple-300">
                  SYSTEM CORE
                </span>
              </div>
              <p className="text-xs text-slate-400">
                Configure timezone detection, solar ephemeris coordinates, and substrate clock synchronization.
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-slate-800/80 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {loading ? (
          <div className="p-12 flex flex-col items-center justify-center space-y-3 text-slate-400">
            <RefreshCw className="w-8 h-8 animate-spin text-purple-400" />
            <span className="text-sm font-mono">Loading system substrate configuration...</span>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Live Diagnostics Card */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="p-4 rounded-xl bg-[#111928] border border-slate-800 space-y-2">
                <div className="flex items-center justify-between text-xs font-mono text-slate-400">
                  <span>LOCAL CLOCK</span>
                  <span className="text-emerald-400 flex items-center space-x-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    <span>SYNCHRONIZED</span>
                  </span>
                </div>
                <div className="text-2xl font-bold font-mono text-white tracking-wider">
                  {formatTimezoneString(settings.timezone)}
                </div>
                <div className="text-xs font-mono text-slate-400 truncate">
                  {settings.timezone} (UTC{settings.utc_offset_hours >= 0 ? `+${settings.utc_offset_hours}` : settings.utc_offset_hours})
                </div>
              </div>

              <div className="p-4 rounded-xl bg-[#111928] border border-slate-800 space-y-2">
                <div className="flex items-center justify-between text-xs font-mono text-slate-400">
                  <span>COORDINATES & CENTROID</span>
                  <MapPin className="w-3.5 h-3.5 text-cyan-400" />
                </div>
                <div className="text-base font-bold text-white truncate">
                  {settings.location_name}
                </div>
                <div className="text-xs font-mono text-cyan-400">
                  Lat: {settings.latitude.toFixed(4)}° • Lon: {settings.longitude.toFixed(4)}°
                </div>
              </div>

              <div className="p-4 rounded-xl bg-[#111928] border border-slate-800 space-y-2">
                <div className="flex items-center justify-between text-xs font-mono text-slate-400">
                  <span>ACTIVE HORA PREVIEW</span>
                  <Sparkles className="w-3.5 h-3.5 text-amber-400" />
                </div>
                {solarTimes && solarTimes.active_hour ? (
                  <>
                    <div className="flex items-center space-x-2">
                      <span className="text-xl font-bold" style={{ color: solarTimes.active_hour.color_hex || "#38bdf8" }}>
                        {solarTimes.active_hour.planet_name}
                      </span>
                      <span className="text-xs font-mono text-slate-400">
                        ({solarTimes.active_hour.metal_symbol} {solarTimes.active_hour.sacred_metal})
                      </span>
                    </div>
                    <div className="text-[11px] font-mono text-slate-400">
                      Day Lord: {solarTimes.day_lord} • {solarTimes.remaining_minutes?.toFixed(0)}m remaining
                    </div>
                  </>
                ) : (
                  <div className="text-xs text-slate-500 font-mono">Calculating...</div>
                )}
              </div>
            </div>

            {/* Auto-Detection & Instant Re-sync Banner */}
            <div className="p-4 rounded-2xl bg-gradient-to-r from-purple-950/40 via-[#111928] to-cyan-950/40 border border-purple-800/40 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div className="flex items-center space-x-3">
                <div className="p-2 rounded-xl bg-purple-900/60 text-purple-300">
                  <Globe className="w-5 h-5" />
                </div>
                <div>
                  <h4 className="text-sm font-bold text-white">Browser Auto-Detection Engine</h4>
                  <p className="text-xs text-slate-400">
                    Automatically derives your physical timezone and coordinates directly from device sensors.
                  </p>
                </div>
              </div>

              <button
                onClick={handleAutoDetect}
                disabled={detecting}
                className="flex items-center space-x-2 px-4 py-2.5 rounded-xl bg-purple-600 hover:bg-purple-500 text-white font-medium text-xs shadow-lg transition disabled:opacity-50 cursor-pointer flex-shrink-0"
              >
                <RefreshCw className={`w-3.5 h-3.5 ${detecting ? "animate-spin" : ""}`} />
                <span>{detecting ? "Detecting..." : "Re-Detect Location Now"}</span>
              </button>
            </div>

            {/* Manual Configuration Form */}
            <div className="p-5 rounded-2xl bg-[#111928] border border-slate-800 space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
                <h3 className="text-sm font-bold text-white flex items-center space-x-2">
                  <Compass className="w-4 h-4 text-cyan-400" />
                  <span>Manual Timezone & Coordinate Override</span>
                </h3>
                <span className="text-xs font-mono text-slate-500">BoltDB Key: system:settings</span>
              </div>

              {/* Quick Select Buttons */}
              <div className="space-y-1.5">
                <label className="text-xs font-mono text-slate-400 block">Quick City Preset Centroids:</label>
                <div className="flex flex-wrap gap-1.5">
                  {[
                    { id: "america/toronto", label: "St. Catharines / Niagara (Sovereign Genesis)" },
                    { id: "america/new_york", label: "New York" },
                    { id: "america/chicago", label: "Chicago" },
                    { id: "america/los_angeles", label: "Los Angeles" },
                    { id: "europe/london", label: "London" },
                    { id: "europe/paris", label: "Paris" },
                    { id: "asia/tokyo", label: "Tokyo" },
                    { id: "asia/kolkata", label: "New Delhi" },
                    { id: "australia/sydney", label: "Sydney" },
                    { id: "utc", label: "Greenwich UTC" },
                  ].map((city) => (
                    <button
                      key={city.id}
                      type="button"
                      onClick={() => selectQuickCity(city.id)}
                      className={`px-2.5 py-1 rounded-lg text-xs font-mono border transition ${
                        settings.timezone.toLowerCase() === city.id
                          ? "bg-cyan-950/80 border-cyan-500 text-cyan-300 font-semibold"
                          : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white hover:bg-slate-800"
                      }`}
                    >
                      {city.label}
                    </button>
                  ))}
                </div>
              </div>

              {/* Input Grid */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 text-xs font-mono pt-2">
                <div>
                  <label className="text-slate-400 block mb-1">IANA Timezone</label>
                  <input
                    type="text"
                    value={settings.timezone}
                    onChange={(e) => setSettings({ ...settings, timezone: e.target.value, auto_detect: false })}
                    placeholder="e.g. America/New_York"
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="text-slate-400 block mb-1">UTC Offset (Hours)</label>
                  <input
                    type="number"
                    step="0.5"
                    value={settings.utc_offset_hours}
                    onChange={(e) => setSettings({ ...settings, utc_offset_hours: parseFloat(e.target.value) || 0, auto_detect: false })}
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="text-slate-400 block mb-1">Latitude (°N/S)</label>
                  <input
                    type="number"
                    step="0.0001"
                    value={settings.latitude}
                    onChange={(e) => setSettings({ ...settings, latitude: parseFloat(e.target.value) || 0, auto_detect: false })}
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                  />
                </div>

                <div>
                  <label className="text-slate-400 block mb-1">Longitude (°E/W)</label>
                  <input
                    type="number"
                    step="0.0001"
                    value={settings.longitude}
                    onChange={(e) => setSettings({ ...settings, longitude: parseFloat(e.target.value) || 0, auto_detect: false })}
                    className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="text-xs font-mono text-slate-400 block mb-1">Location Label</label>
                <input
                  type="text"
                  value={settings.location_name}
                  onChange={(e) => setSettings({ ...settings, location_name: e.target.value, auto_detect: false })}
                  placeholder="e.g. New York, USA"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-xs font-mono text-slate-200 focus:border-cyan-500 focus:outline-none"
                />
              </div>

              {/* Action Bar */}
              <div className="flex items-center justify-between pt-3 border-t border-slate-800/80">
                <div className="flex items-center space-x-2">
                  <button
                    onClick={handleManualSave}
                    disabled={saving}
                    className="flex items-center space-x-2 px-5 py-2 rounded-xl bg-gradient-to-r from-cyan-600 to-emerald-600 hover:from-cyan-500 hover:to-emerald-500 text-white font-medium text-xs shadow transition disabled:opacity-50 cursor-pointer"
                  >
                    <Save className="w-3.5 h-3.5" />
                    <span>{saving ? "Saving..." : "Save to BoltDB Substrate"}</span>
                  </button>

                  <button
                    onClick={onClose}
                    className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs transition"
                  >
                    Done
                  </button>
                </div>

                {saveSuccess && (
                  <div className="flex items-center space-x-1 text-emerald-400 text-xs font-mono">
                    <CheckCircle2 className="w-4 h-4" />
                    <span>Settings persisted to BoltDB</span>
                  </div>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
