export interface StatusCacheStats {
  total_entries: number;
  pending_webhooks: number;
  suppressed_count: number;
  flush_count: number;
}

export interface StatusCacheMetrics {
  cache_hit_rate: number;
  webhook_reduction_rate: number;
  avg_flush_duration_ms: number;
}
