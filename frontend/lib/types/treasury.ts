// AMRA Sovereign Treasury & Studio Type Definitions

export interface FinancialMetrics {
  active_subscribers: number;
  saas_mrr: number;
  youtube_accrued_30d: number;
  total_gross_ecosystem: number;
  ledger_total_transactions: number;
  timestamp: string;
}

export interface AMRAPlan {
  id: string;
  name: string;
  tier: string;
  price_cents: number;
  currency: string;
  interval: string;
  entitlements: string[];
  max_seats: number;
}

export interface LedgerTransaction {
  id: string;
  customer_id: string;
  subscription_id?: string;
  amount_cents: number;
  currency: string;
  provider: string;
  status: string;
  description: string;
  idempotency_key: string;
  created_at: string;
}

export interface AnalyticsReport {
  start_date: string;
  end_date: string;
  total_views: number;
  total_minutes_watched: number;
  average_view_duration_secs: number;
  subscribers_gained: number;
  subscribers_lost: number;
  net_subscribers: number;
  total_comments: number;
  total_likes: number;
  daily_rows: Array<{
    day: string;
    views: number;
    minutes_watched: number;
    average_view_duration_secs: number;
    subscribers_gained: number;
    subscribers_lost: number;
    likes: number;
  }>;
}

export interface FinancialReport {
  start_date: string;
  end_date: string;
  monetized: boolean;
  currency: string;
  total_estimated_revenue: number;
  total_estimated_ad_revenue: number;
  total_estimated_red_revenue: number;
  total_gross_revenue: number;
  average_cpm: number;
  average_playback_based_cpm: number;
  total_monetized_playbacks: number;
  status_message?: string;
  daily_rows: Array<{
    day: string;
    estimated_revenue: number;
    estimated_ad_revenue: number;
    estimated_red_partner_revenue: number;
    gross_revenue: number;
    cpm: number;
    playback_based_cpm: number;
    monetized_playbacks: number;
  }>;
}

export interface ArishadvargaVector {
  key?: string;
  index: number;
  name: string;
  meaning: string;
  shadow_distortion?: string;
  teaching?: string;
  angle_rad: number;
  distance: number;
  coord: { x: number; y: number };
  transmuted_as: string;
}

export interface EsotericPhilosophy {
  key?: string;
  title: string;
  subtitle?: string;
  karma_phala: string;
  purna_kumbha: string;
  amara_amrta?: string;
  summary?: string;
  tags?: string[];
  amra_significance?: string;
}

export interface ArishadvargaDemonDoc {
  index: number;
  key: string;
  name: string;
  shadow_distortion: string;
  transmuted_virtue: string;
  teaching: string;
}

export interface TransmutationStage {
  stage: string;
  sanskrit: string;
  title: string;
  description: string;
}

export interface TransmutationTriadDoc {
  key: string;
  title: string;
  stages: TransmutationStage[];
  churning_law: string;
}

export interface MathematicalFoundationsDoc {
  key: string;
  title: string;
  parametric_x: string;
  parametric_y: string;
  crest_envelope: string;
  pratyahara_spiral: string;
  samudra_manthan_law: string;
  formulas: Record<string, string>;
}

export interface TransmutationTriad {
  samudra_manthan: string;
  solar_fire: string;
  golden_equilibrium: string;
}

export interface TransmutationState {
  state: string;
  description: string;
  shadow_tension: number;
  solar_fire: number;
  devic_ratio: number;
  asuric_ratio: number;
  demons: ArishadvargaVector[];
  golden_equilibrium: boolean;
}

export interface AmraGeometryData {
  body: { svg_path: string; arc_length: number; golden_ratio_phi: number };
  bija?: { svg_path: string };
  leaves?: Array<{ index: number; angle_deg: number; length: number; svg_path: string }>;
  hook?: { svg_path: string; curvature_radius: number };
  stem?: { svg_path: string; tilt_angle_deg: number };
  leaf?: { svg_path: string; veins_count: number };
  purna_kumbha?: { svg_path: string; water_level: number };
  arishadvarga?: ArishadvargaVector[];
  triad?: TransmutationTriad;
  transmutation: TransmutationState;
  mathematical_formulas?: Record<string, string>;
  philosophy?: EsotericPhilosophy;
}

export interface GCloudAccount {
  account: string;
  account_type: string;
  authenticated: boolean;
}

export interface GCloudProject {
  project_id: string;
  project_number: string;
  region: string;
  environment: string;
}

export interface GCloudBillingAccount {
  billing_account_id: string;
  display_name: string;
  open: boolean;
  currency: string;
}

export interface GCloudServiceSpend {
  service_name: string;
  category: string;
  amount_dollars: number;
  percentage: number;
  unit_summary: string;
}

export interface GCloudCostReport {
  month: string;
  accrued_cost_dollars: number;
  budget_limit_dollars: number;
  budget_used_percent: number;
  forecast_cost_dollars: number;
  scale_to_zero_optimized: boolean;
  idle_burn_rate_dollars_per_hr: number;
  services: GCloudServiceSpend[];
  updated_at: string;
}

export interface SovereignMargin {
  gross_revenue_dollars: number;
  gcloud_cost_dollars: number;
  net_sovereign_yield_dollars: number;
  profit_margin_percent: number;
  status: string;
}

export interface GCloudBillingResponse {
  status: {
    account: GCloudAccount;
    project: GCloudProject;
    billing_account: GCloudBillingAccount;
    timestamp: string;
  };
  cost_report: GCloudCostReport;
  sovereign_margin: SovereignMargin;
  ledger_sync_key: string;
  ledger_synced: boolean;
}
