/**
 * Compliance worker — consumes the `compliance-check` queue and runs the
 * DETERMINISTIC rule engine over the inspection's normalized declarations and
 * measurements, persisting checks + violations and updating the outcome.
 */
import { Worker } from 'bullmq';
import { bullConnection } from '../config/redis';
import { logger } from '../config/logger';
import { QUEUE_NAMES, type ComplianceJobData } from '../config/queues';
import { runCompliance } from '../modules/compliance/compliance.service';

export function startComplianceWorker(): Worker<ComplianceJobData> {
  const worker = new Worker<ComplianceJobData>(
    QUEUE_NAMES.COMPLIANCE,
    async (job) => {
      logger.info({ jobId: job.id, ...job.data }, 'Running compliance evaluation');
      const summary = await runCompliance(job.data.inspectionId);
      logger.info({ jobId: job.id, ...summary }, 'Compliance evaluation complete');
      return summary;
    },
    { connection: bullConnection, concurrency: 8 },
  );

  worker.on('failed', (job, err) => logger.error({ jobId: job?.id, err }, 'Compliance job failed'));
  return worker;
}
