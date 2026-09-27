// Dasha Observatory & Astronomical Type Definitions

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

export interface DashaPeriod {
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

export interface StoryContext {
  active_archetype: string;
  thematic_phase: string;
  alchemical_element: string;
  chakra_focus: string;
  resonant_frequency_hz: number;
  hermetic_principles: string[];
  narrative_prompt: string;
}

export interface ActiveSnapshot {
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

export interface DashaProfile {
  id: string;
  name: string;
  birth_date?: string;
  birth_time?: string;
  timezone_offset?: number;
  mode?: "ephemeris" | "nakshatra";
  nakshatra_index?: number;
  pada_number?: number;
  birth_time_utc: string;
  moon_degree: number;
  janma_nakshatra: Nakshatra;
  janma_pada: number;
  balance_years: number;
  starting_lord: string;
  active_snapshot: ActiveSnapshot;
  timeline: DashaPeriod[];
}

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
  color_hex: string;
  sacred_metal: string;
  metal_symbol: string;
  magnum_opus_stage: string;
  story_archetype: string;
  element: string;
  chakra_center: string;
  active: boolean;
  sub_horas?: SubHora[];
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

export interface PulseData {
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
