/**
 * Client for the internal Python ML service (OpenCV / PaddleOCR / custom models).
 *
 * This is the ONLY place Node talks to the ML service. Swapping the ML
 * implementation must not require changes outside this file and ml.types.ts.
 */
import { env } from '../../config/env';
import { httpJson } from '../../common/utils/http-client';
import type {
  MlImageRef,
  OcrResult,
  DetectLabelResult,
  ExtractFieldsResult,
  AnalyzeResult,
} from './ml.types';

function headers(): Record<string, string> {
  return env.ML_SERVICE_API_KEY ? { 'x-api-key': env.ML_SERVICE_API_KEY } : {};
}

function post<T>(path: string, body: MlImageRef): Promise<T> {
  return httpJson<T>(`${env.ML_SERVICE_URL}${path}`, {
    method: 'POST',
    headers: headers(),
    body,
    timeoutMs: env.ML_SERVICE_TIMEOUT_MS,
  });
}

export const mlClient = {
  ocr: (ref: MlImageRef) => post<OcrResult>('/ml/ocr', ref),
  detectLabel: (ref: MlImageRef) => post<DetectLabelResult>('/ml/detect-label', ref),
  extractFields: (ref: MlImageRef) => post<ExtractFieldsResult>('/ml/extract-fields', ref),
  analyze: (ref: MlImageRef) => post<AnalyzeResult>('/ml/analyze', ref),
  health: () =>
    httpJson<{ status: string; modelVersion: string }>(`${env.ML_SERVICE_URL}/health`, {
      timeoutMs: 5000,
    }),
};
