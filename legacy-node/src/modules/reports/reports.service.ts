/**
 * Report management. Generation is ASYNC: the API creates a Report row and
 * enqueues a job; the report worker collects the data, renders the file, uploads
 * it to object storage and marks the report READY. Downloads are served via a
 * short-lived signed URL.
 */
import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { reportQueue } from '../../config/queues';
import { getSignedDownloadUrl } from '../../config/storage';
import { ensureInspection } from '../inspections/inspections.service';

export async function requestReport(inspectionId: string, type: 'PDF' | 'DOCX', requestedById?: string) {
  await ensureInspection(inspectionId);

  const report = await prisma.report.create({
    data: { inspectionId, type: type as never, status: 'PENDING', generatedById: requestedById },
  });

  const job = await reportQueue.add('generate', {
    inspectionId,
    reportId: report.id,
    type,
    requestedByUserId: requestedById,
  });

  return { reportId: report.id, jobId: job.id, status: 'PENDING' as const };
}

export async function listByInspection(inspectionId: string) {
  await ensureInspection(inspectionId);
  return prisma.report.findMany({ where: { inspectionId }, orderBy: { createdAt: 'desc' } });
}

export async function getReport(id: string) {
  const report = await prisma.report.findUnique({ where: { id } });
  if (!report) throw new NotFoundError('Report');
  const downloadUrl =
    report.status === 'READY' && report.storageKey ? await getSignedDownloadUrl(report.storageKey) : null;
  return { ...report, downloadUrl };
}

// ---- Worker-side helpers ----------------------------------------------------

/** Collect everything the report needs from PostgreSQL. */
export async function collectReportData(inspectionId: string) {
  const inspection = await prisma.inspection.findUnique({
    where: { id: inspectionId },
    include: {
      product: { include: { category: true, manufacturer: true } },
      inspector: { select: { fullName: true, email: true } },
      reviewer: { select: { fullName: true, email: true } },
      declarations: true,
      checks: { include: { rule: { select: { code: true, title: true } } } },
      violations: { include: { rule: { select: { code: true, actReference: true } } } },
      evidence: true,
    },
  });
  if (!inspection) throw new NotFoundError('Inspection');
  return inspection;
}

export function markGenerating(reportId: string) {
  return prisma.report.update({ where: { id: reportId }, data: { status: 'GENERATING' } });
}

export function markReady(
  reportId: string,
  data: { storageKey: string; fileName: string; sizeBytes: number; checksum?: string },
) {
  return prisma.report.update({
    where: { id: reportId },
    data: {
      status: 'READY',
      storageKey: data.storageKey,
      fileName: data.fileName,
      sizeBytes: data.sizeBytes,
      checksum: data.checksum,
      generatedAt: new Date(),
    },
  });
}

export function markFailed(reportId: string, error: string) {
  return prisma.report.update({ where: { id: reportId }, data: { status: 'FAILED', error } });
}
