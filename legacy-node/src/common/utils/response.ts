import type { Response } from 'express';
import type { ApiSuccess } from '../types';

/** Send a standard success envelope. */
export function sendSuccess<T>(
  res: Response,
  data: T,
  statusCode = 200,
  meta?: Record<string, unknown>,
): void {
  const body: ApiSuccess<T> = { success: true, data, ...(meta ? { meta } : {}) };
  res.status(statusCode).json(body);
}

export function sendCreated<T>(res: Response, data: T, meta?: Record<string, unknown>): void {
  sendSuccess(res, data, 201, meta);
}

export function sendNoContent(res: Response): void {
  res.status(204).send();
}
