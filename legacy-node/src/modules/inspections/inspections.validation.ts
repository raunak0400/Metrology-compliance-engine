import { z } from 'zod';

export const idParamSchema = z.object({ id: z.string().uuid() });

export const createInspectionSchema = z.object({
  productId: z.string().uuid().optional(),
  categoryId: z.string().uuid().optional(),
  source: z.enum(['FIELD', 'MARKETPLACE', 'LAB', 'OTHER']).optional(),
  notes: z.string().max(2000).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

export const updateInspectionSchema = z
  .object({
    status: z
      .enum(['DRAFT', 'QUEUED', 'PROCESSING', 'AWAITING_REVIEW', 'COMPLETED', 'FAILED'])
      .optional(),
    outcome: z.enum(['PENDING', 'COMPLIANT', 'NON_COMPLIANT', 'MANUAL_REVIEW']).optional(),
    reviewerId: z.string().uuid().optional(),
    notes: z.string().max(2000).optional(),
    metadata: z.record(z.string(), z.unknown()).optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: 'At least one field must be provided' });

export const listInspectionsQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional(),
  limit: z.coerce.number().int().positive().max(100).optional(),
  status: z
    .enum(['DRAFT', 'QUEUED', 'PROCESSING', 'AWAITING_REVIEW', 'COMPLETED', 'FAILED'])
    .optional(),
  outcome: z.enum(['PENDING', 'COMPLIANT', 'NON_COMPLIANT', 'MANUAL_REVIEW']).optional(),
  inspectorId: z.string().uuid().optional(),
  productId: z.string().uuid().optional(),
});
