-- Note: PostgreSQL cannot remove a value from an enum type (no ALTER TYPE
-- ... DROP VALUE). The WARNING / NOT_APPLICABLE / BEST_BEFORE / USE_BY_DATE
-- enum additions from the .up migration are therefore NOT reversed here —
-- reversing them would require recreating each enum type and every column/
-- function that depends on it. This is a standard, accepted limitation of
-- additive enum migrations; down-migrating past this file in an environment
-- that has already used these values will require a manual data migration.

ALTER TABLE inspections DROP CONSTRAINT IF EXISTS chk_inspections_latitude;
ALTER TABLE inspections DROP CONSTRAINT IF EXISTS chk_inspections_longitude;

ALTER TABLE inspections
  DROP COLUMN IF EXISTS location,
  DROP COLUMN IF EXISTS latitude,
  DROP COLUMN IF EXISTS longitude,
  DROP COLUMN IF EXISTS is_finalized,
  DROP COLUMN IF EXISTS finalized_at,
  DROP COLUMN IF EXISTS finalized_by_id;
