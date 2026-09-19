/**
 * Redis connections (ioredis).
 *
 * - `redis`         : general-purpose client (cache, locks, idempotency, rate limit).
 * - `bullConnection`: connection options for BullMQ (needs maxRetriesPerRequest=null).
 *
 * BullMQ requires its own connections with specific settings, so we export the
 * options and let queues/workers create their own clients from them.
 */
import { Redis, type RedisOptions } from 'ioredis';
import { env } from './env';
import { logger } from './logger';

export const redis = new Redis(env.REDIS_URL, {
  maxRetriesPerRequest: 3,
  lazyConnect: false,
});

redis.on('connect', () => logger.info('Redis connected'));
redis.on('error', (err) => logger.error({ err }, 'Redis error'));

// BullMQ blocking commands require maxRetriesPerRequest = null.
export const bullConnection: RedisOptions & { url?: string } = (() => {
  const url = new URL(env.REDIS_URL);
  return {
    host: url.hostname,
    port: Number(url.port || 6379),
    password: url.password || undefined,
    username: url.username || undefined,
    db: url.pathname && url.pathname.length > 1 ? Number(url.pathname.slice(1)) : 0,
    maxRetriesPerRequest: null,
  };
})();

export async function disconnectRedis(): Promise<void> {
  await redis.quit();
  logger.info('Redis disconnected');
}
