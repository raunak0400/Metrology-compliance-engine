/**
 * Request validation using Zod. Validate body, query and/or params against a
 * schema; on success the parsed (coerced) values replace the originals.
 */
import type { Request, Response, NextFunction } from 'express';
import { ZodError, type ZodTypeAny } from 'zod';
import { AppError } from '../common/errors';
import { ERROR_CODES } from '../common/constants';

export interface ValidationSchemas {
  body?: ZodTypeAny;
  query?: ZodTypeAny;
  params?: ZodTypeAny;
}

export function validate(schemas: ValidationSchemas) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    try {
      if (schemas.params) req.params = schemas.params.parse(req.params) as typeof req.params;
      if (schemas.query) {
        // req.query has an inherited getter in Express 4; assign parsed onto it.
        Object.assign(req.query, schemas.query.parse(req.query));
      }
      if (schemas.body) req.body = schemas.body.parse(req.body);
      next();
    } catch (err) {
      if (err instanceof ZodError) {
        const details = err.issues.map((i) => ({
          field: i.path.join('.'),
          message: i.message,
          code: i.code,
        }));
        return next(new AppError(422, ERROR_CODES.VALIDATION_ERROR, 'Validation failed', details));
      }
      next(err);
    }
  };
}
