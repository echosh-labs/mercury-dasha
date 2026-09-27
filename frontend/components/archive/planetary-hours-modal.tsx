"use client";

import React, { useEffect, useState } from "react";
import { 
  Clock, 
  Sun, 
  Moon, 
  Sparkles, 
  CheckCircle2, 
  XCircle, 
  X, 
  Layers, 
  Calendar, 
  Flame,
  ArrowRight,
  ShieldCheck,
  Compass,
  Gauge,
  Activity,
  Zap,
  MapPin
} from "lucide-react";

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

interface PlanetaryHoursModalProps {
  isOpen: boolean;
  onClose: () => void;
  profileId?: string;
}

export default function PlanetaryHoursModal({ isOpen, onClose, profileId }: PlanetaryHoursModalProps) {
  const [schedule, setSchedule] = useState<PlanetaryDaySchedule | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [selectedHour, setSelectedHour] = useState<PlanetaryHour | null>(null);
  const [filterMode, setFilterMode] = useState<"all" | "diurnal" | "nocturnal">("all");
  const [showSubHoras, setShowSubHoras] = useState<boolean>(true);

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

  useEffect(() => {
    if (!isOpen) return;

    const fetchSchedule = async () => {
      setLoading(true);
      try {
        let url = "/api/v1/hora/schedule";
        const params = new URLSearchParams();
        if (profileId) {
          params.set("profile_id", profileId);
        }
        if (typeof window !== "undefined") {
          const storedLat = localStorage.getItem("mercury_geo_lat");
          const storedLon = localStorage.getItem("mercury_geo_lon");
          if (storedLat && storedLon) {
            params.set("lat", storedLat);
            params.set("lon", storedLon);
          }
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

    fetchSchedule();
  }, [isOpen, profileId]);

  if (!isOpen) return null;

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

  const displayedHours = (schedule?.hours || []).filter((h) => {
    if (filterMode === "diurnal") return h.diurnal;
    if (filterMode === "nocturnal") return !h.diurnal;
    return true;
  });

  return (
    <div
      role="dialog"
      aria-modal="true"
      aria-labelledby="planetary-hours-title"
      className="fixed inset-0 z-50 flex items-center justify-center p-3 sm:p-6 bg-black/80 backdrop-blur-md animate-in fade-in duration-200"
    >
      <div 
        className="relative w-full max-w-5xl max-h-[92vh] bg-[#0c1322] border border-cyan-900/60 rounded-2xl shadow-2xl overflow-hidden flex flex-col font-sans"
        onClick={(e) => e.stopPropagation()}
      >
        {/* Header Strip */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-slate-800/80 bg-slate-900/60">
          <div className="flex items-center space-x-3">
            <div className="p-2 rounded-xl bg-cyan-950/80 border border-cyan-700/50 text-cyan-400">
              <Compass className="w-5 h-5" />
            </div>
            <div>
              <div className="flex items-center space-x-2">
                <h2 id="planetary-hours-title" className="text-lg font-bold text-white tracking-tight">
                  Alchemical Planetary Hours & Oblique Ascension Substrate
                </h2>
                <span className="text-xs font-mono px-2 py-0.5 rounded-full bg-cyan-950/70 border border-cyan-800 text-cyan-300">
                  NON-LINEAR EPHEMERIS
                </span>
              </div>
              <p className="text-xs text-slate-400">
                Astronomical temporal intervals dynamically anchored to Apparent Sunrise, True Solar Noon & Oblique Ascension (Lagna arc).
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

        {loading || !schedule ? (
          <div className="p-12 flex flex-col items-center justify-center space-y-3 text-slate-400">
            <Clock className="w-8 h-8 animate-spin text-cyan-400" />
            <span className="text-sm font-mono">Computing topocentric solar ephemeris & oblique ascension matrices...</span>
          </div>
        ) : (
          <div className="flex-1 overflow-y-auto p-6 space-y-6">
            {/* Observer Topocentric Anchor Strip */}
            <div className="flex flex-wrap items-center justify-between gap-2 px-4 py-2.5 rounded-xl bg-gradient-to-r from-cyan-950/40 via-[#111928] to-slate-900/60 border border-cyan-800/40 text-xs font-mono">
              <div className="flex items-center space-x-2">
                <MapPin className="w-3.5 h-3.5 text-cyan-400" />
                <span className="text-slate-400">Profile Anchor:</span>
                <span className="text-white font-bold">{profileId || "Sovereign Genesis"}</span>
                <span className="text-slate-600">•</span>
                <span className="text-cyan-300">📍 St. Catharines / Niagara (43.1594° N, 79.2469° W)</span>
                <span className="text-slate-600">•</span>
                <span className="text-slate-400">America/Toronto (UTC-4)</span>
              </div>
              <span className="text-[10px] px-2 py-0.5 rounded-full bg-emerald-950/80 border border-emerald-700/60 text-emerald-300">
                TOPOCENTRIC CALIBRATED
              </span>
            </div>

            {/* Top Overview Cards */}
            <div className="grid grid-cols-1 md:grid-cols-4 gap-3.5">
              {/* Day Lord & Solar Anchors */}
              <div className="p-4 rounded-xl bg-[#111928] border border-slate-800 space-y-2.5">
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
                <div className="grid grid-cols-3 gap-1 text-[10px] font-mono text-slate-400 pt-1 border-t border-slate-800/80">
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

              {/* Astronomical Corrections (EoT, LAST, Lagna) */}
              <div className="p-4 rounded-xl bg-[#111928] border border-slate-800 space-y-2.5">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-cyan-400 font-mono tracking-wider uppercase flex items-center space-x-1">
                    <Gauge className="w-3 h-3 text-cyan-400" />
                    <span>SOLAR CORRECTIONS</span>
                  </span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-cyan-950 text-cyan-300 border border-cyan-800 font-mono">
                    LAST / EoT
                  </span>
                </div>
                <div>
                  <div className="text-xs text-slate-400 font-mono">Apparent Solar Time (LAST):</div>
                  <div className="text-lg font-bold text-cyan-300 font-mono">
                    {schedule.local_apparent_solar_time || "--:--:--"}
                  </div>
                </div>
                <div className="grid grid-cols-2 gap-2 text-[10px] font-mono text-slate-400 pt-1 border-t border-slate-800/80">
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

              {/* Active Hora Status */}
              <div 
                className="p-4 rounded-xl bg-[#111928] border space-y-2.5"
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
                    <span 
                      className="text-xl font-bold"
                      style={{ color: schedule.active_hour?.color_hex || "#38bdf8" }}
                    >
                      {schedule.active_hour?.planet_name}
                    </span>
                    <span className="text-base font-bold text-slate-300">
                      {schedule.active_hour?.metal_symbol}
                    </span>
                  </div>
                  <span className="text-[10px] font-mono px-2 py-0.5 rounded bg-slate-800/80 text-slate-300 border border-slate-700">
                    {schedule.active_hour?.diurnal ? "☀️ Diurnal" : "🌙 Nocturnal"} H{schedule.active_hour?.index}
                  </span>
                </div>

                <div className="space-y-1">
                  <div className="flex justify-between text-[10px] font-mono text-slate-400">
                    <span>{schedule.elapsed_minutes.toFixed(0)}m elapsed</span>
                    <span className="text-cyan-400 font-semibold">{schedule.remaining_minutes.toFixed(0)}m left</span>
                  </div>
                  <div className="h-1.5 w-full bg-slate-800 rounded-full overflow-hidden">
                    <div 
                      className="h-full bg-gradient-to-r from-cyan-500 to-emerald-400 transition-all duration-500"
                      style={{ width: `${schedule.progress_percent}%` }}
                    />
                  </div>
                </div>
              </div>

              {/* Active Sub-Hora Dialectic */}
              <div className="p-4 rounded-xl bg-[#111928] border border-slate-800 space-y-2.5">
                <div className="flex items-center justify-between">
                  <span className="text-[10px] text-purple-400 font-mono uppercase tracking-wider flex items-center space-x-1">
                    <Zap className="w-3 h-3 text-purple-400" />
                    <span>ACTIVE SUB-HORA</span>
                  </span>
                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-950/70 text-purple-300 border border-purple-800/60 font-mono">
                    1/7 LAGNA ARC
                  </span>
                </div>
                {schedule.active_hour?.active_sub_hora ? (
                  <div>
                    <div className="flex items-center space-x-2">
                      <span className="text-lg font-bold text-white">
                        {schedule.active_hour.active_sub_hora.planet_name}
                      </span>
                      <span className="text-sm font-bold text-slate-300">
                        {schedule.active_hour.active_sub_hora.metal_symbol}
                      </span>
                      <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-slate-800 text-slate-300">
                        Sub {schedule.active_hour.active_sub_hora.index}/7
                      </span>
                    </div>
                    <div className="grid grid-cols-2 gap-1 text-[10px] font-mono text-slate-400 pt-1.5 mt-1 border-t border-slate-800/80">
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

            {/* Selected Hour Details & Sub-Hora Decomposition */}
            {selectedHour && (
              <div 
                className="p-5 rounded-2xl bg-gradient-to-br from-[#0f172a] to-[#111928] border space-y-4 shadow-lg"
                style={{ borderColor: selectedHour.color_hex ? `${selectedHour.color_hex}44` : "#334155" }}
              >
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-800 pb-3">
                  <div className="flex items-center space-x-3">
                    <span className="text-xl font-bold" style={{ color: selectedHour.color_hex || "#38bdf8" }}>
                      Hour {selectedHour.index}: {selectedHour.planet_name}
                    </span>
                    <span className="text-xs px-2 py-0.5 rounded bg-slate-800 text-slate-300 border border-slate-700 font-mono">
                      {formatTime(selectedHour.start_time)} – {formatTime(selectedHour.end_time)} ({selectedHour.duration_minutes.toFixed(1)} mins)
                    </span>
                    {selectedHour.active && (
                      <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 font-mono">
                        ACTIVE NOW
                      </span>
                    )}
                  </div>

                  <div className="flex items-center space-x-2 text-xs font-mono text-slate-400">
                    <span>Metal: <strong className="text-slate-200">{selectedHour.sacred_metal}</strong> {selectedHour.metal_symbol}</span>
                    <span>•</span>
                    <span>Element: <strong className="text-slate-200">{selectedHour.element}</strong></span>
                    <span>•</span>
                    <span>Chakra: <strong className="text-slate-200">{selectedHour.chakra_center}</strong></span>
                  </div>
                </div>

                {/* Brief Application */}
                <div className="p-3 rounded-xl bg-cyan-950/30 border border-cyan-900/50 flex items-start space-x-2.5">
                  <Sparkles className="w-4 h-4 text-cyan-400 flex-shrink-0 mt-0.5" />
                  <div>
                    <span className="text-xs font-semibold text-cyan-300 uppercase tracking-wider block font-mono">
                      Suitable Daily Living Applications:
                    </span>
                    <p className="text-xs text-slate-200 mt-0.5 leading-relaxed font-medium">
                      {selectedHour.brief_application}
                    </p>
                  </div>
                </div>

                {/* Non-Linear Sub-Horas (Oblique Ascension Dialectic) */}
                {selectedHour.sub_horas && selectedHour.sub_horas.length > 0 && (
                  <div className="p-3.5 rounded-xl bg-[#090e1a] border border-slate-800 space-y-2.5">
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

                    <div className="grid grid-cols-1 sm:grid-cols-7 gap-2">
                      {selectedHour.sub_horas.map((sub) => (
                        <div
                          key={sub.index}
                          className={`p-2 rounded-lg border text-left transition-all ${
                            sub.active
                              ? "bg-purple-950/50 border-purple-500 ring-1 ring-purple-500/50"
                              : "bg-[#111928] border-slate-800 hover:border-slate-700"
                          }`}
                        >
                          <div className="flex items-center justify-between">
                            <span className="text-[10px] font-mono text-slate-400">
                              Sub {sub.index}
                            </span>
                            <span className="text-[11px] font-bold text-slate-200">
                              {sub.metal_symbol}
                            </span>
                          </div>
                          <div 
                            className="text-xs font-bold mt-0.5"
                            style={{ color: sub.color_hex || "#c084fc" }}
                          >
                            {sub.planet_name}
                          </div>
                          <div className="text-[10px] font-mono text-slate-300 mt-1">
                            {sub.duration_minutes.toFixed(2)}m
                          </div>
                          <div className="text-[9px] font-mono text-slate-500 mt-0.5">
                            {formatTimeShort(sub.start_time)}–{formatTimeShort(sub.end_time)}
                          </div>
                          <div className="text-[9px] font-mono text-emerald-400 mt-1 flex justify-between border-t border-slate-800 pt-1">
                            <span>Lagna:</span>
                            <span>{sub.lagna_arc_start.toFixed(1)}°</span>
                          </div>
                          <div className="text-[9px] font-mono text-amber-400 flex justify-between">
                            <span>Asu:</span>
                            <span>{sub.asu_count.toFixed(0)}</span>
                          </div>
                          {sub.active && (
                            <div className="mt-1 text-[9px] font-mono text-purple-300 font-bold bg-purple-900/60 px-1 py-0.5 rounded text-center animate-pulse">
                              ACTIVE
                            </div>
                          )}
                        </div>
                      ))}
                    </div>
                  </div>
                )}

                {/* Detailed Guidance: Suitable & Unsuitable Activities */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
                  <div className="p-3 rounded-xl bg-slate-900/60 border border-emerald-900/30 space-y-2">
                    <span className="font-semibold text-emerald-400 flex items-center space-x-1.5 font-mono">
                      <CheckCircle2 className="w-3.5 h-3.5" />
                      <span>FAVORABLE & ALIGNED ACTIONS</span>
                    </span>
                    <ul className="space-y-1.5 text-slate-300">
                      {selectedHour.suitable_activities.map((act, idx) => (
                        <li key={idx} className="flex items-start space-x-1.5">
                          <span className="text-emerald-500 mt-0.5">•</span>
                          <span>{act}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div className="p-3 rounded-xl bg-slate-900/60 border border-rose-900/30 space-y-2">
                    <span className="font-semibold text-rose-400 flex items-center space-x-1.5 font-mono">
                      <XCircle className="w-3.5 h-3.5" />
                      <span>UNFAVORABLE / CAUTIONARY ADVISORIES</span>
                    </span>
                    <ul className="space-y-1.5 text-slate-300">
                      {selectedHour.unsuitable_activities.map((act, idx) => (
                        <li key={idx} className="flex items-start space-x-1.5">
                          <span className="text-rose-500 mt-0.5">•</span>
                          <span>{act}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>

                {/* Alchemical stage & Hermetic Axiom */}
                <div className="flex flex-wrap items-center justify-between text-[11px] font-mono text-slate-400 pt-2 border-t border-slate-800">
                  <div>
                    <span className="text-slate-500">Magnum Opus Stage: </span>
                    <span className="text-amber-400 font-semibold">{selectedHour.magnum_opus_stage}</span>
                  </div>
                  <div>
                    <span className="text-slate-500">Hermetic Axiom: </span>
                    <span className="text-cyan-300 italic">&ldquo;{selectedHour.hermetic_axiom}&rdquo;</span>
                  </div>
                </div>
              </div>
            )}

            {/* 24-Hour Timetable Header & Filter */}
            <div className="space-y-3">
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                <div>
                  <h3 className="text-sm font-bold text-white tracking-wide">
                    Complete 24-Hour Non-Linear Planetary Ephemeris
                  </h3>
                  <p className="text-xs text-slate-400">
                    Divided into 12 diurnal hours (L_morning & L_afternoon anchored at True Solar Noon) and 12 nocturnal hours.
                  </p>
                </div>

                <div className="flex items-center space-x-1 p-1 bg-slate-900 rounded-lg border border-slate-800 text-xs font-mono">
                  <button
                    onClick={() => setFilterMode("all")}
                    className={`px-3 py-1 rounded transition ${filterMode === "all" ? "bg-slate-800 text-cyan-300" : "text-slate-400 hover:text-white"}`}
                  >
                    All 24h
                  </button>
                  <button
                    onClick={() => setFilterMode("diurnal")}
                    className={`px-3 py-1 rounded flex items-center space-x-1 transition ${filterMode === "diurnal" ? "bg-amber-950/60 text-amber-400 border border-amber-800/50" : "text-slate-400 hover:text-white"}`}
                  >
                    <Sun className="w-3 h-3" />
                    <span>Day (1-12)</span>
                  </button>
                  <button
                    onClick={() => setFilterMode("nocturnal")}
                    className={`px-3 py-1 rounded flex items-center space-x-1 transition ${filterMode === "nocturnal" ? "bg-indigo-950/60 text-indigo-400 border border-indigo-800/50" : "text-slate-400 hover:text-white"}`}
                  >
                    <Moon className="w-3 h-3" />
                    <span>Night (13-24)</span>
                  </button>
                </div>
              </div>

              {/* Timetable Grid with Astronomical Meridian Anchor Markers */}
              <div className="space-y-3">
                {/* Morning Diurnal Block (Hours 1 to 6) */}
                {(filterMode === "all" || filterMode === "diurnal") && (
                  <>
                    <div className="text-xs font-mono text-amber-400/90 flex items-center space-x-2 pt-1">
                      <Sun className="w-3.5 h-3.5 text-amber-400" />
                      <span>MORNING DIURNAL PHASE (Sunrise → Solar Noon • L_morning: {schedule.morning_hour_duration_minutes?.toFixed(1) || schedule.day_hour_duration_minutes.toFixed(1)}m)</span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-2">
                      {displayedHours.filter(h => h.index >= 1 && h.index <= 6).map((h) => renderHourCard(h))}
                    </div>

                    {/* True Solar Noon Boundary Marker */}
                    <div className="my-2 p-2 rounded-xl bg-gradient-to-r from-amber-950/60 via-amber-900/40 to-amber-950/60 border border-amber-600/40 flex items-center justify-between text-xs font-mono text-amber-200">
                      <div className="flex items-center space-x-2">
                        <Sun className="w-4 h-4 text-amber-400 animate-spin" />
                        <span className="font-bold tracking-wider">TRUE SOLAR NOON ANCHOR (Meridian Transit)</span>
                        <span className="text-slate-400 text-[11px]">• Hour 6 ends / Hour 7 begins</span>
                      </div>
                      <div className="text-amber-300 font-bold">
                        {formatTime(schedule.solar_noon)} (LAST 12:00:00)
                      </div>
                    </div>

                    {/* Afternoon Diurnal Block (Hours 7 to 12) */}
                    <div className="text-xs font-mono text-amber-400/90 flex items-center space-x-2 pt-1">
                      <Sun className="w-3.5 h-3.5 text-amber-400" />
                      <span>AFTERNOON DIURNAL PHASE (Solar Noon → Sunset • L_afternoon: {schedule.afternoon_hour_duration_minutes?.toFixed(1) || schedule.day_hour_duration_minutes.toFixed(1)}m)</span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-2">
                      {displayedHours.filter(h => h.index >= 7 && h.index <= 12).map((h) => renderHourCard(h))}
                    </div>
                  </>
                )}

                {/* Sunset Boundary Marker */}
                {filterMode === "all" && (
                  <div className="my-2 p-2 rounded-xl bg-gradient-to-r from-indigo-950/60 via-purple-950/40 to-indigo-950/60 border border-indigo-700/40 flex items-center justify-between text-xs font-mono text-indigo-200">
                    <div className="flex items-center space-x-2">
                      <Moon className="w-4 h-4 text-indigo-400" />
                      <span className="font-bold tracking-wider">APPARENT SUNSET (Day/Night Boundary)</span>
                      <span className="text-slate-400 text-[11px]">• Hour 12 ends / Hour 13 begins</span>
                    </div>
                    <div className="text-indigo-300 font-bold">
                      {formatTime(schedule.sunset)}
                    </div>
                  </div>
                )}

                {/* Nocturnal Block (Hours 13 to 24) */}
                {(filterMode === "all" || filterMode === "nocturnal") && (
                  <>
                    <div className="text-xs font-mono text-indigo-400/90 flex items-center space-x-2 pt-1">
                      <Moon className="w-3.5 h-3.5 text-indigo-400" />
                      <span>NOCTURNAL PHASE (Sunset → Next Sunrise • L_night: {schedule.night_hour_duration_minutes.toFixed(1)}m)</span>
                    </div>
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-6 gap-2">
                      {displayedHours.filter(h => h.index >= 13 && h.index <= 24).map((h) => renderHourCard(h))}
                    </div>
                  </>
                )}
              </div>
            </div>
          </div>
        )}
      </div>
    </div>
  );

  function renderHourCard(h: PlanetaryHour) {
    const isCurrent = h.active;
    const isSelected = selectedHour?.index === h.index;
    return (
      <div
        key={h.index}
        onClick={() => setSelectedHour(h)}
        className={`p-3 rounded-xl border text-left cursor-pointer transition-all ${
          isCurrent
            ? "bg-cyan-950/40 border-cyan-500 shadow-md ring-1 ring-cyan-500/50"
            : isSelected
            ? "bg-slate-800/90 border-slate-500"
            : "bg-[#111928]/80 border-slate-800 hover:border-slate-700 hover:bg-slate-800/40"
        }`}
      >
        <div className="flex items-center justify-between">
          <span className="text-[11px] font-mono text-slate-400 flex items-center space-x-1">
            {h.diurnal ? <Sun className="w-3 h-3 text-amber-400" /> : <Moon className="w-3 h-3 text-indigo-400" />}
            <span>H{h.index}</span>
          </span>
          <span className="text-[10px] font-mono text-slate-500">
            {formatTimeShort(h.start_time)}
          </span>
        </div>

        <div className="flex items-center justify-between mt-1.5">
          <span 
            className="font-bold text-sm"
            style={{ color: h.color_hex || "#38bdf8" }}
          >
            {h.planet_name}
          </span>
          <span className="text-xs font-bold text-slate-300">
            {h.metal_symbol}
          </span>
        </div>

        <div className="text-[10px] font-mono text-slate-400 mt-1 flex justify-between">
          <span>{h.duration_minutes.toFixed(0)}m</span>
          <span className="text-slate-500">{h.sanskrit_name}</span>
        </div>

        {isCurrent && (
          <div className="mt-2 text-[10px] font-mono font-semibold text-emerald-400 flex items-center space-x-1">
            <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-ping" />
            <span>ACTIVE • {schedule?.remaining_minutes.toFixed(0)}m left</span>
          </div>
        )}
      </div>
    );
  }
}
