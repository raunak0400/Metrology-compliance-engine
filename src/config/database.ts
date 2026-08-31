/**
 * PostgreSQL access via Prisma. PostgreSQL is the authoritative source of truth.
 * Exposes a single shared PrismaClient instance (avoids exhausting connections
 * during dev hot-reload).
 */
import { PrismaClient } from '@prisma/client';
import { env, isProd } from './env';
import { logger } from './logger';

const globalForPrisma = globalThis as unknown as { prisma?: PrismaClient };

export const prisma =
  globalForPrisma.prisma ??
  new PrismaClient({
    log: isProd ? ['warn', 'error'] : ['warn', 'error'],
  });

if (!isProd) globalForPrisma.prisma = prisma;

export async function connectPostgres(): Promise<void> {
  await prisma.$connect();
  logger.info('PostgreSQL (Prisma) connected');
}

export async function disconnectPostgres(): Promise<void> {
  await prisma.$disconnect();
  logger.info('PostgreSQL (Prisma) disconnected');
}

void env; // ensures env is validated on import
