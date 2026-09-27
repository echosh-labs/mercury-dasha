"use client";
import Link from "next/link";
import React, { useEffect, useState, useCallback, useRef } from "react";
import { Orbit, Sparkles, Sun, Moon, ChevronDown, MapPin } from "lucide-react";
import { fetchProfileResonance, type SymbioticResonance } from "@/lib/storehouse";

interface HoraInfo {
  planet_id: string;
  name: string;
  sanskrit_name: string;
  element: string;
  chakra_center: string;
  color_hex: string;
  sacred_metal?: string;
  metal_symbol?: string;
  magnum_opus_stage?: string;
  story_archetype?: string;
  hour_of_day: number;
  day_lord: string;
  diurnal?: boolean;
  start_time?: string;
  end_time?: string;
  elapsed_minutes?: number;
  remaining_minutes?: number;
  progress_percent?: number;
  brief_application?: string;
  suitable_activities?: string[];
  unsuitable_activities?: string[];
  next_hour_planet?: string;
  active_sub_hora?: {
    index: number;
    planet_name: string;
    metal_symbol: string;
    lagna_arc_start: number;
    lagna_arc_end: number;
    asu_count: number;
  };
  current_lagna_degree?: number;
  local_apparent_solar_time?: string;
  equation_of_time_minutes?: number;
}

interface PulseData {
  timestamp: string;
  uptime: string;
  hora: HoraInfo;
  telemetry: {
    alloc_mb: number;
    sys_mb: number;
    goroutines: number;
    num_gc: number;
  };
  db?: {
    size_bytes: number;
    key_count: number;
  };
}

interface ChronoPulseProps {
  onTelemetryUpdate?: (tel: { alloc_mb: number; uptime: string; goroutines: number }) => void;
  profileId?: string;
}

function formatRemaining(totalSeconds: number): string {
  if (totalSeconds <= 0) return "0s";
  const mins = Math.floor(totalSeconds / 60);
  const secs = totalSeconds % 60;
  if (mins < 2) {
    return `${mins > 0 ? `${mins}m ` : ""}${secs}s`;
  }
  return `${mins}m`;
}

