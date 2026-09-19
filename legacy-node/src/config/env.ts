/**
 * Environment configuration. Loads `.env`, validates and coerces variables with
 * Zod, and exports a single typed `env` object. Import this instead of reading
 * `process.env` directly anywhere else in the codebase.
 */
import 'dotenv/config';
import { z } from 'zod';

const booleanish = z
  .string()
  .transform((v) => v === 'true' || v === '1')
  .pipe(z.boolean());

const csv = z
  .string()
  .transform((v) => v.split(',').map((s) => s.trim()).filter(Boolean));

const schema = z.object({
  NODE_ENV: z.enum(['development', 'test', 'production']).default('development'),
  PORT: z.coerce.number().int().positive().default(4000),
  LOG_LEVEL: z.enum(['fatal', 'error', 'warn', 'info', 'debug', 'trace']).default('info'),
  API_PREFIX: z.string().default('/api/v1'),

  DATABASE_URL: z.string().url(),
  MONGODB_URI: z.string().min(1),
  REDIS_URL: z.string().min(1),

  JWT_ACCESS_SECRET: z.string().min(8),
  JWT_REFRESH_SECRET: z.string().min(8),
  JWT_ACCESS_TTL: z.string().default('15m'),
  JWT_REFRESH_TTL: z.string().default('7d'),

  S3_ENDPOINT: z.string().url(),
  S3_REGION: z.string().default('us-east-1'),
  S3_ACCESS_KEY: z.string().min(1),
  S3_SECRET_KEY: z.string().min(1),
  S3_BUCKET: z.string().min(1),
  S3_FORCE_PATH_STYLE: booleanish.default('true' as never),
  S3_PUBLIC_URL: z.string().url().optional(),
  S3_SIGNED_URL_TTL: z.coerce.number().int().positive().default(900),

  ML_SERVICE_URL: z.string().url(),
  ML_SERVICE_TIMEOUT_MS: z.coerce.number().int().positive().default(30000),
  ML_SERVICE_API_KEY: z.string().optional(),

  MAX_UPLOAD_SIZE_BYTES: z.coerce.number().int().positive().default(15_728_640),
  ALLOWED_IMAGE_MIME: csv.default('image/jpeg,image/png,image/webp' as never),

  RATE_LIMIT_WINDOW_MS: z.coerce.number().int().positive().default(60_000),
  RATE_LIMIT_MAX: z.coerce.number().int().positive().default(120),

  CORS_ORIGINS: csv.default('http://localhost:3000' as never),
});

const parsed = schema.safeParse(process.env);

if (!parsed.success) {
  // Fail fast with a readable message; do not start with an invalid config.
  const issues = parsed.error.issues
    .map((i) => `  - ${i.path.join('.')}: ${i.message}`)
    .join('\n');
  // eslint-disable-next-line no-console
  console.error(`Invalid environment configuration:\n${issues}`);
  process.exit(1);
}

export const env = parsed.data;
export type Env = typeof env;

export const isProd = env.NODE_ENV === 'production';
export const isTest = env.NODE_ENV === 'test';
export const isDev = env.NODE_ENV === 'development';
