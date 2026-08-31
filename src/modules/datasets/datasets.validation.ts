import { z } from 'zod';

export const datasetIdParam = z.object({ id: z.string().uuid() });
export const itemIdParam = z.object({ id: z.string().uuid() });

export const createDatasetSchema = z.object({
  name: z.string().min(2).max(120),
  description: z.string().max(1000).optional(),
  purpose: z.string().max(120).optional(),
});

export const addItemSchema = z.object({
  storageKey: z.string().min(1).max(500),
  sourceImageId: z.string().uuid().optional(),
  mimeType: z.string().max(80).optional(),
  split: z.enum(['UNASSIGNED', 'TRAIN', 'VALIDATION', 'TEST']).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

export const addAnnotationSchema = z
  .object({
    fieldLabel: z.string().max(80).optional(),
    classificationLabel: z.string().max(80).optional(),
    bbox: z.record(z.string(), z.unknown()).optional(),
    value: z.string().max(1000).optional(),
    confidence: z.number().min(0).max(1).optional(),
    status: z.enum(['PENDING', 'IN_PROGRESS', 'ANNOTATED', 'REVIEWED', 'REJECTED']).optional(),
    metadata: z.record(z.string(), z.unknown()).optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: 'Annotation payload cannot be empty' });

export const listQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional(),
  limit: z.coerce.number().int().positive().max(100).optional(),
  split: z.enum(['UNASSIGNED', 'TRAIN', 'VALIDATION', 'TEST']).optional(),
  annotationStatus: z.enum(['PENDING', 'IN_PROGRESS', 'ANNOTATED', 'REVIEWED', 'REJECTED']).optional(),
});
