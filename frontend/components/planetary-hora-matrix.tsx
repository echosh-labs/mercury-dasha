"use client";

import React, { useEffect, useState, useTransition } from "react";
import Link from "next/link";
import {
  Compass,
  Clock,
  Sun,
  Moon,
  Sparkles,
  CheckCircle2,
  XCircle,
  Layers,
  MapPin,
  Gauge,
  Zap,
  RotateCcw,
  Settings,
  ChevronRight,
  ShieldCheck,
  ChevronDown,
  Globe,
  RefreshCw,
  Save,
  ArrowLeft,
  Users
} from "lucide-react";
import { DetectClientTemporalEnvironment, IANACentroids, type DetectedLocation } from "@/lib/timezone";

export interface SubHora {
  index: number;
  planet_id: string;
  planet_name: string;
  sanskrit_name: string;
  color_hex: string;
  metal_symbol: string;
  sacred_metal: string;
  start_time: string;
  end_time: string;
  duration_minutes: number;
  lagna_arc_start: number;
  lagna_arc_end: number;
  lagna_arc_traversed: number;
  asu_count: number;
  active: boolean;
}

export interface PlanetaryHour {
  index: number;
  diurnal: boolean;
  phase: string;
  planet_id: string;
  planet_name: string;
  sanskrit_name: string;
  start_time: string;
  end_time: string;
  duration_minutes: number;
  sacred_metal: string;
  metal_symbol: string;
  element: string;
  chakra_center: string;
  color_hex: string;
  hermetic_axiom: string;
  magnum_opus_stage: string;
  suitable_activities: string[];
  unsuitable_activities: string[];
  brief_application: string;
  sub_horas?: SubHora[];
  active_sub_hora?: SubHora;
  current_lagna_degree?: number;
  local_apparent_solar_time?: string;
  equation_of_time_minutes?: number;
  active: boolean;
}

export interface DashaResonance {
  active_mahadasha: string;
  active_antardasha: string;
  resonance_type: string;
  resonance_title: string;
  description: string;
  harmonic_multiplier: number;
}

export interface PlanetaryDaySchedule {
  date: string;
  day_lord: string;
  day_lord_planet_id: string;
  sunrise: string;
  solar_noon: string;
  sunset: string;
  next_sunrise: string;
  solar_midnight?: string;
  equation_of_time_minutes?: number;
  local_apparent_solar_time?: string;
  current_lagna_degree?: number;
  day_hour_duration_minutes: number;
  night_hour_duration_minutes: number;
  morning_hour_duration_minutes?: number;
  afternoon_hour_duration_minutes?: number;
  hours: PlanetaryHour[];
  active_hour: PlanetaryHour;
  elapsed_minutes: number;
  remaining_minutes: number;
  progress_percent: number;
  next_hour_planet: string;
  next_hour_start_time: string;
  dasha_resonance?: DashaResonance;
}

export interface PlanetaryHoraMatrixProps {
  activeProfileId?: string;
  onSelectProfile?: (id: string) => void;
  onNavigateSegment?: (segment: "dasha" | "hora" | "ephemeris" | "characters") => void;
  hideHeaderNav?: boolean;
}

