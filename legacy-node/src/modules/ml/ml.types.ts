/**
 * Contract types for the internal Python ML service. These describe the
 * STRUCTURED results the Node backend consumes. The Node backend never contains
 * ML logic — it only orchestrates and normalizes these results.
 */

export interface BBox {
  x: number;
  y: number;
  width: number;
  height: number;
}

/** Reference to an image the ML service should process. */
export interface MlImageRef {
  inspectionId: string;
  imageId: string;
  storageKey: string;
  /** Short-lived signed URL the ML service can download the image from. */
  imageUrl?: string;
}

export interface OcrBlock {
  text: string;
  confidence: number;
  bbox: BBox;
}

export interface OcrResult {
  inspectionId: string;
  imageId: string;
  modelVersion: string;
  blocks: OcrBlock[];
  preprocessing?: Record<string, unknown>;
  createdAt: string;
}

export interface DetectedRegion {
  type: string; // e.g. PRINCIPAL_DISPLAY_PANEL, LABEL
  confidence: number;
  bbox: BBox;
}

export interface DetectLabelResult {
  inspectionId: string;
  imageId: string;
  modelVersion: string;
  regions: DetectedRegion[];
  createdAt: string;
}

/** A candidate declaration field emitted by NER/classification models. */
export interface ExtractedField {
  field: string; // maps to DeclarationType, e.g. MRP, NET_QUANTITY
  rawValue: string;
  confidence: number;
  bbox?: BBox;
}

export interface ExtractFieldsResult {
  inspectionId: string;
  imageId: string;
  modelVersion: string;
  fields: ExtractedField[];
  createdAt: string;
}

export interface AnalyzeResult {
  inspectionId: string;
  imageId: string;
  modelVersion: string;
  ocr: OcrResult;
  regions: DetectedRegion[];
  fields: ExtractedField[];
  createdAt: string;
}
