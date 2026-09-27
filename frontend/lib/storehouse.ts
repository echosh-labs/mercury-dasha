"use client";

// High-performance lazy hydration service for Mercury sovereign domain catalogs.
// Eliminates heavy blocking fetches on initial page mount and provides in-memory deduplication.

export interface Pada {
  number: number;
  degrees_start: string;
  degrees_end: string;
  navamsha_sign: string;
}

export interface Nakshatra {
  index: number;
  id: string;
  name: string;
  sanskrit_name: string;
  degree_start: number;
  degree_end: number;
  zodiac_span: string;
  ruling_planet: string;
  frequency_hz: number;
  deity: string;
  symbol: string;
  quality: string;
  padas: Pada[];
}

export interface SacredMetal {
  id: string;
  name: string;
  symbol: string;
  governing_planet: string;
  governing_graha: string;
  alchemical_essence: string;
  transmutation_role: string;
  chakra_center: string;
  color_hex: string;
  conductivity_rating: string;
  quicksilver_affinity: string;
}

export interface HermeticAxiom {
  number: number;
  title: string;
  canonical_text: string;
  governing_planet: string;
  alchemical_stage: string;
  esoteric_meaning: string;
  foundations_theme: string;
}

export interface MagnumOpusStage {
  index: number;
  name: string;
  latin_name: string;
  element: string;
  planet: string;
  description: string;
}

export interface LiveAlignment {
  active_planet: string;
  active_metal: SacredMetal;
  governing_axiom: HermeticAxiom;
  magnum_opus_stage: MagnumOpusStage;
  alchemical_motto: string;
}

export interface DashaPeriodSummary {
  level: string;
  planet: string;
  planet_name: string;
  sanskrit_name: string;
  color_hex: string;
  duration_days: number;
  start_date: string;
  end_date: string;
  sacred_metal?: string;
  metal_symbol?: string;
  age_range?: string;
}

export interface VedicPanchanga {
  vara: string;
  vara_lord: string;
  tithi_number: number;
  tithi_name: string;
  paksha: string;
  tithi_percent: number;
  yoga_number: number;
  yoga_name: string;
  yoga_meaning: string;
  karana_number: number;
  karana_name: string;
  karana_type: string;
  sidereal_sun_degree: number;
}

export interface AyurvedicConstitution {
  dosha: string;
  dosha_qualities: string;
  gana: string;
  yoni_totem: string;
  yoni_animal: string;
  nadi: string;
}

export interface NatalLagna {
  degree: number;
  rashi: string;
  rashi_sanskrit: string;
  rashi_index: number;
  degree_in_sign: number;
  lord: string;
  lord_name: string;
  nakshatra: Nakshatra;
  pada: number;
  symbol: string;
  element: string;
}

export interface LuminaryPlacement {
  degree: number;
  rashi: string;
  rashi_index: number;
  degree_in_sign: number;
  rashi_lord: string;
  house_number: number;
  house_name: string;
  nakshatra: Nakshatra;
  pada: number;
}

export interface BhavaDetail {
  house_number: number;
  name: string;
  sanskrit_name: string;
  domain: string;
  rashi: string;
  rashi_lord: string;
  luminaries: string[];
}

export interface GrahaDignity {
  level: string;
  name: string;
  description: string;
  color_hex: string;
}

export interface GrahaPlacement {
  id: string;
  name: string;
  sanskrit_name: string;
  degree: number;
  rashi: string;
  rashi_index: number;
  degree_in_sign: number;
  house_number: number;
  house_name: string;
  nakshatra: Nakshatra;
  pada: number;
  dignity: GrahaDignity;
  is_retrograde: boolean;
}

export interface TripodOfEmbodiment {
  lagna: NatalLagna;
  surya: LuminaryPlacement;
  chandra: LuminaryPlacement;
  bhavas: BhavaDetail[];
}

export interface NatalAstrology {
  sidereal_moon_degree: number;
  sidereal_sun_degree?: number;
  janma_nakshatra: Nakshatra;
  janma_pada: number;
  pada_detail: Pada;
  starting_lord: string;
  starting_lord_name: string;
  balance_years: number;
  elemental_tattva: string;
  panchanga?: VedicPanchanga;
  ayurveda?: AyurvedicConstitution;
  lagna?: NatalLagna;
  tripod?: TripodOfEmbodiment;
  grahas?: GrahaPlacement[];
}

