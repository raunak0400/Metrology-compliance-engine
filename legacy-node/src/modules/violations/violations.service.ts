import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { ensureInspection } from '../inspections/inspections.service';

export async function listByInspection(inspectionId: string) {
  await ensureInspection(inspectionId);
  return prisma.violation.findMany({
    where: { inspectionId },
    include: {
      rule: { select: { code: true, title: true, actReference: true } },
      evidence: true,
    },
    orderBy: [{ severity: 'desc' }, { createdAt: 'asc' }],
  });
}

export async function getViolation(id: string) {
  const violation = await prisma.violation.findUnique({ where: { id }, include: { evidence: true } });
  if (!violation) throw new NotFoundError('Violation');
  return violation;
}

/**
 * Reviewer action on a violation: confirm, dismiss or resolve. Records who acted.
 * Confirming/rejecting a violation is a human decision layered on top of the
 * deterministic engine output — it does not alter the original check.
 */
export async function updateViolation(
  id: string,
  data: { status?: string; reviewNotes?: string; severity?: string },
  resolvedById?: string,
) {
  await getViolation(id);
  const terminal = data.status === 'DISMISSED' || data.status === 'RESOLVED';
  return prisma.violation.update({
    where: { id },
    data: {
      status: data.status as never,
      severity: data.severity as never,
      reviewNotes: data.reviewNotes,
      resolvedById: terminal ? resolvedById : undefined,
      resolvedAt: terminal ? new Date() : undefined,
    },
  });
}
