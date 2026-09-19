import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { ensureInspection } from '../inspections/inspections.service';
import { normalize } from './normalization.service';
import type { ExtractedField } from '../ml/ml.types';

export async function listByInspection(inspectionId: string) {
  await ensureInspection(inspectionId);
  return prisma.extractedDeclaration.findMany({
    where: { inspectionId },
    orderBy: { type: 'asc' },
  });
}

export async function getDeclaration(id: string) {
  const declaration = await prisma.extractedDeclaration.findUnique({ where: { id } });
  if (!declaration) throw new NotFoundError('Declaration');
  return declaration;
}

/**
 * Reviewer edit. Records that a human corrected the value and by whom, without
 * discarding the original raw OCR value (rawValue is left untouched).
 */
export async function updateDeclaration(
  id: string,
  data: { normalizedValue?: string; numericValue?: number; unit?: string; type?: string },
  editedById?: string,
) {
  await getDeclaration(id);
  return prisma.extractedDeclaration.update({
    where: { id },
    data: {
      normalizedValue: data.normalizedValue,
      numericValue: data.numericValue,
      unit: data.unit,
      type: data.type as never,
      source: 'MANUAL',
      isEditedByReviewer: true,
      editedById,
    },
  });
}

/**
 * Persist normalized declarations from raw ML field-extraction output. Called by
 * the pipeline worker. Preserves rawValue/confidence/bbox as evidence and links
 * each declaration back to its source image and the raw Mongo document.
 */
export async function persistExtractedDeclarations(
  inspectionId: string,
  imageId: string,
  fields: ExtractedField[],
  mongoRefId?: string,
) {
  if (fields.length === 0) return [];

  // Idempotency: replace prior ML-sourced declarations for this image.
  await prisma.extractedDeclaration.deleteMany({
    where: { inspectionId, sourceImageId: imageId, source: { in: ['OCR', 'ML_EXTRACTION'] } },
  });

  const data = fields.map((f) => {
    const n = normalize(f);
    return {
      inspectionId,
      sourceImageId: imageId,
      type: n.type as never,
      normalizedValue: n.normalizedValue,
      rawValue: f.rawValue,
      unit: n.unit,
      numericValue: n.numericValue,
      confidence: f.confidence,
      bbox: (f.bbox ?? undefined) as never,
      source: 'ML_EXTRACTION' as never,
      mongoRefId,
    };
  });

  await prisma.extractedDeclaration.createMany({ data });
  return prisma.extractedDeclaration.findMany({ where: { inspectionId, sourceImageId: imageId } });
}
