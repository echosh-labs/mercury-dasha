"use client";

import React, { useEffect, useState } from "react";
import Link from "next/link";
import {
  Sun,
  Moon,
  Compass,
  Clock,
  MapPin,
  ChevronRight,
  Sparkles,
  Layers,
  Users,
  ArrowLeft,
} from "lucide-react";

// ── Hora Types (mirrored from /alignment for standalone use) ──────────────────

interface PlanetaryHour {
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
  brief_application: string;
  active: boolean;
}

interface PlanetaryDaySchedule {
  date: string;
  day_lord: string;
  day_lord_planet_id: string;
  sunrise: string;
  solar_noon: string;
  sunset: string;
  next_sunrise: string;
  morning_hour_duration_minutes?: number;
  afternoon_hour_duration_minutes?: number;
  day_hour_duration_minutes: number;
  night_hour_duration_minutes: number;
  hours: PlanetaryHour[];
  active_hour: PlanetaryHour;
  elapsed_minutes: number;
  remaining_minutes: number;
  progress_percent: number;
  next_hour_planet: string;
}

// ── Page Component ────────────────────────────────────────────────────────────

export default function EphemerisPage() {
  const [schedule, setSchedule] = useState<PlanetaryDaySchedule | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [filterMode, setFilterMode] = useState<"all" | "diurnal" | "nocturnal">("all");
  const [expandedHour, setExpandedHour] = useState<number | null>(null);
  const [locationName, setLocationName] = useState<string>("Observer Location");
  const [currentTime, setCurrentTime] = useState<Date>(new Date());

  // Live clock
  useEffect(() => {
    const timer = setInterval(() => setCurrentTime(new Date()), 1000);
    return () => clearInterval(timer);
  }, []);

  // Fetch schedule on mount + 60s refresh
  useEffect(() => {
    const fetchSchedule = async () => {
      try {
        // Load saved observer coordinates
        const settingsRes = await fetch("/api/v1/settings");
        let lat = "", lon = "";
        if (settingsRes.ok) {
          const s = await settingsRes.json();
          if (s.latitude && s.longitude) {
            lat = s.latitude.toString();
            lon = s.longitude.toString();
            setLocationName(s.location_name || `${s.latitude.toFixed(2)}°N, ${s.longitude.toFixed(2)}°E`);
          }
        }

        const params = new URLSearchParams();
        if (lat) params.set("lat", lat);
        if (lon) params.set("lon", lon);
        const qs = params.toString();
        const url = `/api/v1/hora/schedule${qs ? `?${qs}` : ""}`;

        const res = await fetch(url);
        if (res.ok) {
          const data: PlanetaryDaySchedule = await res.json();
          setSchedule(data);
        }
      } catch (err) {
        console.error("Failed to fetch ephemeris schedule:", err);
      } finally {
        setLoading(false);
      }
    };

    fetchSchedule();
    const interval = setInterval(fetchSchedule, 60_000);
    return () => clearInterval(interval);
  }, []);

  // ── Helpers ──────────────────────────────────────────────────────────────────

  const formatTime = (isoString?: string) => {
    if (!isoString) return "--:--";
    try {
      return new Date(isoString).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit", second: "2-digit" });
    } catch {
      return isoString;
    }
  };

  const formatTimeShort = (isoString?: string) => {
    if (!isoString) return "--:--";
    try {
      return new Date(isoString).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" });
    } catch {
      return isoString;
    }
  };

  const displayedHours = (schedule?.hours || []).filter((h) => {
    if (filterMode === "diurnal") return h.diurnal;
    if (filterMode === "nocturnal") return !h.diurnal;
    return true;
  });

  // ── Render ──────────────────────────────────────────────────────────────────

  return (
    <main className="min-h-screen bg-[#070b14] text-slate-100 font-sans pb-20 selection:bg-amber-500/30">
      {/* Header */}
      <header className="border-b border-slate-800/80 bg-[#0c1322]/80 backdrop-blur-md sticky top-0 z-40">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-3.5 flex flex-wrap items-center justify-between gap-3">
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
              <Sun className="w-5 h-5 text-amber-400" />
              <h1 className="text-base font-bold text-white tracking-tight">
                Daily Planetary Ephemeris
              </h1>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-amber-950 border border-amber-800 text-amber-300">
                24-HOUR CHART
              </span>
            </div>
          </div>

          <div className="flex items-center space-x-3 text-xs font-mono">
            <Link
              href="/alignment/"
              className="px-2.5 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-300 hover:text-purple-300 transition flex items-center space-x-1"
            >
              <Compass className="w-3.5 h-3.5" />
              <span>Deep Alignment</span>
              <ChevronRight className="w-3 h-3" />
            </Link>
            <Link
              href="/characters/"
              className="px-2.5 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-300 hover:text-cyan-300 transition flex items-center space-x-1"
            >
              <Users className="w-3.5 h-3.5" />
              <span>Characters</span>
              <ChevronRight className="w-3 h-3" />
            </Link>
          </div>
        </div>
      </header>

      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 pt-6 space-y-6">
        {loading || !schedule ? (
          <div className="p-16 flex flex-col items-center justify-center space-y-3 text-slate-400 bg-slate-900/40 rounded-2xl border border-slate-800">
            <Clock className="w-8 h-8 animate-spin text-amber-400" />
            <span className="text-sm font-mono">
              Computing topocentric ephemeris &amp; 24-hour planetary matrix...
            </span>
          </div>
        ) : (
          <>
            {/* Day Overview Ribbon */}
            <section className="flex flex-wrap items-center justify-between gap-3 px-5 py-4 rounded-2xl bg-gradient-to-r from-amber-950/30 via-[#111928] to-slate-900/80 border border-amber-800/30 text-xs font-mono">
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <span className="text-slate-400">Day Lord:</span>
                  <span className="text-lg font-bold text-white">{schedule.day_lord}</span>
                  <span className="text-amber-400">({schedule.day_lord_planet_id})</span>
                </div>
                <span className="text-slate-700">•</span>
                <span className="text-slate-300">{schedule.date}</span>
              </div>

              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-3">
                  <div className="text-center">
                    <span className="text-slate-500 block text-[10px]">SUNRISE</span>
                    <span className="text-slate-200 font-semibold">{formatTimeShort(schedule.sunrise)}</span>
                  </div>
                  <div className="text-center">
                    <span className="text-amber-400 block text-[10px] font-semibold">NOON</span>
                    <span className="text-amber-300 font-semibold">{formatTimeShort(schedule.solar_noon)}</span>
                  </div>
                  <div className="text-center">
                    <span className="text-slate-500 block text-[10px]">SUNSET</span>
                    <span className="text-slate-200 font-semibold">{formatTimeShort(schedule.sunset)}</span>
                  </div>
                </div>
                <span className="text-slate-700">•</span>
                <div className="flex items-center space-x-1.5">
                  <MapPin className="w-3.5 h-3.5 text-cyan-400" />
                  <span className="text-cyan-300">{locationName}</span>
                </div>
              </div>
            </section>

            {/* Active Hora Highlight */}
            <section
              className="p-5 rounded-2xl border flex flex-wrap items-center justify-between gap-4"
              style={{ borderColor: schedule.active_hour?.color_hex ? `${schedule.active_hour.color_hex}55` : "#1e293b", backgroundColor: "#0c132266" }}
            >
              <div className="flex items-center space-x-4">
                <div className="flex items-center space-x-2">
                  <span className="w-2.5 h-2.5 rounded-full bg-emerald-400 animate-pulse" />
                  <span className="text-[10px] font-mono text-emerald-400 uppercase tracking-wider font-bold">Active Hora</span>
                </div>
                <span className="text-2xl font-bold" style={{ color: schedule.active_hour?.color_hex || "#38bdf8" }}>
                  {schedule.active_hour?.planet_name}
                </span>
                <span className="text-lg font-bold text-slate-300">{schedule.active_hour?.metal_symbol}</span>
                <span className="text-xs font-mono text-slate-400">
                  {schedule.active_hour?.sacred_metal}
                </span>
              </div>
              <div className="flex items-center space-x-4 text-xs font-mono">
                <span className="text-slate-400">
                  {schedule.active_hour?.diurnal ? "☀️ Diurnal" : "🌙 Nocturnal"} H{schedule.active_hour?.index}
                </span>
                <span className="text-slate-600">•</span>
                <span className="text-cyan-400 font-semibold">{schedule.remaining_minutes.toFixed(0)}m remaining</span>
                <div className="w-24 h-1.5 bg-slate-800 rounded-full overflow-hidden">
                  <div
                    className="h-full bg-gradient-to-r from-cyan-500 to-emerald-400 transition-all duration-500"
                    style={{ width: `${schedule.progress_percent}%` }}
                  />
                </div>
                <span className="text-slate-500">{currentTime.toLocaleTimeString()}</span>
              </div>
            </section>

            {/* 24-Hour Ephemeris Grid */}
            <section className="space-y-4">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
                <div>
                  <h2 className="text-base font-bold text-white tracking-wide">
                    Complete 24-Hour Non-Linear Planetary Ephemeris
                  </h2>
                  <p className="text-xs text-slate-400">
                    Meeus astronomical anomaly with topocentric latitude ascension &amp; apparent solar meridian anchors.
                  </p>
                </div>

                <div className="flex items-center space-x-1.5 p-1 bg-slate-900 rounded-xl border border-slate-800 text-xs font-mono">
                  <button
                    onClick={() => setFilterMode("all")}
                    className={`px-3 py-1.5 rounded-lg transition ${
                      filterMode === "all" ? "bg-slate-800 text-cyan-300 font-bold" : "text-slate-400 hover:text-white"
                    }`}
                  >
                    All 24h
                  </button>
                  <button
                    onClick={() => setFilterMode("diurnal")}
                    className={`px-3 py-1.5 rounded-lg flex items-center space-x-1 transition ${
                      filterMode === "diurnal"
                        ? "bg-amber-950/70 text-amber-400 border border-amber-800 font-bold"
                        : "text-slate-400 hover:text-white"
                    }`}
                  >
                    <Sun className="w-3.5 h-3.5" />
                    <span>Diurnal (1-12)</span>
                  </button>
                  <button
                    onClick={() => setFilterMode("nocturnal")}
                    className={`px-3 py-1.5 rounded-lg flex items-center space-x-1 transition ${
                      filterMode === "nocturnal"
                        ? "bg-indigo-950/70 text-indigo-400 border border-indigo-800 font-bold"
                        : "text-slate-400 hover:text-white"
                    }`}
                  >
                    <Moon className="w-3.5 h-3.5" />
                    <span>Nocturnal (13-24)</span>
                  </button>
                </div>
              </div>

              {/* Timetable Blocks */}
              <div className="space-y-4">
                {/* Morning Diurnal (1-6) */}
                {(filterMode === "all" || filterMode === "diurnal") && (
                  <>
                    <div className="text-xs font-mono text-amber-400/90 flex items-center space-x-2 pt-2">
                      <Sun className="w-4 h-4 text-amber-400" />
                      <span className="font-bold">
                        MORNING DIURNAL PHASE (Sunrise → Solar Noon • L_morning:{" "}
                        {schedule.morning_hour_duration_minutes?.toFixed(1) ||
                          schedule.day_hour_duration_minutes.toFixed(1)}
                        m)
                      </span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3">
                      {displayedHours
                        .filter((h) => h.index >= 1 && h.index <= 6)
                        .map((h) => renderHourCard(h))}
                    </div>

                    {/* Solar Noon Meridian Marker */}
                    <div className="p-3 rounded-xl bg-gradient-to-r from-amber-950/60 via-amber-900/40 to-amber-950/60 border border-amber-600/40 flex items-center justify-between text-xs font-mono text-amber-200 shadow-md">
                      <div className="flex items-center space-x-2">
                        <Sun className="w-4 h-4 text-amber-400 animate-spin" />
                        <span className="font-bold tracking-wider">TRUE SOLAR NOON ANCHOR (Meridian Transit)</span>
                        <span className="text-slate-400 text-[11px]">• Hour 6 ends / Hour 7 begins</span>
                      </div>
                      <div className="text-amber-300 font-bold">
                        {formatTime(schedule.solar_noon)}
                      </div>
                    </div>

                    {/* Afternoon Diurnal (7-12) */}
                    <div className="text-xs font-mono text-amber-400/90 flex items-center space-x-2 pt-2">
                      <Sun className="w-4 h-4 text-amber-400" />
                      <span className="font-bold">
                        AFTERNOON DIURNAL PHASE (Solar Noon → Sunset • L_afternoon:{" "}
                        {schedule.afternoon_hour_duration_minutes?.toFixed(1) ||
                          schedule.day_hour_duration_minutes.toFixed(1)}
                        m)
                      </span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3">
                      {displayedHours
                        .filter((h) => h.index >= 7 && h.index <= 12)
                        .map((h) => renderHourCard(h))}
                    </div>
                  </>
                )}

                {/* Sunset Boundary Marker */}
                {filterMode === "all" && (
                  <div className="p-3 rounded-xl bg-gradient-to-r from-indigo-950/60 via-purple-950/40 to-indigo-950/60 border border-indigo-700/40 flex items-center justify-between text-xs font-mono text-indigo-200 shadow-md">
                    <div className="flex items-center space-x-2">
                      <Moon className="w-4 h-4 text-indigo-400" />
                      <span className="font-bold tracking-wider">APPARENT SUNSET (Day/Night Boundary)</span>
                      <span className="text-slate-400 text-[11px]">• Hour 12 ends / Hour 13 begins</span>
                    </div>
                    <div className="text-indigo-300 font-bold">{formatTime(schedule.sunset)}</div>
                  </div>
                )}

                {/* Nocturnal Block (13-24) */}
                {(filterMode === "all" || filterMode === "nocturnal") && (
                  <>
                    <div className="text-xs font-mono text-indigo-400/90 flex items-center space-x-2 pt-2">
                      <Moon className="w-4 h-4 text-indigo-400" />
                      <span className="font-bold">
                        NOCTURNAL PHASE (Sunset → Next Sunrise • L_night:{" "}
                        {schedule.night_hour_duration_minutes.toFixed(1)}m)
                      </span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-3">
                      {displayedHours
                        .filter((h) => h.index >= 13 && h.index <= 24)
                        .map((h) => renderHourCard(h))}
                    </div>
                  </>
                )}
              </div>
            </section>

            {/* Explore Deeper CTA */}
            <section className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <Link
                href="/alignment/"
                className="p-5 rounded-2xl bg-gradient-to-br from-purple-950/40 to-[#111928] border border-purple-800/40 hover:border-purple-500/60 transition group"
              >
                <div className="flex items-center space-x-3 mb-2">
                  <Compass className="w-5 h-5 text-purple-400 group-hover:rotate-45 transition-transform" />
                  <span className="font-bold text-white text-sm">Temporal Alignment &amp; Deep Inspection</span>
                  <ChevronRight className="w-4 h-4 text-purple-400 group-hover:translate-x-1 transition-transform" />
                </div>
                <p className="text-xs text-slate-400 font-sans leading-relaxed">
                  Drill into any hora with 7 non-linear oblique ascension sub-horas, Lagna arc traversal,
                  Vedic Asu time units, favorable &amp; cautionary guidance, and topocentric observer calibration.
                </p>
              </Link>

              <Link
                href="/characters/"
                className="p-5 rounded-2xl bg-gradient-to-br from-cyan-950/40 to-[#111928] border border-cyan-800/40 hover:border-cyan-500/60 transition group"
              >
                <div className="flex items-center space-x-3 mb-2">
                  <Users className="w-5 h-5 text-cyan-400 group-hover:scale-110 transition-transform" />
                  <span className="font-bold text-white text-sm">Character Sanctuary &amp; 120-Year Timeline</span>
                  <ChevronRight className="w-4 h-4 text-cyan-400 group-hover:translate-x-1 transition-transform" />
                </div>
                <p className="text-xs text-slate-400 font-sans leading-relaxed">
                  Create a sovereign character profile and discover your complete Vedic Panchanga, Nava Grahas,
                  Essential Dignities, and 120-year Vimshottari master timeline with video chronicles.
                </p>
              </Link>
            </section>
          </>
        )}
      </div>
    </main>
  );

  // ── Hour Card Renderer ──────────────────────────────────────────────────────

  function renderHourCard(h: PlanetaryHour) {
    const isCurrent = h.active;
    const isExpanded = expandedHour === h.index;

    return (
      <div key={h.index} className="space-y-0">
        <div
          onClick={() => setExpandedHour(isExpanded ? null : h.index)}
          className={`p-3.5 rounded-xl border text-left cursor-pointer transition-all ${
            isCurrent
              ? "bg-cyan-950/50 border-cyan-400 shadow-lg ring-1 ring-cyan-400/50"
              : isExpanded
              ? "bg-slate-800/90 border-slate-400 shadow-md"
              : "bg-[#0c1322] border-slate-800/90 hover:border-slate-700 hover:bg-slate-800/40"
          }`}
        >
          <div className="flex items-center justify-between">
            <span className="text-[11px] font-mono text-slate-400 flex items-center space-x-1">
              {h.diurnal ? <Sun className="w-3 h-3 text-amber-400" /> : <Moon className="w-3 h-3 text-indigo-400" />}
              <span>H{h.index}</span>
            </span>
            <span className="text-[10px] font-mono text-slate-500">{formatTimeShort(h.start_time)}</span>
          </div>

          <div className="flex items-center justify-between mt-2">
            <span className="font-bold text-sm" style={{ color: h.color_hex || "#38bdf8" }}>
              {h.planet_name}
            </span>
            <span className="text-xs font-bold text-slate-300">{h.metal_symbol}</span>
          </div>

          <div className="text-[10px] font-mono text-slate-400 mt-1.5 flex justify-between">
            <span>{h.duration_minutes.toFixed(0)}m</span>
            <span className="text-slate-500">{h.sanskrit_name}</span>
          </div>

          {isCurrent && (
            <div className="mt-2 text-[10px] font-mono font-bold text-emerald-400 flex items-center space-x-1">
              <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping" />
              <span>ACTIVE • {schedule?.remaining_minutes.toFixed(0)}m left</span>
            </div>
          )}
        </div>

        {/* Expanded Detail */}
        {isExpanded && (
          <div className="mt-1.5 p-3.5 rounded-xl bg-[#111928] border border-slate-700/80 text-xs font-mono space-y-2.5 animate-in fade-in duration-150">
            <div className="flex flex-wrap items-center gap-3 text-slate-300">
              <span>
                Metal: <strong className="text-amber-300">{h.sacred_metal}</strong> {h.metal_symbol}
              </span>
              <span className="text-slate-600">•</span>
              <span>
                Element: <strong className="text-slate-200">{h.element}</strong>
              </span>
              <span className="text-slate-600">•</span>
              <span>
                Chakra: <strong className="text-slate-200">{h.chakra_center}</strong>
              </span>
            </div>
            <div className="flex items-center justify-between text-slate-400">
              <div>
                <span className="text-slate-500">Opus Stage: </span>
                <span className="text-amber-400 font-bold">{h.magnum_opus_stage}</span>
              </div>
              <span className="text-[10px] text-slate-500">
                {formatTimeShort(h.start_time)} – {formatTimeShort(h.end_time)}
              </span>
            </div>
            <div className="text-cyan-300 italic">
              &ldquo;{h.hermetic_axiom}&rdquo;
            </div>
            {h.brief_application && (
              <div className="p-2.5 rounded-lg bg-cyan-950/30 border border-cyan-900/40">
                <div className="flex items-start space-x-2">
                  <Sparkles className="w-3.5 h-3.5 text-cyan-400 shrink-0 mt-0.5" />
                  <p className="text-slate-200 font-sans text-[11px] leading-relaxed">
                    {h.brief_application}
                  </p>
                </div>
              </div>
            )}
          </div>
        )}
      </div>
    );
  }
}
