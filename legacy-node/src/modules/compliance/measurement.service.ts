/**
 * Measurement service.
 *
 * Deterministic geometry only — kept deliberately SEPARATE from OCR and from the
 * legal rule engine. It estimates real-world sizes (e.g. declaration text height)
 * from bounding boxes, but ONLY when a valid scale reference exists. Without a
 * valid scale it returns `hasValidScale: false` and a null measurement: we never
 * pretend pixel-to-mm conversion is accurate. The rule engine consumes these
 * results; it does not compute them.
 */
import { prisma } from '../../config/database';
import type { EngineMeasurement } from './rule-engine/types';

export interface ScaleReference {
  /** Direct pixels-per-mm, if known. */
  pxPerMm?: number;
  /** Or derived from a reference object of known real size. */
  referenceObjectMm?: number;
  referenceObjectPx?: number;
  source?: string;
}

export interface BBox {
  x: number;
  y: number;
  width: number;
  height: number;
}

/** Resolve a pixels-per-mm ratio from a scale reference, or null if invalid. */
export function resolvePxPerMm(scale: ScaleReference | null | undefined): number | null {
  if (!scale) return null;
  if (typeof scale.pxPerMm === 'number' && scale.pxPerMm > 0) return scale.pxPerMm;
  if (
    typeof scale.referenceObjectMm === 'number' &&
    typeof scale.referenceObjectPx === 'number' &&
    scale.referenceObjectMm > 0 &&
    scale.referenceObjectPx > 0
  ) {
    return scale.referenceObjectPx / scale.referenceObjectMm;
  }
  return null;
}

/** Estimate text height in mm from a bbox; null when no valid scale reference. */
export function estimateTextHeightMm(
  bbox: BBox | null | undefined,
  scale: ScaleReference | null | undefined,
): { heightMm: number | null; hasValidScale: boolean } {
  const pxPerMm = resolvePxPerMm(scale);
  if (!bbox || pxPerMm === null) return { heightMm: null, hasValidScale: false };
  return { heightMm: Number((bbox.height / pxPerMm).toFixed(3)), hasValidScale: true };
}

/**
 * Build measurement results for an inspection from its declarations' bounding
 * boxes and each source image's scale reference.
 */
export async function computeMeasurements(inspectionId: string): Promise<EngineMeasurement[]> {
  const declarations = await prisma.extractedDeclaration.findMany({
    where: { inspectionId },
    select: { id: true, type: true, bbox: true, sourceImage: { select: { scaleReference: true } } },
  });

  return declarations.map((d) => {
    const scale = (d.sourceImage?.scaleReference ?? null) as ScaleReference | null;
    const { heightMm, hasValidScale } = estimateTextHeightMm(d.bbox as BBox | null, scale);
    return { declarationId: d.id, type: d.type, heightMm, hasValidScale };
  });
}
