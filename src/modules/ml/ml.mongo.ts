/**
 * Mongoose models for raw ML/OCR/CV documents stored in MongoDB.
 *
 * These are intentionally schema-loose: they capture the raw, evolving output of
 * the ML models (OCR blocks, bounding boxes, confidences, detected regions,
 * preprocessing metadata, field extractions). PostgreSQL stores the normalized,
 * authoritative data; these documents are the evidence trail behind it.
 */
import { Schema, model, type InferSchemaType } from 'mongoose';

const bboxSchema = new Schema(
  { x: Number, y: Number, width: Number, height: Number },
  { _id: false },
);

// ---- Raw OCR document -------------------------------------------------------
const ocrBlockSchema = new Schema(
  {
    text: String,
    confidence: Number,
    bbox: bboxSchema,
  },
  { _id: false },
);

const ocrDocumentSchema = new Schema(
  {
    inspectionId: { type: String, required: true, index: true },
    imageId: { type: String, required: true, index: true },
    modelVersion: String,
    blocks: [ocrBlockSchema],
    preprocessing: Schema.Types.Mixed,
    createdAt: { type: Date, default: Date.now },
  },
  { collection: 'ocr_documents' },
);

export type OcrDocument = InferSchemaType<typeof ocrDocumentSchema>;
export const OcrDocumentModel = model('OcrDocument', ocrDocumentSchema);

// ---- Generic ML result document (detection / fields / analyze) --------------
const mlResultSchema = new Schema(
  {
    inspectionId: { type: String, required: true, index: true },
    imageId: { type: String, required: true, index: true },
    jobType: { type: String, index: true }, // OCR | LABEL_DETECTION | FIELD_EXTRACTION | ANALYZE
    modelVersion: String,
    // Raw, unmodified payload returned by the ML service.
    payload: Schema.Types.Mixed,
    createdAt: { type: Date, default: Date.now },
  },
  { collection: 'ml_results' },
);

export type MlResultDocument = InferSchemaType<typeof mlResultSchema>;
export const MlResultModel = model('MlResult', mlResultSchema);