export type CharacterProfile = SovereignProfile;

export interface NatalAlchemy {
  sacred_metal: SacredMetal;
  governing_axiom: HermeticAxiom;
  magnum_opus_stage: MagnumOpusStage;
  chakra_anchor: string;
  resonant_frequency_hz: number;
  quicksilver_affinity: string;
  alchemical_motto: string;
}

export interface ActiveAlchemicalState {
  mahadasha_metal: SacredMetal;
  mahadasha_axiom: HermeticAxiom;
  antardasha_metal: SacredMetal;
  antardasha_axiom: HermeticAxiom;
  transmutation_vessel: string;
}

export interface SymbioticResonance {
  transit_hora_planet: string;
  transit_hora_planet_name: string;
  transit_metal: SacredMetal;
  transit_axiom: HermeticAxiom;
  is_janma_resonant: boolean;
  is_mahadasha_resonant: boolean;
  is_antardasha_resonant: boolean;
  resonance_tier: string;
  transmutation_guidance: string;
}

export interface ProfileSummary {
  id: string;
  name: string;
  birth_date: string;
  location_name: string;
  nakshatra_name: string;
  nakshatra_index: number;
  pada_number: number;
  starting_lord: string;
  natal_metal: string;
  active_mahadasha: string;
  dosha?: string;
  gana?: string;
  tithi?: string;
  lagna_rashi?: string;
  sun_rashi?: string;
  updated_at: string;
}

export interface SovereignProfile {
  id: string;
  name: string;
  title?: string;
  backstory?: string;
  avatar_icon?: string;
  chronicles?: CharacterChronicleRef[];
  birth_date?: string;
  birth_time?: string;
  timezone?: string;
  timezone_offset?: number;
  latitude?: number;
  longitude?: number;
  location_name?: string;
  mode?: "ephemeris" | "nakshatra";
  nakshatra_index?: number;
  pada_number?: number;
  birth_time_utc: string;
  moon_degree: number;
  janma_nakshatra: Nakshatra;
  janma_pada: number;
  balance_years: number;
  starting_lord: string;
  active_snapshot: any;
  timeline?: any[];
  astrology: NatalAstrology;
  alchemy: NatalAlchemy;
  active_alchemy: ActiveAlchemicalState;
  timeline_summary: DashaPeriodSummary[];
  symbiotic_resonance?: SymbioticResonance;
  created_at?: string;
  updated_at?: string;
}

export interface AlchemyData {
  philosophy: string;
  sovereign_element: string;
  live_alignment: LiveAlignment;
  sacred_metals: SacredMetal[];
  hermetic_axioms: HermeticAxiom[];
  magnum_opus_stages: MagnumOpusStage[];
  natal_alchemy?: NatalAlchemy;
  active_alchemy?: ActiveAlchemicalState;
  symbiotic_resonance?: SymbioticResonance;
  active_profile_id?: string;
  active_profile_name?: string;
}

// In-memory single-flight promise cache
let nakshatrasPromise: Promise<Nakshatra[]> | null = null;
let cachedNakshatras: Nakshatra[] | null = null;

let alchemyPromise: Promise<AlchemyData | null> | null = null;
let cachedAlchemy: AlchemyData | null = null;
let lastAlchemyPlanet: string | null = null;

let activeProfilePromise: Promise<SovereignProfile | null> | null = null;
let cachedActiveProfile: SovereignProfile | null = null;

/**
 * Lazily hydrates the 27 Nakshatras and 108 Padas from BoltDB storehouse.
 * Single-flight deduplication guarantees at most one network request per session.
 */
export async function fetchNakshatrasLazy(): Promise<Nakshatra[]> {
  if (cachedNakshatras && cachedNakshatras.length > 0) {
    return cachedNakshatras;
  }

  if (nakshatrasPromise) {
    return nakshatrasPromise;
  }

  nakshatrasPromise = (async () => {
    try {
      const res = await fetch("/api/dasha/nakshatras");
      if (!res.ok) {
        throw new Error(`Failed to load nakshatras: ${res.statusText}`);
      }
      const data = await res.json();
      const list: Nakshatra[] = data.nakshatras || [];
      cachedNakshatras = list;
      return list;
    } catch (err) {
      console.error("[Storehouse] Nakshatras lazy hydration error:", err);
      nakshatrasPromise = null;
      return [];
    }
  })();

  return nakshatrasPromise;
}

