/**
 * Express application setup: security middleware, structured logging, body
 * parsing, health checks, API routes and centralized error handling.
 * (HTTP server bootstrap lives in server.ts.)
 */
import express, { type Application } from 'express';
import helmet from 'helmet';
import cors from 'cors';
import { env } from './config/env';
import { httpLogger } from './config/logger';
import { globalRateLimiter } from './middleware/rate-limit.middleware';
import { errorHandler, notFoundHandler } from './middleware/error.middleware';
import { apiRouter } from './routes';

export function createApp(): Application {
  const app = express();

  // Trust the first proxy (needed for correct req.ip behind a load balancer).
  app.set('trust proxy', 1);

  // Security headers + CORS.
  app.use(helmet());
  app.use(
    cors({
      origin: env.CORS_ORIGINS,
      credentials: true,
    }),
  );

  // Structured request logging (adds x-request-id).
  app.use(httpLogger);

  // Body parsing.
  app.use(express.json({ limit: '1mb' }));
  app.use(express.urlencoded({ extended: true }));

  // Liveness/readiness (excluded from rate limiting and auth).
  app.get('/health', (_req, res) => {
    res.json({ success: true, data: { status: 'ok', uptime: process.uptime() } });
  });

  // Global rate limiting for the API surface.
  app.use(env.API_PREFIX, globalRateLimiter, apiRouter);

  // 404 + centralized error handling (must be last).
  app.use(notFoundHandler);
  app.use(errorHandler);

  return app;
}

export default createApp;
