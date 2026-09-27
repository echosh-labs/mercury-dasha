"use client";
import Link from "next/link";

import React, { useState, useEffect } from "react";
import { 
  Sparkles, 
  Calendar, 
  Clock, 
  Orbit, 
  ChevronRight, 
  ChevronDown, 
  Bookmark, 
  Save, 
  RotateCcw, 
  Compass, 
  Flame, 
  Layers, 
  Radio, 
  CheckCircle2, 
  AlertCircle,
  MapPin,
  Users,
  Video
} from "lucide-react";

import {
  fetchNakshatrasLazy,
  type Nakshatra,
  type Pada,
  type SovereignProfile,
  type ProfileSummary,
  type SymbioticResonance,
  type NatalAstrology,
  type NatalAlchemy,
  type ActiveAlchemicalState,
  type DashaPeriodSummary
} from "@/lib/storehouse";

interface DashaPeriod {
  level: string;
  planet: string;
  planet_name: string;
  sanskrit_name: string;
  color_hex: string;
  duration_days: number;
  start_date: string;
  end_date: string;
  sub_periods?: DashaPeriod[];
}

interface StoryContext {
  active_archetype: string;
  thematic_phase: string;
  alchemical_element: string;
  chakra_focus: string;
  resonant_frequency_hz: number;
  hermetic_principles: string[];
  narrative_prompt: string;
}

interface ActiveSnapshot {
  reference_time: string;
  mahadasha: DashaPeriod;
  antardasha: DashaPeriod;
  pratyantardasha: DashaPeriod;
  maha_elapsed_days: number;
  maha_remaining_days: number;
  maha_percent_done: number;
  antar_elapsed_days: number;
  antar_remaining_days: number;
  antar_percent_done: number;
  story_context: StoryContext;
}

export type DashaProfile = SovereignProfile;

