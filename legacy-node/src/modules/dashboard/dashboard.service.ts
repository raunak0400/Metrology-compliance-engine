import { prisma } from '../../config/database';

/** High-level counts for the compliance dashboard. */
export async function getSummary() {
  const [totalInspections, byOutcome, byStatus, openViolations, totalProducts, totalRules] =
    await Promise.all([
      prisma.inspection.count(),
      prisma.inspection.groupBy({ by: ['outcome'], _count: { _all: true } }),
      prisma.inspection.groupBy({ by: ['status'], _count: { _all: true } }),
      prisma.violation.count({ where: { status: 'OPEN' } }),
      prisma.product.count(),
      prisma.legalRule.count({ where: { isActive: true } }),
    ]);

  const toMap = (rows: { _count: { _all: number } }[], key: string) =>
    Object.fromEntries(rows.map((r) => [(r as Record<string, unknown>)[key], r._count._all]));

  return {
    totalInspections,
    totalProducts,
    activeRules: totalRules,
    openViolations,
    inspectionsByOutcome: toMap(byOutcome, 'outcome'),
    inspectionsByStatus: toMap(byStatus, 'status'),
  };
}

/** Violation breakdown by severity + status, plus most-triggered rules. */
export async function getViolationStats() {
  const [bySeverity, byStatus, topRules] = await Promise.all([
    prisma.violation.groupBy({ by: ['severity'], _count: { _all: true } }),
    prisma.violation.groupBy({ by: ['status'], _count: { _all: true } }),
    prisma.violation.groupBy({
      by: ['ruleId'],
      _count: { _all: true },
      orderBy: { _count: { ruleId: 'desc' } },
      take: 10,
    }),
  ]);

  return {
    bySeverity: Object.fromEntries(bySeverity.map((r) => [r.severity, r._count._all])),
    byStatus: Object.fromEntries(byStatus.map((r) => [r.status, r._count._all])),
    topRules,
  };
}

/** Recent inspections + a simple daily created-count for the last 14 days. */
export async function getInspectionStats() {
  const since = new Date(Date.now() - 14 * 24 * 60 * 60 * 1000);
  const [recent, recentCount] = await Promise.all([
    prisma.inspection.findMany({
      take: 10,
      orderBy: { createdAt: 'desc' },
      select: { id: true, code: true, status: true, outcome: true, createdAt: true },
    }),
    prisma.inspection.count({ where: { createdAt: { gte: since } } }),
  ]);

  return { recent, last14DaysCount: recentCount };
}
