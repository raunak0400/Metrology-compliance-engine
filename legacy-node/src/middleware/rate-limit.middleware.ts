/**
 * Rate limiting backed by Redis so limits are shared across API instances.
 */
import rateLimit, { type RateLimitRequestHandler } from 'express-rate-limit';
import { RedisStore } from 'rate-limit-redis';
import { redis } from '../config/redis';
import { env } from '../config/env';
import { ERROR_CODES } from '../common/constants';

function makeLimiter(opts: { windowMs: number; max: number; prefix: string }): RateLimitRequestHandler {
  return rateLimit({
    windowMs: opts.windowMs,
    max: opts.max,
    standardHeaders: true,
    legacyHeaders: false,
    store: new RedisStore({
      prefix: `rl:${opts.prefix}:`,
      // ioredis: forward commands to the shared client.
      sendCommand: (command: string, ...args: string[]) =>
        redis.call(command, ...args) as Promise<never>,
    }),
    keyGenerator: (req) => req.user?.id ?? req.ip ?? 'anonymous',
    message: {
      success: false,
      error: { code: ERROR_CODES.RATE_LIMITED, message: 'Too many requests, please try again later.' },
    },
  });
}

/** Global API limiter. */
export const globalRateLimiter = makeLimiter({
  windowMs: env.RATE_LIMIT_WINDOW_MS,
  max: env.RATE_LIMIT_MAX,
  prefix: 'global',
});

/** Stricter limiter for auth endpoints (brute-force protection). */
export const authRateLimiter = makeLimiter({
  windowMs: 15 * 60 * 1000,
  max: 20,
  prefix: 'auth',
});
