/**
 * HTTP server bootstrap. Connects infrastructure (PostgreSQL, MongoDB, storage),
 * starts the Express app, and wires graceful shutdown.
 *
 * Note: BullMQ WORKERS are started separately (src/workers) so the API and the
 * worker fleet can scale independently.
 */
import { createApp } from './app';
import { env } from './config/env';
import { logger } from './config/logger';
import { connectPostgres, disconnectPostgres } from './config/database';
import { connectMongo, disconnectMongo } from './config/mongodb';
import { ensureBucket } from './config/storage';
import { disconnectRedis } from './config/redis';
import { closeQueues } from './config/queues';

async function bootstrap(): Promise<void> {
  await connectPostgres();
  await connectMongo();
  // Best-effort: ensure the dev bucket exists (no-op if it already does).
  await ensureBucket().catch((err) => logger.warn({ err }, 'Could not ensure storage bucket'));

  const app = createApp();
  const server = app.listen(env.PORT, () => {
    logger.info(`API listening on http://localhost:${env.PORT}${env.API_PREFIX}`);
  });

  const shutdown = async (signal: string) => {
    logger.info({ signal }, 'Shutting down...');
    server.close();
    await Promise.allSettled([closeQueues(), disconnectPostgres(), disconnectMongo(), disconnectRedis()]);
    process.exit(0);
  };

  process.on('SIGTERM', () => void shutdown('SIGTERM'));
  process.on('SIGINT', () => void shutdown('SIGINT'));
}

bootstrap().catch((err) => {
  logger.fatal({ err }, 'Failed to start server');
  process.exit(1);
});
