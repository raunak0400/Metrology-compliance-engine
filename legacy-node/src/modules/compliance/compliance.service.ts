/**
 * Compliance orchestration.
 *
 * Loads the normalized declarations + measurements + applicable legal rules from
 * PostgreSQL, runs the DETERMINISTIC rule engine, and persists the resulting
 * compliance checks and violations. Then it updates the inspection outcome.
 *
 * Pipeline position:
 *   declarations + measurements + rules -> engine -> checks -> violations
 * The engine (not any ML/LLM) decides compliance.
 */
import { prisma } from '../../config/database';
import { MANUAL_REVIEW_CONFIDENCE_THRESHOLD } from '../../common/constants';
import { complianceQueue } from '../../config/queues';
import { ensureInspection } from '../inspections/inspections.service';
import { computeMeasurements } from './measurement.service';
import { evaluate } from './rule-engine/engine';
import type {
  EngineRule,
  EngineDeclaration,
  CheckResult,
  OutcomeName,
} from './rule-engine/types';

/** Load active legal rules applicable to a category (category-mapped or global). */
async function loadApplicableRules(categoryId: string | null): Promise<EngineRule[]> {
  const where = categoryId
    ? {
        isActive: true,
        OR: [{ categories: { some: { categoryId } } }, { categories: { none: {} } }],
      }
    : { isActive: true };

  const rules = await prisma.legalRule.findMany({
    where,
    include: { requirements: { orderBy: { orderIndex: 'asc' } } },
  });

  return rules.map((r) => ({
    id: r.id,
    code: r.code,
    title: r.title,
    actReference: r.actReference,
    severityDefault: r.severityDefault as EngineRule['severityDefault'],
    requirements: r.requirements.map((q) => ({
      id: q.id,
      key: q.key,
      declarationType: q.declarationType,
      operator: q.operator as never,
      expected: (q.expected as Record<string, unknown> | null) ?? null,
      unit: q.unit,
      severity: (q.severity as EngineRule['severityDefault'] | null) ?? null,
      message: q.message,
      isMandatory: q.isMandatory,
    })),
  }));
}

async function loadDeclarations(inspectionId: string): Promise<EngineDeclaration[]> {
  const declarations = await prisma.extractedDeclaration.findMany({ where: { inspectionId } });
  return declarations.map((d) => ({
    declarationId: d.id,
    type: d.type,
    normalizedValue: d.normalizedValue,
    numericValue: d.numericValue !== null ? Number(d.numericValue) : null,
    unit: d.unit,
    confidence: d.confidence,
  }));
}

export interface ComplianceRunSummary {
  inspectionId: string;
  outcome: OutcomeName;
  totalChecks: number;
  passed: number;
  failed: number;
  manualReview: number;
  violations: number;
}

/** Run the full deterministic compliance evaluation and persist the results. */
export async function runCompliance(inspectionId: string): Promise<ComplianceRunSummary> {
  const inspection = await ensureInspection(inspectionId);

  const [declarations, measurements, rules] = await Promise.all([
    loadDeclarations(inspectionId),
    computeMeasurements(inspectionId),
    loadApplicableRules(inspection.categoryId),
  ]);

  const isImported = declarations.some(
    (d) => d.type === 'IMPORTER' && d.normalizedValue && d.normalizedValue.trim().length > 0,
  );

  const output = evaluate(rules, {
    productCategoryCode: null,
    isImported,
    declarations,
    measurements,
    manualReviewThreshold: MANUAL_REVIEW_CONFIDENCE_THRESHOLD,
  });

  await persistResults(inspectionId, output.checks, output.outcome);

  const passed = output.checks.filter((c) => c.status === 'PASS').length;
  const failed = output.checks.filter((c) => c.status === 'FAIL').length;
  const manualReview = output.checks.filter((c) => c.status === 'MANUAL_REVIEW').length;
  const violations = output.checks.filter((c) => c.status === 'FAIL').length;

  return {
    inspectionId,
    outcome: output.outcome,
    totalChecks: output.checks.length,
    passed,
    failed,
    manualReview,
    violations,
  };
}

/** Persist checks + violations atomically and update the inspection outcome. */
async function persistResults(inspectionId: string, checks: CheckResult[], outcome: OutcomeName) {
  await prisma.$transaction(async (tx) => {
    // Idempotent re-run: clear prior evaluation.
    await tx.violation.deleteMany({ where: { inspectionId } });
    await tx.complianceCheck.deleteMany({ where: { inspectionId } });

    for (const c of checks) {
      const check = await tx.complianceCheck.create({
        data: {
          inspectionId,
          ruleId: c.ruleId,
          requirementId: c.requirementId,
          declarationId: c.declarationId,
          status: c.status as never,
          actualValue: c.actualValue,
          expectedValue: c.expectedValue,
          reason: c.reason,
          confidence: c.confidence,
          requiresManualReview: c.requiresManualReview,
          details: { ruleCode: c.ruleCode, requirementKey: c.requirementKey } as never,
        },
      });

      // Only definite failures become violations. Manual-review checks do NOT.
      if (c.status === 'FAIL') {
        await tx.violation.create({
          data: {
            inspectionId,
            ruleId: c.ruleId,
            checkId: check.id,
            severity: c.severity as never,
            status: 'OPEN',
            title: c.title,
            reason: c.reason,
            legalReference: c.legalReference,
          },
        });
      }
    }

    const status = outcome === 'MANUAL_REVIEW' ? 'AWAITING_REVIEW' : 'COMPLETED';
    await tx.inspection.update({
      where: { id: inspectionId },
      data: { outcome: outcome as never, status: status as never, completedAt: new Date() },
    });
  });
}

/**
 * Enqueue a compliance evaluation on the compliance-check queue. The engine is
 * deterministic and fast, but we still run it via the worker so status
 * transitions and retries are consistent with the rest of the pipeline.
 */
export async function enqueueCompliance(inspectionId: string, triggeredByUserId?: string) {
  await ensureInspection(inspectionId);
  const job = await complianceQueue.add('run', { inspectionId, triggeredByUserId });
  await prisma.inspection.update({ where: { id: inspectionId }, data: { status: 'PROCESSING' } });
  return { jobId: job.id };
}

export async function listChecks(inspectionId: string) {
  await ensureInspection(inspectionId);
  return prisma.complianceCheck.findMany({
    where: { inspectionId },
    include: { rule: { select: { code: true, title: true } } },
    orderBy: { createdAt: 'asc' },
  });
}
