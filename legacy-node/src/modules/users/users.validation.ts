import { z } from 'zod';

export const idParamSchema = z.object({ id: z.string().uuid() });

export const listUsersQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional(),
  limit: z.coerce.number().int().positive().max(100).optional(),
  isActive: z.enum(['true', 'false']).optional(),
  search: z.string().max(120).optional(),
});

export const updateUserSchema = z
  .object({
    fullName: z.string().min(2).max(120).optional(),
    phone: z.string().min(6).max(20).optional(),
    isActive: z.boolean().optional(),
    roleIds: z.array(z.string().uuid()).optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: 'At least one field must be provided' });
