import { z } from 'zod';

export const idParamSchema = z.object({ id: z.string().uuid() });

export const createProductSchema = z.object({
  name: z.string().min(1).max(200),
  sku: z.string().max(80).optional(),
  brand: z.string().max(120).optional(),
  barcode: z.string().max(64).optional(),
  categoryId: z.string().uuid(),
  manufacturerId: z.string().uuid().optional(),
  netQuantityDeclared: z.string().max(80).optional(),
  mrpDeclared: z.number().nonnegative().optional(),
  attributes: z.record(z.string(), z.unknown()).optional(),
});

export const updateProductSchema = createProductSchema.partial().refine(
  (v) => Object.keys(v).length > 0,
  { message: 'At least one field must be provided' },
);

export const listProductsQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional(),
  limit: z.coerce.number().int().positive().max(100).optional(),
  categoryId: z.string().uuid().optional(),
  manufacturerId: z.string().uuid().optional(),
  search: z.string().max(120).optional(),
});
