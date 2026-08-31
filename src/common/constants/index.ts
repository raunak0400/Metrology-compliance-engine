/** Shared constants: error codes, role names, permission scopes, misc. */

export const ERROR_CODES = {
  VALIDATION_ERROR: 'VALIDATION_ERROR',
  UNAUTHORIZED: 'UNAUTHORIZED',
  FORBIDDEN: 'FORBIDDEN',
  NOT_FOUND: 'NOT_FOUND',
  CONFLICT: 'CONFLICT',
  RATE_LIMITED: 'RATE_LIMITED',
  BAD_REQUEST: 'BAD_REQUEST',
  UNSUPPORTED_MEDIA_TYPE: 'UNSUPPORTED_MEDIA_TYPE',
  PAYLOAD_TOO_LARGE: 'PAYLOAD_TOO_LARGE',
  INTERNAL_ERROR: 'INTERNAL_ERROR',
  SERVICE_UNAVAILABLE: 'SERVICE_UNAVAILABLE',
} as const;

export type ErrorCode = (typeof ERROR_CODES)[keyof typeof ERROR_CODES];

/** Built-in role names (roles are extensible; these are the seeded defaults). */
export const ROLES = {
  ADMIN: 'ADMIN',
  INSPECTOR: 'INSPECTOR',
  REVIEWER: 'REVIEWER',
  SUPERVISOR: 'SUPERVISOR',
} as const;

export type RoleName = (typeof ROLES)[keyof typeof ROLES];

/** Coarse permission scopes used by policy functions. `*` grants everything. */
export const PERMISSIONS = {
  INSPECTIONS_READ: 'inspections:read',
  INSPECTIONS_WRITE: 'inspections:write',
  INSPECTIONS_FINALIZE: 'inspections:finalize',
  PRODUCTS_READ: 'products:read',
  PRODUCTS_WRITE: 'products:write',
  SCANS_WRITE: 'scans:write',
  DECLARATIONS_WRITE: 'declarations:write',
  VIOLATIONS_WRITE: 'violations:write',
  RULES_READ: 'rules:read',
  RULES_WRITE: 'rules:write',
  REPORTS_READ: 'reports:read',
  DATASETS_WRITE: 'datasets:write',
  AUDIT_READ: 'audit:read',
  DASHBOARD_READ: 'dashboard:read',
} as const;

export const PAGINATION = {
  DEFAULT_PAGE: 1,
  DEFAULT_LIMIT: 20,
  MAX_LIMIT: 100,
} as const;

/**
 * Confidence threshold below which an extracted declaration / check is routed to
 * manual review rather than being turned into an automatic legal violation.
 */
export const MANUAL_REVIEW_CONFIDENCE_THRESHOLD = 0.7;