export default function PlanetaryHoraMatrix({
  activeProfileId: propProfileId,
  onSelectProfile,
  onNavigateSegment,
  hideHeaderNav = false
}: PlanetaryHoraMatrixProps = {}) {
  const [schedule, setSchedule] = useState<PlanetaryDaySchedule | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [selectedHour, setSelectedHour] = useState<PlanetaryHour | null>(null);

  const [profileList, setProfileList] = useState<{ id: string; name: string }[]>([]);
  const [activeProfileId, setActiveProfileId] = useState<string>(propProfileId || "");
  const [showConfig, setShowConfig] = useState<boolean>(false);

  useEffect(() => {
    if (propProfileId !== undefined && propProfileId !== activeProfileId) {
      setActiveProfileId(propProfileId);
    }
  }, [propProfileId]);

  // Topocentric Settings State
  const [settings, setSettings] = useState<DetectedLocation>({
    timezone: "America/Toronto",
    utc_offset_hours: -4.0,
    latitude: 43.1594,
    longitude: -79.2469,
    auto_detect: true,
    location_name: "St. Catharines, ON",
  });
  const [detecting, setDetecting] = useState<boolean>(false);
  const [savingSettings, setSavingSettings] = useState<boolean>(false);
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);
  const [currentTime, setCurrentTime] = useState<Date>(new Date());

  // Live seconds ticker
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  // Fetch profiles on mount
  useEffect(() => {
    const fetchProfiles = async () => {
      try {
        const res = await fetch("/api/v1/profiles");
        if (res.ok) {
          const data = await res.json();
          if (data.summaries && data.summaries.length > 0) {
            setProfileList(data.summaries.map((s: any) => ({ id: s.id, name: s.name })));
          } else if (data.profiles && data.profiles.length > 0) {
            setProfileList(data.profiles.map((p: string) => ({ id: p, name: p })));
          }
          const stored = localStorage.getItem("mercury_active_profile");
          if (stored) setActiveProfileId(stored);
        }
      } catch (err) {
        console.error("Failed to load profiles:", err);
      }
    };

    const fetchSettings = async () => {
      try {
        const res = await fetch("/api/v1/settings");
        if (res.ok) {
          const s = await res.json();
          setSettings(s);
        }
      } catch (err) {
        console.error("Failed to load settings:", err);
      }
    };

    fetchProfiles();
    fetchSettings();
  }, []);

  // Fetch schedule whenever profile or coordinates change
  const fetchSchedule = async () => {
    setLoading(true);
    try {
      let url = "/api/v1/hora/schedule";
      const params = new URLSearchParams();
      if (activeProfileId) params.set("profile_id", activeProfileId);
      if (settings.latitude && settings.longitude) {
        params.set("lat", settings.latitude.toString());
        params.set("lon", settings.longitude.toString());
      }
      const queryString = params.toString();
      if (queryString) url += `?${queryString}`;

      const res = await fetch(url);
      if (res.ok) {
        const data: PlanetaryDaySchedule = await res.json();
        setSchedule(data);
        setSelectedHour(data.active_hour || (data.hours && data.hours[0]) || null);
      }
    } catch (err) {
      console.error("Failed to fetch planetary schedule:", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchSchedule();
  }, [activeProfileId, settings.latitude, settings.longitude]);

  const handleAutoDetect = async () => {
    setDetecting(true);
    setSaveSuccess(false);
    try {
      const detected = await DetectClientTemporalEnvironment();
      setSettings(detected);
      await persistSettings(detected);
      setSaveSuccess(true);
    } catch (err) {
      console.error("Auto-detect failed:", err);
    } finally {
      setDetecting(false);
    }
  };

  const persistSettings = async (toSave: DetectedLocation) => {
    setSavingSettings(true);
    try {
      const res = await fetch("/api/v1/settings", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(toSave),
      });
      if (res.ok) {
        localStorage.setItem("mercury_geo_lat", toSave.latitude.toString());
        localStorage.setItem("mercury_geo_lon", toSave.longitude.toString());
        localStorage.setItem("mercury_timezone", toSave.timezone);
        setSaveSuccess(true);
        setTimeout(() => setSaveSuccess(false), 3000);
      }
    } catch (err) {
      console.error("Failed to persist settings:", err);
    } finally {
      setSavingSettings(false);
    }
  };

  const selectQuickCity = (normTz: string) => {
    const centroid = IANACentroids[normTz];
    if (centroid) {
      const updated: DetectedLocation = {
        ...settings,
        timezone: normTz,
        latitude: centroid.latitude,
        longitude: centroid.longitude,
        location_name: centroid.name,
        auto_detect: false,
      };
      setSettings(updated);
      persistSettings(updated);
    }
  };

  const formatTime = (isoString?: string) => {
    if (!isoString) return "--:--";
    try {
      const date = new Date(isoString);
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
    } catch {
      return isoString;
    }
  };

  const formatTimeShort = (isoString?: string) => {
    if (!isoString) return "--:--";
    try {
      const date = new Date(isoString);
      return date.toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    } catch {
      return isoString;
    }
  };

  return (
    <div className="space-y-6">
      {/* Route Sub-Header & Controls */}
      {!hideHeaderNav && (
        <div className="p-4 sm:p-5 rounded-2xl bg-[#0c1322]/90 border border-slate-800 shadow-xl flex flex-wrap items-center justify-between gap-3">
          <div className="flex items-center space-x-3">
            <Link
              href="/"
              className="p-2 rounded-xl bg-slate-800/80 hover:bg-slate-700/80 text-slate-300 hover:text-white transition flex items-center space-x-1 text-xs font-mono"
            >
              <ArrowLeft className="w-4 h-4" />
              <span>Observatory</span>
            </Link>
            <div className="h-4 w-px bg-slate-800" />
            <div className="flex items-center space-x-2">
              <Compass className="w-5 h-5 text-cyan-400" />
              <h1 className="text-base font-bold text-white tracking-tight">
                Temporal Alignment &amp; Planetary Ephemeris
              </h1>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-cyan-950 border border-cyan-800 text-cyan-300">
                24-HOUR CHALDEAN CLOCK
              </span>
            </div>
          </div>

          <div className="flex items-center space-x-3 text-xs font-mono">
            {/* Character Profile Selector */}
            <div className="flex items-center space-x-1.5 bg-slate-900 border border-slate-800 rounded-lg px-2.5 py-1.5">
              <Users className="w-3.5 h-3.5 text-cyan-400" />
              <select
                value={activeProfileId}
                onChange={(e) => {
                  const val = e.target.value;
                  setActiveProfileId(val);
                  if (onSelectProfile) onSelectProfile(val);
                  if (typeof window !== "undefined") {
                    localStorage.setItem("mercury_active_profile", val);
                  }
                }}
                className="bg-transparent text-slate-200 focus:outline-none cursor-pointer"
              >
                <option value="">Sovereign Default (Observer)</option>
                {profileList.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
            </div>

            {/* Topocentric Config Toggle */}
            <button
              onClick={() => setShowConfig(!showConfig)}
              className={`px-3 py-1.5 rounded-lg border transition flex items-center space-x-1.5 ${
                showConfig
                  ? "bg-cyan-950 border-cyan-600 text-cyan-300"
                  : "bg-slate-900 border-slate-800 text-slate-300 hover:text-white"
              }`}
            >
              <Settings className="w-3.5 h-3.5 text-cyan-400" />
              <span>Calibrate Location</span>
              <ChevronDown className={`w-3 h-3 transition-transform ${showConfig ? "rotate-180" : ""}`} />
            </button>
          </div>
        </div>
      )}

      <div className="space-y-6">
        {/* Topocentric Calibration Drawer (when opened) */}
        {showConfig && (
          <section className="p-5 rounded-2xl bg-gradient-to-br from-[#0c1322] to-[#111928] border border-purple-800/60 shadow-xl space-y-4 animate-in fade-in duration-200">
            <div className="flex items-center justify-between border-b border-slate-800/80 pb-3">
              <div className="flex items-center space-x-2">
                <Globe className="w-4 h-4 text-purple-400" />
                <h3 className="text-sm font-bold text-white">
                  Topocentric Observer Coordinates &amp; Temporal Engine Calibration
                </h3>
              </div>
              <div className="flex items-center space-x-2">
                <button
                  onClick={handleAutoDetect}
                  disabled={detecting}
                  className="px-3 py-1 rounded-lg bg-purple-950/70 hover:bg-purple-900/80 border border-purple-700/60 text-purple-300 text-xs font-mono flex items-center space-x-1 transition disabled:opacity-50"
                >
                  <RefreshCw className={`w-3 h-3 ${detecting ? "animate-spin" : ""}`} />
                  <span>{detecting ? "Detecting..." : "Browser Auto-Detect"}</span>
                </button>
                {saveSuccess && (
                  <span className="text-xs font-mono text-emerald-400 flex items-center space-x-1">
                    <CheckCircle2 className="w-3 h-3" />
                    <span>Saved!</span>
                  </span>
                )}
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-4 gap-4 text-xs font-mono">
              <div>
                <label className="text-slate-400 block mb-1">Location Label</label>
                <input
                  type="text"
                  value={settings.location_name}
                  onChange={(e) => setSettings({ ...settings, location_name: e.target.value })}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-slate-200 focus:border-purple-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-slate-400 block mb-1">Latitude (°N)</label>
                <input
                  type="number"
                  step="0.0001"
                  value={settings.latitude}
                  onChange={(e) => setSettings({ ...settings, latitude: parseFloat(e.target.value) || 0 })}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-slate-200 focus:border-purple-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-slate-400 block mb-1">Longitude (°E)</label>
                <input
                  type="number"
                  step="0.0001"
                  value={settings.longitude}
                  onChange={(e) => setSettings({ ...settings, longitude: parseFloat(e.target.value) || 0 })}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-slate-200 focus:border-purple-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-slate-400 block mb-1">IANA Timezone</label>
                <input
                  type="text"
                  value={settings.timezone}
                  onChange={(e) => setSettings({ ...settings, timezone: e.target.value })}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-1.5 text-slate-200 focus:border-purple-500 focus:outline-none"
                />
              </div>
            </div>

            {/* Quick Cities */}
            <div className="flex flex-wrap items-center gap-2 pt-2 border-t border-slate-800/80">
              <span className="text-[11px] font-mono text-slate-400">Quick Centroids:</span>
              {Object.keys(IANACentroids).slice(0, 8).map((key) => {
                const c = IANACentroids[key];
                return (
                  <button
                    key={key}
                    onClick={() => selectQuickCity(key)}
                    className={`px-2.5 py-1 rounded text-[11px] font-mono border transition ${
                      settings.timezone === key
                        ? "bg-purple-900/60 border-purple-500 text-purple-200"
                        : "bg-slate-900 border-slate-800 text-slate-400 hover:text-slate-200"
                    }`}
                  >
                    {c.name}
                  </button>
                );
              })}
              <button
                onClick={() => persistSettings(settings)}
                disabled={savingSettings}
                className="ml-auto px-4 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-mono font-bold transition flex items-center space-x-1.5"
              >
                <Save className="w-3.5 h-3.5" />
                <span>{savingSettings ? "Persisting..." : "Apply & Persist"}</span>
              </button>
            </div>
          </section>
        )}

        {/* Top Observer Anchor Status Ribbon */}
        <section className="flex flex-wrap items-center justify-between gap-3 px-5 py-3 rounded-2xl bg-gradient-to-r from-cyan-950/40 via-[#111928] to-slate-900/80 border border-cyan-800/40 text-xs font-mono">
          <div className="flex items-center space-x-2">
            <MapPin className="w-4 h-4 text-cyan-400" />
            <span className="text-slate-400">Observer Grounding:</span>
            <span className="text-white font-bold">{settings.location_name}</span>
            <span className="text-slate-600">•</span>
            <span className="text-cyan-300">
              ({settings.latitude.toFixed(4)}°N, {settings.longitude.toFixed(4)}°E)
            </span>
            <span className="text-slate-600">•</span>
            <span className="text-slate-400">{settings.timezone}</span>
          </div>

          <div className="flex items-center space-x-3">
            <span className="text-slate-400">Local Clock:</span>
            <span className="text-white font-bold">{currentTime.toLocaleTimeString()}</span>
            <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-950/80 border border-emerald-700/60 text-emerald-300">
              TOPOCENTRIC CALIBRATED
            </span>
          </div>
        </section>

        {loading || !schedule ? (
          <div className="p-16 flex flex-col items-center justify-center space-y-3 text-slate-400 bg-slate-900/40 rounded-2xl border border-slate-800">
            <Clock className="w-8 h-8 animate-spin text-cyan-400" />
            <span className="text-sm font-mono">
              Computing topocentric ephemeris, oblique ascension &amp; 24-hour planetary matrix...
            </span>
          </div>
        ) : (
          <>
            {/* Top Stat Matrix: 4 Strategic Metric Tiles */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
              {/* Tile 1: Day Lord & Solar Milestones */}
              <div className="p-5 rounded-2xl bg-[#0c1322] border border-slate-800 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-slate-400 font-mono tracking-wider uppercase">DAY LORD</span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-amber-500/10 text-amber-400 border border-amber-500/30 font-mono">
                    {schedule.date}
                  </span>
                </div>
                <div className="flex items-center space-x-2">
                  <span className="text-2xl font-bold text-white">{schedule.day_lord}</span>
                  <span className="text-xs text-cyan-400 font-mono">({schedule.day_lord_planet_id})</span>
                </div>
                <div className="grid grid-cols-3 gap-1 text-[10px] font-mono text-slate-400 pt-2 border-t border-slate-800/80">
                  <div>
                    <span className="text-slate-500 block">SUNRISE</span>
                    <span className="text-slate-200 font-semibold">{formatTimeShort(schedule.sunrise)}</span>
                  </div>
                  <div>
                    <span className="text-amber-400 block font-semibold">NOON</span>
                    <span className="text-amber-300 font-semibold">{formatTimeShort(schedule.solar_noon)}</span>
                  </div>
                  <div>
                    <span className="text-slate-500 block">SUNSET</span>
                    <span className="text-slate-200 font-semibold">{formatTimeShort(schedule.sunset)}</span>
                  </div>
                </div>
              </div>

              {/* Tile 2: Astronomical Solar Corrections */}
              <div className="p-5 rounded-2xl bg-[#0c1322] border border-slate-800 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-cyan-400 font-mono tracking-wider uppercase flex items-center space-x-1">
                    <Gauge className="w-3.5 h-3.5" />
                    <span>SOLAR CORRECTIONS</span>
                  </span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-cyan-950 text-cyan-300 border border-cyan-800 font-mono">
                    LAST / EoT
                  </span>
                </div>
                <div>
                  <div className="text-xs text-slate-400 font-mono">Apparent Solar Time (LAST):</div>
                  <div className="text-xl font-bold text-cyan-300 font-mono">
                    {schedule.local_apparent_solar_time || "--:--:--"}
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-2 text-[10px] font-mono text-slate-400 pt-2 border-t border-slate-800/80">
                  <div>
                    <span className="text-slate-500 block">EQ OF TIME</span>
                    <span className="text-slate-200 font-semibold">
                      {schedule.equation_of_time_minutes !== undefined
                        ? `${schedule.equation_of_time_minutes >= 0 ? "+" : ""}${schedule.equation_of_time_minutes.toFixed(2)}m`
                        : "--"}
                    </span>
                  </div>
                  <div>
                    <span className="text-slate-500 block">ASCENDANT</span>
                    <span className="text-emerald-400 font-semibold">
                      {schedule.current_lagna_degree !== undefined
                        ? `${schedule.current_lagna_degree.toFixed(2)}°`
                        : "--"}
                    </span>
                  </div>
                </div>
              </div>

              {/* Tile 3: Active Live Hora */}
              <div
                className="p-5 rounded-2xl bg-[#0c1322] border space-y-3"
                style={{ borderColor: schedule.active_hour?.color_hex ? `${schedule.active_hour.color_hex}55` : "#1e293b" }}
              >
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-slate-400 font-mono uppercase tracking-wider">ACTIVE HORA</span>
                  <span className="text-[10px] px-2 py-0.5 rounded-full font-mono bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 flex items-center space-x-1">
                    <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                    <span>LIVE</span>
                  </span>
                </div>
                <div className="flex items-center justify-between">
                  <div className="flex items-center space-x-2">
                    <span className="text-2xl font-bold" style={{ color: schedule.active_hour?.color_hex || "#38bdf8" }}>
                      {schedule.active_hour?.planet_name}
                    </span>
                    <span className="text-lg font-bold text-slate-300">
                      {schedule.active_hour?.metal_symbol}
                    </span>
                  </div>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700">
                    {schedule.active_hour?.diurnal ? "☀️ Diurnal" : "🌙 Nocturnal"} H{schedule.active_hour?.index}
                  </span>
                </div>
                <div className="space-y-1.5 pt-1">
                  <div className="flex justify-between text-[10px] font-mono text-slate-400">
                    <span>{schedule.elapsed_minutes.toFixed(0)}m elapsed</span>
                    <span className="text-cyan-400 font-semibold">{schedule.remaining_minutes.toFixed(0)}m remaining</span>
                  </div>
                  <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
                    <div
                      className="h-full bg-gradient-to-r from-cyan-500 to-emerald-400 transition-all duration-500"
                      style={{ width: `${schedule.progress_percent}%` }}
                    />
                  </div>
                </div>
              </div>

              {/* Tile 4: Active Oblique Sub-Hora */}
              <div className="p-5 rounded-2xl bg-[#0c1322] border border-slate-800 space-y-3">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-purple-400 font-mono uppercase tracking-wider flex items-center space-x-1">
                    <Zap className="w-3.5 h-3.5" />
                    <span>ACTIVE SUB-HORA</span>
                  </span>
                  <span className="text-[10px] px-2 py-0.5 rounded bg-purple-950 text-purple-300 border border-purple-800 font-mono">
                    1/7 LAGNA ARC
                  </span>
                </div>
                {schedule.active_hour?.active_sub_hora ? (
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-xl font-bold text-white">
                        {schedule.active_hour.active_sub_hora.planet_name}
                      </span>
                      <span className="text-base font-bold text-slate-300">
                        {schedule.active_hour.active_sub_hora.metal_symbol}
                      </span>
                      <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800 text-slate-300">
                        Sub {schedule.active_hour.active_sub_hora.index}/7
                      </span>
                    </div>
                    <div className="grid grid-cols-2 gap-2 text-[10px] font-mono text-slate-400 pt-2 mt-2 border-t border-slate-800/80">
                      <div>
                        <span className="text-slate-500 block">VEDIC ASU</span>
                        <span className="text-amber-400 font-semibold">
                          {schedule.active_hour.active_sub_hora.asu_count.toFixed(0)} Asu
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">LAGNA ARC</span>
                        <span className="text-emerald-400 font-semibold">
                          {schedule.active_hour.active_sub_hora.lagna_arc_start.toFixed(1)}°–{schedule.active_hour.active_sub_hora.lagna_arc_end.toFixed(1)}°
                        </span>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="text-xs text-slate-500 font-mono py-2">
                    Calculating oblique sub-interval...
                  </div>
                )}
              </div>
            </div>

            {/* Selected Hour Detailed Focus Panel */}
            {selectedHour && (
              <section
                className="p-6 rounded-2xl bg-gradient-to-br from-[#0c1322] to-[#111928] border space-y-5 shadow-2xl"
                style={{ borderColor: selectedHour.color_hex ? `${selectedHour.color_hex}55` : "#334155" }}
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800/80 pb-4">
                  <div className="flex items-center space-x-3">
                    <span className="text-2xl font-bold" style={{ color: selectedHour.color_hex || "#38bdf8" }}>
                      Hour {selectedHour.index}: {selectedHour.planet_name} ({selectedHour.sanskrit_name})
                    </span>
                    <span className="text-xs px-2.5 py-1 rounded bg-slate-800 text-slate-300 border border-slate-700 font-mono">
                      {formatTime(selectedHour.start_time)} – {formatTime(selectedHour.end_time)} ({selectedHour.duration_minutes.toFixed(1)} mins)
                    </span>
                    {selectedHour.active && (
                      <span className="text-xs px-2.5 py-1 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 font-mono font-bold animate-pulse">
                        ACTIVE NOW
                      </span>
                    )}
                  </div>

                  <div className="flex flex-wrap items-center gap-3 text-xs font-mono text-slate-400">
                    <span>
                      Metal: <strong className="text-slate-200">{selectedHour.sacred_metal}</strong> {selectedHour.metal_symbol}
                    </span>
                    <span>•</span>
                    <span>
                      Element: <strong className="text-slate-200">{selectedHour.element}</strong>
                    </span>
                    <span>•</span>
                    <span>
                      Chakra: <strong className="text-slate-200">{selectedHour.chakra_center}</strong>
                    </span>
                  </div>
                </div>

                {/* Brief Application */}
                <div className="p-4 rounded-xl bg-cyan-950/30 border border-cyan-900/50 flex items-start space-x-3">
                  <Sparkles className="w-5 h-5 text-cyan-400 shrink-0 mt-0.5" />
                  <div>
                    <span className="text-xs font-bold text-cyan-300 uppercase tracking-wider block font-mono">
                      Alchemical Living Application &amp; Archetypal Directive:
                    </span>
                    <p className="text-sm text-slate-200 mt-1 leading-relaxed font-medium">
                      {selectedHour.brief_application}
                    </p>
                  </div>
                </div>

                {/* 7 Non-Linear Sub-Horas */}
                {selectedHour.sub_horas && selectedHour.sub_horas.length > 0 && (
                  <div className="p-4 rounded-xl bg-[#080d1a] border border-slate-800/80 space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Layers className="w-4 h-4 text-purple-400" />
                        <span className="text-xs font-bold text-white font-mono uppercase tracking-wider">
                          7 Non-Linear Sub-Horas (Oblique Ascension Lagna Arc)
                        </span>
                      </div>
                      <span className="text-[10px] font-mono text-slate-400">
                        1 Asu = 4s = 1&apos; equatorial ascension
                      </span>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-7 gap-2.5">
                      {selectedHour.sub_horas.map((sub) => (
                        <div
                          key={sub.index}
                          className={`p-3 rounded-xl border text-left transition-all ${
                            sub.active
                              ? "bg-purple-950/60 border-purple-500 ring-2 ring-purple-500/40 shadow-lg"
                              : "bg-[#111928] border-slate-800 hover:border-slate-700"
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-[10px] font-mono text-slate-400">Sub {sub.index}</span>
                            <span className="text-xs font-bold text-slate-200">{sub.metal_symbol}</span>
                          </div>
                          <div
                            className="text-sm font-bold mt-1"
                            style={{ color: sub.color_hex || "#c084fc" }}
                          >
                            {sub.planet_name}
                          </div>
                          <div className="text-[10px] font-mono text-slate-300 mt-1">
                            {sub.duration_minutes.toFixed(2)}m
                          </div>
                          <div className="text-[10px] font-mono text-slate-500">
                            {formatTimeShort(sub.start_time)}–{formatTimeShort(sub.end_time)}
                          </div>
                          <div className="text-[10px] font-mono text-emerald-400 mt-1.5 flex justify-between border-t border-slate-800 pt-1">
                            <span>Lagna:</span>
                            <span>{sub.lagna_arc_start.toFixed(1)}°</span>
                          </div>
                          <div className="text-[10px] font-mono text-amber-400 flex justify-between">
                            <span>Asu:</span>
                            <span>{sub.asu_count.toFixed(0)}</span>
                          </div>
                          {sub.active && (
                            <div className="mt-1.5 text-[9px] font-mono text-purple-300 font-bold bg-purple-900/80 px-1 py-0.5 rounded text-center animate-pulse">
                              ACTIVE NOW
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Guidance Matrix: Favorable vs Unfavorable */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
                  <div className="p-4 rounded-xl bg-slate-900/60 border border-emerald-900/40 space-y-2.5">
                    <span className="font-bold text-emerald-400 flex items-center space-x-1.5 font-mono">
                      <CheckCircle2 className="w-4 h-4" />
                      <span>FAVORABLE &amp; HARMONIC ENDEAVORS</span>
                    </span>
                    <ul className="space-y-1.5 text-slate-300">
                      {selectedHour.suitable_activities.map((act, idx) => (
                        <li key={idx} className="flex items-start space-x-2">
                          <span className="text-emerald-500 mt-0.5 font-bold">•</span>
                          <span>{act}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div className="p-4 rounded-xl bg-slate-900/60 border border-rose-900/40 space-y-2.5">
                    <span className="font-bold text-rose-400 flex items-center space-x-1.5 font-mono">
                      <XCircle className="w-4 h-4" />
                      <span>CAUTIONARY &amp; DISSONANT ACTIVITIES</span>
                    </span>
                    <ul className="space-y-1.5 text-slate-300">
                      {selectedHour.unsuitable_activities.map((act, idx) => (
                        <li key={idx} className="flex items-start space-x-2">
                          <span className="text-rose-500 mt-0.5 font-bold">•</span>
                          <span>{act}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>

                {/* Alchemical Metaphysics Footer */}
                <div className="flex flex-wrap items-center justify-between text-xs font-mono text-slate-400 pt-3 border-t border-slate-800/80">
                  <div>
                    <span className="text-slate-500">Magnum Opus Stage: </span>
                    <span className="text-amber-400 font-bold">{selectedHour.magnum_opus_stage}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Hermetic Principle: </span>
                    <span className="text-cyan-300 italic">&ldquo;{selectedHour.hermetic_axiom}&rdquo;</span>
                  </div>
                </div>
              </section>
            )}

            {/* Link to Full 24-Hour Ephemeris Route */}
            <Link
              href="/ephemeris/"
              className="block p-5 rounded-2xl bg-gradient-to-r from-amber-950/30 via-[#111928] to-amber-950/30 border border-amber-800/40 hover:border-amber-500/60 transition group"
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center space-x-3">
                  <Sun className="w-5 h-5 text-amber-400 group-hover:rotate-12 transition-transform" />
                  <div>
                    <span className="font-bold text-white text-sm block">
                      Complete 24-Hour Non-Linear Planetary Ephemeris
                    </span>
                    <span className="text-xs text-slate-400 font-sans">
                      View the full daily chart — all 24 planetary hours with sacred metals, Hermetic axioms, and cosmic alignment guidance.
                    </span>
                  </div>
                </div>
                <ChevronRight className="w-5 h-5 text-amber-400 group-hover:translate-x-1 transition-transform shrink-0" />
              </div>
            </Link>
          </>
        )}
      </div>
    </div>
  );
}
