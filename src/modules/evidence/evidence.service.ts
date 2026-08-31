import { prisma } from '../../config/database';
import { buildObjectKey, putObject, getSignedDownloadUrl } from '../../config/storage';
import { ensureInspection } from '../inspections/inspections.service';
import type { UploadedFile } from '../scans/scans.service';

export interface AddEvidenceInput {
  type?: string;
  violationId?: string;
  imageId?: string;
  bbox?: Record<string, unknown>;
  caption?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Attach evidence to an inspection. Evidence may be an uploaded file (stored in
 * object storage) and/or a reference to an existing inspection image + bbox.
 */
export async function addEvidence(inspectionId: string, input: AddEvidenceInput, file?: UploadedFile) {
  await ensureInspection(inspectionId);

  let storageKey: string | undefined;
  if (file) {
    storageKey = buildObjectKey(`inspections/${inspectionId}/evidence`, file.originalname);
    await putObject({ key: storageKey, body: file.buffer, contentType: file.mimetype });
  }

  return prisma.evidence.create({
    data: {
      inspectionId,
      violationId: input.violationId,
      imageId: input.imageId,
      type: (input.type as never) ?? 'ORIGINAL_IMAGE',
      storageKey,
      bbox: input.bbox as never,
      caption: input.caption,
      metadata: input.metadata as never,
    },
  });
}

export async function listByInspection(inspectionId: string) {
  await ensureInspection(inspectionId);
  const items = await prisma.evidence.findMany({
    where: { inspectionId },
    orderBy: { createdAt: 'desc' },
  });

  // Attach short-lived signed download URLs for any stored artifacts.
  return Promise.all(
    items.map(async (e) => ({
      ...e,
      downloadUrl: e.storageKey ? await getSignedDownloadUrl(e.storageKey) : null,
    })),
  );
}
