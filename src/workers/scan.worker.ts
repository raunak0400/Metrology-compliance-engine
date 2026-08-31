/**
 * Scan worker — consumes the `scan-processing` queue and runs the full image
 * pipeline (ML call -> raw Mongo -> normalized declarations -> compliance handoff).
 */
import { Worker } from 'bullmq';
import { bullConnection } from '../config/redis';
import { logger } from '../config/logger';
import { QUEUE_NAMES, type ScanJobData } from '../config/queues';
import { processScanJob } from '../modules/scans/scan.pipeline';

export function startScanWorker(): Worker<ScanJobData> {
  const worker = new Worker<ScanJobData>(
    QUEUE_NAMES.SCAN,
    async (job) => {
      logger.info({ jobId: job.id, ...job.data }, 'Processing scan job');
      await processScanJob(job.data);
    },
    { connection: bullConnection, concurrency: 4 },
  );

  worker.on('failed', (job, err) =>
    logger.error({ jobId: job?.id, err }, 'Scan job failed'),
  );
  return worker;
}
