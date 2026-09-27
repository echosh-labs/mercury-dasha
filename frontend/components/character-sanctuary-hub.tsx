"use client";

import React, { useState, useEffect, useTransition } from "react";
import Link from "next/link";
import {
  Users,
  Compass,
  Sparkles,
  Calendar,
  Clock,
  MapPin,
  CheckCircle2,
  AlertCircle,
  Video,
  Layers,
  Flame,
  Orbit,
  ArrowLeft,
  ChevronRight,
  ChevronDown,
  Plus,
  Trash2,
  RotateCcw,
  Film,
  Music,
  Activity,
  Maximize2,
  FileCode,
  Share2,
  Sliders,
  ExternalLink,
  ShieldCheck,
  Zap,
  Save,
  Eye,
  Info
} from "lucide-react";
import {
  fetchCharacters,
  fetchCharacter,
  fetchCharacterTimeline,
  generateCharacterChronicle,
  type ProfileSummary,
  type CharacterProfile,
  type DashaPeriod,
  type TimelineManifest,
  type CharacterChronicleRef,
  type ChronicleRequest
} from "@/lib/storehouse";
import ChronoPulse from "@/components/chrono-pulse";

export interface CharacterSanctuaryHubProps {
  activeCharacterId?: string;
  onSelectCharacter?: (id: string) => void;
  onNavigateSegment?: (segment: "dasha" | "hora" | "ephemeris" | "characters") => void;
  hideHeaderNav?: boolean;
}

