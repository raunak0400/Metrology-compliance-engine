/**
 * Deterministic normalization of raw ML/OCR field values into structured
 * declarations. This is pure, testable logic — NO ML and NO legal decisions
 * happen here. It converts messy raw strings into { normalizedValue, numericValue,
 * unit } that the deterministic rule engine can consume.
 *
 * The original raw value, confidence and bbox are always preserved by the caller;
 * nothing here discards OCR evidence.
 */
import type { ExtractedField } from '../ml/ml.types';

export type DeclarationTypeName =
  | 'MRP'
  | 'NET_QUANTITY'
  | 'MANUFACTURER'
  | 'PACKER'
  | 'IMPORTER'
  | 'MANUFACTURE_DATE'
  | 'PACKING_DATE'
  | 'EXPIRY_DATE'
  | 'CONSUMER_CARE'
  | 'COUNTRY_OF_ORIGIN'
  | 'DIMENSIONS'
  | 'UNIT_SALE_PRICE'
  | 'GENERIC_NAME'
  | 'OTHER';

export interface NormalizedDeclaration {
  type: DeclarationTypeName;
  normalizedValue: string | null;
  numericValue: number | null;
  unit: string | null;
}

const KNOWN_TYPES: Set<string> = new Set([
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
]);

function coerceType(field: string): DeclarationTypeName {
  const upper = field.trim().toUpperCase().replace(/\s+/g, '_');
  return (KNOWN_TYPES.has(upper) ? upper : 'OTHER') as DeclarationTypeName;
}

/** Extract the first monetary amount from a string, e.g. "MRP Rs. 199.00" -> 199. */
function parseMoney(raw: string): number | null {
  const m = raw.replace(/,/g, '').match(/(\d+(?:\.\d+)?)/);
  return m ? Number(m[1]) : null;
}

/** Extract a quantity + unit, e.g. "Net Qty: 500 g" -> { value: 500, unit: "g" }. */
function parseQuantity(raw: string): { value: number | null; unit: string | null } {
  const m = raw
    .replace(/,/g, '')
    .match(/(\d+(?:\.\d+)?)\s*(kg|g|mg|l|ml|cl|pcs|n|no\.?|units?)?/i);
  if (!m) return { value: null, unit: null };
  return { value: Number(m[1]), unit: m[2] ? m[2].toLowerCase() : null };
}

export function normalize(field: ExtractedField): NormalizedDeclaration {
  const type = coerceType(field.field);
  const raw = (field.rawValue ?? '').trim();

  switch (type) {
    case 'MRP':
    case 'UNIT_SALE_PRICE': {
      const value = parseMoney(raw);
      return { type, normalizedValue: value !== null ? value.toFixed(2) : raw, numericValue: value, unit: 'INR' };
    }
    case 'NET_QUANTITY': {
      const { value, unit } = parseQuantity(raw);
      return {
        type,
        normalizedValue: value !== null ? `${value}${unit ? ' ' + unit : ''}` : raw,
        numericValue: value,
        unit,
      };
    }
    case 'MANUFACTURE_DATE':
    case 'PACKING_DATE':
    case 'EXPIRY_DATE': {
      // Normalize common date-ish tokens; keep raw if unparseable.
      const normalized = raw.replace(/\s+/g, ' ');
      return { type, normalizedValue: normalized || null, numericValue: null, unit: null };
    }
    default:
      return { type, normalizedValue: raw || null, numericValue: null, unit: null };
  }
}
