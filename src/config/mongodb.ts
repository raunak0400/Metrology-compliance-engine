/**
 * MongoDB (Mongoose) connection.
 *
 * MongoDB stores flexible, schema-loose ML/OCR/CV documents: raw OCR output,
 * text blocks, bounding boxes, confidences, detected regions, preprocessing
 * metadata and model predictions. It is NOT the source of truth for compliance
 * results — those live in PostgreSQL.
 */
import mongoose from 'mongoose';
import { env } from './env';
import { logger } from './logger';

mongoose.set('strictQuery', true);

export async function connectMongo(): Promise<typeof mongoose> {
  const conn = await mongoose.connect(env.MONGODB_URI, {
    serverSelectionTimeoutMS: 10_000,
  });
  logger.info('MongoDB connected');
  return conn;
}

export async function disconnectMongo(): Promise<void> {
  await mongoose.disconnect();
  logger.info('MongoDB disconnected');
}

export { mongoose };
