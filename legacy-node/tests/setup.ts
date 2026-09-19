/**
 * Jest global setup. Provides safe default environment variables so importing
 * modules that validate env (config/env.ts) does not abort the test process.
 * Real infrastructure is only needed for integration tests (opt-in).
 */
process.env.NODE_ENV = process.env.NODE_ENV ?? 'test';
process.env.DATABASE_URL =
  process.env.DATABASE_URL ?? 'postgresql://postgres:postgres@localhost:5432/legal_metrology_test?schema=public';
process.env.MONGODB_URI = process.env.MONGODB_URI ?? 'mongodb://localhost:27017/legal_metrology_ml_test';
process.env.REDIS_URL = process.env.REDIS_URL ?? 'redis://localhost:6379';
process.env.JWT_ACCESS_SECRET = process.env.JWT_ACCESS_SECRET ?? 'test-access-secret';
process.env.JWT_REFRESH_SECRET = process.env.JWT_REFRESH_SECRET ?? 'test-refresh-secret';
process.env.S3_ENDPOINT = process.env.S3_ENDPOINT ?? 'http://localhost:9000';
process.env.S3_ACCESS_KEY = process.env.S3_ACCESS_KEY ?? 'minioadmin';
process.env.S3_SECRET_KEY = process.env.S3_SECRET_KEY ?? 'minioadmin';
process.env.S3_BUCKET = process.env.S3_BUCKET ?? 'legal-metrology-test';
process.env.ML_SERVICE_URL = process.env.ML_SERVICE_URL ?? 'http://localhost:8000';
