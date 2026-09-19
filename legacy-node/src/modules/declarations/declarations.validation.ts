import { z } from 'zod';

export const inspectionIdParamSchema = z.object({ id: z.string().uuid() });
export const declarationIdParamSchema = z.object({ id: z.string().uuid() });

export const updateDeclarationSchema = z
  .object({
    normalizedValue: z.string().max(1000).optional(),
    numericValue: z.number().optional(),
    unit: z.string().max(20).optional(),
    type: z
      .enum([
        'MRP',
        'NET_QUANTITY',
        'MANUFACTURER',
        'PACKER',
        'IMPORTER',
        'MANUFACTURE_DATE',
        'PACKING_DATE',
        'EXPIRY_DATE',
        'CONSUMER_CARE',
        'COUNTRY_OF_ORIGIN',
        'DIMENSIONS',
        'UNIT_SALE_PRICE',
        'GENERIC_NAME',
        'OTHER',
      ])
      .optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: 'At least one field must be provided' });
