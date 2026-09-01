-- =============================================================================
-- Legal Metrology Packaged Commodities Compliance Checker
-- PostgreSQL schema (source of truth)
--
-- Domain: Department of Consumer Affairs / Legal Metrology.
-- Compliance under the Legal Metrology Act, 2009 and the Legal Metrology
-- (Packaged Commodities) Rules, 2011.
--
-- PostgreSQL holds authoritative, normalized, relational data. Raw OCR / CV / ML
-- output lives in MongoDB; binary files live in S3/MinIO. This schema stores only
-- storage keys and Mongo document references, never binary blobs.
--
-- Ported 1:1 from the original Prisma schema (prisma/schema.prisma) as part of
-- the Node -> Go migration. See backend/README.md for the migration mapping.
-- =============================================================================

CREATE EXTENSION IF NOT EXISTS pgcrypto;

-- ============================ ENUMS ==========================================

CREATE TYPE inspection_status AS ENUM ('DRAFT', 'QUEUED', 'PROCESSING', 'AWAITING_REVIEW', 'COMPLETED', 'FAILED');
CREATE TYPE inspection_outcome AS ENUM ('PENDING', 'COMPLIANT', 'NON_COMPLIANT', 'MANUAL_REVIEW');
CREATE TYPE media_type AS ENUM ('IMAGE', 'DOCUMENT', 'OTHER');
CREATE TYPE image_kind AS ENUM ('ORIGINAL', 'PROCESSED', 'CROPPED_PANEL', 'ANNOTATED');
CREATE TYPE image_processing_status AS ENUM ('UPLOADED', 'QUEUED', 'PROCESSING', 'PROCESSED', 'FAILED');
CREATE TYPE declaration_type AS ENUM (
  'MRP', 'NET_QUANTITY', 'MANUFACTURER', 'PACKER', 'IMPORTER', 'MANUFACTURE_DATE',
  'PACKING_DATE', 'EXPIRY_DATE', 'CONSUMER_CARE', 'COUNTRY_OF_ORIGIN', 'DIMENSIONS',
  'UNIT_SALE_PRICE', 'GENERIC_NAME', 'OTHER'
);
CREATE TYPE declaration_source AS ENUM ('OCR', 'ML_EXTRACTION', 'MANUAL');
CREATE TYPE compliance_status AS ENUM ('PASS', 'FAIL', 'MANUAL_REVIEW', 'NOT_APPLICABLE');
CREATE TYPE requirement_operator AS ENUM (
  'EXISTS', 'NOT_EMPTY', 'REGEX_MATCH', 'NUMERIC_MIN', 'NUMERIC_MAX', 'NUMERIC_RANGE',
  'MIN_TEXT_HEIGHT_MM', 'ENUM_IN', 'DATE_VALID', 'DATE_NOT_FUTURE', 'REQUIRED_IF_IMPORTED', 'CUSTOM'
);
CREATE TYPE violation_severity AS ENUM ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL');
CREATE TYPE violation_status AS ENUM ('OPEN', 'CONFIRMED', 'DISMISSED', 'RESOLVED');
CREATE TYPE ml_job_type AS ENUM ('PREPROCESS', 'OCR', 'LABEL_DETECTION', 'FIELD_EXTRACTION', 'ANALYZE');
CREATE TYPE ml_job_status AS ENUM ('PENDING', 'QUEUED', 'RUNNING', 'SUCCEEDED', 'FAILED');
CREATE TYPE evidence_type AS ENUM ('ORIGINAL_IMAGE', 'CROPPED_PANEL', 'ANNOTATED_IMAGE', 'DOCUMENT', 'OTHER');
CREATE TYPE report_type AS ENUM ('PDF', 'DOCX');
CREATE TYPE report_status AS ENUM ('PENDING', 'GENERATING', 'READY', 'FAILED');
CREATE TYPE dataset_split AS ENUM ('UNASSIGNED', 'TRAIN', 'VALIDATION', 'TEST');
CREATE TYPE annotation_status AS ENUM ('PENDING', 'IN_PROGRESS', 'ANNOTATED', 'REVIEWED', 'REJECTED');
CREATE TYPE audit_action AS ENUM ('CREATE', 'UPDATE', 'DELETE', 'LOGIN', 'LOGOUT', 'PROCESS', 'REVIEW', 'FINALIZE', 'EXPORT');

