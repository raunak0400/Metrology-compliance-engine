import { z } from 'zod';

const OPERATORS = [
  'EXISTS',
  'NOT_EMPTY',
  'REGEX_MATCH',
  'NUMERIC_MIN',
  'NUMERIC_MAX',
  'NUMERIC_RANGE',
  'MIN_TEXT_HEIGHT_MM',
  'ENUM_IN',
  'DATE_VALID',
  'DATE_NOT_FUTURE',
  'REQUIRED_IF_IMPORTED',
  'CUSTOM',
] as const;

const DECLARATION_TYPES = [
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
] as const;

const SEVERITY = ['LOW', 'MEDIUM', 'HIGH', 'CRITICAL'] as const;

export const idParamSchema = z.object({ id: z.string().uuid() });

const requirementSchema = z.object({
  key: z.string().min(1).max(120),
  declarationType: z.enum(DECLARATION_TYPES).optional(),
  operator: z.enum(OPERATORS),
  expected: z.record(z.string(), z.unknown()).optional(),
  unit: z.string().max(20).optional(),
  severity: z.enum(SEVERITY).optional(),
  message: z.string().max(500).optional(),
  isMandatory: z.boolean().optional(),
  orderIndex: z.number().int().nonnegative().optional(),
});

export const createRuleSchema = z.object({
  code: z.string().min(2).max(80),
  title: z.string().min(2).max(200),
  description: z.string().max(2000).optional(),
  actReference: z.string().max(300).optional(),
  section: z.string().max(80).optional(),
  severityDefault: z.enum(SEVERITY).optional(),
  effectiveFrom: z.string().datetime().optional(),
  effectiveTo: z.string().datetime().optional(),
  isActive: z.boolean().optional(),
  params: z.record(z.string(), z.unknown()).optional(),
  categoryIds: z.array(z.string().uuid()).optional(),
  requirements: z.array(requirementSchema).optional(),
});

export const updateRuleSchema = createRuleSchema
  .partial()
  .refine((v) => Object.keys(v).length > 0, { message: 'At least one field must be provided' });

export const listRulesQuerySchema = z.object({
  page: z.coerce.number().int().positive().optional(),
  limit: z.coerce.number().int().positive().max(100).optional(),
  isActive: z.enum(['true', 'false']).optional(),
  search: z.string().max(120).optional(),
});
