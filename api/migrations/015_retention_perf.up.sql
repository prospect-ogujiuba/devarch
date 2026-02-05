-- Keep high-churn metrics table lean and fast.

CREATE INDEX IF NOT EXISTS idx_container_metrics_recorded_at_brin
  ON container_metrics USING BRIN (recorded_at);

CREATE INDEX IF NOT EXISTS idx_images_last_synced_at
  ON images(last_synced_at);

CREATE INDEX IF NOT EXISTS idx_vulnerabilities_created_at
  ON vulnerabilities(created_at);

ALTER TABLE container_metrics SET (
  autovacuum_vacuum_scale_factor = 0.02,
  autovacuum_analyze_scale_factor = 0.01,
  autovacuum_vacuum_threshold = 5000,
  autovacuum_analyze_threshold = 5000
);