/**
 * Lazily hydrates the Alchemical Laboratory catalog from BoltDB storehouse.
 * If forceRefresh is true, fetches fresh live_alignment for the active hora ruler.
 */
export async function fetchAlchemyLazy(forceRefresh = false): Promise<AlchemyData | null> {
  if (!forceRefresh && cachedAlchemy) {
    return cachedAlchemy;
  }

  if (!forceRefresh && alchemyPromise) {
    return alchemyPromise;
  }

  alchemyPromise = (async () => {
    try {
      let url = "/api/dasha/alchemy";
      if (typeof window !== "undefined") {
        const lat = localStorage.getItem("mercury_geo_lat");
        const lon = localStorage.getItem("mercury_geo_lon");
        if (lat && lon) {
          url += `?lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`;
        }
      }

      const res = await fetch(url);
      if (!res.ok) {
        throw new Error(`Failed to load alchemy catalog: ${res.statusText}`);
      }
      const data: AlchemyData = await res.json();
      cachedAlchemy = data;
      if (data.live_alignment) {
        lastAlchemyPlanet = data.live_alignment.active_planet;
      }
      return data;
    } catch (err) {
      console.error("[Storehouse] Alchemy lazy hydration error:", err);
      alchemyPromise = null;
      return null;
    }
  })();

  return alchemyPromise;
}

/**
 * Invalidates and flushes the in-memory storehouse cache.
 */
export function invalidateStorehouseCache(): void {
  cachedNakshatras = null;
  nakshatrasPromise = null;
  cachedAlchemy = null;
  alchemyPromise = null;
  lastAlchemyPlanet = null;
  cachedActiveProfile = null;
  activeProfilePromise = null;
}

/**
 * Fetches the active profile (defaults to sovereign-genesis or localStorage selection)
 * with unified astrology and alchemy matrices and live symbiotic resonance.
 */
export async function fetchActiveProfile(
  id?: string,
  forceRefresh = false
): Promise<SovereignProfile | null> {
  const targetId = id || (typeof window !== "undefined" ? localStorage.getItem("mercury_active_profile") : null) || "sovereign-genesis";

  if (!forceRefresh && cachedActiveProfile && cachedActiveProfile.id === targetId) {
    return cachedActiveProfile;
  }

  if (!forceRefresh && activeProfilePromise) {
    return activeProfilePromise;
  }

  activeProfilePromise = (async () => {
    try {
      let url = `/api/v1/profiles/${encodeURIComponent(targetId)}`;
      if (typeof window !== "undefined") {
        const lat = localStorage.getItem("mercury_geo_lat");
        const lon = localStorage.getItem("mercury_geo_lon");
        if (lat && lon) {
          url += `?lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`;
        }
      }

      const res = await fetch(url);
      if (!res.ok) {
        throw new Error(`Failed to load profile ${targetId}: ${res.statusText}`);
      }
      const data: SovereignProfile = await res.json();
      cachedActiveProfile = data;
      return data;
    } catch (err) {
      console.error("[Storehouse] Active profile hydration error:", err);
      activeProfilePromise = null;
      return null;
    }
  })();

  return activeProfilePromise;
}

/**
 * Fetches lightweight profile summaries for fast UI selection and menus.
 */
export async function fetchProfileSummaries(): Promise<ProfileSummary[]> {
  try {
    const res = await fetch("/api/v1/profiles");
    if (!res.ok) {
      throw new Error(`Failed to list profiles: ${res.statusText}`);
    }
    const data = await res.json();
    return data.summaries || [];
  } catch (err) {
    console.error("[Storehouse] Profile summaries fetch error:", err);
    return [];
  }
}

/**
 * Fetches real-time symbiotic resonance between a profile and the live transit hora.
 */
