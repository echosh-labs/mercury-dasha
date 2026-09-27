"use client";

import React, { createContext, useContext, useState, useEffect, useCallback, ReactNode } from "react";

export interface SubHoraInfo {
  index: number;
  planet_name: string;
  metal_symbol: string;
  lagna_arc_start: number;
  lagna_arc_end: number;
  asu_count: number;
}

export interface HoraInfo {
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
  active_sub_hora?: SubHoraInfo;
  current_lagna_degree?: number;
  local_apparent_solar_time?: string;
  equation_of_time_minutes?: number;
}

export interface TelemetryMetrics {
  alloc_mb: number;
  sys_mb: number;
  goroutines: number;
  uptime: string;
}

interface TemporalContextType {
  activeHora: HoraInfo | null;
  dayLord: string;
  connected: boolean;
  uptime: string;
  remainingSeconds: number | null;
  telemetry: TelemetryMetrics | null;
  locationLabel: string;
  setLocationLabel: (label: string) => void;
  updateActiveHora: (hora: HoraInfo) => void;
  realignTemporal: () => void;
}

const TemporalContext = createContext<TemporalContextType | undefined>(undefined);

export function TemporalProvider({ children }: { children: ReactNode }) {
  const [activeHora, setActiveHora] = useState<HoraInfo | null>(null);
  const [dayLord, setDayLord] = useState<string>("Venus");
  const [connected, setConnected] = useState<boolean>(false);
  const [uptime, setUptime] = useState<string>("Active");
  const [remainingSeconds, setRemainingSeconds] = useState<number | null>(null);
  const [telemetry, setTelemetry] = useState<TelemetryMetrics | null>(null);
  const [locationLabel, setLocationLabel] = useState<string>("St. Catharines, ON");

  const updateActiveHora = useCallback((hora: HoraInfo) => {
    setActiveHora(hora);
    if (hora.day_lord) {
      setDayLord(hora.day_lord);
    }
    if (hora.end_time) {
      const endMs = new Date(hora.end_time).getTime();
      const diffSecs = Math.max(0, Math.floor((endMs - Date.now()) / 1000));
      setRemainingSeconds(diffSecs);
    }
  }, []);

  const realignTemporal = useCallback(() => {
    if (typeof window !== "undefined") {
      window.dispatchEvent(new CustomEvent("mercury_settings_updated"));
    }
  }, []);

  // Listen for global real-time hora transitions broadcasted by ChronoPulse or background SSE
  useEffect(() => {
    const handleHoraUpdated = (e: Event) => {
      const custom = e as CustomEvent<HoraInfo>;
      if (custom.detail) {
        updateActiveHora(custom.detail);
      }
    };

    const handleSettingsUpdated = () => {
      if (typeof window !== "undefined") {
        const storedName = localStorage.getItem("mercury_geo_name");
        if (storedName) {
          setLocationLabel(storedName);
        }
      }
    };

    if (typeof window !== "undefined") {
      window.addEventListener("mercury_hora_updated", handleHoraUpdated);
      window.addEventListener("mercury_settings_updated", handleSettingsUpdated);

      const storedName = localStorage.getItem("mercury_geo_name");
      if (storedName) {
        setLocationLabel(storedName);
      }
    }

    return () => {
      if (typeof window !== "undefined") {
        window.removeEventListener("mercury_hora_updated", handleHoraUpdated);
        window.removeEventListener("mercury_settings_updated", handleSettingsUpdated);
      }
    };
  }, [updateActiveHora]);

  return (
    <TemporalContext.Provider
      value={{
        activeHora,
        dayLord,
        connected,
        uptime,
        remainingSeconds,
        telemetry,
        locationLabel,
        setLocationLabel,
        updateActiveHora,
        realignTemporal,
      }}
    >
      {children}
    </TemporalContext.Provider>
  );
}

export function useTemporal(): TemporalContextType {
  const context = useContext(TemporalContext);
  if (!context) {
    throw new Error("useTemporal must be used within a TemporalProvider");
  }
  return context;
}
