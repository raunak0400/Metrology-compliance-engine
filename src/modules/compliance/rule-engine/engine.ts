/**
 * The deterministic legal rule engine.
 *
 * Pure and side-effect free. Given the normalized declarations, measurement
 * results and the applicable legal rules, it emits one compliance check per rule
 * requirement and an overall outcome. Low-confidence extractions and ambiguous
 * measurement cases are routed to MANUAL_REVIEW — never auto-converted into
 * legal violations.
 */
import { applyOperator } from './operators';
import type {
  EngineContext,
  EngineRule,
  EngineDeclaration,
  EngineMeasurement,
  CheckResult,
  EngineOutput,
  ComplianceStatusName,
  OutcomeName,
} from './types';

const MANUAL_REVIEW_PREFIX = 'MANUAL_REVIEW:';

function findByType<T extends { type: string }>(items: T[], type: string | null): T | undefined {
  if (!type) return undefined;
  return items.find((i) => i.type === type);
}

export function evaluate(rules: EngineRule[], context: EngineContext): EngineOutput {
  const checks: CheckResult[] = [];

  for (const rule of rules) {
    for (const req of rule.requirements) {
      const declaration: EngineDeclaration | undefined = findByType(
        context.declarations,
        req.declarationType,
      );
      const measurement: EngineMeasurement | undefined = findByType(
        context.measurements,
        req.declarationType,
      );

      const result = applyOperator(req.operator, {
        declaration,
        measurement,
        expected: req.expected,
        context,
      });

      const isManualByOperator = result.reason?.startsWith(MANUAL_REVIEW_PREFIX) ?? false;
      const lowConfidence =
        declaration?.confidence != null && declaration.confidence < context.manualReviewThreshold;

      let status: ComplianceStatusName;
      let requiresManualReview = false;
      let reason = result.reason;

      if (isManualByOperator) {
        status = 'MANUAL_REVIEW';
        requiresManualReview = true;
        reason = result.reason!.slice(MANUAL_REVIEW_PREFIX.length);
      } else if (result.pass) {
        status = 'PASS';
      } else if (lowConfidence) {
        // A failing check on a low-confidence extraction is ambiguous — do NOT
        // auto-create a violation; send it for human review instead.
        status = 'MANUAL_REVIEW';
        requiresManualReview = true;
        reason = `${result.reason ?? 'Failed check'} (low OCR confidence ${declaration?.confidence})`;
      } else {
        status = 'FAIL';
      }

      checks.push({
        ruleId: rule.id,
        ruleCode: rule.code,
        requirementId: req.id,
        requirementKey: req.key,
        declarationId: declaration?.declarationId ?? null,
        status,
        severity: req.severity ?? rule.severityDefault,
        actualValue: result.actual,
        expectedValue: result.expected,
        reason: reason ?? null,
        confidence: declaration?.confidence ?? null,
        requiresManualReview,
        legalReference: rule.actReference,
        title: req.message ?? rule.title,
      });
    }
  }

  return { checks, outcome: deriveOutcome(rules, checks) };
}

/** Overall outcome: any manual review dominates; else any fail => non-compliant. */
function deriveOutcome(rules: EngineRule[], checks: CheckResult[]): OutcomeName {
  const mandatoryKeys = new Set(
    rules.flatMap((r) => r.requirements.filter((q) => q.isMandatory).map((q) => q.id)),
  );
  const relevant = checks.filter((c) => mandatoryKeys.has(c.requirementId));

  if (relevant.some((c) => c.status === 'MANUAL_REVIEW')) return 'MANUAL_REVIEW';
  if (relevant.some((c) => c.status === 'FAIL')) return 'NON_COMPLIANT';
  return 'COMPLIANT';
}
