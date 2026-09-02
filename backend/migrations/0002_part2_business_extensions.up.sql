-- =============================================================================
-- Part 2 business-layer extensions (additive; see backend/README.md).
--
-- - inspections gain location/lat/long (explicit field-of-service.md
--   inspection metadata) and a finalize/reopen workflow (is_finalized,
--   finalized_at, finalized_by_id) so a completed inspection can't be
--   silently mutated by a later PATCH.
-- - compliance_status gains WARNING (a failed but non-mandatory requirement
--   — distinct from FAIL, which only fires for mandatory requirements).
-- - inspection_outcome gains NOT_APPLICABLE (no rules apply to the
--   inspection's category, so there is nothing to comply or fail with).
-- - declaration_type gains BEST_BEFORE and USE_BY_DATE as distinct concepts
--   from a generic EXPIRY_DATE.
-- =============================================================================

ALTER TYPE compliance_status ADD VALUE IF NOT EXISTS 'WARNING';
ALTER TYPE inspection_outcome ADD VALUE IF NOT EXISTS 'NOT_APPLICABLE';
ALTER TYPE declaration_type ADD VALUE IF NOT EXISTS 'BEST_BEFORE';
ALTER TYPE declaration_type ADD VALUE IF NOT EXISTS 'USE_BY_DATE';

ALTER TABLE inspections
  ADD COLUMN IF NOT EXISTS location TEXT,
  ADD COLUMN IF NOT EXISTS latitude DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS longitude DOUBLE PRECISION,
  ADD COLUMN IF NOT EXISTS is_finalized BOOLEAN NOT NULL DEFAULT FALSE,
  ADD COLUMN IF NOT EXISTS finalized_at TIMESTAMPTZ,
  ADD COLUMN IF NOT EXISTS finalized_by_id UUID REFERENCES users(id) ON DELETE SET NULL;

ALTER TABLE inspections
  ADD CONSTRAINT chk_inspections_latitude CHECK (latitude IS NULL OR (latitude >= -90 AND latitude <= 90)),
  ADD CONSTRAINT chk_inspections_longitude CHECK (longitude IS NULL OR (longitude >= -180 AND longitude <= 180));
