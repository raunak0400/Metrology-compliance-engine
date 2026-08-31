/**
 * Persistence helpers for raw ML output in MongoDB. Returns the Mongo document
 * id so a reference can be stored on the authoritative PostgreSQL rows
 * (InspectionImage.mongoOcrDocId, ExtractedDeclaration.mongoRefId, MLJob.resultMongoId).
 */
import { OcrDocumentModel, MlResultModel } from './ml.mongo';
import type { OcrResult, AnalyzeResult, DetectLabelResult, ExtractFieldsResult } from './ml.types';

export async function saveOcrDocument(result: OcrResult): Promise<string> {
  const doc = await OcrDocumentModel.create({
    inspectionId: result.inspectionId,
    imageId: result.imageId,
    modelVersion: result.modelVersion,
    blocks: result.blocks,
    preprocessing: result.preprocessing,
    createdAt: new Date(result.createdAt),
  });
  return doc.id as string;
}

export async function saveMlResult(
  jobType: 'OCR' | 'LABEL_DETECTION' | 'FIELD_EXTRACTION' | 'ANALYZE',
  payload: OcrResult | AnalyzeResult | DetectLabelResult | ExtractFieldsResult,
): Promise<string> {
  const doc = await MlResultModel.create({
    inspectionId: payload.inspectionId,
    imageId: payload.imageId,
    jobType,
    modelVersion: payload.modelVersion,
    payload,
  });
  return doc.id as string;
}

export function getOcrDocument(id: string) {
  return OcrDocumentModel.findById(id).lean();
}
