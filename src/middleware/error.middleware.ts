/**
 * Centralized error handling. Produces the consistent error envelope:
 *   { success: false, error: { code, message, details, requestId } }
 * Stack traces are never exposed in production.
 */
import type { Request, Response, NextFunction } from 'express';
import { ZodError } from 'zod';
import { Prisma } from '@prisma/client';
import { AppError, InternalError, NotFoundError } from '../common/errors';
import { ERROR_CODES } from '../common/constants';
import { isProd } from '../config/env';
import { logger } from '../config/logger';
import type { ApiError } from '../common/types';

/** 404 handler for unmatched routes. */
export function notFoundHandler(req: Request, _res: Response, next: NextFunction): void {
  next(new NotFoundError(`Route ${req.method} ${req.path}`));
}

function mapError(err: unknown): AppError {
  if (err instanceof AppError) return err;

  if (err instanceof ZodError) {
    const details = err.issues.map((i) => ({ field: i.path.join('.'), message: i.message, code: i.code }));
    return new AppError(422, ERROR_CODES.VALIDATION_ERROR, 'Validation failed', details);
  }

  if (err instanceof Prisma.PrismaClientKnownRequestError) {
    switch (err.code) {
      case 'P2002':
        return new AppError(
          409,
          ERROR_CODES.CONFLICT,
          `Unique constraint violated on ${(err.meta?.target as string[])?.join(', ') ?? 'field'}`,
        );
      case 'P2025':
        return new AppError(404, ERROR_CODES.NOT_FOUND, 'Record not found');
      case 'P2003':
        return new AppError(409, ERROR_CODES.CONFLICT, 'Foreign key constraint failed');
      default:
        return new AppError(400, ERROR_CODES.BAD_REQUEST, `Database error (${err.code})`);
    }
  }

  if (err instanceof Prisma.PrismaClientValidationError) {
    return new AppError(400, ERROR_CODES.BAD_REQUEST, 'Invalid database query');
  }

  return new InternalError();
}

// Express error middleware MUST have four args to be recognized as such.
// eslint-disable-next-line @typescript-eslint/no-unused-vars
export function errorHandler(err: unknown, req: Request, res: Response, _next: NextFunction): void {
  const appError = mapError(err);
  const requestId = req.id !== undefined ? String(req.id) : undefined;

  if (!appError.isOperational || appError.statusCode >= 500) {
    logger.error({ err, requestId }, 'Unhandled error');
  } else {
    logger.warn({ code: appError.code, msg: appError.message, requestId }, 'Handled error');
  }

  const body: ApiError = {
    success: false,
    error: {
      code: appError.code,
      message: appError.message,
      ...(appError.details ? { details: appError.details } : {}),
      ...(requestId ? { requestId } : {}),
    },
  };

  // Only include a stack outside production, for debugging.
  if (!isProd && appError.statusCode >= 500 && err instanceof Error) {
    (body.error as Record<string, unknown>).stack = err.stack;
  }

  res.status(appError.statusCode).json(body);
}
