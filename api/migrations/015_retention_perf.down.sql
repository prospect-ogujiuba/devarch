DROP INDEX IF EXISTS idx_container_metrics_recorded_at_brin;
DROP INDEX IF EXISTS idx_images_last_synced_at;
DROP INDEX IF EXISTS idx_vulnerabilities_created_at;

ALTER TABLE container_metrics RESET (
  autovacuum_vacuum_scale_factor,
  autovacuum_analyze_scale_factor,
  autovacuum_vacuum_threshold,
  autovacuum_analyze_threshold
);
