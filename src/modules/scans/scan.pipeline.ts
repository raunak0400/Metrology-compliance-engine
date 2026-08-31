/**
 * Scan processing pipeline (worker-side orchestration).
 *
 * Enforces the mandated separation:
 *   IMAGE -> (Python ML: preprocess/detect/OCR/extract) -> RAW ML RESULT (MongoDB)
 *         -> NORMALIZATION -> STRUCTURED DECLARATIONS (PostgreSQL)
 *         -> enqueue DETERMINISTIC compliance evaluation
 *
 * Node never runs ML here — it calls the ML service, stores raw output in Mongo,
 * normalizes into PostgreSQL, then hands off to the compliance queue.
 */
import { prisma } from '../../config/database';
import { logger } from '../../config/logger';
import { getSignedDownloadUrl } from '../../config/storage';
import { complianceQueue } from '../../config/queues';
import { mlClient } from '../ml/ml.client';
import { saveOcrDocument, saveMlResult } from '../ml/ml.repository';
import { persistExtractedDeclarations } from '../declarations/declarations.service';
import type { ScanJobData } from '../../config/queues';
import type { MlImageRef } from '../ml/ml.types';

export async function processScanJob(data: ScanJobData): Promise<void> {
  const { inspectionId, imageId, mlJobId } = data;
  const log = logger.child({ inspectionId, imageId, mlJobId });

  const image = await prisma.inspectionImage.findUnique({ where: { id: imageId } });
  if (!image) throw new Error(`InspectionImage ${imageId} not found`);

  await prisma.mLJob.update({
    where: { id: mlJobId },
    data: { status: 'RUNNING', startedAt: new Date() },
  });
  await prisma.inspectionImage.update({ where: { id: imageId }, data: { status: 'PROCESSING' } });

  try {
    // The ML service downloads the image via a short-lived signed URL.
    const imageUrl = await getSignedDownloadUrl(image.storageKey);
    const ref: MlImageRef = { inspectionId, imageId, storageKey: image.storageKey, imageUrl };

    // 1) Call the Python ML service (preprocess + detect + OCR + extract).
    const analysis = await mlClient.analyze(ref);

    // 2) Persist RAW ML output to MongoDB (evidence trail, not source of truth).
    const ocrDocId = await saveOcrDocument(analysis.ocr);
    const analyzeDocId = await saveMlResult('ANALYZE', analysis);

    // 3) Normalize fields into authoritative PostgreSQL declarations.
    await persistExtractedDeclarations(inspectionId, imageId, analysis.fields, analyzeDocId);

    // 4) Update authoritative rows with references to the raw Mongo documents.
    await prisma.inspectionImage.update({
      where: { id: imageId },
      data: { status: 'PROCESSED', mongoOcrDocId: ocrDocId },
    });
    await prisma.mLJob.update({
      where: { id: mlJobId },
      data: { status: 'SUCCEEDED', finishedAt: new Date(), resultMongoId: analyzeDocId },
    });

    // 5) Hand off to the DETERMINISTIC compliance engine (separate queue).
    await complianceQueue.add('run', { inspectionId });

    log.info('Scan pipeline completed; compliance evaluation queued');
  } catch (err) {
    log.error({ err }, 'Scan pipeline failed');
    await prisma.inspectionImage.update({ where: { id: imageId }, data: { status: 'FAILED' } });
    await prisma.mLJob.update({
      where: { id: mlJobId },
      data: {
        status: 'FAILED',
        finishedAt: new Date(),
        error: (err as Error).message,
        attempts: { increment: 1 },
      },
    });
    throw err; // let BullMQ apply retry/backoff
  }
}
