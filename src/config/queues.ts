/**
 * BullMQ queue definitions.
 *
 * The API layer MUST NOT run expensive OCR/ML/report work synchronously. It
 * enqueues jobs here and returns a job id; workers (src/workers) consume them.
 */
import { Queue, type JobsOptions } from 'bullmq';
import { bullConnection } from './redis';

export const QUEUE_NAMES = {
  SCAN: 'scan-processing',
  OCR: 'ocr-processing',
  ML_INFERENCE: 'ml-inference',
  COMPLIANCE: 'compliance-check',
  REPORT: 'report-generation',
} as const;

export type QueueName = (typeof QUEUE_NAMES)[keyof typeof QUEUE_NAMES];

// ---- Job payload contracts (shared by producers and workers) ----------------

export interface ScanJobData {
  inspectionId: string;
  imageId: string;
  mlJobId: string;
}

export interface OcrJobData {
  inspectionId: string;
  imageId: string;
  mlJobId: string;
}

export interface ComplianceJobData {
  inspectionId: string;
  triggeredByUserId?: string;
}

export interface ReportJobData {
  inspectionId: string;
  reportId: string;
  type: 'PDF' | 'DOCX';
  requestedByUserId?: string;
}

const defaultJobOptions: JobsOptions = {
  attempts: 3,
  backoff: { type: 'exponential', delay: 5_000 },
  removeOnComplete: { age: 86_400, count: 1000 },
  removeOnFail: { age: 604_800 },
};

function makeQueue<T>(name: QueueName): Queue<T> {
  return new Queue<T>(name, { connection: bullConnection, defaultJobOptions });
}

export const scanQueue = makeQueue<ScanJobData>(QUEUE_NAMES.SCAN);
export const ocrQueue = makeQueue<OcrJobData>(QUEUE_NAMES.OCR);
export const mlQueue = makeQueue<OcrJobData>(QUEUE_NAMES.ML_INFERENCE);
export const complianceQueue = makeQueue<ComplianceJobData>(QUEUE_NAMES.COMPLIANCE);
export const reportQueue = makeQueue<ReportJobData>(QUEUE_NAMES.REPORT);

export const allQueues = [scanQueue, ocrQueue, mlQueue, complianceQueue, reportQueue];

export async function closeQueues(): Promise<void> {
  await Promise.all(allQueues.map((q) => q.close()));
}
