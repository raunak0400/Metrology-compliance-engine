/**
 * OCR worker — consumes the `ocr-processing` queue for the granular OCR-only step.
 *
 * The default pipeline uses the combined `analyze` call in the scan worker; this
 * worker exists for the ML team's alternative flow where OCR is run independently
 * (e.g. re-OCR with a different model version). It stores raw OCR in MongoDB and
 * updates the corresponding MLJob — it does NOT make compliance decisions.
 */
import { Worker } from 'bullmq';
import { bullConnection } from '../config/redis';
import { logger } from '../config/logger';
import { prisma } from '../config/database';
import { getSignedDownloadUrl } from '../config/storage';
import { QUEUE_NAMES, type OcrJobData } from '../config/queues';
import { mlClient } from '../modules/ml/ml.client';
import { saveOcrDocument } from '../modules/ml/ml.repository';
import type { MlImageRef } from '../modules/ml/ml.types';

export function startOcrWorker(): Worker<OcrJobData> {
  const worker = new Worker<OcrJobData>(
    QUEUE_NAMES.OCR,
    async (job) => {
      const { inspectionId, imageId, mlJobId } = job.data;
      const image = await prisma.inspectionImage.findUnique({ where: { id: imageId } });
      if (!image) throw new Error(`InspectionImage ${imageId} not found`);

      await prisma.mLJob.update({ where: { id: mlJobId }, data: { status: 'RUNNING', startedAt: new Date() } });

      const imageUrl = await getSignedDownloadUrl(image.storageKey);
      const ref: MlImageRef = { inspectionId, imageId, storageKey: image.storageKey, imageUrl };
      const ocr = await mlClient.ocr(ref);
      const ocrDocId = await saveOcrDocument(ocr);

      await prisma.inspectionImage.update({ where: { id: imageId }, data: { mongoOcrDocId: ocrDocId } });
      await prisma.mLJob.update({
        where: { id: mlJobId },
        data: { status: 'SUCCEEDED', finishedAt: new Date(), resultMongoId: ocrDocId },
      });
    },
    { connection: bullConnection, concurrency: 4 },
  );

  worker.on('failed', (job, err) => logger.error({ jobId: job?.id, err }, 'OCR job failed'));
  return worker;
}
