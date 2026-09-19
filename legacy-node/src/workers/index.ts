/**
 * Worker fleet bootstrap. Connects the infrastructure the workers need
 * (PostgreSQL + MongoDB; Redis via BullMQ) and starts all BullMQ workers.
 *
 * Run separately from the API: `npm run dev:workers` / `npm run start:workers`.
 */
import { logger } from '../config/logger';
import { connectPostgres, disconnectPostgres } from '../config/database';
import { connectMongo, disconnectMongo } from '../config/mongodb';
import { disconnectRedis } from '../config/redis';
import { startScanWorker } from './scan.worker';
import { startOcrWorker } from './ocr.worker';
import { startComplianceWorker } from './compliance.worker';
import { startReportWorker } from './report.worker';

async function bootstrap(): Promise<void> {
  await connectPostgres();
  await connectMongo();

  const workers = [startScanWorker(), startOcrWorker(), startComplianceWorker(), startReportWorker()];
  logger.info(`Started ${workers.length} workers: scan, ocr, compliance, report`);

  const shutdown = async (signal: string) => {
    logger.info({ signal }, 'Shutting down workers...');
    await Promise.allSettled(workers.map((w) => w.close()));
    await Promise.allSettled([disconnectPostgres(), disconnectMongo(), disconnectRedis()]);
    process.exit(0);
  };

  process.on('SIGTERM', () => void shutdown('SIGTERM'));
  process.on('SIGINT', () => void shutdown('SIGINT'));
}

bootstrap().catch((err) => {
  logger.fatal({ err }, 'Failed to start workers');
  process.exit(1);
});
