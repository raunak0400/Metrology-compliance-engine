/**
 * File upload handling with Multer (in-memory), enforcing MIME allow-list and
 * size limits. The buffer is handed to the storage service; nothing is written
 * to local disk.
 */
import type { Request, Response, NextFunction } from 'express';
import multer, { MulterError } from 'multer';
import { env } from '../config/env';
import { PayloadTooLargeError, UnsupportedMediaTypeError } from '../common/errors';

const storage = multer.memoryStorage();

const uploader = multer({
  storage,
  limits: { fileSize: env.MAX_UPLOAD_SIZE_BYTES, files: 1 },
  fileFilter: (_req, file, cb) => {
    if (env.ALLOWED_IMAGE_MIME.includes(file.mimetype)) {
      cb(null, true);
    } else {
      cb(new UnsupportedMediaTypeError(`Unsupported image type: ${file.mimetype}`));
    }
  },
});

/** Single-image upload on field `image`, with error translation. */
export function uploadSingleImage(fieldName = 'image') {
  const handler = uploader.single(fieldName);
  return (req: Request, res: Response, next: NextFunction): void => {
    handler(req, res, (err: unknown) => {
      if (err instanceof MulterError) {
        if (err.code === 'LIMIT_FILE_SIZE') return next(new PayloadTooLargeError('Image exceeds size limit'));
        return next(new UnsupportedMediaTypeError(err.message));
      }
      if (err) return next(err);
      next();
    });
  };
}
