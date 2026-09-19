/**
 * Types for the DETERMINISTIC legal rule engine.
 *
 * The engine is a pure function: (declarations + measurements + applicable rules)
 * -> (compliance checks + overall outcome). It performs NO database access, NO
 * network calls, and contains NO ML. An LLM/ML model must never make the
 * compliance decision — only this engine does, from structured inputs.
 */

export type ComplianceStatusName = 'PASS' | 'FAIL' | 'MANUAL_REVIEW' | 'NOT_APPLICABLE';
export type OutcomeName = 'COMPLIANT' | 'NON_COMPLIANT' | 'MANUAL_REVIEW';
export type SeverityName = 'LOW' | 'MEDIUM' | 'HIGH' | 'CRITICAL';

export type OperatorName =
  | 'EXISTS'
  | 'NOT_EMPTY'
  | 'REGEX_MATCH'
  | 'NUMERIC_MIN'
  | 'NUMERIC_MAX'
  | 'NUMERIC_RANGE'
  | 'MIN_TEXT_HEIGHT_MM'
  | 'ENUM_IN'
  | 'DATE_VALID'
  | 'DATE_NOT_FUTURE'
  | 'REQUIRED_IF_IMPORTED'
  | 'CUSTOM';

/** Normalized declaration as the engine sees it (subset of the DB row). */
export interface EngineDeclaration {
  declarationId?: string;
  type: string;
  normalizedValue: string | null;
  numericValue: number | null;
  unit: string | null;
  confidence: number | null;
}

/** A measurement result the engine consumes (produced by measurement.service). */
export interface EngineMeasurement {
  declarationId?: string;
  type: string;
  /** Estimated real-world height in mm, or null when no valid scale reference. */
  heightMm: number | null;
  hasValidScale: boolean;
}

export interface EngineRequirement {
  id: string;
  key: string;
  declarationType: string | null;
  operator: OperatorName;
  expected: Record<string, unknown> | null;
  unit: string | null;
  severity: SeverityName | null;
  message: string | null;
  isMandatory: boolean;
}

export interface EngineRule {
  id: string;
  code: string;
  title: string;
  actReference: string | null;
  severityDefault: SeverityName;
  requirements: EngineRequirement[];
}

export interface EngineContext {
  productCategoryCode?: string | null;
  isImported?: boolean;
  declarations: EngineDeclaration[];
  measurements: EngineMeasurement[];
  /** Below this confidence, a check is routed to MANUAL_REVIEW, never auto-FAIL. */
  manualReviewThreshold: number;
}

export interface CheckResult {
  ruleId: string;
  ruleCode: string;
  requirementId: string;
  requirementKey: string;
  declarationId: string | null;
  status: ComplianceStatusName;
  severity: SeverityName;
  actualValue: string | null;
  expectedValue: string | null;
  reason: string | null;
  confidence: number | null;
  requiresManualReview: boolean;
  legalReference: string | null;
  title: string;
}

export interface EngineOutput {
  checks: CheckResult[];
  outcome: OutcomeName;
}
