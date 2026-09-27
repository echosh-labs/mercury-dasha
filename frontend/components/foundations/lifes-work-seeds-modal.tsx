"use client";

import React, { useState, useEffect, useCallback } from "react";
import {
  X,
  Sparkles,
  BookOpen,
  Mic,
  Calendar,
  Layers,
  ArrowRight,
  CheckCircle2,
  Bookmark,
  Compass,
  Radio,
  Clock,
  Search,
  RefreshCw,
  Film,
  Zap,
} from "lucide-react";
import {
  StorySeed,
  ResonantExcerpt,
  ResonantTransitResponse,
} from "@/lib/types/foundations";
import { useToast } from "@/lib/toast-context";

interface LifesWorkSeedsModalProps {
  isOpen: boolean;
  onClose: () => void;
  onLoadDraft: (draft: {
    title: string;
    description: string;
    tags: string[];
    filePath: string;
  }) => void;
}

export default function LifesWorkSeedsModal({
  isOpen,
  onClose,
  onLoadDraft,
}: LifesWorkSeedsModalProps) {
  const { success: toastSuccess, error: toastError } = useToast();

  const [activeMode, setActiveMode] = useState<"catalog" | "resonant" | "saved">("resonant");
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedSanctuary, setSelectedSanctuary] = useState<string>("all");

  const [catalogItems, setCatalogItems] = useState<any[]>([]);
  const [resonantData, setResonantData] = useState<ResonantTransitResponse | null>(null);
  const [savedSeeds, setSavedSeeds] = useState<StorySeed[]>([]);
  const [loading, setLoading] = useState<boolean>(false);

  // Selected item for preview/synthesis
  const [selectedItem, setSelectedItem] = useState<any | null>(null);
  const [currentSeed, setCurrentSeed] = useState<StorySeed | null>(null);
  const [synthesizing, setSynthesizing] = useState<boolean>(false);
  const [savingSeed, setSavingSeed] = useState<boolean>(false);

  // 1. Fetch Resonant Transit Excerpts
  const fetchResonant = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/foundations/seeds/resonant?limit=15");
      if (res.ok) {
        const data: ResonantTransitResponse = await res.json();
        setResonantData(data);
        if (data.excerpts && data.excerpts.length > 0 && !selectedItem) {
          handleSelectItem(data.excerpts[0]);
        }
      }
    } catch (err) {
      console.error("Failed to fetch resonant transit seeds:", err);
    } finally {
      setLoading(false);
    }
  }, [selectedItem]);

  // 2. Fetch Catalog Items
  const fetchCatalog = useCallback(async () => {
    setLoading(true);
    try {
      const url = new URL("/api/v1/text/catalog", window.location.origin);
      url.searchParams.set("limit", "50");
      if (searchQuery) url.searchParams.set("q", searchQuery);
      if (selectedSanctuary !== "all") url.searchParams.set("sanctuary", selectedSanctuary);

      const res = await fetch(url.toString());
      if (res.ok) {
        const data = await res.json();
        setCatalogItems(data.items || []);
        if (data.items && data.items.length > 0 && !selectedItem) {
          handleSelectItem(data.items[0]);
        }
      }
    } catch (err) {
      console.error("Failed to fetch text catalog:", err);
    } finally {
      setLoading(false);
    }
  }, [searchQuery, selectedSanctuary, selectedItem]);

  // 3. Fetch Saved Seeds
  const fetchSaved = useCallback(async () => {
    setLoading(true);
    try {
      const res = await fetch("/api/v1/foundations/seeds");
      if (res.ok) {
        const data = await res.json();
        setSavedSeeds(data.seeds || []);
        if (data.seeds && data.seeds.length > 0 && !selectedItem) {
          setCurrentSeed(data.seeds[0]);
          setSelectedItem(data.seeds[0]);
        }
      }
    } catch (err) {
      console.error("Failed to fetch saved story seeds:", err);
    } finally {
      setLoading(false);
    }
  }, [selectedItem]);

  useEffect(() => {
    if (!isOpen) return;
    if (activeMode === "resonant") {
      fetchResonant();
    } else if (activeMode === "catalog") {
      fetchCatalog();
    } else if (activeMode === "saved") {
      fetchSaved();
    }
  }, [isOpen, activeMode, fetchResonant, fetchCatalog, fetchSaved]);

  // Handle selecting an item and auto-synthesizing
  const handleSelectItem = async (item: any) => {
    setSelectedItem(item);
    // If it's already a full seed (from saved mode)
    if (item.scene_beats && item.cosmic_alignment) {
      setCurrentSeed(item);
      return;
    }

    setSynthesizing(true);
    try {
      const body: any = {};
      if (item.catalog_id?.startsWith("recorder:") || item.id?.startsWith("recorder:")) {
        body.id = item.catalog_id || item.id;
      } else {
        body.path = item.path;
      }

      const res = await fetch("/api/v1/foundations/seeds/from-text", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify(body),
      });

      if (res.ok) {
        const seed: StorySeed = await res.json();
        setCurrentSeed(seed);
      } else {
        toastError("Could not synthesize seed for this document.");
      }
    } catch (err: any) {
      toastError(`Synthesis error: ${err.message}`);
    } finally {
      setSynthesizing(false);
    }
  };

  // Save Current Seed to BoltDB
  const handleSaveSeed = async () => {
    if (!currentSeed) return;
    setSavingSeed(true);
    try {
      const res = await fetch("/api/v1/foundations/seeds/from-text", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({
          path: currentSeed.source_path,
          id: currentSeed.source_path.startsWith("audio/recorder/")
            ? `recorder:${currentSeed.source_path.replace("audio/recorder/", "").replace(".txt", "")}`
            : undefined,
          title: currentSeed.title,
          note_date: currentSeed.note_date,
          save: true,
        }),
      });

      if (res.ok) {
        toastSuccess("Story Seed saved to Sovereign Storehouse.");
        fetchSaved();
      } else {
        toastError("Failed to save seed.");
      }
    } catch (err: any) {
      toastError(`Save error: ${err.message}`);
    } finally {
      setSavingSeed(false);
    }
  };

  // 1-Click Load into Video Draft
  const handleLoadDraft = () => {
    if (!currentSeed) return;

    let desc = `Foundations Chronicle • ${currentSeed.title}\n\n`;
    desc += `Core Thesis: ${currentSeed.narrative_hook}\n\n`;
    if (currentSeed.dialogue_anchor) {
      desc += `Dialogue Quote: "${currentSeed.dialogue_anchor}"\n\n`;
    }
    desc += `Vimshottari Dasha: ${currentSeed.cosmic_alignment.mahadasha} (${currentSeed.cosmic_alignment.mahadasha_sanskrit}) / ${currentSeed.cosmic_alignment.antardasha} (${currentSeed.cosmic_alignment.antardasha_sanskrit})\n`;
    desc += `Sacred Metal: ${currentSeed.cosmic_alignment.sacred_metal} • Hermetic Axiom: ${currentSeed.cosmic_alignment.hermetic_axiom}\n`;
    desc += `Magnum Opus Stage: ${currentSeed.cosmic_alignment.magnum_opus_stage} • Solfeggio: ${currentSeed.cosmic_alignment.resonant_frequency_hz} Hz\n\n`;
    desc += `Transmuted by echosh-labs Mercury Dasha Engine.`;

    const tags = [
      "MercuryDasha",
      "echosh-labs",
      currentSeed.cosmic_alignment.mahadasha,
      currentSeed.sanctuary_id,
      "FoundationsStorySeed",
    ];

    let filePath = "";
    if (currentSeed.has_audio && currentSeed.audio_url) {
      filePath = currentSeed.source_path;
    }

    onLoadDraft({
      title: currentSeed.title,
      description: desc,
      tags,
      filePath,
    });

    toastSuccess(`Story Seed loaded into Video Pipeline: "${currentSeed.title}"`);
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 sm:p-6 bg-slate-950/80 backdrop-blur-md animate-in fade-in duration-200">
      <div className="bg-slate-900 border border-slate-800 rounded-3xl w-full max-w-6xl max-h-[90vh] flex flex-col shadow-2xl overflow-hidden">
        {/* 1. Modal Header */}
        <div className="p-6 border-b border-slate-800 flex items-center justify-between bg-slate-950/60">
          <div className="flex items-center space-x-3">
            <span className="p-2.5 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-400">
              <Sparkles className="w-5 h-5" />
            </span>
            <div>
              <div className="flex items-center space-x-2">
                <h3 className="text-lg font-bold text-white font-serif tracking-tight">
                  Import from Life&apos;s Work &amp; Lore Matrix
                </h3>
                <span className="bg-amber-500/10 text-amber-300 border border-amber-500/30 text-[10px] font-mono px-2 py-0.5 rounded-full flex items-center space-x-1">
                  <Radio className="w-3 h-3 text-amber-400 animate-pulse" />
                  <span>FOUNDATIONS SYNTHESIS</span>
                </span>
              </div>
              <p className="text-xs text-slate-400 mt-0.5">
                Transform personal writings and spoken audio chronicles into video drafts, narrative seeds, and timeline manifests.
              </p>
            </div>
          </div>

          <button
            onClick={onClose}
            className="p-2 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-400 hover:text-white transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* 2. Mode Navigation & Sub-Tabs */}
        <div className="px-6 py-3 border-b border-slate-800/80 bg-slate-950/30 flex flex-wrap items-center justify-between gap-4">
          <div className="flex items-center space-x-2">
            <button
              onClick={() => setActiveMode("resonant")}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-mono transition flex items-center space-x-1.5 ${
                activeMode === "resonant"
                  ? "bg-amber-500/20 text-amber-300 border border-amber-500/40"
                  : "bg-slate-800/60 text-slate-400 hover:text-slate-200 border border-slate-700/50"
              }`}
            >
              <Zap className="w-3.5 h-3.5 text-amber-400" />
              <span>Transit Resonance</span>
            </button>

            <button
              onClick={() => setActiveMode("catalog")}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-mono transition flex items-center space-x-1.5 ${
                activeMode === "catalog"
                  ? "bg-cyan-500/20 text-cyan-300 border border-cyan-500/40"
                  : "bg-slate-800/60 text-slate-400 hover:text-slate-200 border border-slate-700/50"
              }`}
            >
              <BookOpen className="w-3.5 h-3.5 text-cyan-400" />
              <span>Full Archive (926 Notes)</span>
            </button>

            <button
              onClick={() => setActiveMode("saved")}
              className={`px-3.5 py-1.5 rounded-xl text-xs font-mono transition flex items-center space-x-1.5 ${
                activeMode === "saved"
                  ? "bg-purple-500/20 text-purple-300 border border-purple-500/40"
                  : "bg-slate-800/60 text-slate-400 hover:text-slate-200 border border-slate-700/50"
              }`}
            >
              <Bookmark className="w-3.5 h-3.5 text-purple-400" />
              <span>Saved Story Seeds</span>
            </button>
          </div>

          {activeMode === "catalog" && (
            <div className="flex items-center space-x-2 text-xs font-mono">
              <div className="relative">
                <Search className="w-3.5 h-3.5 text-slate-500 absolute left-2.5 top-2.5" />
                <input
                  type="text"
                  placeholder="Filter notes..."
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="bg-slate-900 border border-slate-800 rounded-lg pl-8 pr-3 py-1.5 text-slate-200 focus:outline-none focus:border-cyan-500/50"
                />
              </div>

              <select
                value={selectedSanctuary}
                onChange={(e) => setSelectedSanctuary(e.target.value)}
                className="bg-slate-900 border border-slate-800 rounded-lg px-2.5 py-1.5 text-slate-300 focus:outline-none focus:border-cyan-500/50"
              >
                <option value="all">All Sanctuaries</option>
                <option value="blessings">Blessings</option>
                <option value="hermetic_kybalion">Kybalion / Hermetic</option>
                <option value="martial_discipline">Martial Discipline</option>
                <option value="dated_journals">Dated Journals</option>
                <option value="spoken_transcripts">Spoken Transcripts</option>
                <option value="foundations_lore">Foundations Lore</option>
              </select>
            </div>
          )}
        </div>

        {/* 3. Main Split View: Left List / Right Preview */}
        <div className="grid grid-cols-1 lg:grid-cols-12 flex-1 overflow-hidden divide-y lg:divide-y-0 lg:divide-x divide-slate-800">
          {/* Left Column: List of items */}
          <div className="lg:col-span-5 p-4 overflow-y-auto space-y-2.5 max-h-[60vh] lg:max-h-full">
            {loading ? (
              <div className="py-12 flex flex-col items-center justify-center text-xs font-mono text-slate-500 space-y-2">
                <RefreshCw className="w-5 h-5 animate-spin text-amber-400" />
                <span>Querying sovereign archives...</span>
              </div>
            ) : activeMode === "resonant" ? (
              <div>
                {resonantData?.current_transit && (
                  <div className="p-3 mb-3 rounded-xl bg-amber-950/30 border border-amber-500/20 text-xs font-mono space-y-1">
                    <div className="flex items-center justify-between text-amber-300 font-bold">
                      <span>Active Transit Alignment</span>
                      <span className="text-[10px] bg-amber-500/10 px-2 py-0.5 rounded border border-amber-500/30">
                        {resonantData.current_transit.mahadasha} / {resonantData.current_transit.antardasha}
                      </span>
                    </div>
                    <div className="text-[11px] text-slate-400">
                      Sacred Metal: <span className="text-slate-200">{resonantData.current_transit.sacred_metal}</span> • Axiom: <span className="text-slate-200">{resonantData.current_transit.hermetic_axiom}</span>
                    </div>
                  </div>
                )}

                {resonantData?.excerpts && resonantData.excerpts.length > 0 ? (
                  resonantData.excerpts.map((item) => (
                    <div
                      key={item.catalog_id}
                      onClick={() => handleSelectItem(item)}
                      className={`p-3.5 rounded-xl border cursor-pointer transition text-xs font-mono space-y-1.5 ${
                        selectedItem?.catalog_id === item.catalog_id
                          ? "bg-amber-500/10 border-amber-500/50 shadow-md"
                          : "bg-slate-950/60 border-slate-800 hover:border-slate-700"
                      }`}
                    >
                      <div className="flex items-center justify-between">
                        <span className="font-bold text-slate-200 truncate pr-2">
                          {item.title}
                        </span>
                        {item.has_audio && (
                          <span className="p-1 rounded bg-red-500/10 text-red-400 shrink-0" title="Audio Attached">
                            <Mic className="w-3 h-3" />
                          </span>
                        )}
                      </div>

                      <div className="text-[11px] text-amber-400 font-sans flex items-center space-x-1">
                        <Zap className="w-3 h-3 shrink-0" />
                        <span className="truncate">{item.resonance_reason}</span>
                      </div>

                      <div className="flex items-center justify-between text-[10px] text-slate-500">
                        <span>{item.sanctuary_label}</span>
                        <span>{item.note_date || item.year}</span>
                      </div>
                    </div>
                  ))
                ) : (
                  <div className="py-12 text-center text-xs font-mono text-slate-500">
                    No resonant items found for current transit.
                  </div>
                )}
              </div>
            ) : activeMode === "catalog" ? (
              catalogItems.length > 0 ? (
                catalogItems.map((item) => (
                  <div
                    key={item.id}
                    onClick={() => handleSelectItem(item)}
                    className={`p-3.5 rounded-xl border cursor-pointer transition text-xs font-mono space-y-1.5 ${
                      selectedItem?.id === item.id
                        ? "bg-cyan-500/10 border-cyan-500/50 shadow-md"
                        : "bg-slate-950/60 border-slate-800 hover:border-slate-700"
                    }`}
                  >
                    <div className="flex items-center justify-between">
                      <span className="font-bold text-slate-200 truncate pr-2">{item.title}</span>
                      {item.has_audio && (
                        <span className="p-1 rounded bg-red-500/10 text-red-400 shrink-0" title="Audio Attached">
                          <Mic className="w-3 h-3" />
                        </span>
                      )}
                    </div>
                    <div className="text-[11px] text-slate-400 font-sans line-clamp-2">
                      {item.snippet}
                    </div>
                    <div className="flex items-center justify-between text-[10px] text-slate-500">
                      <span>{item.sanctuary_label}</span>
                      <span>{item.note_date || item.year}</span>
                    </div>
                  </div>
                ))
              ) : (
                <div className="py-12 text-center text-xs font-mono text-slate-500">
                  No catalog items found.
                </div>
              )
            ) : savedSeeds.length > 0 ? (
              savedSeeds.map((seed) => (
                <div
                  key={seed.id}
                  onClick={() => handleSelectItem(seed)}
                  className={`p-3.5 rounded-xl border cursor-pointer transition text-xs font-mono space-y-1.5 ${
                    selectedItem?.id === seed.id
                      ? "bg-purple-500/10 border-purple-500/50 shadow-md"
                      : "bg-slate-950/60 border-slate-800 hover:border-slate-700"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-slate-200 truncate pr-2">{seed.title}</span>
                    <span className="text-[10px] px-2 py-0.5 rounded bg-purple-500/10 text-purple-300 border border-purple-500/20">
                      SEED
                    </span>
                  </div>
                  <div className="text-[11px] text-slate-400 font-sans line-clamp-2">
                    {seed.narrative_hook}
                  </div>
                  <div className="flex items-center justify-between text-[10px] text-slate-500">
                    <span>{seed.cosmic_alignment.mahadasha} Epoch</span>
                    <span>{seed.note_date}</span>
                  </div>
                </div>
              ))
            ) : (
              <div className="py-12 text-center text-xs font-mono text-slate-500">
                No saved seeds yet. Synthesize and save your first seed.
              </div>
            )}
          </div>

          {/* Right Column: Synthesis Blueprint & Actions */}
          <div className="lg:col-span-7 p-6 overflow-y-auto max-h-[60vh] lg:max-h-full space-y-5 bg-slate-950/40">
            {synthesizing ? (
              <div className="py-24 flex flex-col items-center justify-center text-xs font-mono text-slate-500 space-y-3">
                <RefreshCw className="w-8 h-8 animate-spin text-amber-400" />
                <span>Synthesizing Foundations Story Seed...</span>
                <span className="text-[11px] text-slate-600">Extracting narrative hook &amp; resolving cosmic Dasha era</span>
              </div>
            ) : currentSeed ? (
              <div className="space-y-5">
                {/* Seed Header Badge */}
                <div className="p-4 rounded-2xl bg-gradient-to-r from-slate-900 to-slate-950 border border-slate-800 space-y-3 shadow-lg">
                  <div className="flex flex-wrap items-center justify-between gap-2">
                    <div className="flex items-center space-x-2">
                      <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono bg-amber-500/10 text-amber-300 border border-amber-500/30">
                        {currentSeed.cosmic_alignment.mahadasha} ({currentSeed.cosmic_alignment.mahadasha_sanskrit})
                      </span>
                      <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono bg-cyan-500/10 text-cyan-300 border border-cyan-500/30">
                        {currentSeed.cosmic_alignment.sacred_metal}
                      </span>
                    </div>

                    <span className="text-xs font-mono text-slate-400 flex items-center space-x-1">
                      <Calendar className="w-3.5 h-3.5" />
                      <span>{currentSeed.note_date || currentSeed.year}</span>
                    </span>
                  </div>

                  <h4 className="text-base font-bold text-white font-serif">
                    {currentSeed.title}
                  </h4>

                  <div className="text-xs font-mono text-slate-400 flex flex-wrap gap-x-4 gap-y-1 pt-1 border-t border-slate-800/80 text-[11px]">
                    <span>Protagonist: <strong className="text-slate-200">{currentSeed.protagonist_role}</strong></span>
                    <span>Solfeggio: <strong className="text-cyan-300">{currentSeed.cosmic_alignment.resonant_frequency_hz} Hz</strong></span>
                    <span>Words: <strong className="text-slate-300">{currentSeed.word_count}</strong></span>
                  </div>
                </div>

                {/* Narrative Hook & Dialogue Anchor */}
                <div className="space-y-3">
                  <div className="p-3.5 rounded-xl bg-slate-900/90 border border-slate-800 space-y-1">
                    <div className="text-[10px] font-mono uppercase text-amber-400 font-semibold tracking-wider">
                      Narrative Hook / Opening Thesis
                    </div>
                    <p className="text-xs text-slate-200 font-sans leading-relaxed">
                      {currentSeed.narrative_hook}
                    </p>
                  </div>

                  {currentSeed.dialogue_anchor && (
                    <div className="p-3.5 rounded-xl bg-slate-900/90 border border-slate-800 space-y-1">
                      <div className="text-[10px] font-mono uppercase text-cyan-400 font-semibold tracking-wider">
                        Dialogue Anchor / Voice Quote
                      </div>
                      <p className="text-xs text-cyan-100 font-serif italic leading-relaxed">
                        &ldquo;{currentSeed.dialogue_anchor}&rdquo;
                      </p>
                    </div>
                  )}
                </div>

                {/* 3-Act Scene Progression */}
                <div className="space-y-2">
                  <div className="text-[11px] font-mono uppercase text-slate-400 font-bold flex items-center space-x-1.5">
                    <Layers className="w-3.5 h-3.5 text-amber-400" />
                    <span>3-Act Alchemical Scene Progression</span>
                  </div>

                  <div className="space-y-2">
                    {currentSeed.scene_beats.map((beat) => (
                      <div
                        key={beat.act}
                        className="p-3 rounded-xl bg-slate-900/60 border border-slate-800/80 text-xs font-mono space-y-1"
                      >
                        <div className="flex items-center justify-between text-slate-300 font-bold">
                          <span className="text-amber-300">{beat.title}</span>
                          <span className="text-[10px] font-normal text-slate-500">{beat.emotional_tone}</span>
                        </div>
                        <p className="text-[11px] text-slate-400 font-sans">{beat.premise}</p>
                        <div className="text-[10px] text-slate-500 pt-1 flex items-center space-x-1">
                          <Film className="w-3 h-3 text-slate-400 shrink-0" />
                          <span className="truncate">{beat.prompt_directive}</span>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Visual Directives */}
                <div className="p-3.5 rounded-xl bg-slate-900/70 border border-slate-800 space-y-2 text-xs font-mono">
                  <div className="flex items-center justify-between">
                    <span className="text-slate-400 uppercase text-[10px]">Alchemical Color Palette</span>
                    <div className="flex space-x-1.5">
                      {currentSeed.visual_directives.color_palette.map((color, idx) => (
                        <span
                          key={idx}
                          className="w-5 h-5 rounded-full border border-white/20 shadow-sm"
                          style={{ backgroundColor: color }}
                          title={color}
                        />
                      ))}
                    </div>
                  </div>
                  <div className="text-[11px] text-slate-400">
                    Atmosphere: <span className="text-slate-200">{currentSeed.visual_directives.atmosphere}</span>
                  </div>
                </div>

                {/* Actions Bar */}
                <div className="flex flex-wrap items-center gap-3 pt-2">
                  <button
                    onClick={handleLoadDraft}
                    className="flex-1 py-3 px-4 rounded-xl bg-gradient-to-r from-red-600 to-amber-600 hover:from-red-500 hover:to-amber-500 text-white font-bold text-xs font-mono flex items-center justify-center space-x-2 transition shadow-lg shadow-red-900/30"
                  >
                    <span>Load into Video Production Draft</span>
                    <ArrowRight className="w-4 h-4" />
                  </button>

                  <button
                    onClick={handleSaveSeed}
                    disabled={savingSeed}
                    className="py-3 px-4 rounded-xl bg-slate-800 hover:bg-slate-700 text-slate-200 font-mono text-xs transition flex items-center space-x-1.5 border border-slate-700"
                  >
                    <Bookmark className={`w-4 h-4 ${savingSeed ? "animate-pulse text-purple-400" : ""}`} />
                    <span>{savingSeed ? "Saving..." : "Save Seed"}</span>
                  </button>
                </div>
              </div>
            ) : (
              <div className="py-24 text-center text-xs font-mono text-slate-500">
                Select a document or transcript from the left column to synthesize its Foundations Story Seed.
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
}