export default function DashaCalculator() {
  const [mode, setMode] = useState<"ephemeris" | "nakshatra">("ephemeris");
  const [profileName, setProfileName] = useState<string>("Sovereign Genesis");
  const [birthDate, setBirthDate] = useState<string>("1992-06-15");
  const [birthTime, setBirthTime] = useState<string>("12:00");
  const [timezoneOffset, setTimezoneOffset] = useState<number>(-4); // EDT
  const [locationName, setLocationName] = useState<string>("St. Catharines, ON");
  const [latitude, setLatitude] = useState<number>(43.1594);
  const [longitude, setLongitude] = useState<number>(-79.2469);
  const [timezone, setTimezone] = useState<string>("America/Toronto");
  const [nakshatraIndex, setNakshatraIndex] = useState<number>(9); // Ashlesha (Mercury)
  const [padaNumber, setPadaNumber] = useState<number>(1);
  const [nakshatrasList, setNakshatrasList] = useState<Nakshatra[]>([]);

  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [profile, setProfile] = useState<DashaProfile | null>(null);

  // Expandable timeline accordion state
  const [expandedMaha, setExpandedMaha] = useState<Record<number, boolean>>({});
  const [expandedAntar, setExpandedAntar] = useState<Record<string, boolean>>({});

  // Saved profiles list
  const [savedProfiles, setSavedProfiles] = useState<string[]>([]);
  const [profileSummaries, setProfileSummaries] = useState<ProfileSummary[]>([]);
  const [selectedSaved, setSelectedSaved] = useState<string>("");
  const [saveSuccess, setSaveSuccess] = useState<boolean>(false);
  const [activeHoraRuler, setActiveHoraRuler] = useState<string>("");

  const handleLoadProfile = async (id: string) => {
    if (!id) return;
    setSelectedSaved(id);
    setLoading(true);
    setError(null);
    try {
      let url = `/api/v1/profiles/${encodeURIComponent(id)}`;
      if (typeof window !== "undefined") {
        const lat = localStorage.getItem("mercury_geo_lat");
        const lon = localStorage.getItem("mercury_geo_lon");
        if (lat && lon) {
          url += `?lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`;
        }
      }
      const res = await fetch(url);
      if (res.ok) {
        const data: DashaProfile = await res.json();
        setProfile(data);
        setProfileName(data.name || "Loaded Profile");
        if (typeof window !== "undefined") {
          localStorage.setItem("mercury_last_profile", id);
          localStorage.setItem("mercury_active_profile", id);
        }

        if (data.mode) {
          setMode(data.mode);
        }
        if (data.nakshatra_index) {
          setNakshatraIndex(data.nakshatra_index);
        }
        if (data.pada_number) {
          setPadaNumber(data.pada_number);
        }
        if (data.timezone_offset !== undefined && data.timezone_offset !== null) {
          setTimezoneOffset(data.timezone_offset);
        }

        if (data.location_name) {
          setLocationName(data.location_name);
        }
        if (data.latitude !== undefined && data.latitude !== null && data.latitude !== 0) {
          setLatitude(data.latitude);
        }
        if (data.longitude !== undefined && data.longitude !== null && data.longitude !== 0) {
          setLongitude(data.longitude);
        }
        if (data.timezone) {
          setTimezone(data.timezone);
        }

        if (data.birth_date && data.birth_time) {
          setBirthDate(data.birth_date);
          setBirthTime(data.birth_time);
        } else if (data.birth_time_utc) {
          // Legacy profile fallback: compute local time using the timezone offset
          const offsetHours = (data.timezone_offset !== undefined && data.timezone_offset !== null)
            ? data.timezone_offset
            : -4; // Default to EDT offset if unspecified
          const utcDate = new Date(data.birth_time_utc);
          const localMs = utcDate.getTime() + offsetHours * 3600 * 1000;
          const localDate = new Date(localMs);
          setBirthDate(localDate.toISOString().slice(0, 10));
          setBirthTime(localDate.toISOString().slice(11, 16));
        }

        // Auto-expand active Mahadasha
        if (data.timeline && data.active_snapshot) {
          const activeIdx = data.timeline.findIndex(
            (m) => m.planet === data.active_snapshot.mahadasha.planet
          );
          if (activeIdx >= 0) {
            setExpandedMaha({ [activeIdx]: true });
          }
        }
      }
    } catch (err) {
      console.error("Failed to load profile:", err);
    } finally {
      setLoading(false);
    }
  };

  const handleCalculate = async () => {
    setLoading(true);
    setError(null);
    setSaveSuccess(false);

    try {
      const payload: any = {
        name: profileName,
        birth_date: birthDate,
        birth_time: birthTime,
        timezone_offset: timezoneOffset,
      };

      if (mode === "nakshatra") {
        payload.nakshatra_index = nakshatraIndex;
        payload.pada_number = padaNumber;
      }

      const res = await fetch("/api/dasha/calculate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });

      if (!res.ok) {
        const errData = await res.json();
        throw new Error(errData.error || "Calculation failed");
      }

      const data: DashaProfile = await res.json();
      const updatedProfile: DashaProfile = {
        ...data,
        id: selectedSaved || data.id,
        name: profileName,
        birth_date: birthDate,
        birth_time: birthTime,
        timezone_offset: timezoneOffset,
        location_name: locationName,
        latitude: latitude,
        longitude: longitude,
        timezone: timezone,
        mode,
        nakshatra_index: nakshatraIndex,
        pada_number: padaNumber,
      };
      setProfile(updatedProfile);

      // Auto-expand the active Mahadasha
      const activeIdx = updatedProfile.timeline?.findIndex(
        (m) => m.planet === updatedProfile.active_snapshot?.mahadasha?.planet
      ) ?? -1;
      if (activeIdx >= 0) {
        setExpandedMaha({ [activeIdx]: true });
      }
    } catch (err: any) {
      setError(err.message || "An error occurred during calculation");
    } finally {
      setLoading(false);
    }
  };

  const handleSaveProfile = async () => {
    setLoading(true);
    setError(null);
    setSaveSuccess(false);

    try {
      const payload: any = {
        name: profileName,
        birth_date: birthDate,
        birth_time: birthTime,
        timezone_offset: timezoneOffset,
      };

      if (mode === "nakshatra") {
        payload.nakshatra_index = nakshatraIndex;
        payload.pada_number = padaNumber;
      }

      // 1. Calculate authoritative astronomical ephemeris for latest inputs
      const calcRes = await fetch("/api/dasha/calculate", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(payload)
      });

      if (!calcRes.ok) {
        const errData = await calcRes.json();
        throw new Error(errData.error || "Calculation failed before saving");
      }

      const calculatedData: DashaProfile = await calcRes.json();
      const id = `profile:${profileName.toLowerCase().replace(/[^a-z0-9]/g, "-")}`;
      const toSave: DashaProfile = {
        ...calculatedData,
        id,
        name: profileName,
        birth_date: birthDate,
        birth_time: birthTime,
        timezone_offset: timezoneOffset,
        location_name: locationName,
        latitude: latitude,
        longitude: longitude,
        timezone: timezone,
        mode,
        nakshatra_index: nakshatraIndex,
        pada_number: padaNumber,
      };

      // 2. Persist to BoltDB substrate
      const saveRes = await fetch(`/api/v1/profiles/${id}`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(toSave)
      });

      if (!saveRes.ok) {
        const errData = await saveRes.json();
        throw new Error(errData.error || "Failed to save profile to database");
      }

      setProfile(toSave);
      setSelectedSaved(id);
      setSaveSuccess(true);
      if (!savedProfiles.includes(id)) {
        setSavedProfiles([...savedProfiles, id]);
      }
      if (typeof window !== "undefined") {
        localStorage.setItem("mercury_last_profile", id);
        localStorage.setItem("mercury_geo_lat", latitude.toString());
        localStorage.setItem("mercury_geo_lon", longitude.toString());
        localStorage.setItem("mercury_geo_name", locationName);
        localStorage.setItem("mercury_timezone", timezone);
        window.dispatchEvent(new Event("mercury_settings_updated"));
        window.dispatchEvent(new Event("mercury_profile_updated"));
      }

      // Refresh summaries list in background
      fetch("/api/v1/profiles")
        .then((r) => (r.ok ? r.json() : null))
        .then((d) => {
          if (d?.summaries) setProfileSummaries(d.summaries);
          if (d?.profiles) setSavedProfiles(d.profiles);
        })
        .catch(() => {});

      // Auto-expand active Mahadasha
      const activeIdx = toSave.timeline?.findIndex(
        (m) => m.planet === toSave.active_snapshot?.mahadasha?.planet
      ) ?? -1;
      if (activeIdx >= 0) {
        setExpandedMaha({ [activeIdx]: true });
      }

      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      setError(err.message || "Failed to save profile");
    } finally {
      setLoading(false);
    }
  };

  const handleNewProfile = () => {
    setSelectedSaved("");
    setProfileName("New Sovereign");
    setBirthDate("1992-06-15");
    setBirthTime("12:00");
    setTimezoneOffset(-4);
    setMode("ephemeris");
  };

  // Fetch nakshatras catalog and saved profiles on mount
  useEffect(() => {
    let mounted = true;

    const initData = async () => {
      try {
        const [nakshatras, pRes] = await Promise.all([
          fetchNakshatrasLazy(),
          fetch("/api/v1/profiles")
        ]);

        if (mounted && nakshatras && nakshatras.length > 0) {
          setNakshatrasList(nakshatras);
        }

        let profileLoaded = false;
        if (pRes.ok) {
          const pData = await pRes.json();
          const profiles: string[] = pData.profiles || [];
          const summaries: ProfileSummary[] = pData.summaries || [];
          if (mounted) {
            setSavedProfiles(profiles);
            setProfileSummaries(summaries);
          }

          // Restore last viewed profile or the primary saved profile if available
          const lastProfile = typeof window !== "undefined" ? localStorage.getItem("mercury_last_profile") : null;
          const targetProfile = (lastProfile && profiles.includes(lastProfile))
            ? lastProfile
            : (profiles.length > 0 ? profiles[0] : null);

          if (targetProfile && mounted) {
            await handleLoadProfile(targetProfile);
            profileLoaded = true;
          }
        }

        if (!profileLoaded && mounted) {
          await handleCalculate();
        }
      } catch (err) {
        console.error("Failed to load initial data:", err);
        if (mounted) handleCalculate();
      }
    };

    initData();

    const handleHoraUpdated = (e: any) => {
      if (e?.detail?.ruling_planet && mounted) {
        setActiveHoraRuler(e.detail.ruling_planet);
      }
    };

    window.addEventListener("mercury_hora_updated", handleHoraUpdated);
    return () => {
      mounted = false;
      window.removeEventListener("mercury_hora_updated", handleHoraUpdated);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);

  const toggleMaha = (idx: number) => {
    setExpandedMaha((prev) => ({ ...prev, [idx]: !prev[idx] }));
  };

  const toggleAntar = (key: string) => {
    setExpandedAntar((prev) => ({ ...prev, [key]: !prev[key] }));
  };

  const activeSnap = profile?.active_snapshot;
  const story = activeSnap?.story_context;

  return (
    <div className="space-y-8">
      {/* Studio Header & Configuration */}
      <div className="bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-6">
        <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800 pb-4">
          <div>
            <div className="flex items-center space-x-2">
              <Compass className="w-5 h-5 text-cyan-400" />
              <h2 className="text-xl font-bold text-white tracking-tight">
                Vimshottari Dasha Engine Studio
              </h2>
              <span className="px-2 py-0.5 rounded text-[10px] font-mono bg-cyan-950/70 border border-cyan-800 text-cyan-300">
                120-Year Master Clock
              </span>
            </div>
            <p className="text-xs text-slate-400 mt-1">
              Precision temporal ephemeris calculation & foundational storytelling archetype synthesis.
            </p>
          </div>

          {/* Saved Profiles Quick Recall */}
          <div className="flex items-center space-x-2">
            <Bookmark className="w-4 h-4 text-slate-400" />
            <select
              value={selectedSaved}
              onChange={(e) => handleLoadProfile(e.target.value)}
              className="bg-slate-900 border border-slate-700 text-xs rounded-lg px-3 py-1.5 text-slate-300 font-mono focus:border-cyan-500 focus:outline-none max-w-[280px] truncate"
            >
              <option value="">Load Saved Profile...</option>
              {profileSummaries.length > 0 ? (
                profileSummaries.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name} — {s.nakshatra_name} ({s.natal_metal}) [{s.active_mahadasha} Maha]
                  </option>
                ))
              ) : (
                savedProfiles.map((p) => (
                  <option key={p} value={p}>
                    {p}
                  </option>
                ))
              )}
            </select>
            <button
              onClick={handleNewProfile}
              title="Start a new profile draft"
              className="bg-slate-900 hover:bg-slate-800 text-slate-400 hover:text-cyan-300 text-xs px-2.5 py-1.5 rounded-lg border border-slate-700 transition flex items-center space-x-1"
            >
              <RotateCcw className="w-3 h-3" />
              <span>New</span>
            </button>
            <Link
              href="/characters/"
              title="Open full Character Sanctuary management"
              className="bg-purple-950/70 hover:bg-purple-900/80 text-purple-300 hover:text-white text-xs px-2.5 py-1.5 rounded-lg border border-purple-700/60 transition flex items-center space-x-1 font-mono"
            >
              <Users className="w-3 h-3 text-purple-400" />
              <span>Sanctuary</span>
            </Link>
          </div>
        </div>

        {/* Mode Selector */}
        <div className="flex space-x-3 text-xs font-mono">
          <button
            onClick={() => setMode("ephemeris")}
            className={`px-3 py-1.5 rounded-lg border transition ${
              mode === "ephemeris"
                ? "bg-cyan-950/70 border-cyan-500 text-cyan-300 font-semibold"
                : "bg-slate-900/80 border-slate-800 text-slate-400 hover:text-slate-200"
            }`}
          >
            Astronomical Ephemeris (Date/Time)
          </button>
          <button
            onClick={() => setMode("nakshatra")}
            className={`px-3 py-1.5 rounded-lg border transition ${
              mode === "nakshatra"
                ? "bg-cyan-950/70 border-cyan-500 text-cyan-300 font-semibold"
                : "bg-slate-900/80 border-slate-800 text-slate-400 hover:text-slate-200"
            }`}
          >
            Direct Janma Nakshatra & Pada
          </button>
        </div>

        {/* Input Parameters Grid */}
        <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-xs font-mono">
          <div>
            <label className="text-slate-400 block mb-1">Profile / Chart Label</label>
            <input
              type="text"
              value={profileName}
              onChange={(e) => setProfileName(e.target.value)}
              className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
            />
          </div>

          {mode === "ephemeris" ? (
            <>
              <div>
                <label className="text-slate-400 block mb-1">Birth Date (YYYY-MM-DD)</label>
                <input
                  type="date"
                  value={birthDate}
                  onChange={(e) => setBirthDate(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-slate-400 block mb-1">Birth Time (HH:MM)</label>
                <input
                  type="time"
                  value={birthTime}
                  onChange={(e) => setBirthTime(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                />
              </div>
              <div>
                <label className="text-slate-400 block mb-1">Timezone Offset (Hours)</label>
                <input
                  type="number"
                  step="0.5"
                  value={timezoneOffset}
                  onChange={(e) => setTimezoneOffset(parseFloat(e.target.value) || 0)}
                  placeholder="e.g. -4 for EDT"
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                />
              </div>
            </>
          ) : (
            <>
              <div>
                <label className="text-slate-400 block mb-1">Janma Nakshatra</label>
                <select
                  value={nakshatraIndex}
                  onChange={(e) => setNakshatraIndex(parseInt(e.target.value) || 1)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                >
                  {nakshatrasList.map((n) => (
                    <option key={n.index} value={n.index}>
                      {n.index}. {n.name} ({n.ruling_planet})
                    </option>
                  ))}
                </select>
              </div>
              <div>
                <label className="text-slate-400 block mb-1">Pada (1 - 4)</label>
                <select
                  value={padaNumber}
                  onChange={(e) => setPadaNumber(parseInt(e.target.value) || 1)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                >
                  <option value={1}>Pada 1 (1st Quarter)</option>
                  <option value={2}>Pada 2 (2nd Quarter)</option>
                  <option value={3}>Pada 3 (3rd Quarter)</option>
                  <option value={4}>Pada 4 (4th Quarter)</option>
                </select>
              </div>
              <div>
                <label className="text-slate-400 block mb-1">Target Epoch</label>
                <input
                  type="date"
                  value={birthDate}
                  onChange={(e) => setBirthDate(e.target.value)}
                  className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
                />
              </div>
            </>
          )}
        </div>

        {/* Topocentric Observer Coordinates & Planetary Hours Alignment */}
        <div className="pt-4 border-t border-slate-800/80 space-y-3">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
            <div className="flex items-center space-x-2">
              <MapPin className="w-4 h-4 text-cyan-400" />
              <span className="text-xs font-bold text-slate-200 uppercase tracking-wider font-mono">
                Topocentric Observer Coordinates & Planetary Hours Alignment
              </span>
              <span className="text-[10px] px-2 py-0.5 rounded-full bg-cyan-950/70 border border-cyan-800 text-cyan-300 font-mono">
                PROFILE ANCHOR
              </span>
            </div>
            <div className="flex flex-wrap items-center gap-1.5 text-[11px] font-mono">
              <span className="text-slate-500 mr-1">Presets:</span>
              <button
                type="button"
                onClick={() => {
                  setLocationName("St. Catharines, ON");
                  setLatitude(43.1594);
                  setLongitude(-79.2469);
                  setTimezone("America/Toronto");
                  setTimezoneOffset(-4);
                }}
                className={`px-2 py-0.5 rounded border transition ${
                  latitude === 43.1594 && longitude === -79.2469
                    ? "bg-cyan-950 border-cyan-500 text-cyan-300 font-bold"
                    : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                }`}
              >
                📍 Niagara / St. Catharines (Default)
              </button>
              <button
                type="button"
                onClick={() => {
                  setLocationName("Toronto, ON");
                  setLatitude(43.6532);
                  setLongitude(-79.3832);
                  setTimezone("America/Toronto");
                  setTimezoneOffset(-4);
                }}
                className={`px-2 py-0.5 rounded border transition ${
                  latitude === 43.6532 && longitude === -79.3832
                    ? "bg-cyan-950 border-cyan-500 text-cyan-300 font-bold"
                    : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                }`}
              >
                🏙️ Toronto
              </button>
              <button
                type="button"
                onClick={() => {
                  setLocationName("New York, NY");
                  setLatitude(40.7128);
                  setLongitude(-74.0060);
                  setTimezone("America/New_York");
                  setTimezoneOffset(-4);
                }}
                className={`px-2 py-0.5 rounded border transition ${
                  latitude === 40.7128 && longitude === -74.0060
                    ? "bg-cyan-950 border-cyan-500 text-cyan-300 font-bold"
                    : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                }`}
              >
                🗽 New York
              </button>
              <button
                type="button"
                onClick={() => {
                  if (typeof navigator !== "undefined" && navigator.geolocation) {
                    navigator.geolocation.getCurrentPosition(
                      (pos) => {
                        const lat = parseFloat(pos.coords.latitude.toFixed(4));
                        const lon = parseFloat(pos.coords.longitude.toFixed(4));
                        setLatitude(lat);
                        setLongitude(lon);
                        setLocationName(`GPS: ${lat}°, ${lon}°`);
                      },
                      (err) => console.error("GPS detection error:", err),
                      { enableHighAccuracy: true, timeout: 8000 }
                    );
                  }
                }}
                className="px-2 py-0.5 rounded border bg-slate-900 border-slate-800 text-slate-400 hover:text-white flex items-center space-x-1"
              >
                <Compass className="w-3 h-3 text-emerald-400" />
                <span>GPS Auto-Detect</span>
              </button>
            </div>
          </div>

          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 text-xs font-mono">
            <div>
              <label className="text-slate-400 block mb-1">Observer Location Name</label>
              <input
                type="text"
                value={locationName}
                onChange={(e) => setLocationName(e.target.value)}
                placeholder="e.g. St. Catharines, ON"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="text-slate-400 block mb-1">Latitude (°N/S)</label>
              <input
                type="number"
                step="0.0001"
                value={latitude}
                onChange={(e) => setLatitude(parseFloat(e.target.value) || 0)}
                placeholder="e.g. 43.1594"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="text-slate-400 block mb-1">Longitude (°E/W)</label>
              <input
                type="number"
                step="0.0001"
                value={longitude}
                onChange={(e) => setLongitude(parseFloat(e.target.value) || 0)}
                placeholder="e.g. -79.2469"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
              />
            </div>
            <div>
              <label className="text-slate-400 block mb-1">IANA Timezone</label>
              <input
                type="text"
                value={timezone}
                onChange={(e) => setTimezone(e.target.value)}
                placeholder="e.g. America/Toronto"
                className="w-full bg-slate-900 border border-slate-700 rounded-lg px-3 py-2 text-slate-200 focus:border-cyan-500 focus:outline-none"
              />
            </div>
          </div>
          <p className="text-[11px] text-slate-400 font-sans">
            Coordinates stored with the user profile (<strong className="text-cyan-300 font-mono">profile:sovereign-genesis</strong>) govern topocentric apparent sunrise, solar noon, and the planetary Hora schedule broadcasted live across the site.
          </p>
        </div>

        {/* Actions Bar */}
        <div className="flex items-center justify-between pt-2">
          <div className="flex items-center space-x-2">
            <button
              onClick={handleCalculate}
              disabled={loading}
              className="flex items-center space-x-2 bg-gradient-to-r from-cyan-600 to-emerald-600 hover:from-cyan-500 hover:to-emerald-500 text-white font-medium text-xs px-5 py-2.5 rounded-lg shadow transition disabled:opacity-50"
            >
              <Sparkles className="w-4 h-4" />
              <span>{loading ? "Calculating..." : "Compute Dasha Progression"}</span>
            </button>

            <button
              onClick={handleSaveProfile}
              disabled={loading}
              className="flex items-center space-x-1.5 bg-slate-800 hover:bg-slate-700 text-slate-300 text-xs px-3.5 py-2.5 rounded-lg border border-slate-700 transition disabled:opacity-50"
            >
              <Save className="w-3.5 h-3.5 text-cyan-400" />
              <span>{loading ? "Saving..." : "Save to BoltDB"}</span>
            </button>
          </div>

          {saveSuccess && (
            <div className="flex items-center space-x-1 text-emerald-400 text-xs font-mono">
              <CheckCircle2 className="w-4 h-4" />
              <span>Saved to BoltDB Vault</span>
            </div>
          )}
        </div>

        {error && (
          <div className="flex items-center space-x-2 bg-rose-950/50 border border-rose-800 text-rose-300 text-xs p-3 rounded-lg">
            <AlertCircle className="w-4 h-4 text-rose-400" />
            <span>{error}</span>
          </div>
        )}
      </div>

      {/* Active Progression & Storytelling Hologram Card */}
      {profile && activeSnap && (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Active Arc Card */}
          <div className="lg:col-span-1 bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-5">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <span className="text-xs font-mono uppercase text-slate-400">Current Temporal Arc</span>
              <span className="h-2 w-2 rounded-full bg-emerald-400 animate-ping" />
            </div>

            <div className="space-y-4 font-mono">
              {/* Mahadasha Ruler */}
              <div>
                <div className="text-[10px] uppercase text-slate-500">Active Mahadasha (Major Era)</div>
                <div className="text-xl font-bold flex items-center space-x-2" style={{ color: activeSnap.mahadasha.color_hex }}>
                  <Orbit className="w-5 h-5" />
                  <span>{activeSnap.mahadasha.planet_name}</span>
                </div>
                <div className="text-xs text-slate-400 mt-1">
                  {new Date(activeSnap.mahadasha.start_date).toLocaleDateString()} — {new Date(activeSnap.mahadasha.end_date).toLocaleDateString()}
                </div>
                {/* Progress bar */}
                <div className="w-full bg-slate-800 h-2 rounded-full overflow-hidden mt-2">
                  <div
                    className="h-full rounded-full transition-all duration-500"
                    style={{
                      width: `${activeSnap.maha_percent_done}%`,
                      backgroundColor: activeSnap.mahadasha.color_hex || "#10b981",
                    }}
                  />
                </div>
                <div className="flex justify-between text-[10px] text-slate-400 mt-1">
                  <span>{activeSnap.maha_percent_done}% traversed</span>
                  <span>{Math.round(activeSnap.maha_remaining_days)} days left</span>
                </div>
              </div>

              {/* Antardasha Ruler */}
              <div className="pt-2 border-t border-slate-800/80">
                <div className="text-[10px] uppercase text-slate-500">Active Antardasha (Sub-Chapter)</div>
                <div className="text-base font-bold flex items-center space-x-2" style={{ color: activeSnap.antardasha.color_hex }}>
                  <Flame className="w-4 h-4" />
                  <span>{activeSnap.antardasha.planet_name}</span>
                </div>
                <div className="text-xs text-slate-400 mt-0.5">
                  {new Date(activeSnap.antardasha.start_date).toLocaleDateString()} — {new Date(activeSnap.antardasha.end_date).toLocaleDateString()}
                </div>
                {/* Antardasha progress bar */}
                <div className="w-full bg-slate-800 h-1.5 rounded-full overflow-hidden mt-2">
                  <div
                    className="h-full rounded-full transition-all duration-500"
                    style={{
                      width: `${activeSnap.antar_percent_done}%`,
                      backgroundColor: activeSnap.antardasha.color_hex || "#ec4899",
                    }}
                  />
                </div>
                <div className="flex justify-between text-[10px] text-slate-400 mt-1">
                  <span>{activeSnap.antar_percent_done}% traversed</span>
                  <span>{Math.round(activeSnap.antar_remaining_days)} days left</span>
                </div>
              </div>

              {/* Pratyantardasha */}
              <div className="pt-2 border-t border-slate-800/80 flex items-center justify-between text-xs">
                <span className="text-slate-500">Pratyantardasha (Sub-Sub):</span>
                <span className="font-bold text-slate-200" style={{ color: activeSnap.pratyantardasha.color_hex }}>
                  {activeSnap.pratyantardasha.planet_name}
                </span>
              </div>

              {/* Live Planetary Hora & Temporal Harmonic Resonance */}
              {activeHoraRuler && (
                <div className="pt-2 border-t border-slate-800/80 flex items-center justify-between text-xs">
                  <span className="text-slate-500">Live Planetary Hora:</span>
                  <div className="flex items-center space-x-1.5">
                    <span className="font-semibold text-amber-300 font-mono">{activeHoraRuler}</span>
                    {(activeHoraRuler.toLowerCase() === activeSnap.mahadasha.planet.toLowerCase() ||
                      activeHoraRuler.toLowerCase() === activeSnap.antardasha.planet.toLowerCase() ||
                      activeHoraRuler.toLowerCase() === activeSnap.pratyantardasha.planet.toLowerCase()) && (
                      <span className="px-1.5 py-0.5 rounded text-[9px] bg-amber-500/20 text-amber-300 border border-amber-500/40 font-mono animate-pulse">
                        HARMONIC RESONANCE
                      </span>
                    )}
                  </div>
                </div>
              )}
            </div>

            {/* Unified Sovereign Astrological & Alchemical Signature Matrix */}
            <div className="pt-3 border-t border-slate-800 text-xs font-mono space-y-3 text-slate-400 bg-slate-900/60 p-3.5 rounded-xl border border-slate-800/80">
              <div className="flex items-center justify-between border-b border-slate-800 pb-1.5">
                <span className="text-[10px] text-cyan-400 font-bold uppercase tracking-wider">
                  Sovereign Matrix
                </span>
                <span className="text-[10px] text-slate-400">
                  {profile.janma_nakshatra?.name} • Pada {profile.janma_pada}
                </span>
              </div>

              {/* The Tripod of Embodiment (Lagna, Surya, Chandra) */}
              {profile.astrology?.tripod && profile.astrology.tripod.lagna?.rashi && (
                <div className="bg-slate-950/80 p-2 rounded-lg border border-slate-800 space-y-1.5">
                  <div className="text-[10px] text-cyan-400 uppercase font-bold tracking-wider flex items-center justify-between">
                    <span>Tripod of Embodiment</span>
                    <span className="text-slate-500 font-normal">Lagna • Surya • Chandra</span>
                  </div>
                  <div className="grid grid-cols-3 gap-1.5 text-[10px]">
                    <div className="p-1 rounded bg-slate-900/80 border border-cyan-900/40 text-center">
                      <span className="text-cyan-400 block font-bold">Rising Lagna</span>
                      <span className="text-slate-200 font-semibold">{profile.astrology.tripod.lagna.symbol} {profile.astrology.tripod.lagna.rashi_sanskrit}</span>
                      <span className="text-slate-500 block text-[9px]">{profile.astrology.tripod.lagna.degree_in_sign.toFixed(1)}°</span>
                    </div>
                    <div className="p-1 rounded bg-slate-900/80 border border-amber-900/40 text-center">
                      <span className="text-amber-400 block font-bold">Surya (Sun)</span>
                      <span className="text-slate-200 font-semibold">{profile.astrology.tripod.surya.rashi?.split(" ")[0]}</span>
                      <span className="text-slate-500 block text-[9px]">H{profile.astrology.tripod.surya.house_number} ({profile.astrology.tripod.surya.degree_in_sign.toFixed(1)}°)</span>
                    </div>
                    <div className="p-1 rounded bg-slate-900/80 border border-emerald-900/40 text-center">
                      <span className="text-emerald-400 block font-bold">Chandra (Moon)</span>
                      <span className="text-slate-200 font-semibold">{profile.astrology.tripod.chandra.rashi?.split(" ")[0]}</span>
                      <span className="text-slate-500 block text-[9px]">H{profile.astrology.tripod.chandra.house_number} ({profile.astrology.tripod.chandra.degree_in_sign.toFixed(1)}°)</span>
                    </div>
                  </div>
                </div>
              )}

              <div className="space-y-1 text-[11px]">
                <div className="flex justify-between">
                  <span className="text-slate-500">Sidereal Moon:</span>
                  <span className="text-slate-200">{profile.moon_degree.toFixed(2)}°</span>
                </div>
                {profile.astrology?.panchanga?.sidereal_sun_degree ? (
                  <div className="flex justify-between">
                    <span className="text-slate-500">Sidereal Sun:</span>
                    <span className="text-amber-300">{profile.astrology.panchanga.sidereal_sun_degree.toFixed(2)}°</span>
                  </div>
                ) : null}
                {profile.astrology?.lagna?.degree ? (
                  <div className="flex justify-between">
                    <span className="text-slate-500">Sidereal Lagna:</span>
                    <span className="text-cyan-300">{profile.astrology.lagna.degree.toFixed(2)}° ({profile.astrology.lagna.rashi})</span>
                  </div>
                ) : null}
                <div className="flex justify-between">
                  <span className="text-slate-500">Starting Lord:</span>
                  <span className="text-slate-200">{profile.starting_lord} ({profile.balance_years.toFixed(2)} yrs)</span>
                </div>
                {profile.janma_nakshatra?.deity && (
                  <div className="flex justify-between">
                    <span className="text-slate-500">Presiding Deity:</span>
                    <span className="text-cyan-300">{profile.janma_nakshatra.deity}</span>
                  </div>
                )}
                {profile.astrology?.pada_detail?.navamsha_sign && (
                  <div className="flex justify-between">
                    <span className="text-slate-500">Navamsha Pada Sign:</span>
                    <span className="text-cyan-300">{profile.astrology.pada_detail.navamsha_sign}</span>
                  </div>
                )}
                {profile.astrology?.elemental_tattva && (
                  <div className="flex justify-between">
                    <span className="text-slate-500">Elemental Tattva:</span>
                    <span className="text-emerald-300">{profile.astrology.elemental_tattva}</span>
                  </div>
                )}
              </div>

              {/* Vedic Panchanga (The 5 Cosmic Limbs) */}
              {profile.astrology?.panchanga && profile.astrology.panchanga.tithi_name && (
                <div className="pt-2 border-t border-slate-800 space-y-1.5">
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] text-amber-400 font-bold uppercase tracking-wider">
                      Vedic Panchanga
                    </span>
                    <span className="text-[10px] text-slate-500">5 Cosmic Limbs</span>
                  </div>

                  <div className="grid grid-cols-2 gap-1.5 text-[10px]">
                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">1. Vara (Day):</span>
                      <span className="text-slate-200 font-semibold">{profile.astrology.panchanga.vara?.split(" ")[0]}</span>
                      <span className="text-slate-500 block text-[9px]">{profile.astrology.panchanga.vara_lord}</span>
                    </div>

                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">2. Tithi (Phase):</span>
                      <span className="text-cyan-300 font-semibold">{profile.astrology.panchanga.tithi_name}</span>
                      <span className="text-slate-500 block text-[9px]">{profile.astrology.panchanga.paksha} ({profile.astrology.panchanga.tithi_percent?.toFixed(0)}%)</span>
                    </div>

                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">4. Yoga:</span>
                      <span className="text-emerald-300 font-semibold">{profile.astrology.panchanga.yoga_name}</span>
                      <span className="text-slate-500 block text-[9px]">{profile.astrology.panchanga.yoga_meaning}</span>
                    </div>

                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800">
                      <span className="text-slate-500 block">5. Karana:</span>
                      <span className="text-purple-300 font-semibold">{profile.astrology.panchanga.karana_name}</span>
                      <span className="text-slate-500 block text-[9px]">{profile.astrology.panchanga.karana_type}</span>
                    </div>
                  </div>
                </div>
              )}

              {/* Ayurvedic Biological Triad */}
              {profile.astrology?.ayurveda && profile.astrology.ayurveda.dosha && (
                <div className="pt-2 border-t border-slate-800 space-y-1.5">
                  <div className="flex items-center justify-between">
                    <span className="text-[10px] text-emerald-400 font-bold uppercase tracking-wider">
                      Ayurvedic Triad
                    </span>
                    <span className="text-[10px] font-semibold text-emerald-300">
                      {profile.astrology.ayurveda.dosha} Dosha
                    </span>
                  </div>

                  <div className="grid grid-cols-3 gap-1.5 text-[10px]">
                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800 text-center">
                      <span className="text-slate-500 block text-[9px]">Gana</span>
                      <span className="text-slate-200 font-semibold">{profile.astrology.ayurveda.gana}</span>
                    </div>

                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800 text-center">
                      <span className="text-slate-500 block text-[9px]">Yoni Totem</span>
                      <span className="text-amber-300 font-semibold">{profile.astrology.ayurveda.yoni_animal}</span>
                    </div>

                    <div className="bg-slate-950/70 p-1.5 rounded border border-slate-800 text-center">
                      <span className="text-slate-500 block text-[9px]">Nadi</span>
                      <span className="text-cyan-300 font-semibold">{profile.astrology.ayurveda.nadi}</span>
                    </div>
                  </div>

                  {profile.astrology.ayurveda.dosha_qualities && (
                    <div className="text-[9px] text-slate-400 italic pt-0.5">
                      {profile.astrology.ayurveda.dosha_qualities}
                    </div>
                  )}
                </div>
              )}

              {profile.alchemy?.sacred_metal && (
                <div className="pt-2 border-t border-slate-800 space-y-1.5">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-500 text-[10px] uppercase">Natal Sacred Metal</span>
                    <span
                      className="px-2 py-0.5 rounded text-[10px] font-bold bg-slate-950 border border-slate-800"
                      style={{ color: profile.alchemy.sacred_metal.color_hex }}
                    >
                      {profile.alchemy.sacred_metal.symbol} {profile.alchemy.sacred_metal.name}
                    </span>
                  </div>
                  {profile.alchemy.governing_axiom && (
                    <div className="text-[10px] text-slate-300 italic">
                      Axiom: &ldquo;{profile.alchemy.governing_axiom.title}&rdquo;
                    </div>
                  )}
                  {profile.alchemy.alchemical_motto && (
                    <p className="text-[10px] text-slate-400 font-sans leading-tight">
                      &ldquo;{profile.alchemy.alchemical_motto}&rdquo;
                    </p>
                  )}
                </div>
              )}

              {/* Active Symbiotic Resonance Indicator */}
              {profile.symbiotic_resonance && (profile.symbiotic_resonance.is_janma_resonant || profile.symbiotic_resonance.is_mahadasha_resonant || profile.symbiotic_resonance.is_antardasha_resonant) && (
                <div className="mt-2 p-2 rounded-lg bg-amber-950/40 border border-amber-500/50 text-[11px] text-amber-300 animate-pulse">
                  <div className="flex items-center space-x-1.5 font-bold">
                    <Sparkles className="w-3 h-3 text-amber-400" />
                    <span>{profile.symbiotic_resonance.resonance_tier?.replace(/_/g, " ")}</span>
                    <span className="text-[10px] font-mono text-cyan-300">({profile.symbiotic_resonance.transit_hora_planet_name} Hora)</span>
                  </div>
                  <div className="text-[10px] text-amber-200/90 font-sans mt-0.5">
                    {profile.symbiotic_resonance.transmutation_guidance}
                  </div>
                </div>
              )}
            </div>

            {/* Symbiotic Alchemical Hora Alignment Link */}
            <div className="pt-2">
              <Link
                href={profile?.id ? `/alignment/?profile_id=${encodeURIComponent(profile.id)}` : "/alignment/"}
                className="w-full flex items-center justify-between px-3.5 py-2.5 rounded-xl bg-gradient-to-r from-purple-950/60 via-slate-900 to-cyan-950/60 border border-purple-800/40 hover:border-purple-600 text-xs font-mono text-purple-200 transition group cursor-pointer shadow-sm"
              >
                <div className="flex items-center space-x-2">
                  <Sparkles className="w-3.5 h-3.5 text-purple-400 group-hover:scale-110 transition-transform" />
                  <span>Symbiotic Daily Hora Alignment</span>
                </div>
                <span className="text-[11px] text-cyan-400 group-hover:translate-x-0.5 transition-transform flex items-center space-x-1">
                  <span>View 24h</span>
                  <ChevronRight className="w-3 h-3" />
                </span>
              </Link>
            </div>
          </div>

          {/* Foundations Studio & Archetype Gateway Card */}
          {story && (
            <div className="lg:col-span-2 bg-gradient-to-br from-[#111928] via-[#0e1726] to-[#0c1424] border border-cyan-900/40 rounded-2xl p-6 shadow-xl flex flex-col justify-between space-y-4">
              <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                <div className="flex items-center space-x-2">
                  <Sparkles className="w-5 h-5 text-cyan-400" />
                  <h3 className="text-base font-bold text-white">
                    Foundations Chronicle Resonance
                  </h3>
                </div>
                <span className="text-xs font-mono text-cyan-400 bg-cyan-950/60 px-2 py-0.5 rounded border border-cyan-800">
                  Resonance: {story.resonant_frequency_hz} Hz
                </span>
              </div>

              <div className="grid grid-cols-1 sm:grid-cols-2 gap-3 text-xs font-mono">
                <div className="p-3 bg-slate-900/70 border border-slate-800 rounded-xl">
                  <div className="text-slate-500 uppercase text-[10px]">Active Archetype</div>
                  <div className="text-sm font-bold text-cyan-300 mt-0.5">{story.active_archetype}</div>
                </div>
                <div className="p-3 bg-slate-900/70 border border-slate-800 rounded-xl">
                  <div className="text-slate-500 uppercase text-[10px]">Alchemical Element</div>
                  <div className="text-sm font-semibold text-emerald-300 mt-0.5">{story.alchemical_element}</div>
                </div>
              </div>

              <div className="pt-2">
                <Link
                  href="/foundations/"
                  className="w-full flex items-center justify-between px-4 py-3 rounded-xl bg-gradient-to-r from-red-950/50 via-slate-900 to-cyan-950/50 border border-red-800/40 hover:border-red-600 text-xs font-mono text-red-200 transition group cursor-pointer shadow-sm"
                >
                  <div className="flex items-center space-x-2">
                    <Video className="w-4 h-4 text-red-400 group-hover:scale-110 transition-transform" />
                    <span>Open Foundations Studio &amp; Audiovisual Chronicles</span>
                  </div>
                  <span className="text-[11px] text-cyan-400 group-hover:translate-x-0.5 transition-transform flex items-center space-x-1">
                    <span>Synthesize Video</span>
                    <ChevronRight className="w-3 h-3" />
                  </span>
                </Link>
              </div>
            </div>
          )}
        </div>
      )}

      {/* 12-Bhava Sovereign House Matrix */}
      {profile && profile.astrology?.tripod?.bhavas && profile.astrology.tripod.bhavas.length === 12 && (
        <div className="bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2 border-b border-slate-800 pb-3">
            <div className="flex items-center space-x-2">
              <Compass className="w-5 h-5 text-cyan-400" />
              <h3 className="text-base font-bold text-white">
                12-Bhava Sovereign House Matrix (Whole Sign)
              </h3>
            </div>
            <div className="flex items-center space-x-3 text-xs font-mono text-slate-400">
              <span className="flex items-center space-x-1">
                <span className="w-2 h-2 rounded-full bg-cyan-400 inline-block" />
                <span>Lagna (Body)</span>
              </span>
              <span className="flex items-center space-x-1">
                <span className="w-2 h-2 rounded-full bg-amber-400 inline-block" />
                <span>Surya (Soul)</span>
              </span>
              <span className="flex items-center space-x-1">
                <span className="w-2 h-2 rounded-full bg-emerald-400 inline-block" />
                <span>Chandra (Mind)</span>
              </span>
            </div>
          </div>

          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 xl:grid-cols-6 gap-3 font-mono text-xs">
            {profile.astrology.tripod.bhavas.map((bhava) => {
              const hasLagna = bhava.luminaries?.some((l) => l.startsWith("Lagna"));
              const hasSurya = bhava.luminaries?.some((l) => l.startsWith("Surya"));
              const hasChandra = bhava.luminaries?.some((l) => l.startsWith("Chandra"));

              return (
                <div
                  key={bhava.house_number}
                  className={`p-3 rounded-xl border flex flex-col justify-between transition-all ${
                    hasLagna
                      ? "bg-cyan-950/30 border-cyan-500/60 shadow-md shadow-cyan-950/50"
                      : hasSurya || hasChandra
                      ? "bg-slate-900/90 border-slate-700"
                      : "bg-slate-900/40 border-slate-800/80 hover:border-slate-700"
                  }`}
                >
                  <div className="space-y-1.5">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] uppercase font-bold text-slate-500">
                        H{bhava.house_number} • {bhava.sanskrit_name}
                      </span>
                      {hasLagna && (
                        <span className="px-1.5 py-0.5 rounded text-[9px] font-bold bg-cyan-500/20 text-cyan-300 border border-cyan-500/40">
                          LAGNA
                        </span>
                      )}
                    </div>

                    <div className="font-bold text-slate-200 text-xs truncate">
                      {bhava.rashi}
                    </div>
                    <div className="text-[10px] text-slate-400">
                      Lord: {bhava.rashi_lord?.split(" ")[0]}
                    </div>

                    {bhava.luminaries && bhava.luminaries.length > 0 && (
                      <div className="pt-1 space-y-1">
                        {bhava.luminaries.map((lum, lIdx) => {
                          const isSun = lum.startsWith("Surya");
                          const isMoon = lum.startsWith("Chandra");
                          return (
                            <div
                              key={lIdx}
                              className={`text-[9px] px-1.5 py-0.5 rounded font-bold border ${
                                isSun
                                  ? "bg-amber-500/20 text-amber-300 border-amber-500/40"
                                  : isMoon
                                  ? "bg-emerald-500/20 text-emerald-300 border-emerald-500/40"
                                  : "bg-cyan-500/20 text-cyan-300 border-cyan-500/40"
                              }`}
                            >
                              {lum}
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>

                  <div className="text-[9px] text-slate-400/80 font-sans leading-tight pt-2 border-t border-slate-800/60 mt-2">
                    {bhava.domain}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* 120-Year Master Timeline Accordion */}
      {profile && profile.timeline && (
        <div className="bg-[#111928] border border-slate-800 rounded-2xl p-6 shadow-xl space-y-4">
          <div className="flex items-center justify-between border-b border-slate-800 pb-3">
            <div className="flex items-center space-x-2">
              <Calendar className="w-5 h-5 text-cyan-400" />
              <h3 className="text-base font-bold text-white">
                120-Year Vimshottari Master Timeline
              </h3>
            </div>
            <span className="text-xs font-mono text-slate-400">
              9 Planetary Mahadasha Epochs
            </span>
          </div>

          <div className="space-y-2 font-mono text-xs">
            {profile.timeline.map((maha: any, mIdx: number) => {
              const isActive = activeSnap?.mahadasha.planet === maha.planet;
              const isExpanded = expandedMaha[mIdx];
              const isCompleted = new Date(maha.end_date) < new Date();

              return (
                <div
                  key={mIdx}
                  className={`border rounded-xl transition overflow-hidden ${
                    isActive
                      ? "border-cyan-500/60 bg-cyan-950/20"
                      : "border-slate-800/80 bg-slate-900/40"
                  }`}
                >
                  {/* Mahadasha Row Header */}
                  <div
                    onClick={() => toggleMaha(mIdx)}
                    className="p-3.5 flex items-center justify-between cursor-pointer hover:bg-slate-800/40 transition"
                  >
                    <div className="flex items-center space-x-3">
                      {isExpanded ? (
                        <ChevronDown className="w-4 h-4 text-cyan-400" />
                      ) : (
                        <ChevronRight className="w-4 h-4 text-slate-500" />
                      )}
                      <span
                        className="font-bold text-sm"
                        style={{ color: maha.color_hex || "#38bdf8" }}
                      >
                        {maha.planet_name}
                      </span>
                      <span className="text-slate-400 hidden sm:inline">
                        ({maha.sanskrit_name})
                      </span>
                    </div>

                    <div className="flex items-center space-x-3">
                      <span className="text-slate-400 text-xs">
                        {new Date(maha.start_date).getFullYear()} — {new Date(maha.end_date).getFullYear()}
                      </span>
                      {isActive ? (
                        <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-emerald-500/20 text-emerald-400 border border-emerald-500/30">
                          ACTIVE ERA
                        </span>
                      ) : isCompleted ? (
                        <span className="px-2 py-0.5 rounded text-[10px] bg-slate-800 text-slate-500 border border-slate-700">
                          COMPLETED
                        </span>
                      ) : (
                        <span className="px-2 py-0.5 rounded text-[10px] bg-slate-800/50 text-slate-400 border border-slate-800">
                          FUTURE
                        </span>
                      )}
                    </div>
                  </div>

                  {/* Sub-Antardashas List */}
                  {isExpanded && maha.sub_periods && (
                    <div className="border-t border-slate-800/80 bg-slate-950/60 p-3 space-y-1.5 pl-8">
                      {maha.sub_periods.map((antar: any, aIdx: number) => {
                        const antarKey = `${mIdx}-${aIdx}`;
                        const isAntarActive =
                          isActive && activeSnap?.antardasha.planet === antar.planet;
                        const isAntarExpanded = expandedAntar[antarKey];

                        return (
                          <div
                            key={aIdx}
                            className={`rounded-lg border p-2 transition ${
                              isAntarActive
                                ? "border-cyan-500/40 bg-cyan-950/30"
                                : "border-slate-800/60 bg-slate-900/30"
                            }`}
                          >
                            <div
                              onClick={() => toggleAntar(antarKey)}
                              className="flex items-center justify-between cursor-pointer"
                            >
                              <div className="flex items-center space-x-2">
                                <span className="text-slate-500 text-[10px]">{aIdx + 1}.</span>
                                <span
                                  className="font-semibold text-xs"
                                  style={{ color: antar.color_hex }}
                                >
                                  {maha.planet_name} / {antar.planet_name}
                                </span>
                              </div>
                              <div className="flex items-center space-x-2 text-[11px] text-slate-400">
                                <span>
                                  {new Date(antar.start_date).toLocaleDateString()} —{" "}
                                  {new Date(antar.end_date).toLocaleDateString()}
                                </span>
                                {isAntarActive && (
                                  <span className="px-1.5 py-0.2 rounded text-[9px] font-bold bg-emerald-400/20 text-emerald-400 border border-emerald-400/30">
                                    CURRENT
                                  </span>
                                )}
                              </div>
                            </div>

                            {/* Pratyantardashas Sub-Sub-Periods */}
                            {isAntarExpanded && antar.sub_periods && (
                              <div className="mt-2 pt-2 border-t border-slate-800/60 grid grid-cols-1 sm:grid-cols-3 gap-1.5 text-[10px] text-slate-400">
                                {antar.sub_periods.map((prat: any, pIdx: number) => (
                                  <div
                                    key={pIdx}
                                    className="p-1.5 rounded bg-slate-900/60 border border-slate-800/60 flex items-center justify-between"
                                  >
                                    <span style={{ color: prat.color_hex }}>{prat.planet_name}</span>
                                    <span>{new Date(prat.start_date).toLocaleDateString().slice(0, 5)}</span>
                                  </div>
                                ))}
                              </div>
                            )}
                          </div>
                        );
                      })}
                    </div>
                  )}
                </div>
              );
            })}
          </div>
        </div>
      )}


    </div>
  );
}