export async function fetchProfileResonance(
  id: string
): Promise<SymbioticResonance | null> {
  try {
    let url = `/api/v1/profiles/${encodeURIComponent(id)}/resonance`;
    if (typeof window !== "undefined") {
      const lat = localStorage.getItem("mercury_geo_lat");
      const lon = localStorage.getItem("mercury_geo_lon");
      if (lat && lon) {
        url += `?lat=${encodeURIComponent(lat)}&lon=${encodeURIComponent(lon)}`;
      }
    }
    const res = await fetch(url);
    if (!res.ok) {
      throw new Error(`Failed to fetch resonance: ${res.statusText}`);
    }
    return await res.json();
  } catch (err) {
    console.error("[Storehouse] Profile resonance fetch error:", err);
    return null;
  }
}


export interface CharacterChronicleRef {
  id: string;
  title: string;
  dasha_level: string;
  planet: string;
  duration_sec: number;
  orientation: string;
  created_at: string;
}

export interface CanvasConfig {
  orientation: "16:9" | "9:16" | "1:1";
  width: number;
  height: number;
  fps: number;
  background_color: string;
}

export interface AudioDucking {
  enabled: boolean;
  attenuate_to: number;
  attack_ms: number;
  release_ms: number;
  sidechain_from: string;
}

export interface AudioTrack {
  id: string;
  role: "voice" | "ambient" | "sfx";
  source_path: string;
  start_sec: number;
  end_sec?: number;
  volume: number;
  fade_in_sec?: number;
  fade_out_sec?: number;
  loop?: boolean;
  ducking?: AudioDucking;
}

export interface VisualScene {
  id: string;
  start_sec: number;
  duration_sec: number;
  source_type: string;
  source_path: string;
  motion: string;
  motion_scale?: number;
  transition_in: string;
  transition_sec?: number;
  caption?: string;
  subtitle?: string;
}

export interface GlobalOverlay {
  id: string;
  type: "waveform" | "chrono_badge" | "lower_third" | "watermark";
  position: string;
  start_sec: number;
  end_sec?: number;
  config?: Record<string, any>;
}

export interface ChronoBinding {
  vimshottari_lord: string;
  planetary_hora: string;
  sacred_metal: string;
  hermetic_axiom: string;
  magnum_opus_stage: string;
}

export interface TimelineManifest {
  id: string;
  title: string;
  description?: string;
  tags?: string[];
  canvas: CanvasConfig;
  duration_sec: number;
  audio_tracks: AudioTrack[];
  scenes: VisualScene[];
  overlays?: GlobalOverlay[];
  chrono_meta?: ChronoBinding;
  created_at: string;
  updated_at: string;
}

export interface ChronicleRequest {
  dasha_planet?: string;
  orientation?: "16:9" | "9:16" | "1:1";
  duration_sec?: number;
  voice_audio_path?: string;
  ambient_music_path?: string;
}

export interface DashaPeriod {
  level: string;
  planet: string;
  planet_name: string;
  sanskrit_name: string;
  color_hex: string;
  duration_days: number;
  start_date: string;
  end_date: string;
  sacred_metal?: string;
  metal_symbol?: string;
  hermetic_axiom?: string;
  story_archetype?: string;
  frequency_hz?: number;
  age_start?: number;
  age_end?: number;
  sub_periods?: DashaPeriod[];
}


export async function fetchCharacters(): Promise<ProfileSummary[]> {
  try {
    const res = await fetch("/api/v1/characters");
    if (!res.ok) return [];
    const data = await res.json();
    return data.summaries || [];
  } catch {
    return [];
  }
}

export async function fetchCharacter(id: string): Promise<CharacterProfile | null> {
  try {
    const res = await fetch(`/api/v1/characters/${encodeURIComponent(id)}`);
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

export async function fetchCharacterTimeline(id: string): Promise<DashaPeriod[]> {
  try {
    const res = await fetch(`/api/v1/characters/${encodeURIComponent(id)}/timeline`);
    if (!res.ok) return [];
    const data = await res.json();
    return data.timeline || [];
  } catch {
    return [];
  }
}

export async function generateCharacterChronicle(
  charId: string,
  req: ChronicleRequest
): Promise<{ status: string; chronicle: CharacterChronicleRef; manifest: TimelineManifest } | null> {
  try {
    const res = await fetch(`/api/v1/characters/${encodeURIComponent(charId)}/chronicle`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(req),
    });
    if (!res.ok) return null;
    return res.json();
  } catch {
    return null;
  }
}

// Re-export BoltDB esoteric storehouse client
export * from "./esoteric";