export default function ChronoPulse({ onTelemetryUpdate, profileId }: ChronoPulseProps) {
  const [pulse, setPulse] = useState<PulseData | null>(null);
  const [connected, setConnected] = useState<boolean>(false);
  const [locationLabel, setLocationLabel] = useState<string>("St. Catharines, ON");
  const [showTooltip, setShowTooltip] = useState<boolean>(false);
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);
  const [resonance, setResonance] = useState<SymbioticResonance | null>(null);

  const horaRef = useRef<HoraInfo | null>(null);
  const fetchingRef = useRef<boolean>(false);

  const updateResonance = useCallback(async () => {
    try {
      const targetId = profileId || (typeof window !== "undefined" ? localStorage.getItem("mercury_active_profile") : null) || "sovereign-genesis";
      const res = await fetchProfileResonance(targetId);
      if (res) {
        setResonance(res);
      }
    } catch {
      // non-blocking
    }
  }, [profileId]);

  useEffect(() => {
    updateResonance();
  }, [updateResonance]);

  const updateLocationLabel = useCallback(() => {
    if (typeof window !== "undefined") {
      const storedName = localStorage.getItem("mercury_geo_name");
      if (storedName) {
        setLocationLabel(storedName);
        return;
      }
    }
    fetch("/api/v1/settings")
      .then((res) => (res.ok ? res.json() : null))
      .then((data) => {
        if (data && data.location_name) {
          setLocationLabel(data.location_name);
        }
      })
      .catch(() => {});
  }, []);

  useEffect(() => {
    updateLocationLabel();
  }, [updateLocationLabel]);

  const getGeoParams = useCallback((): string => {
    if (typeof window === "undefined") return "";
    const lat = localStorage.getItem("mercury_geo_lat");
    const lon = localStorage.getItem("mercury_geo_lon");
    if (lat && lon) {
      return `lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`;
    }
    return "";
  }, []);

  const fetchActiveHora = useCallback(async () => {
    if (fetchingRef.current) return;
    fetchingRef.current = true;
    try {
      let url = "/api/dasha/hora";
      const params = new URLSearchParams();
      if (profileId) params.set("profile_id", profileId);
      const geo = getGeoParams();
      if (geo) {
        const p = new URLSearchParams(geo);
        p.forEach((v, k) => params.set(k, v));
      }
      const qs = params.toString();
      if (qs) url += `?${qs}`;

      const res = await fetch(url);
      if (res.ok) {
        const data = await res.json();
        if (data && data.hora) {
          horaRef.current = data.hora;
          if (typeof window !== "undefined") {
            window.dispatchEvent(new CustomEvent("mercury_hora_updated", { detail: data.hora }));
          }
          if (data.hora.end_time) {
            const endMs = new Date(data.hora.end_time).getTime();
            const diff = Math.max(0, Math.floor((endMs - Date.now()) / 1000));
            setRemainingSeconds(diff);
          }
          setPulse((prev) =>
            prev
              ? { ...prev, hora: data.hora }
              : {
                  timestamp: data.timestamp || new Date().toISOString(),
                  uptime: "Active",
                  hora: data.hora,
                  telemetry: { alloc_mb: 0, sys_mb: 0, goroutines: 0, num_gc: 0 },
                }
          );
        }
        updateResonance();
      }
    } catch (err) {
      console.error("Failed to fetch fresh hora:", err);
    } finally {
      fetchingRef.current = false;
    }
  }, [profileId, getGeoParams, updateResonance]);

  // Local 1-second high-efficiency countdown ticker
  useEffect(() => {
    const tick = () => {
      const activeHora = horaRef.current;
      if (!activeHora || !activeHora.end_time) {
        setRemainingSeconds(null);
        return;
      }
      const endMs = new Date(activeHora.end_time).getTime();
      const nowMs = Date.now();
      const diffSecs = Math.floor((endMs - nowMs) / 1000);

      if (diffSecs <= 0) {
        setRemainingSeconds(0);
        // Boundary crossed! Immediately fetch next authoritative hora
        fetchActiveHora();
      } else {
        setRemainingSeconds(diffSecs);
      }
    };

    tick();
    const interval = setInterval(tick, 1000);
    return () => clearInterval(interval);
  }, [fetchActiveHora]);

  useEffect(() => {
    let eventSource: EventSource | null = null;
    let fallbackInterval: NodeJS.Timeout | null = null;

    const connectSSE = () => {
      if (eventSource) {
        eventSource.close();
      }
      try {
        const geo = getGeoParams();
        const pulseUrl = geo ? `/api/v1/stream/pulse?${geo}` : "/api/v1/stream/pulse";
        eventSource = new EventSource(pulseUrl);

        eventSource.onopen = () => {
          setConnected(true);
        };

        eventSource.addEventListener("pulse", (e: MessageEvent) => {
          try {
            const data: PulseData = JSON.parse(e.data);
            setPulse(data);
            if (data.hora) {
              horaRef.current = data.hora;
              if (typeof window !== "undefined") {
                window.dispatchEvent(new CustomEvent("mercury_hora_updated", { detail: data.hora }));
              }
              if (data.hora.end_time) {
                const endMs = new Date(data.hora.end_time).getTime();
                const diffSecs = Math.max(0, Math.floor((endMs - Date.now()) / 1000));
                setRemainingSeconds(diffSecs);
              }
            }
            setConnected(true);
            if (onTelemetryUpdate && data.telemetry) {
              onTelemetryUpdate({
                alloc_mb: data.telemetry.alloc_mb,
                uptime: data.uptime,
                goroutines: data.telemetry.goroutines,
              });
            }
          } catch (err) {
            console.error("Failed to parse pulse event:", err);
          }
        });

        eventSource.onerror = () => {
          setConnected(false);
          eventSource?.close();
          // Fallback to fetch every 5s if SSE disconnects
          if (!fallbackInterval) {
            fallbackInterval = setInterval(fetchFallback, 5000);
          }
        };
      } catch (err) {
        console.error("SSE connection error:", err);
        setConnected(false);
        if (!fallbackInterval) {
          fallbackInterval = setInterval(fetchFallback, 5000);
        }
      }
    };

    const fetchFallback = () => {
      fetchActiveHora();
    };

    connectSSE();

    const handleSettingsUpdated = () => {
      updateLocationLabel();
      connectSSE();
      fetchActiveHora();
      updateResonance();
    };

    if (typeof window !== "undefined") {
      window.addEventListener("mercury_settings_updated", handleSettingsUpdated);
      window.addEventListener("mercury_profile_updated", handleSettingsUpdated);
      window.addEventListener("storage", handleSettingsUpdated);
    }

    return () => {
      if (eventSource) eventSource.close();
      if (fallbackInterval) clearInterval(fallbackInterval);
      if (typeof window !== "undefined") {
        window.removeEventListener("mercury_settings_updated", handleSettingsUpdated);
        window.removeEventListener("mercury_profile_updated", handleSettingsUpdated);
        window.removeEventListener("storage", handleSettingsUpdated);
      }
    };
  }, [onTelemetryUpdate, getGeoParams, updateLocationLabel, fetchActiveHora, updateResonance]);

  const hora = pulse?.hora;

  return (
    <>
      <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
        {/* Live Chrono-Pulse Indicator */}
        <div className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-[#0d1527] border border-cyan-900/60 shadow-inner">
          <span
            className={`h-2 w-2 rounded-full ${
              connected ? "bg-emerald-400 animate-ping" : "bg-amber-400"
            }`}
          />
          <span className="text-cyan-400 font-semibold tracking-wider">CHRONO-PULSE</span>
          <span className="text-slate-500">•</span>
          <span className="text-slate-300">
            {pulse ? pulse.uptime : "Syncing..."}
          </span>
        </div>

        {/* Planetary Hora Ruler (Subtle & Designated Interactive Spot) */}
        {hora && (
          <div className="relative">
            <Link
              href="/alignment/"
              onMouseEnter={() => setShowTooltip(true)}
              onMouseLeave={() => setShowTooltip(false)}
              className="flex items-center space-x-2 px-3 py-1.5 rounded-lg bg-[#111928] border transition-all shadow-sm hover:brightness-110 cursor-pointer group"
              style={{ borderColor: hora.color_hex ? `${hora.color_hex}55` : "#334155" }}
              title="Click to view 24-hour planetary hours schedule & alchemical guidance"
            >
              <Orbit className="w-3.5 h-3.5 group-hover:rotate-45 transition-transform duration-300" style={{ color: hora.color_hex || "#10b981" }} />
              <span className="text-slate-400">Hora:</span>
              <span className="font-bold" style={{ color: hora.color_hex || "#10b981" }}>
                {hora.name}
              </span>
              {hora.metal_symbol && (
                <span className="font-bold text-slate-300" title={`Sacred Metal: ${hora.sacred_metal}`}>
                  {hora.metal_symbol}
                </span>
              )}
              <span className="text-[10px] px-1.5 py-0.5 rounded bg-slate-800 text-slate-400 border border-slate-700 flex items-center space-x-1">
                {hora.diurnal !== undefined && (
                  hora.diurnal ? <Sun className="w-2.5 h-2.5 text-amber-400" /> : <Moon className="w-2.5 h-2.5 text-indigo-400" />
                )}
                <span>H{hora.hour_of_day}</span>
                {remainingSeconds !== null && remainingSeconds > 0 ? (
                  <span className="text-cyan-400 font-semibold">• {formatRemaining(remainingSeconds)}</span>
                ) : hora.remaining_minutes !== undefined && hora.remaining_minutes > 0 ? (
                  <span className="text-cyan-400 font-semibold">• {hora.remaining_minutes.toFixed(0)}m</span>
                ) : null}
              </span>
              <ChevronDown className="w-3 h-3 text-slate-500 group-hover:text-slate-300 transition-colors" />
            </Link>

            {/* Subtle Hover Tooltip with Brief Application */}
            {showTooltip && hora.brief_application && (
              <div className="absolute top-full left-0 mt-2 w-72 p-2.5 rounded-xl bg-[#0d1527] border border-cyan-800/80 shadow-2xl z-40 text-left font-sans animate-in fade-in zoom-in-95 duration-150 pointer-events-none">
                <div className="flex items-center space-x-1.5 text-[10px] font-mono text-cyan-400 font-semibold uppercase tracking-wider mb-1">
                  <Sparkles className="w-3 h-3 text-amber-400" />
                  <span>Suitable Applications</span>
                </div>
                <p className="text-[11px] text-slate-200 leading-snug font-medium">
                  {hora.brief_application}
                </p>
                {hora.active_sub_hora && (
                  <div className="mt-1.5 pt-1.5 border-t border-slate-800/80 flex items-center justify-between text-[10px] font-mono text-purple-300">
                    <span>Sub-Hora: {hora.active_sub_hora.planet_name} {hora.active_sub_hora.metal_symbol}</span>
                    <span className="text-amber-400 font-semibold">{hora.active_sub_hora.asu_count.toFixed(0)} Asu</span>
                  </div>
                )}
                <div className="mt-1.5 pt-1.5 border-t border-slate-800/80 flex items-center justify-between text-[10px] font-mono text-slate-400">
                  <span>Metal: {hora.sacred_metal}</span>
                  <span className="text-cyan-300">Click for 24h Timetable →</span>
                </div>
              </div>
            )}
          </div>
        )}

        {/* Symbiotic Resonance Indicator */}
        {resonance && (resonance.is_janma_resonant || resonance.is_mahadasha_resonant || resonance.is_antardasha_resonant) && (
          <div
            className="flex items-center space-x-1.5 px-2.5 py-1.5 rounded-lg bg-amber-950/40 border border-amber-500/50 text-amber-300 animate-pulse shadow-sm cursor-help"
            title={`${resonance.resonance_tier?.replace(/_/g, " ")}: ${resonance.transmutation_guidance}`}
          >
            <Sparkles className="w-3.5 h-3.5 text-amber-400" />
            <span className="font-semibold text-[11px]">{resonance.resonance_tier?.replace(/_/g, " ")}</span>
          </div>
        )}

        {/* Topocentric Observer Location Badge */}
        <Link
          href="/alignment/"
          className="flex items-center space-x-1.5 px-2.5 py-1.5 rounded-lg bg-[#111928] border border-cyan-900/40 hover:border-cyan-500/60 transition-all shadow-sm cursor-pointer group text-slate-300 hover:text-cyan-300"
          title="Click to view or calibrate topocentric observer location & system coordinates"
        >
          <MapPin className="w-3.5 h-3.5 text-cyan-400 group-hover:scale-110 transition-transform" />
          <span className="font-semibold">{locationLabel}</span>
          <ChevronDown className="w-3 h-3 text-slate-500 group-hover:text-slate-300 transition-colors" />
        </Link>
      </div>


    </>
  );
}
