/**
 * Application error hierarchy. Throw these from services/controllers; the
 * centralized error middleware maps them to consistent JSON responses.
 */
import { ERROR_CODES, type ErrorCode } from '../constants';

export interface ErrorDetail {
  field?: string;
  message: string;
  code?: string;
}

export class AppError extends Error {
  public readonly statusCode: number;
  public readonly code: ErrorCode;
  public readonly details?: ErrorDetail[];
  public readonly isOperational: boolean;

  constructor(
    statusCode: number,
    code: ErrorCode,
    message: string,
    details?: ErrorDetail[],
    isOperational = true,
  ) {
    super(message);
    this.name = this.constructor.name;
    this.statusCode = statusCode;
    this.code = code;
    this.details = details;
    this.isOperational = isOperational;
    Error.captureStackTrace?.(this, this.constructor);
  }
}

export class ValidationError extends AppError {
  constructor(message = 'Validation failed', details?: ErrorDetail[]) {
    super(422, ERROR_CODES.VALIDATION_ERROR, message, details);
  }
}

export class BadRequestError extends AppError {
  constructor(message = 'Bad request', details?: ErrorDetail[]) {
    super(400, ERROR_CODES.BAD_REQUEST, message, details);
  }
}

export class UnauthorizedError extends AppError {
  constructor(message = 'Authentication required') {
    super(401, ERROR_CODES.UNAUTHORIZED, message);
  }
}

export class ForbiddenError extends AppError {
  constructor(message = 'You do not have permission to perform this action') {
    super(403, ERROR_CODES.FORBIDDEN, message);
  }
}

export class NotFoundError extends AppError {
  constructor(resource = 'Resource') {
    super(404, ERROR_CODES.NOT_FOUND, `${resource} not found`);
  }
}

export class ConflictError extends AppError {
  constructor(message = 'Resource already exists') {
    super(409, ERROR_CODES.CONFLICT, message);
  }
}

export class UnsupportedMediaTypeError extends AppError {
  constructor(message = 'Unsupported media type') {
    super(415, ERROR_CODES.UNSUPPORTED_MEDIA_TYPE, message);
  }
}

export class PayloadTooLargeError extends AppError {
  constructor(message = 'Payload too large') {
    super(413, ERROR_CODES.PAYLOAD_TOO_LARGE, message);
  }
}

export class ServiceUnavailableError extends AppError {
  constructor(message = 'Upstream service unavailable') {
    super(503, ERROR_CODES.SERVICE_UNAVAILABLE, message);
  }
}

export class InternalError extends AppError {
  constructor(message = 'Internal server error') {
    super(500, ERROR_CODES.INTERNAL_ERROR, message, undefined, false);
  }
}
