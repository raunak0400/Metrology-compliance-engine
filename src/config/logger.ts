/**
 * Structured logging with Pino. Use `logger` for app-wide logging and
 * `httpLogger` as Express middleware (adds a request id and per-request logs).
 */
import pino from 'pino';
import { pinoHttp } from 'pino-http';
import { randomUUID } from 'node:crypto';
import { env, isProd, isTest } from './env';

export const logger = pino({
  level: env.LOG_LEVEL,
  base: { service: 'legal-metrology-api' },
  redact: {
    paths: [
      'req.headers.authorization',
      'req.headers.cookie',
      'password',
      'passwordHash',
      '*.password',
      '*.passwordHash',
    ],
    censor: '[REDACTED]',
  },
  // Pretty transport only in local development. In production we emit JSON; in
  // test we avoid the transport worker thread (keeps Jest teardown clean).
  transport:
    isProd || isTest
      ? undefined
      : { target: 'pino-pretty', options: { colorize: true, translateTime: 'SYS:standard' } },
});

export const httpLogger = pinoHttp({
  logger,
  // Reuse an inbound x-request-id or generate one; exposed as req.id.
  genReqId: (req, res) => {
    const existing = (req.headers['x-request-id'] as string) || randomUUID();
    res.setHeader('x-request-id', existing);
    return existing;
  },
  customProps: (req) => ({
    // Populated by auth.middleware when a valid token is present.
    userId: (req as { user?: { id?: string } }).user?.id,
  }),
  customLogLevel: (_req, res, err) => {
    if (err || res.statusCode >= 500) return 'error';
    if (res.statusCode >= 400) return 'warn';
    return 'info';
  },
  autoLogging: {
    ignore: (req) => req.url === '/health' || req.url === '/healthz',
  },
});