-- ============================ IDENTITY & RBAC =================================

CREATE TABLE users (
  id             UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email          TEXT NOT NULL UNIQUE,
  password_hash  TEXT NOT NULL,
  full_name      TEXT NOT NULL,
  phone          TEXT,
  is_active      BOOLEAN NOT NULL DEFAULT TRUE,
  last_login_at  TIMESTAMPTZ,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at     TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_users_is_active ON users (is_active);

CREATE TABLE roles (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name         TEXT NOT NULL UNIQUE,
  description  TEXT,
  is_system    BOOLEAN NOT NULL DEFAULT FALSE,
  permissions  TEXT[] NOT NULL DEFAULT '{}',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE user_roles (
  user_id        UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id        UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  assigned_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (user_id, role_id)
);
CREATE INDEX idx_user_roles_role_id ON user_roles (role_id);

CREATE TABLE refresh_tokens (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash      TEXT NOT NULL UNIQUE,
  jti             TEXT NOT NULL UNIQUE,
  user_agent      TEXT,
  ip              TEXT,
  expires_at      TIMESTAMPTZ NOT NULL,
  revoked_at      TIMESTAMPTZ,
  replaced_by_id  UUID,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);

-- ============================ CATALOG ========================================

CREATE TABLE product_categories (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT NOT NULL UNIQUE,
  code        TEXT NOT NULL UNIQUE,
  description TEXT,
  parent_id   UUID REFERENCES product_categories(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_product_categories_parent_id ON product_categories (parent_id);

CREATE TABLE manufacturers (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name       TEXT NOT NULL,
  address    TEXT,
  country    TEXT,
  license_no TEXT,
  contact    TEXT,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_manufacturers_name ON manufacturers (name);

CREATE TABLE products (
  id                    UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name                  TEXT NOT NULL,
  sku                   TEXT UNIQUE,
  brand                 TEXT,
  barcode               TEXT,
  category_id           UUID NOT NULL REFERENCES product_categories(id) ON DELETE RESTRICT,
  manufacturer_id       UUID REFERENCES manufacturers(id) ON DELETE SET NULL,
  net_quantity_declared TEXT,
  mrp_declared          NUMERIC(12, 2),
  attributes            JSONB,
  created_by_id         UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at            TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_products_category_id ON products (category_id);
CREATE INDEX idx_products_manufacturer_id ON products (manufacturer_id);

CREATE TABLE product_media (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  product_id  UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
  storage_key TEXT NOT NULL,
  mime_type   TEXT NOT NULL,
  kind        media_type NOT NULL DEFAULT 'IMAGE',
  size_bytes  INTEGER,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_product_media_product_id ON product_media (product_id);

-- ============================ INSPECTIONS ====================================

CREATE TABLE inspections (
  id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code         TEXT NOT NULL UNIQUE,
  product_id   UUID REFERENCES products(id) ON DELETE SET NULL,
  category_id  UUID REFERENCES product_categories(id) ON DELETE SET NULL,
  inspector_id UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  reviewer_id  UUID REFERENCES users(id) ON DELETE SET NULL,
  status       inspection_status NOT NULL DEFAULT 'DRAFT',
  outcome      inspection_outcome NOT NULL DEFAULT 'PENDING',
  source       TEXT,
  notes        TEXT,
  metadata     JSONB,
  started_at   TIMESTAMPTZ,
  completed_at TIMESTAMPTZ,
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_inspections_status ON inspections (status);
CREATE INDEX idx_inspections_outcome ON inspections (outcome);
CREATE INDEX idx_inspections_inspector_id ON inspections (inspector_id);
CREATE INDEX idx_inspections_product_id ON inspections (product_id);

CREATE TABLE inspection_images (
  id                 UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id      UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  storage_key        TEXT NOT NULL UNIQUE,
  original_filename  TEXT,
  mime_type          TEXT NOT NULL,
  size_bytes         INTEGER,
  width              INTEGER,
  height             INTEGER,
  kind               image_kind NOT NULL DEFAULT 'ORIGINAL',
  status             image_processing_status NOT NULL DEFAULT 'UPLOADED',
  checksum           TEXT,
  mongo_ocr_doc_id   TEXT,
  scale_reference    JSONB,
  created_at         TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at         TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_inspection_images_inspection_id ON inspection_images (inspection_id);
CREATE INDEX idx_inspection_images_status ON inspection_images (status);

CREATE TABLE extracted_declarations (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id        UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  source_image_id      UUID REFERENCES inspection_images(id) ON DELETE SET NULL,
  type                 declaration_type NOT NULL,
  normalized_value     TEXT,
  raw_value            TEXT,
  unit                 TEXT,
  numeric_value        NUMERIC(18, 4),
  confidence           DOUBLE PRECISION,
  bbox                 JSONB,
  source               declaration_source NOT NULL DEFAULT 'OCR',
  extraction_metadata  JSONB,
  mongo_ref_id         TEXT,
  is_edited_by_reviewer BOOLEAN NOT NULL DEFAULT FALSE,
  edited_by_id         UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_extracted_declarations_inspection_type ON extracted_declarations (inspection_id, type);

-- ============================ LEGAL RULES ====================================

CREATE TABLE legal_rules (
  id               UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code             TEXT NOT NULL UNIQUE,
  title            TEXT NOT NULL,
  description      TEXT,
  act_reference    TEXT,
  section          TEXT,
  version          INTEGER NOT NULL DEFAULT 1,
  effective_from   TIMESTAMPTZ,
  effective_to     TIMESTAMPTZ,
  is_active        BOOLEAN NOT NULL DEFAULT TRUE,
  severity_default violation_severity NOT NULL DEFAULT 'MEDIUM',
  params           JSONB,
  created_by_id    UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at       TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_legal_rules_is_active ON legal_rules (is_active);

CREATE TABLE rule_requirements (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  rule_id           UUID NOT NULL REFERENCES legal_rules(id) ON DELETE CASCADE,
  key               TEXT NOT NULL,
  declaration_type  declaration_type,
  operator          requirement_operator NOT NULL,
  expected          JSONB,
  unit              TEXT,
  severity          violation_severity,
  message           TEXT,
  is_mandatory      BOOLEAN NOT NULL DEFAULT TRUE,
  order_index       INTEGER NOT NULL DEFAULT 0,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rule_requirements_rule_id ON rule_requirements (rule_id);

CREATE TABLE rule_categories (
  rule_id     UUID NOT NULL REFERENCES legal_rules(id) ON DELETE CASCADE,
  category_id UUID NOT NULL REFERENCES product_categories(id) ON DELETE CASCADE,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (rule_id, category_id)
);
CREATE INDEX idx_rule_categories_category_id ON rule_categories (category_id);

-- ============================ COMPLIANCE =====================================

CREATE TABLE compliance_checks (
  id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id          UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  rule_id                UUID NOT NULL REFERENCES legal_rules(id) ON DELETE RESTRICT,
  requirement_id         UUID REFERENCES rule_requirements(id) ON DELETE SET NULL,
  declaration_id         UUID REFERENCES extracted_declarations(id) ON DELETE SET NULL,
  status                 compliance_status NOT NULL,
  actual_value           TEXT,
  expected_value         TEXT,
  reason                 TEXT,
  confidence             DOUBLE PRECISION,
  requires_manual_review BOOLEAN NOT NULL DEFAULT FALSE,
  details                JSONB,
  evaluated_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  created_at             TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_compliance_checks_inspection_id ON compliance_checks (inspection_id);
CREATE INDEX idx_compliance_checks_rule_id ON compliance_checks (rule_id);
CREATE INDEX idx_compliance_checks_status ON compliance_checks (status);

CREATE TABLE violations (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id   UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  rule_id         UUID REFERENCES legal_rules(id) ON DELETE SET NULL,
  check_id        UUID UNIQUE REFERENCES compliance_checks(id) ON DELETE SET NULL,
  severity        violation_severity NOT NULL DEFAULT 'MEDIUM',
  status          violation_status NOT NULL DEFAULT 'OPEN',
  title           TEXT NOT NULL,
  description     TEXT,
  legal_reference TEXT,
  reason          TEXT,
  resolved_by_id  UUID REFERENCES users(id) ON DELETE SET NULL,
  resolved_at     TIMESTAMPTZ,
  review_notes    TEXT,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_violations_inspection_status ON violations (inspection_id, status);
CREATE INDEX idx_violations_severity ON violations (severity);

CREATE TABLE evidence (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  violation_id  UUID REFERENCES violations(id) ON DELETE SET NULL,
  image_id      UUID REFERENCES inspection_images(id) ON DELETE SET NULL,
  type          evidence_type NOT NULL DEFAULT 'ORIGINAL_IMAGE',
  storage_key   TEXT,
  bbox          JSONB,
  caption       TEXT,
  metadata      JSONB,
  created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_evidence_inspection_id ON evidence (inspection_id);
CREATE INDEX idx_evidence_violation_id ON evidence (violation_id);

-- ============================ ML JOB STATE ===================================

CREATE TABLE ml_jobs (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id   UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  image_id        UUID REFERENCES inspection_images(id) ON DELETE SET NULL,
  type            ml_job_type NOT NULL,
  status          ml_job_status NOT NULL DEFAULT 'PENDING',
  queue_name      TEXT,
  bull_job_id     TEXT,
  attempts        INTEGER NOT NULL DEFAULT 0,
  request_payload JSONB,
  result_mongo_id TEXT,
  error           TEXT,
  started_at      TIMESTAMPTZ,
  finished_at     TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_ml_jobs_inspection_id ON ml_jobs (inspection_id);
CREATE INDEX idx_ml_jobs_status ON ml_jobs (status);
CREATE INDEX idx_ml_jobs_bull_job_id ON ml_jobs (bull_job_id);

-- ============================ REPORTS ========================================

CREATE TABLE reports (
  id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  inspection_id   UUID NOT NULL REFERENCES inspections(id) ON DELETE CASCADE,
  type            report_type NOT NULL DEFAULT 'PDF',
  status          report_status NOT NULL DEFAULT 'PENDING',
  storage_key     TEXT,
  file_name       TEXT,
  size_bytes      INTEGER,
  checksum        TEXT,
  generated_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
  error           TEXT,
  requested_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  generated_at    TIMESTAMPTZ,
  created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_reports_inspection_id ON reports (inspection_id);
CREATE INDEX idx_reports_status ON reports (status);

-- ============================ DATASETS =======================================

CREATE TABLE datasets (
  id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name          TEXT NOT NULL UNIQUE,
  slug          TEXT NOT NULL UNIQUE,
  description   TEXT,
  purpose       TEXT,
  created_by_id UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE dataset_versions (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  dataset_id UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
  version    TEXT NOT NULL,
  notes      TEXT,
  is_frozen  BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  UNIQUE (dataset_id, version)
);

CREATE TABLE dataset_items (
  id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  dataset_id        UUID NOT NULL REFERENCES datasets(id) ON DELETE CASCADE,
  version_id        UUID REFERENCES dataset_versions(id) ON DELETE SET NULL,
  storage_key       TEXT NOT NULL,
  source_image_id   UUID REFERENCES inspection_images(id) ON DELETE SET NULL,
  mime_type         TEXT,
  split             dataset_split NOT NULL DEFAULT 'UNASSIGNED',
  annotation_status annotation_status NOT NULL DEFAULT 'PENDING',
  metadata          JSONB,
  created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at        TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_dataset_items_dataset_split ON dataset_items (dataset_id, split);
CREATE INDEX idx_dataset_items_annotation_status ON dataset_items (annotation_status);

CREATE TABLE dataset_annotations (
  id                   UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  item_id              UUID NOT NULL REFERENCES dataset_items(id) ON DELETE CASCADE,
  annotator_id         UUID REFERENCES users(id) ON DELETE SET NULL,
  field_label          TEXT,
  classification_label TEXT,
  bbox                 JSONB,
  value                TEXT,
  confidence           DOUBLE PRECISION,
  status               annotation_status NOT NULL DEFAULT 'PENDING',
  metadata             JSONB,
  created_at           TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at           TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_dataset_annotations_item_id ON dataset_annotations (item_id);

-- ============================ AUDIT ==========================================

CREATE TABLE audit_logs (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  actor_id    UUID REFERENCES users(id) ON DELETE SET NULL,
  action      audit_action NOT NULL,
  entity_type TEXT NOT NULL,
  entity_id   TEXT,
  old_state   JSONB,
  new_state   JSONB,
  ip          TEXT,
  user_agent  TEXT,
  request_id  TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_entity ON audit_logs (entity_type, entity_id);
CREATE INDEX idx_audit_logs_actor_id ON audit_logs (actor_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs (created_at);
