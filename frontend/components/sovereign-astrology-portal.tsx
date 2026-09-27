"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  Compass,
  Clock,
  Sun,
  Users,
  Sparkles,
  ShieldCheck,
  Calendar,
  Layers,
  ChevronRight,
  ExternalLink,
  RotateCcw
} from "lucide-react";
import DashaCalculator from "./dasha-calculator";
import PlanetaryHoraMatrix from "./planetary-hora-matrix";
import DailyEphemerisChart from "./daily-ephemeris-chart";
import CharacterSanctuaryHub from "./character-sanctuary-hub";
import ChronoPulse from "./chrono-pulse";
import { fetchCharacters, type ProfileSummary } from "@/lib/storehouse";

export type AstrologySegment = "dasha" | "hora" | "ephemeris" | "characters";

interface SovereignAstrologyPortalProps {
  initialSegment?: AstrologySegment;
}

export default function SovereignAstrologyPortal({
  initialSegment = "dasha",
}: SovereignAstrologyPortalProps) {
  const [activeSegment, setActiveSegment] = useState<AstrologySegment>(initialSegment);
  const [activeProfileId, setActiveProfileId] = useState<string>("");
  const [profiles, setProfiles] = useState<ProfileSummary[]>([]);
  const [loadingProfiles, setLoadingProfiles] = useState<boolean>(true);

  // Sync segment from URL query parameter if present
  useEffect(() => {
    if (typeof window !== "undefined") {
      const params = new URLSearchParams(window.location.search);
      const segParam = params.get("segment") as AstrologySegment | null;
      if (segParam && ["dasha", "hora", "ephemeris", "characters"].includes(segParam)) {
        setActiveSegment(segParam);
      }
      const charParam = params.get("id") || params.get("profile_id");
      if (charParam) {
        setActiveProfileId(charParam);
      } else {
        const stored = localStorage.getItem("mercury_active_profile");
        if (stored) setActiveProfileId(stored);
      }
    }
  }, []);

  // Fetch registered sovereign characters/profiles for unified dropdown
  const loadProfiles = useCallback(async () => {
    try {
      const list = await fetchCharacters();
      setProfiles(list);
      if (list.length > 0 && !activeProfileId) {
        const stored = typeof window !== "undefined" ? localStorage.getItem("mercury_active_profile") : null;
        if (stored && list.some((p) => p.id === stored)) {
          setActiveProfileId(stored);
        } else {
          setActiveProfileId(list[0].id);
        }
      }
    } catch (err) {
      console.error("Failed to load profiles in portal:", err);
    } finally {
      setLoadingProfiles(false);
    }
  }, [activeProfileId]);

  useEffect(() => {
    loadProfiles();
  }, [loadProfiles]);

  const handleProfileChange = (id: string) => {
    setActiveProfileId(id);
    if (typeof window !== "undefined") {
      localStorage.setItem("mercury_active_profile", id);
      const url = new URL(window.location.href);
      url.searchParams.set("profile_id", id);
      window.history.replaceState({}, "", url.toString());
    }
  };

  const handleSegmentChange = (seg: AstrologySegment) => {
    setActiveSegment(seg);
    if (typeof window !== "undefined") {
      const url = new URL(window.location.href);
      url.searchParams.set("segment", seg);
      window.history.replaceState({}, "", url.toString());
    }
  };

  const activeProfile = profiles.find((p) => p.id === activeProfileId);

  return (
    <div className="space-y-6">
      {/* Consolidated Sovereign Observatory Header */}
      <div className="flex flex-col lg:flex-row lg:items-center justify-between gap-4 bg-gradient-to-r from-slate-900 via-[#101726] to-[#0c1424] p-5 sm:p-6 rounded-2xl border border-slate-800 shadow-2xl">
        <div className="space-y-2">
          <div className="flex flex-wrap items-center gap-2">
            <div className="flex items-center space-x-2">
              <Compass className="w-5 h-5 text-cyan-400" />
              <h1 className="text-xl sm:text-2xl font-bold text-white tracking-tight font-serif">
                Mercury Dasha Observatory
              </h1>
            </div>
            <span className="bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-xs font-mono px-2.5 py-0.5 rounded-full flex items-center space-x-1">
              <ShieldCheck className="w-3 h-3" />
              <span>SOVEREIGN PORTAL</span>
            </span>
          </div>
          <p className="text-xs text-slate-400 max-w-2xl leading-relaxed">
            Consolidated Astrological &amp; Character Engine: Vimshottari Dasha calculations, 24h Chaldean planetary hora, topocentric ephemeris, and protagonist sanctuary dossiers for <span className="text-cyan-300 font-medium">echosh-labs</span>.
          </p>
        </div>

        {/* Global Chrono Telemetry & Character Selector */}
        <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
          {/* Active Character Selector */}
          <div className="flex items-center space-x-1.5 bg-slate-900/90 border border-slate-800 rounded-xl px-3 py-1.5 shadow-sm">
            <Users className="w-3.5 h-3.5 text-purple-400" />
            <span className="text-slate-400 text-[11px] hidden sm:inline">Active Subject:</span>
            <select
              value={activeProfileId}
              onChange={(e) => handleProfileChange(e.target.value)}
              className="bg-transparent text-slate-200 font-bold focus:outline-none cursor-pointer max-w-[150px] sm:max-w-[180px] truncate"
            >
              {profiles.map((p) => (
                <option key={p.id} value={p.id} className="bg-slate-900 text-slate-100">
                  {p.name} ({p.nakshatra_name || "Natal"})
                </option>
              ))}
              {profiles.length === 0 && (
                <option value="">Sovereign Genesis (Default)</option>
              )}
            </select>
          </div>

          <ChronoPulse />
        </div>
      </div>

      {/* Primary Astrological Segment Navigation Bar */}
      <div className="flex flex-wrap items-center gap-1.5 sm:gap-2 p-1.5 rounded-xl bg-slate-900/80 border border-slate-800/80 backdrop-blur text-xs font-mono shadow-md">
        <button
          onClick={() => handleSegmentChange("dasha")}
          className={`flex items-center space-x-2 px-3.5 py-2 rounded-lg transition-all ${
            activeSegment === "dasha"
              ? "bg-cyan-950/80 text-cyan-300 border border-cyan-700/60 font-semibold shadow-inner"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/50"
          }`}
        >
          <Sparkles className="w-4 h-4 text-cyan-400" />
          <span>Vimshottari Dasha</span>
          {activeProfile && (
            <span className="text-[10px] hidden md:inline px-1.5 py-0.5 rounded bg-cyan-900/40 text-cyan-400">
              {activeProfile.active_mahadasha || "Active"}
            </span>
          )}
        </button>

        <button
          onClick={() => handleSegmentChange("hora")}
          className={`flex items-center space-x-2 px-3.5 py-2 rounded-lg transition-all ${
            activeSegment === "hora"
              ? "bg-purple-950/80 text-purple-300 border border-purple-700/60 font-semibold shadow-inner"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/50"
          }`}
        >
          <Clock className="w-4 h-4 text-purple-400" />
          <span>Planetary Hora (24h)</span>
          <span className="text-[10px] hidden md:inline px-1.5 py-0.5 rounded bg-purple-900/40 text-purple-400">
            Chaldean
          </span>
        </button>

        <button
          onClick={() => handleSegmentChange("ephemeris")}
          className={`flex items-center space-x-2 px-3.5 py-2 rounded-lg transition-all ${
            activeSegment === "ephemeris"
              ? "bg-amber-950/80 text-amber-300 border border-amber-700/60 font-semibold shadow-inner"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/50"
          }`}
        >
          <Sun className="w-4 h-4 text-amber-400" />
          <span>Daily Ephemeris</span>
          <span className="text-[10px] hidden md:inline px-1.5 py-0.5 rounded bg-amber-900/40 text-amber-400">
            Solar Anchors
          </span>
        </button>

        <button
          onClick={() => handleSegmentChange("characters")}
          className={`flex items-center space-x-2 px-3.5 py-2 rounded-lg transition-all ${
            activeSegment === "characters"
              ? "bg-rose-950/80 text-rose-300 border border-rose-700/60 font-semibold shadow-inner"
              : "text-slate-400 hover:text-slate-200 hover:bg-slate-800/50"
          }`}
        >
          <Users className="w-4 h-4 text-rose-400" />
          <span>Character Sanctuary</span>
          <span className="text-[10px] hidden md:inline px-1.5 py-0.5 rounded bg-rose-900/40 text-rose-400">
            {profiles.length} Profiles
          </span>
        </button>
      </div>

      {/* Active Segment Content */}
      <div className="animate-in fade-in duration-150">
        {activeSegment === "dasha" && (
          <DashaCalculator
            activeProfileId={activeProfileId}
            onProfileChange={handleProfileChange}
            onOpenSegment={handleSegmentChange}
          />
        )}

        {activeSegment === "hora" && (
          <PlanetaryHoraMatrix
            activeProfileId={activeProfileId}
            onSelectProfile={handleProfileChange}
            onNavigateSegment={handleSegmentChange}
            hideHeaderNav={true}
          />
        )}

        {activeSegment === "ephemeris" && (
          <DailyEphemerisChart
            onNavigateSegment={handleSegmentChange}
            hideHeaderNav={true}
          />
        )}

        {activeSegment === "characters" && (
          <CharacterSanctuaryHub
            activeCharacterId={activeProfileId}
            onSelectCharacter={handleProfileChange}
            onNavigateSegment={handleSegmentChange}
            hideHeaderNav={true}
          />
        )}
      </div>
    </div>
  );
}