export default function CharacterSanctuaryHub({
  activeCharacterId: propCharId,
  onSelectCharacter,
  onNavigateSegment,
  hideHeaderNav = false
}: CharacterSanctuaryHubProps = {}) {
  const [characterList, setCharacterList] = useState<ProfileSummary[]>([]);
  const [activeCharId, setActiveCharId] = useState<string>(propCharId || "");
  const [character, setCharacter] = useState<CharacterProfile | null>(null);
  const [timeline, setTimeline] = useState<DashaPeriod[]>([]);
  const [loading, setLoading] = useState<boolean>(true);
  const [timelineLoading, setTimelineLoading] = useState<boolean>(false);
  const [activeTab, setActiveTab] = useState<"blueprint" | "alchemy" | "timeline" | "chronicles">("blueprint");

  useEffect(() => {
    if (propCharId !== undefined && propCharId !== activeCharId) {
      setActiveCharId(propCharId);
      loadCharacterDossier(propCharId);
    }
  }, [propCharId]);

  // Timeline accordion state
  const [expandedMaha, setExpandedMaha] = useState<Record<number, boolean>>({});
  const [expandedAntar, setExpandedAntar] = useState<Record<string, boolean>>({});

  // Video Generation Chronicle State
  const [selectedEpochPlanet, setSelectedEpochPlanet] = useState<string>("");
  const [chronicleOrientation, setChronicleOrientation] = useState<"16:9" | "9:16" | "1:1">("16:9");
  const [chronicleDuration, setChronicleDuration] = useState<number>(60);
  const [compilingChronicle, setCompilingChronicle] = useState<boolean>(false);
  const [compiledManifest, setCompiledManifest] = useState<TimelineManifest | null>(null);
  const [chronicleSuccess, setChronicleSuccess] = useState<boolean>(false);
  const [showManifestJson, setShowManifestJson] = useState<boolean>(false);

  // Character Forge State
  const [showForge, setShowForge] = useState<boolean>(false);
  const [forgeName, setForgeName] = useState<string>("");
  const [forgeTitle, setForgeTitle] = useState<string>("");
  const [forgeBackstory, setForgeBackstory] = useState<string>("");
  const [forgeDate, setForgeDate] = useState<string>("1992-06-15");
  const [forgeTime, setForgeTime] = useState<string>("12:00");
  const [forgeTz, setForgeTz] = useState<number>(-4.0);
  const [forgeCity, setForgeCity] = useState<string>("St. Catharines, ON");
  const [forgeLat, setForgeLat] = useState<number>(43.1594);
  const [forgeLon, setForgeLon] = useState<number>(-79.2469);
  const [forging, setForging] = useState<boolean>(false);
  const [forgeError, setForgeError] = useState<string | null>(null);

  // 1. Initial Load: Fetch Characters List
  useEffect(() => {
    const init = async () => {
      setLoading(true);
      try {
        const summaries = await fetchCharacters();
        setCharacterList(summaries);

        // Check URL query param or localStorage
        let targetId = "";
        if (typeof window !== "undefined") {
          const params = new URLSearchParams(window.location.search);
          const qId = params.get("id");
          const qTab = params.get("tab");
          if (qTab && ["blueprint", "alchemy", "timeline", "chronicles"].includes(qTab)) {
            setActiveTab(qTab as any);
          }
          const stored = localStorage.getItem("mercury_active_profile");
          if (qId && summaries.some((s) => s.id === qId)) {
            targetId = qId;
          } else if (stored && summaries.some((s) => s.id === stored)) {
            targetId = stored;
          } else if (summaries.length > 0) {
            targetId = summaries[0].id;
          }
        }

        if (targetId) {
          setActiveCharId(targetId);
          await loadCharacterDossier(targetId);
        }
      } catch (err) {
        console.error("Failed to load characters:", err);
      } finally {
        setLoading(false);
      }
    };

    init();
  }, []);

  // 2. Load Single Character Dossier & Timeline
  const loadCharacterDossier = async (id: string) => {
    if (!id) return;
    try {
      const char = await fetchCharacter(id);
      if (char) {
        setCharacter(char);
        if (typeof window !== "undefined") {
          localStorage.setItem("mercury_active_profile", id);
        }
        // If timeline is not in dossier, fetch drilldown
        if (char.timeline && char.timeline.length > 0) {
          setTimeline(char.timeline);
        } else {
          setTimelineLoading(true);
          const tl = await fetchCharacterTimeline(id);
          setTimeline(tl);
          setTimelineLoading(false);
        }
        // Set default epoch for chronicle
        if (char.active_snapshot?.mahadasha?.planet) {
          setSelectedEpochPlanet(char.active_snapshot.mahadasha.planet);
        }
      }
    } catch (err) {
      console.error("Failed to load character dossier:", err);
    }
  };

  const handleSelectCharacter = async (id: string) => {
    setActiveCharId(id);
    if (onSelectCharacter) onSelectCharacter(id);
    if (typeof window !== "undefined") {
      localStorage.setItem("mercury_active_profile", id);
    }
    await loadCharacterDossier(id);
  };

  // 3. Compile Video Timeline Chronicle
  const handleCompileChronicle = async () => {
    if (!character) return;
    setCompilingChronicle(true);
    setChronicleSuccess(false);
    try {
      const req: ChronicleRequest = {
        dasha_planet: selectedEpochPlanet || character.active_snapshot?.mahadasha?.planet,
        orientation: chronicleOrientation,
        duration_sec: chronicleDuration,
      };
      const res = await generateCharacterChronicle(character.id, req);
      if (res && res.manifest) {
        setCompiledManifest(res.manifest);
        setChronicleSuccess(true);
        // Refresh character dossier to display newly recorded chronicle
        await loadCharacterDossier(character.id);
      }
    } catch (err) {
      console.error("Failed to compile chronicle:", err);
    } finally {
      setCompilingChronicle(false);
    }
  };

  // 4. Forge New Sovereign Character
  const handleForgeCharacter = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!forgeName || !forgeDate) {
      setForgeError("Character name and birth date are required.");
      return;
    }
    setForging(true);
    setForgeError(null);
    try {
      const res = await fetch("/api/v1/characters", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          name: forgeName,
          title: forgeTitle,
          backstory: forgeBackstory,
          birth_date: forgeDate,
          birth_time: forgeTime,
          timezone_offset: forgeTz,
          location_name: forgeCity,
          latitude: forgeLat,
          longitude: forgeLon,
        }),
      });
      if (res.ok) {
        const newChar: CharacterProfile = await res.json();
        setShowForge(false);
        setForgeName("");
        setForgeTitle("");
        setForgeBackstory("");
        // Reload list and activate
        const summaries = await fetchCharacters();
        setCharacterList(summaries);
        setActiveCharId(newChar.id);
        await loadCharacterDossier(newChar.id);
      } else {
        const errData = await res.json();
        setForgeError(errData.error || "Failed to forge character");
      }
    } catch (err) {
      setForgeError("Network error forging character");
    } finally {
      setForging(false);
    }
  };

  // 5. Delete Character
  const handleDeleteCharacter = async (id: string) => {
    if (!confirm("Are you certain you wish to dissolve this sovereign character record from the storehouse?")) return;
    try {
      const res = await fetch(`/api/v1/characters/${encodeURIComponent(id)}`, { method: "DELETE" });
      if (res.ok) {
        const summaries = await fetchCharacters();
        setCharacterList(summaries);
        if (summaries.length > 0) {
          setActiveCharId(summaries[0].id);
          await loadCharacterDossier(summaries[0].id);
        } else {
          setCharacter(null);
          setTimeline([]);
        }
      }
    } catch (err) {
      console.error("Failed to delete character:", err);
    }
  };

  const toggleMaha = (idx: number) => {
    setExpandedMaha((prev) => ({ ...prev, [idx]: !prev[idx] }));
  };

  const toggleAntar = (key: string) => {
    setExpandedAntar((prev) => ({ ...prev, [key]: !prev[key] }));
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
              <Users className="w-5 h-5 text-purple-400" />
              <h1 className="text-base font-bold text-white tracking-tight">
                Character Sanctuary
              </h1>
              <span className="text-[10px] font-mono px-2 py-0.5 rounded-full bg-purple-950 border border-purple-800 text-purple-300">
                LIVING PROTAGONISTS
              </span>
            </div>
          </div>

          <div className="flex items-center space-x-3 text-xs font-mono">
            <Link
              href="/alignment/"
              className="px-2.5 py-1.5 rounded-lg bg-slate-900 hover:bg-slate-800 border border-slate-800 text-slate-300 hover:text-cyan-300 transition flex items-center space-x-1"
            >
              <Compass className="w-3.5 h-3.5 text-cyan-400" />
              <span>24h Alignment</span>
            </Link>

            <button
              onClick={() => setShowForge(true)}
              className="px-3 py-1.5 rounded-lg bg-purple-600 hover:bg-purple-500 text-white font-bold transition flex items-center space-x-1.5 shadow-md shadow-purple-950/40"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Forge Character</span>
            </button>
          </div>
        </div>
      )}

      <div className="space-y-6">
        {/* Horizontal Character Selection Ribbon */}
        <section className="space-y-2">
          <div className="flex items-center justify-between text-xs font-mono text-slate-400">
            <span className="uppercase tracking-wider">Registered Sovereigns ({characterList.length})</span>
            <span>Click to switch active character</span>
          </div>

          <div className="flex items-center space-x-3 overflow-x-auto pb-2 scrollbar-thin scrollbar-thumb-slate-800">
            {characterList.map((c) => {
              const isSelected = c.id === activeCharId;
              return (
                <button
                  key={c.id}
                  onClick={() => handleSelectCharacter(c.id)}
                  className={`px-4 py-2.5 rounded-2xl border text-left transition-all shrink-0 flex items-center space-x-3 ${
                    isSelected
                      ? "bg-purple-950/60 border-purple-500 ring-2 ring-purple-500/40 shadow-lg"
                      : "bg-[#0c1322] border-slate-800 hover:border-slate-700 hover:bg-slate-800/50"
                  }`}
                >
                  <div className="w-9 h-9 rounded-xl bg-purple-900/40 border border-purple-700/60 flex items-center justify-center font-bold text-purple-300 font-mono text-sm">
                    {c.name.charAt(0)}
                  </div>
                  <div>
                    <div className="flex items-center space-x-1.5">
                      <span className="font-bold text-sm text-white">{c.name}</span>
                      {isSelected && (
                        <span className="w-2 h-2 rounded-full bg-emerald-400 animate-pulse" />
                      )}
                    </div>
                    <div className="text-[11px] font-mono text-slate-400 flex items-center space-x-1 mt-0.5">
                      <span className="text-cyan-400">{c.nakshatra_name}</span>
                      <span>•</span>
                      <span className="text-amber-400">{c.natal_metal}</span>
                      <span>•</span>
                      <span className="text-slate-500">[{c.active_mahadasha} Maha]</span>
                    </div>
                  </div>
                </button>
              );
            })}

            <button
              onClick={() => setShowForge(true)}
              className="px-4 py-2.5 rounded-2xl border border-dashed border-slate-700 hover:border-purple-500 hover:bg-purple-950/20 text-slate-400 hover:text-purple-300 transition shrink-0 flex items-center space-x-2 text-xs font-mono"
            >
              <Plus className="w-4 h-4" />
              <span>Forge New Character</span>
            </button>
          </div>
        </section>

        {loading || !character ? (
          <div className="p-16 flex flex-col items-center justify-center space-y-3 text-slate-400 bg-slate-900/40 rounded-2xl border border-slate-800">
            <Users className="w-8 h-8 animate-spin text-purple-400" />
            <span className="text-sm font-mono">Hydrating Sovereign Character Dossier &amp; Astrological Matrix...</span>
          </div>
        ) : (
          <>
            {/* Active Character Dossier Header Card */}
            <section className="p-6 rounded-3xl bg-gradient-to-br from-[#0e1628] via-[#0c1322] to-[#070c18] border border-cyan-900/50 shadow-2xl space-y-5">
              <div className="flex flex-col md:flex-row md:items-center justify-between gap-4 border-b border-slate-800/80 pb-5">
                <div className="flex items-start space-x-4">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-purple-900/80 to-cyan-950 border border-purple-500/50 flex items-center justify-center text-2xl font-bold font-mono text-white shadow-inner">
                    {character.name.charAt(0)}
                  </div>
                  <div>
                    <div className="flex flex-wrap items-center gap-2">
                      <h2 className="text-2xl font-extrabold text-white tracking-tight">
                        {character.name}
                      </h2>
                      {character.title && (
                        <span className="text-xs px-2.5 py-1 rounded-full bg-cyan-950/80 border border-cyan-700 text-cyan-300 font-mono">
                          {character.title}
                        </span>
                      )}
                      <span className="text-xs px-2.5 py-1 rounded-full bg-purple-950/80 border border-purple-700 text-purple-300 font-mono">
                        SOVEREIGN CHARACTER
                      </span>
                    </div>

                    {character.backstory && (
                      <p className="text-xs text-slate-300 mt-1 max-w-2xl leading-relaxed italic">
                        &ldquo;{character.backstory}&rdquo;
                      </p>
                    )}

                    <div className="flex flex-wrap items-center gap-3 text-xs font-mono text-slate-400 mt-2">
                      <span className="flex items-center space-x-1">
                        <Calendar className="w-3.5 h-3.5 text-cyan-400" />
                        <span>{character.birth_date} {character.birth_time || "12:00"} (UTC{character.timezone_offset !== undefined && character.timezone_offset >= 0 ? `+${character.timezone_offset}` : character.timezone_offset ?? -4})</span>
                      </span>
                      <span>•</span>
                      <span className="flex items-center space-x-1">
                        <MapPin className="w-3.5 h-3.5 text-cyan-400" />
                        <span>{character.location_name || "St. Catharines / Niagara"}</span>
                      </span>
                      <span>•</span>
                      <span className="text-emerald-400">
                        {character.astrology?.janma_nakshatra?.name} (Pada {character.janma_pada})
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex flex-wrap items-center gap-2">
                  <button
                    onClick={() => handleDeleteCharacter(character.id)}
                    title="Dissolve character from storehouse"
                    className="p-2 rounded-xl bg-rose-950/40 hover:bg-rose-900/60 border border-rose-800 text-rose-400 hover:text-white transition"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>

              {/* Dossier Navigation Tabs */}
              <div className="flex flex-wrap items-center gap-2 text-xs font-mono">
                <button
                  onClick={() => setActiveTab("blueprint")}
                  className={`px-4 py-2 rounded-xl transition flex items-center space-x-2 ${
                    activeTab === "blueprint"
                      ? "bg-cyan-950 border border-cyan-500 text-cyan-300 font-bold shadow-md"
                      : "bg-slate-900 border border-slate-800 text-slate-400 hover:text-white"
                  }`}
                >
                  <Compass className="w-4 h-4" />
                  <span>1. Cosmic Blueprint &amp; Tripod</span>
                </button>

                <button
                  onClick={() => setActiveTab("alchemy")}
                  className={`px-4 py-2 rounded-xl transition flex items-center space-x-2 ${
                    activeTab === "alchemy"
                      ? "bg-amber-950 border border-amber-500 text-amber-300 font-bold shadow-md"
                      : "bg-slate-900 border border-slate-800 text-slate-400 hover:text-white"
                  }`}
                >
                  <Flame className="w-4 h-4" />
                  <span>2. Alchemical Crucible &amp; Metaphysics</span>
                </button>

                <button
                  onClick={() => setActiveTab("timeline")}
                  className={`px-4 py-2 rounded-xl transition flex items-center space-x-2 ${
                    activeTab === "timeline"
                      ? "bg-purple-950 border border-purple-500 text-purple-300 font-bold shadow-md"
                      : "bg-slate-900 border border-slate-800 text-slate-400 hover:text-white"
                  }`}
                >
                  <Clock className="w-4 h-4" />
                  <span>3. 120-Year Master Timeline</span>
                </button>

                <button
                  onClick={() => setActiveTab("chronicles")}
                  className={`px-4 py-2 rounded-xl transition flex items-center space-x-2 ${
                    activeTab === "chronicles"
                      ? "bg-red-950 border border-red-500 text-red-300 font-bold shadow-md"
                      : "bg-slate-900 border border-slate-800 text-slate-400 hover:text-white"
                  }`}
                >
                  <Film className="w-4 h-4" />
                  <span>4. Audiovisual Chronicles Studio</span>
                  {character.chronicles && character.chronicles.length > 0 && (
                    <span className="px-1.5 py-0.2 rounded-full bg-red-800 text-white text-[10px]">
                      {character.chronicles.length}
                    </span>
                  )}
                </button>
              </div>
            </section>

            {/* TAB 1: COSMIC BLUEPRINT & TRIPOD */}
            {activeTab === "blueprint" && (
              <section className="space-y-6 animate-in fade-in duration-200">
                {/* Tripod of Embodiment */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                  {/* Physical Vessel: Rising Lagna */}
                  <div className="p-5 rounded-2xl bg-[#0c1322] border border-cyan-800/60 space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-mono text-cyan-400 uppercase tracking-wider">
                        PHYSICAL VESSEL (BODY)
                      </span>
                      <span className="text-xs px-2 py-0.5 rounded bg-cyan-950 text-cyan-300 border border-cyan-800 font-mono">
                        BHAVA 1
                      </span>
                    </div>
                    <div>
                      <div className="text-xl font-bold text-white">
                        {character.astrology?.lagna?.rashi || "Ascendant"}
                      </div>
                      <div className="text-xs font-mono text-slate-400 mt-0.5">
                        {character.astrology?.lagna?.degree_in_sign?.toFixed(2)}° in sign • {character.astrology?.lagna?.nakshatra?.name} (Pada {character.astrology?.lagna?.pada})
                      </div>
                    </div>
                    <div className="text-[11px] font-mono text-slate-400 pt-2 border-t border-slate-800/80">
                      Ruler: <strong className="text-slate-200">{character.astrology?.lagna?.lord_name}</strong> ({character.astrology?.lagna?.element})
                    </div>
                  </div>

                  {/* Soul Purpose: Surya / Sun */}
                  <div className="p-5 rounded-2xl bg-[#0c1322] border border-amber-800/60 space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-mono text-amber-400 uppercase tracking-wider">
                        SOUL PURPOSE (ATMA)
                      </span>
                      <span className="text-xs px-2 py-0.5 rounded bg-amber-950 text-amber-300 border border-amber-800 font-mono">
                        BHAVA {character.astrology?.tripod?.surya?.house_number || 1}
                      </span>
                    </div>
                    <div>
                      <div className="text-xl font-bold text-white">
                        {character.astrology?.tripod?.surya?.rashi || "Surya"}
                      </div>
                      <div className="text-xs font-mono text-slate-400 mt-0.5">
                        {character.astrology?.tripod?.surya?.degree_in_sign?.toFixed(2)}° in sign • {character.astrology?.tripod?.surya?.nakshatra?.name}
                      </div>
                    </div>
                    <div className="text-[11px] font-mono text-slate-400 pt-2 border-t border-slate-800/80">
                      House: <strong className="text-slate-200">{character.astrology?.tripod?.surya?.house_name}</strong> • Lord: {character.astrology?.tripod?.surya?.rashi_lord}
                    </div>
                  </div>

                  {/* Mind & Intuition: Chandra / Moon */}
                  <div className="p-5 rounded-2xl bg-[#0c1322] border border-indigo-800/60 space-y-3">
                    <div className="flex items-center justify-between">
                      <span className="text-[10px] font-mono text-indigo-400 uppercase tracking-wider">
                        MIND &amp; PERCEPTION (MANAS)
                      </span>
                      <span className="text-xs px-2 py-0.5 rounded bg-indigo-950 text-indigo-300 border border-indigo-800 font-mono">
                        BHAVA {character.astrology?.tripod?.chandra?.house_number || 1}
                      </span>
                    </div>
                    <div>
                      <div className="text-xl font-bold text-white">
                        {character.astrology?.tripod?.chandra?.rashi || "Chandra"}
                      </div>
                      <div className="text-xs font-mono text-slate-400 mt-0.5">
                        {character.astrology?.janma_nakshatra?.name} (Pada {character.janma_pada}) • {character.astrology?.tripod?.chandra?.degree_in_sign?.toFixed(2)}°
                      </div>
                    </div>
                    <div className="text-[11px] font-mono text-slate-400 pt-2 border-t border-slate-800/80">
                      Janma Lord: <strong className="text-slate-200">{character.starting_lord}</strong> • Tattva: {character.astrology?.elemental_tattva}
                    </div>
                  </div>
                </div>

                {/* 9 Classical Nava Grahas Matrix */}
                <div className="p-6 rounded-3xl bg-[#0c1322] border border-slate-800 space-y-4">
                  <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                    <div className="flex items-center space-x-2">
                      <Orbit className="w-5 h-5 text-cyan-400" />
                      <h3 className="text-sm font-bold text-white tracking-wide">
                        9 Classical Nava Grahas &amp; Essential Dignities (Avasthas)
                      </h3>
                    </div>
                    <span className="text-[10px] font-mono text-slate-400">
                      Meeus orbital anomaly with Lahiri Ayanamsha subtraction
                    </span>
                  </div>

                  <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                    {character.astrology?.grahas?.map((g) => (
                      <div
                        key={g.id}
                        className="p-3.5 rounded-xl bg-[#090e1a] border border-slate-800/80 space-y-2 hover:border-slate-700 transition"
                      >
                        <div className="flex items-center justify-between">
                          <span className="font-bold text-sm text-white flex items-center space-x-1.5">
                            <span>{g.name}</span>
                            <span className="text-xs text-slate-400 font-mono">({g.sanskrit_name})</span>
                          </span>
                          <span
                            className="text-[10px] font-mono font-bold px-2 py-0.5 rounded-full border"
                            style={{
                              borderColor: g.dignity?.color_hex ? `${g.dignity.color_hex}77` : "#334155",
                              color: g.dignity?.color_hex || "#38bdf8",
                              backgroundColor: g.dignity?.color_hex ? `${g.dignity.color_hex}15` : "transparent",
                            }}
                          >
                            {g.dignity?.name || "Neutral"}
                          </span>
                        </div>

                        <div className="grid grid-cols-2 gap-1 text-[11px] font-mono text-slate-300">
                          <div>
                            <span className="text-slate-500 block">SIGN</span>
                            <span>{g.rashi}</span>
                          </div>
                          <div>
                            <span className="text-slate-500 block">HOUSE</span>
                            <span className="text-cyan-300">Bhava {g.house_number} ({g.house_name})</span>
                          </div>
                        </div>

                        <div className="text-[10px] font-mono text-slate-400 pt-1.5 border-t border-slate-800/80 flex justify-between">
                          <span>{g.degree_in_sign.toFixed(2)}° in sign</span>
                          <span className="text-slate-500">{g.nakshatra?.name} P{g.pada}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* 12-Bhava Whole Sign Temple Grid */}
                <div className="p-6 rounded-3xl bg-[#0c1322] border border-slate-800 space-y-4">
                  <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                    <div className="flex items-center space-x-2">
                      <Layers className="w-5 h-5 text-emerald-400" />
                      <h3 className="text-sm font-bold text-white tracking-wide">
                        12-Bhava Whole Sign Temple Matrix
                      </h3>
                    </div>
                    <span className="text-[10px] font-mono text-slate-400">
                      Classical Houses derived from Sidereal Lagna
                    </span>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
                    {character.astrology?.tripod?.bhavas?.map((b) => (
                      <div
                        key={b.house_number}
                        className={`p-3.5 rounded-xl border text-left transition ${
                          b.luminaries && b.luminaries.length > 0
                            ? "bg-[#111c30] border-cyan-700/60 shadow-md"
                            : "bg-[#090e1a] border-slate-800/80"
                        }`}
                      >
                        <div className="flex items-center justify-between">
                          <span className="text-xs font-bold font-mono text-cyan-400">
                            Bhava {b.house_number}: {b.sanskrit_name}
                          </span>
                          <span className="text-[10px] font-mono text-slate-500">
                            Lord: {b.rashi_lord}
                          </span>
                        </div>
                        <div className="text-sm font-bold text-white mt-1">
                          {b.rashi}
                        </div>
                        <div className="text-[10px] text-slate-400 mt-1 line-clamp-1">
                          {b.domain}
                        </div>

                        {b.luminaries && b.luminaries.length > 0 && (
                          <div className="mt-2 pt-1.5 border-t border-slate-800 flex flex-wrap gap-1">
                            {b.luminaries.map((lum, idx) => (
                              <span
                                key={idx}
                                className="text-[9px] font-mono font-bold px-1.5 py-0.2 rounded bg-cyan-950 border border-cyan-800 text-cyan-300"
                              >
                                {lum}
                              </span>
                            ))}
                          </div>
                        )}
                      </div>
                    ))}
                  </div>
                </div>

                {/* Panchanga & Ayurvedic Triad */}
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  {/* Panchanga */}
                  <div className="p-5 rounded-2xl bg-[#0c1322] border border-slate-800 space-y-3">
                    <span className="text-[10px] font-mono text-cyan-400 uppercase tracking-wider block">
                      VEDIC PANCHANGA (THE 5 COSMIC LIMBS)
                    </span>
                    <div className="grid grid-cols-2 gap-2 text-xs font-mono text-slate-300">
                      <div>
                        <span className="text-slate-500 block">VARA (SOLAR DAY)</span>
                        <span className="font-bold text-white">{character.astrology?.panchanga?.vara}</span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">TITHI (LUNAR PHASE)</span>
                        <span className="font-bold text-amber-400">
                          {character.astrology?.panchanga?.tithi_name} ({character.astrology?.panchanga?.paksha})
                        </span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">YOGA (SOLILUNAR)</span>
                        <span className="font-bold text-purple-300">{character.astrology?.panchanga?.yoga_name}</span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">KARANA (HALF-TITHI)</span>
                        <span className="font-bold text-emerald-300">{character.astrology?.panchanga?.karana_name}</span>
                      </div>
                    </div>
                  </div>

                  {/* Ayurveda */}
                  <div className="p-5 rounded-2xl bg-[#0c1322] border border-slate-800 space-y-3">
                    <span className="text-[10px] font-mono text-emerald-400 uppercase tracking-wider block">
                      AYURVEDIC CONSTITUTION &amp; TOTEMS
                    </span>
                    <div className="grid grid-cols-2 gap-2 text-xs font-mono text-slate-300">
                      <div>
                        <span className="text-slate-500 block">DOSHA (BIOLOGICAL)</span>
                        <span className="font-bold text-emerald-400">{character.astrology?.ayurveda?.dosha}</span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">GANA (TEMPERAMENT)</span>
                        <span className="font-bold text-cyan-300">{character.astrology?.ayurveda?.gana}</span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">YONI ANIMAL TOTEM</span>
                        <span className="font-bold text-amber-300">{character.astrology?.ayurveda?.yoni_totem}</span>
                      </div>
                      <div>
                        <span className="text-slate-500 block">NADI (MERIDIAN)</span>
                        <span className="font-bold text-purple-300">{character.astrology?.ayurveda?.nadi}</span>
                      </div>
                    </div>
                  </div>
                </div>
              </section>
            )}

            {/* TAB 2: ALCHEMICAL CRUCIBLE & METAPHYSICS */}
            {activeTab === "alchemy" && (
              <section className="space-y-6 animate-in fade-in duration-200">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  {/* Natal Fixed Crucible */}
                  <div className="p-6 rounded-3xl bg-gradient-to-br from-[#0c1322] to-[#121c2e] border border-amber-900/60 shadow-xl space-y-4">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                      <div className="flex items-center space-x-2">
                        <Flame className="w-5 h-5 text-amber-400" />
                        <h3 className="text-sm font-bold text-white tracking-wide">
                          Natal Vessel (The Fixed Sovereign Anchor)
                        </h3>
                      </div>
                      <span className="text-xs font-mono px-2 py-0.5 rounded bg-amber-950 text-amber-300 border border-amber-800">
                        NATAL CORE
                      </span>
                    </div>

                    <div className="space-y-3">
                      <div className="flex items-center justify-between">
                        <div>
                          <span className="text-xs text-slate-400 font-mono">Sacred Birth Metal:</span>
                          <div className="text-2xl font-bold text-amber-400 flex items-center space-x-2">
                            <span>{character.alchemy?.sacred_metal?.name}</span>
                            <span className="text-xl text-slate-300">{character.alchemy?.sacred_metal?.symbol}</span>
                          </div>
                        </div>
                        <div className="text-right">
                          <span className="text-xs text-slate-400 font-mono">Vibrational Frequency:</span>
                          <div className="text-xl font-bold text-cyan-300 font-mono">
                            {character.alchemy?.resonant_frequency_hz} Hz
                          </div>
                        </div>
                      </div>

                      <div className="p-4 rounded-xl bg-slate-900/80 border border-slate-800 space-y-1.5 text-xs font-mono">
                        <span className="text-slate-400 block font-semibold">Governing Hermetic Axiom:</span>
                        <div className="text-sm font-serif italic text-cyan-200">
                          &ldquo;{character.alchemy?.governing_axiom?.canonical_text}&rdquo;
                        </div>
                        <div className="text-[11px] text-slate-400 mt-1">
                          {character.alchemy?.governing_axiom?.title}
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-2 text-xs font-mono text-slate-300 pt-1">
                        <div>
                          <span className="text-slate-500 block">CHAKRA CENTER</span>
                          <span>{character.alchemy?.chakra_anchor}</span>
                        </div>
                        <div>
                          <span className="text-slate-500 block">MAGNUM OPUS STAGE</span>
                          <span className="text-amber-300 font-bold">{character.alchemy?.magnum_opus_stage?.name}</span>
                        </div>
                      </div>

                      <div className="p-3 rounded-xl bg-amber-950/20 border border-amber-800/40 text-xs font-mono text-amber-300">
                        {character.alchemy?.alchemical_motto}
                      </div>
                    </div>
                  </div>

                  {/* Active Transmutation Sky Vessel */}
                  <div className="p-6 rounded-3xl bg-gradient-to-br from-[#0c1322] to-[#16122e] border border-purple-900/60 shadow-xl space-y-4">
                    <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                      <div className="flex items-center space-x-2">
                        <Sparkles className="w-5 h-5 text-purple-400" />
                        <h3 className="text-sm font-bold text-white tracking-wide">
                          Sky Vessel (The Volatile Transmutation Horizon)
                        </h3>
                      </div>
                      <span className="text-xs font-mono px-2 py-0.5 rounded bg-purple-950 text-purple-300 border border-purple-800 animate-pulse">
                        LIVE TRANSIT
                      </span>
                    </div>

                    <div className="space-y-3">
                      <div className="p-4 rounded-xl bg-purple-950/30 border border-purple-800/60 space-y-1 text-xs font-mono">
                        <span className="text-purple-400 block font-bold uppercase">
                          Active Transmutation Vessel:
                        </span>
                        <div className="text-base font-bold text-white">
                          {character.active_alchemy?.transmutation_vessel}
                        </div>
                      </div>

                      <div className="grid grid-cols-2 gap-3 text-xs font-mono">
                        <div className="p-3 rounded-xl bg-slate-900 border border-slate-800">
                          <span className="text-slate-500 block text-[10px]">MAHADASHA METAL</span>
                          <span className="font-bold text-amber-300 text-sm">
                            {character.active_alchemy?.mahadasha_metal?.name} ({character.active_alchemy?.mahadasha_metal?.symbol})
                          </span>
                        </div>
                        <div className="p-3 rounded-xl bg-slate-900 border border-slate-800">
                          <span className="text-slate-500 block text-[10px]">ANTARDASHA METAL</span>
                          <span className="font-bold text-cyan-300 text-sm">
                            {character.active_alchemy?.antardasha_metal?.name} ({character.active_alchemy?.antardasha_metal?.symbol})
                          </span>
                        </div>
                      </div>

                      {/* Symbiotic Resonance Status */}
                      {character.symbiotic_resonance && (
                        <div className="p-4 rounded-xl bg-slate-900/80 border border-cyan-800/60 space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="text-[10px] font-mono text-cyan-400 uppercase tracking-wider">
                              CURRENT TRANSIT HORA RESONANCE
                            </span>
                            <span className="text-xs font-mono font-bold text-emerald-400">
                              {character.symbiotic_resonance.resonance_tier}
                            </span>
                          </div>
                          <p className="text-xs text-slate-200 leading-relaxed font-sans">
                            {character.symbiotic_resonance.transmutation_guidance}
                          </p>
                        </div>
                      )}
                    </div>
                  </div>
                </div>
              </section>
            )}

            {/* TAB 3: 120-YEAR VIMSHOTTARI MASTER TIMELINE */}
            {activeTab === "timeline" && (
              <section className="space-y-4 animate-in fade-in duration-200">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 rounded-2xl bg-[#0c1322] border border-slate-800 text-xs font-mono">
                  <div>
                    <h3 className="text-sm font-bold text-white tracking-wide">
                      120-Year Vimshottari Master Timeline Hierarchy
                    </h3>
                    <p className="text-slate-400 mt-0.5">
                      Enriched with governing sacred metals, Hermetic principles, and character age progression.
                    </p>
                  </div>

                  <div className="flex items-center space-x-2">
                    <span className="text-cyan-400 font-bold">120 Total Years</span>
                    <span className="text-slate-600">•</span>
                    <span className="text-amber-400">Balance at Birth: {character.balance_years.toFixed(2)}y</span>
                  </div>
                </div>

                {timelineLoading ? (
                  <div className="p-12 flex flex-col items-center justify-center space-y-2 text-slate-400 bg-slate-900/30 rounded-2xl">
                    <Clock className="w-6 h-6 animate-spin text-purple-400" />
                    <span className="text-xs font-mono">Unfolding 120-year sub-period tree...</span>
                  </div>
                ) : (
                  <div className="space-y-3">
                    {timeline.map((maha, mIdx) => {
                      const isExpanded = !!expandedMaha[mIdx];
                      const isActive = character.active_snapshot?.mahadasha?.planet === maha.planet;
                      return (
                        <div
                          key={mIdx}
                          className={`rounded-2xl border transition-all overflow-hidden ${
                            isActive
                              ? "bg-gradient-to-r from-purple-950/40 via-[#0c1322] to-[#0c1322] border-purple-500/80 shadow-lg"
                              : "bg-[#0c1322] border-slate-800"
                          }`}
                        >
                          {/* Mahadasha Header */}
                          <div
                            onClick={() => toggleMaha(mIdx)}
                            className="p-4 flex items-center justify-between cursor-pointer hover:bg-slate-800/40 transition select-none"
                          >
                            <div className="flex items-center space-x-3.5">
                              <div
                                className="w-10 h-10 rounded-xl flex items-center justify-center font-bold text-base text-white font-mono shadow-md"
                                style={{ backgroundColor: maha.color_hex ? `${maha.color_hex}33` : "#3b82f6", border: `1px solid ${maha.color_hex || "#3b82f6"}` }}
                              >
                                {maha.metal_symbol || "☿"}
                              </div>
                              <div>
                                <div className="flex items-center space-x-2">
                                  <span className="font-bold text-base text-white">
                                    {maha.planet_name} Mahadasha
                                  </span>
                                  <span className="text-xs font-mono text-slate-400">
                                    ({maha.sanskrit_name})
                                  </span>
                                  {isActive && (
                                    <span className="text-[10px] font-mono font-bold px-2 py-0.5 rounded-full bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 animate-pulse">
                                      ACTIVE NOW
                                    </span>
                                  )}
                                </div>
                                <div className="text-xs font-mono text-slate-400 flex items-center space-x-2 mt-0.5">
                                  <span>{maha.sacred_metal}</span>
                                  <span>•</span>
                                  <span className="text-amber-300 font-bold">{maha.age_start !== undefined ? `Age ${maha.age_start.toFixed(1)} – ${maha.age_end?.toFixed(1)}` : ""}</span>
                                  <span>•</span>
                                  <span>{new Date(maha.start_date).toLocaleDateString()} to {new Date(maha.end_date).toLocaleDateString()}</span>
                                </div>
                              </div>
                            </div>

                            <div className="flex items-center space-x-3">
                              <button
                                onClick={(e) => {
                                  e.stopPropagation();
                                  setSelectedEpochPlanet(maha.planet);
                                  setActiveTab("chronicles");
                                }}
                                className="px-3 py-1 rounded-lg bg-red-950/60 hover:bg-red-900/80 border border-red-800 text-red-300 hover:text-white text-xs font-mono flex items-center space-x-1.5 transition"
                              >
                                <Film className="w-3.5 h-3.5" />
                                <span>Epoch Chronicle</span>
                              </button>
                              <ChevronDown className={`w-4 h-4 text-slate-400 transition-transform duration-300 ${isExpanded ? "rotate-180" : ""}`} />
                            </div>
                          </div>

                          {/* Antardashas Child Accordion */}
                          {isExpanded && maha.sub_periods && maha.sub_periods.length > 0 && (
                            <div className="border-t border-slate-800/80 bg-[#080d1a] p-4 space-y-2.5">
                              <div className="text-[11px] font-mono text-slate-500 uppercase tracking-wider mb-2">
                                9 Sub-Epochs (Antardashas) of {maha.planet_name}
                              </div>
                              <div className="grid grid-cols-1 md:grid-cols-3 gap-2.5">
                                {maha.sub_periods.map((antar, aIdx) => {
                                  const isAntarActive = character.active_snapshot?.antardasha?.planet === antar.planet && isActive;
                                  return (
                                    <div
                                      key={aIdx}
                                      className={`p-3 rounded-xl border text-left text-xs font-mono transition ${
                                        isAntarActive
                                          ? "bg-purple-950/50 border-purple-500 shadow-md ring-1 ring-purple-500/50"
                                          : "bg-[#0c1322] border-slate-800 hover:border-slate-700"
                                      }`}
                                    >
                                      <div className="flex items-center justify-between">
                                        <span className="font-bold text-white">
                                          {antar.planet_name} Sub
                                        </span>
                                        <span className="text-slate-400 font-bold">
                                          {antar.metal_symbol}
                                        </span>
                                      </div>
                                      <div className="text-[11px] text-amber-300 mt-1">
                                        {antar.age_start !== undefined ? `Age ${antar.age_start.toFixed(1)} – ${antar.age_end?.toFixed(1)}` : ""}
                                      </div>
                                      <div className="text-[10px] text-slate-500 mt-0.5">
                                        {new Date(antar.start_date).toLocaleDateString()} – {new Date(antar.end_date).toLocaleDateString()}
                                      </div>
                                      {isAntarActive && (
                                        <div className="mt-1.5 text-[9px] font-bold text-emerald-400 bg-emerald-950/60 px-1.5 py-0.5 rounded text-center">
                                          CURRENT SUB-CYCLE
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
                    })}
                  </div>
                )}
              </section>
            )}

            {/* TAB 4: AUDIOVISUAL CHRONICLES & VIDEO STUDIO */}
            {activeTab === "chronicles" && (
              <section className="space-y-6 animate-in fade-in duration-200">
                {/* Chronicle Compiler Panel */}
                <div className="p-6 rounded-3xl bg-gradient-to-br from-[#0e172a] to-[#0c1322] border border-red-900/50 shadow-2xl space-y-5">
                  <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 border-b border-slate-800/80 pb-4">
                    <div className="flex items-center space-x-3">
                      <div className="p-2.5 rounded-2xl bg-red-950/80 border border-red-800 text-red-400">
                        <Film className="w-5 h-5" />
                      </div>
                      <div>
                        <h3 className="text-base font-bold text-white tracking-tight">
                          Audiovisual Sovereign Chronicle Compiler
                        </h3>
                        <p className="text-xs text-slate-400">
                          Compile multi-track video generation timeline manifests for YouTube Sovereign Studio from Dasha epochs.
                        </p>
                      </div>
                    </div>

                    <button
                      onClick={handleCompileChronicle}
                      disabled={compilingChronicle}
                      className="px-5 py-2.5 rounded-xl bg-red-600 hover:bg-red-500 text-white font-bold text-xs font-mono flex items-center space-x-2 transition shadow-lg shadow-red-950/50 disabled:opacity-50"
                    >
                      <Video className={`w-4 h-4 ${compilingChronicle ? "animate-spin" : ""}`} />
                      <span>{compilingChronicle ? "Compiling..." : "Compile Presentation Timeline"}</span>
                    </button>
                  </div>

                  {/* Chronicle Configuration Grid */}
                  <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 text-xs font-mono">
                    {/* Target Dasha Epoch */}
                    <div>
                      <label className="text-slate-400 block mb-1.5 font-bold">Target Dasha Epoch</label>
                      <select
                        value={selectedEpochPlanet}
                        onChange={(e) => setSelectedEpochPlanet(e.target.value)}
                        className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-slate-200 focus:border-red-500 focus:outline-none"
                      >
                        {character.timeline_summary?.map((m) => (
                          <option key={m.planet} value={m.planet}>
                            {m.planet_name} Mahadasha ({m.sacred_metal}) • {m.age_range}
                          </option>
                        ))}
                      </select>
                    </div>

                    {/* Canvas Aspect Ratio */}
                    <div>
                      <label className="text-slate-400 block mb-1.5 font-bold">Canvas Orientation</label>
                      <div className="grid grid-cols-3 gap-2">
                        <button
                          type="button"
                          onClick={() => setChronicleOrientation("16:9")}
                          className={`py-2 rounded-xl border text-center transition ${
                            chronicleOrientation === "16:9"
                              ? "bg-red-950/80 border-red-500 text-red-300 font-bold"
                              : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                          }`}
                        >
                          16:9 (YouTube)
                        </button>
                        <button
                          type="button"
                          onClick={() => setChronicleOrientation("9:16")}
                          className={`py-2 rounded-xl border text-center transition ${
                            chronicleOrientation === "9:16"
                              ? "bg-red-950/80 border-red-500 text-red-300 font-bold"
                              : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                          }`}
                        >
                          9:16 (Shorts)
                        </button>
                        <button
                          type="button"
                          onClick={() => setChronicleOrientation("1:1")}
                          className={`py-2 rounded-xl border text-center transition ${
                            chronicleOrientation === "1:1"
                              ? "bg-red-950/80 border-red-500 text-red-300 font-bold"
                              : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                          }`}
                        >
                          1:1 (Square)
                        </button>
                      </div>
                    </div>

                    {/* Presentation Duration */}
                    <div>
                      <label className="text-slate-400 block mb-1.5 font-bold">Presentation Duration</label>
                      <div className="grid grid-cols-4 gap-1.5">
                        {[60, 90, 120, 180].map((d) => (
                          <button
                            key={d}
                            type="button"
                            onClick={() => setChronicleDuration(d)}
                            className={`py-2 rounded-xl border text-center transition ${
                              chronicleDuration === d
                                ? "bg-red-950/80 border-red-500 text-red-300 font-bold"
                                : "bg-slate-900 border-slate-800 text-slate-400 hover:text-white"
                            }`}
                          >
                            {d}s
                          </button>
                        ))}
                      </div>
                    </div>
                  </div>

                  {chronicleSuccess && (
                    <div className="p-3.5 rounded-xl bg-emerald-950/40 border border-emerald-500/40 flex items-center justify-between text-xs font-mono text-emerald-300">
                      <div className="flex items-center space-x-2">
                        <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                        <span>Timeline Manifest successfully compiled and saved to Sovereign Storehouse!</span>
                      </div>
                      <button
                        onClick={() => setShowManifestJson(!showManifestJson)}
                        className="underline hover:text-white transition"
                      >
                        {showManifestJson ? "Hide JSON" : "Inspect Manifest JSON"}
                      </button>
                    </div>
                  )}

                  {/* Multi-Track Visual Timeline Presentation */}
                  {compiledManifest && (
                    <div className="p-5 rounded-2xl bg-[#080d1a] border border-slate-800 space-y-4">
                      <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                        <div className="flex items-center space-x-2">
                          <Activity className="w-4 h-4 text-red-400" />
                          <span className="text-xs font-bold text-white font-mono uppercase tracking-wider">
                            Multi-Track Presentation Canvas ({compiledManifest.canvas.width}x{compiledManifest.canvas.height} • {compiledManifest.duration_sec}s)
                          </span>
                        </div>
                        <span className="text-[10px] font-mono text-cyan-300 bg-cyan-950 px-2 py-0.5 rounded border border-cyan-800">
                          {compiledManifest.chrono_meta?.vimshottari_lord}
                        </span>
                      </div>

                      {/* Track 1: Visual Scenes */}
                      <div className="space-y-1.5">
                        <span className="text-[10px] font-mono text-slate-500 block uppercase">TRACK 1: VISUAL NARRATIVE SCENES</span>
                        <div className="grid grid-cols-1 sm:grid-cols-4 gap-2">
                          {compiledManifest.scenes.map((s, idx) => (
                            <div key={s.id} className="p-3 rounded-xl bg-[#0e1628] border border-slate-800 space-y-1 text-xs font-mono">
                              <div className="flex items-center justify-between text-[10px] text-slate-500">
                                <span>Scene {idx + 1}</span>
                                <span>{s.duration_sec.toFixed(1)}s</span>
                              </div>
                              <div className="font-bold text-white line-clamp-1">{s.caption}</div>
                              <div className="text-[10px] text-slate-400 line-clamp-1">{s.subtitle}</div>
                              <div className="pt-1 text-[9px] text-cyan-400 flex justify-between">
                                <span>{s.motion}</span>
                                <span>{s.transition_in}</span>
                              </div>
                            </div>
                          ))}
                        </div>
                      </div>

                      {/* Track 2: Audio Composition */}
                      <div className="space-y-1.5 pt-2">
                        <span className="text-[10px] font-mono text-slate-500 block uppercase">TRACK 2: HARMONIC AUDIO &amp; DUCKING</span>
                        <div className="p-3 rounded-xl bg-[#0e1628] border border-slate-800 flex flex-wrap items-center justify-between gap-2 text-xs font-mono">
                          <div className="flex items-center space-x-2">
                            <Music className="w-4 h-4 text-purple-400" />
                            <span className="text-white font-bold">Resonant Frequency Drone:</span>
                            <span className="text-cyan-300">{compiledManifest.audio_tracks[0]?.source_path}</span>
                          </div>
                          <span className="text-[10px] text-emerald-400 bg-emerald-950/60 px-2 py-0.5 rounded border border-emerald-800">
                            AUTO-DUCKING ENABLED (-14dB)
                          </span>
                        </div>
                      </div>

                      {/* Track 3: Telemetry Overlays */}
                      <div className="space-y-1.5 pt-2">
                        <span className="text-[10px] font-mono text-slate-500 block uppercase">TRACK 3: CHRONO OVERLAYS &amp; WAVEFORM</span>
                        <div className="flex flex-wrap gap-2 text-[11px] font-mono">
                          {compiledManifest.overlays?.map((o) => (
                            <span key={o.id} className="px-2.5 py-1 rounded-lg bg-slate-900 border border-slate-800 text-slate-300">
                              {o.type} ({o.position})
                            </span>
                          ))}
                        </div>
                      </div>

                      {showManifestJson && (
                        <pre className="p-4 rounded-xl bg-black/60 border border-slate-800 text-[11px] font-mono text-cyan-300 overflow-x-auto max-h-72">
                          {JSON.stringify(compiledManifest, null, 2)}
                        </pre>
                      )}
                    </div>
                  )}
                </div>

                {/* Previously Compiled Chronicles Archive */}
                <div className="p-6 rounded-3xl bg-[#0c1322] border border-slate-800 space-y-4">
                  <div className="flex items-center justify-between border-b border-slate-800 pb-3">
                    <div className="flex items-center space-x-2">
                      <Film className="w-5 h-5 text-red-400" />
                      <h3 className="text-sm font-bold text-white tracking-wide">
                        Saved Chronicles Archive ({character.chronicles?.length || 0})
                      </h3>
                    </div>
                    <Link
                      href="/foundations/"
                      className="text-xs font-mono text-red-400 hover:text-white flex items-center space-x-1"
                    >
                      <span>Foundations Studio →</span>
                    </Link>
                  </div>

                  {character.chronicles && character.chronicles.length > 0 ? (
                    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
                      {character.chronicles.map((c) => (
                        <div key={c.id} className="p-4 rounded-xl bg-[#090e1a] border border-slate-800 space-y-2">
                          <div className="flex items-center justify-between">
                            <span className="font-bold text-sm text-white line-clamp-1">{c.title}</span>
                            <span className="text-[10px] font-mono px-1.5 py-0.5 rounded bg-red-950 text-red-300 border border-red-800">
                              {c.orientation}
                            </span>
                          </div>
                          <div className="text-xs font-mono text-slate-400 flex justify-between">
                            <span>{c.planet} ({c.dasha_level})</span>
                            <span className="text-cyan-400">{c.duration_sec}s</span>
                          </div>
                          <div className="text-[10px] font-mono text-slate-500 pt-1 border-t border-slate-800/80">
                            Created: {new Date(c.created_at).toLocaleString()}
                          </div>
                        </div>
                      ))}
                    </div>
                  ) : (
                    <div className="p-8 text-center text-xs font-mono text-slate-500">
                      No chronicles compiled yet. Select a Dasha epoch above and compile your first sovereign presentation manifest.
                    </div>
                  )}
                </div>
              </section>
            )}
          </>
        )}
      </div>

      {/* Forge Character Slide-Over / Modal */}
      {showForge && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80 backdrop-blur-md animate-in fade-in duration-200">
          <div className="w-full max-w-xl bg-[#0c1322] border border-purple-800/80 rounded-3xl p-6 shadow-2xl space-y-5 font-sans">
            <div className="flex items-center justify-between border-b border-slate-800 pb-3">
              <div className="flex items-center space-x-2.5">
                <Users className="w-5 h-5 text-purple-400" />
                <h3 className="text-base font-bold text-white">Forge Sovereign Character</h3>
              </div>
              <button
                onClick={() => setShowForge(false)}
                className="text-slate-400 hover:text-white transition text-xs font-mono"
              >
                Cancel
              </button>
            </div>

            <form onSubmit={handleForgeCharacter} className="space-y-4 text-xs font-mono">
              {forgeError && (
                <div className="p-3 rounded-xl bg-rose-950/60 border border-rose-800 text-rose-300 flex items-center space-x-2">
                  <AlertCircle className="w-4 h-4" />
                  <span>{forgeError}</span>
                </div>
              )}

              <div>
                <label className="text-slate-400 block mb-1">Character Name *</label>
                <input
                  type="text"
                  required
                  value={forgeName}
                  onChange={(e) => setForgeName(e.target.value)}
                  placeholder="e.g. Aurelius Drake"
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-slate-400 block mb-1">Title / Epithet</label>
                  <input
                    type="text"
                    value={forgeTitle}
                    onChange={(e) => setForgeTitle(e.target.value)}
                    placeholder="e.g. The Sovereign Architect"
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-slate-400 block mb-1">Location Label</label>
                  <input
                    type="text"
                    value={forgeCity}
                    onChange={(e) => setForgeCity(e.target.value)}
                    placeholder="St. Catharines, ON"
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
              </div>

              <div>
                <label className="text-slate-400 block mb-1">Character Lore / Backstory</label>
                <textarea
                  rows={2}
                  value={forgeBackstory}
                  onChange={(e) => setForgeBackstory(e.target.value)}
                  placeholder="The foundational narrative archetype..."
                  className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none resize-none font-sans text-xs"
                />
              </div>

              <div className="grid grid-cols-3 gap-3">
                <div>
                  <label className="text-slate-400 block mb-1">Birth Date *</label>
                  <input
                    type="date"
                    required
                    value={forgeDate}
                    onChange={(e) => setForgeDate(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-slate-400 block mb-1">Birth Time</label>
                  <input
                    type="time"
                    value={forgeTime}
                    onChange={(e) => setForgeTime(e.target.value)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-slate-400 block mb-1">UTC Offset</label>
                  <input
                    type="number"
                    step="0.5"
                    value={forgeTz}
                    onChange={(e) => setForgeTz(parseFloat(e.target.value) || 0)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-slate-400 block mb-1">Latitude (°N)</label>
                  <input
                    type="number"
                    step="0.0001"
                    value={forgeLat}
                    onChange={(e) => setForgeLat(parseFloat(e.target.value) || 0)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
                <div>
                  <label className="text-slate-400 block mb-1">Longitude (°E)</label>
                  <input
                    type="number"
                    step="0.0001"
                    value={forgeLon}
                    onChange={(e) => setForgeLon(parseFloat(e.target.value) || 0)}
                    className="w-full bg-slate-900 border border-slate-700 rounded-xl px-3 py-2 text-white focus:border-purple-500 focus:outline-none"
                  />
                </div>
              </div>

              <div className="pt-3 border-t border-slate-800 flex justify-end space-x-3">
                <button
                  type="button"
                  onClick={() => setShowForge(false)}
                  className="px-4 py-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-300 transition"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  disabled={forging}
                  className="px-5 py-2 rounded-xl bg-purple-600 hover:bg-purple-500 text-white font-bold transition flex items-center space-x-1.5 disabled:opacity-50"
                >
                  <Sparkles className="w-3.5 h-3.5" />
                  <span>{forging ? "Forging..." : "Forge Character"}</span>
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
