/**
 * S3-compatible object storage (MinIO in dev, AWS S3 in prod).
 *
 * Binary files (images, cropped panels, evidence, generated reports, dataset
 * artifacts) live here. PostgreSQL stores only the storage KEY, never the bytes.
 * Downloads are served through short-lived signed URLs.
 */
import {
  S3Client,
  PutObjectCommand,
  GetObjectCommand,
  DeleteObjectCommand,
  HeadBucketCommand,
  CreateBucketCommand,
} from '@aws-sdk/client-s3';
import { getSignedUrl } from '@aws-sdk/s3-request-presigner';
import { randomUUID } from 'node:crypto';
import { env } from './env';
import { logger } from './logger';

export const s3 = new S3Client({
  region: env.S3_REGION,
  endpoint: env.S3_ENDPOINT,
  forcePathStyle: env.S3_FORCE_PATH_STYLE,
  credentials: {
    accessKeyId: env.S3_ACCESS_KEY,
    secretAccessKey: env.S3_SECRET_KEY,
  },
});

/** Build a structured, collision-resistant object key. */
export function buildObjectKey(prefix: string, filename: string): string {
  const safe = filename.replace(/[^a-zA-Z0-9._-]/g, '_');
  const date = new Date().toISOString().slice(0, 10);
  return `${prefix}/${date}/${randomUUID()}-${safe}`;
}

export interface UploadParams {
  key: string;
  body: Buffer | Uint8Array | string;
  contentType: string;
  metadata?: Record<string, string>;
}

export async function putObject(params: UploadParams): Promise<{ key: string }> {
  await s3.send(
    new PutObjectCommand({
      Bucket: env.S3_BUCKET,
      Key: params.key,
      Body: params.body,
      ContentType: params.contentType,
      Metadata: params.metadata,
    }),
  );
  return { key: params.key };
}

export async function deleteObject(key: string): Promise<void> {
  await s3.send(new DeleteObjectCommand({ Bucket: env.S3_BUCKET, Key: key }));
}

/** Time-limited signed URL for downloading a private object. */
export async function getSignedDownloadUrl(key: string, ttlSeconds = env.S3_SIGNED_URL_TTL): Promise<string> {
  return getSignedUrl(s3, new GetObjectCommand({ Bucket: env.S3_BUCKET, Key: key }), {
    expiresIn: ttlSeconds,
  });
}

/** Time-limited signed URL for a direct client -> storage upload. */
export async function getSignedUploadUrl(
  key: string,
  contentType: string,
  ttlSeconds = env.S3_SIGNED_URL_TTL,
): Promise<string> {
  return getSignedUrl(
    s3,
    new PutObjectCommand({ Bucket: env.S3_BUCKET, Key: key, ContentType: contentType }),
    { expiresIn: ttlSeconds },
  );
}

/** Ensure the configured bucket exists (used on startup in dev). */
export async function ensureBucket(): Promise<void> {
  try {
    await s3.send(new HeadBucketCommand({ Bucket: env.S3_BUCKET }));
  } catch {
    await s3.send(new CreateBucketCommand({ Bucket: env.S3_BUCKET }));
    logger.info({ bucket: env.S3_BUCKET }, 'Created storage bucket');
  }
}
