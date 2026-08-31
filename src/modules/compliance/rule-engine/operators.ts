/**
 * Deterministic operator implementations for the rule engine. Each operator is a
 * pure function returning a normalized result. Operators never throw for normal
 * failing input — they return { pass: false, ... }.
 */
import type { EngineDeclaration, EngineMeasurement, EngineContext, OperatorName } from './types';

export interface OperatorResult {
  pass: boolean;
  actual: string | null;
  expected: string | null;
  reason: string | null;
}

function ok(actual: string | null): OperatorResult {
  return { pass: true, actual, expected: null, reason: null };
}
function fail(actual: string | null, expected: string | null, reason: string): OperatorResult {
  return { pass: false, actual, expected, reason };
}

export interface OperatorInput {
  declaration: EngineDeclaration | undefined;
  measurement: EngineMeasurement | undefined;
  expected: Record<string, unknown> | null;
  context: EngineContext;
}

export function applyOperator(operator: OperatorName, input: OperatorInput): OperatorResult {
  const { declaration, measurement, expected, context } = input;
  const value = declaration?.normalizedValue ?? null;

  switch (operator) {
    case 'EXISTS':
      return declaration ? ok(value) : fail(null, 'present', 'Declaration is absent');

    case 'NOT_EMPTY':
      return value && value.trim().length > 0
        ? ok(value)
        : fail(value, 'non-empty', 'Declaration is empty');

    case 'REGEX_MATCH': {
      const pattern = String(expected?.pattern ?? '');
      if (!pattern) return fail(value, null, 'No regex pattern configured');
      if (value === null) return fail(null, pattern, 'No value to match');
      const re = new RegExp(pattern);
      return re.test(value) ? ok(value) : fail(value, pattern, 'Value does not match required pattern');
    }

    case 'NUMERIC_MIN': {
      const min = Number(expected?.min);
      const n = declaration?.numericValue;
      if (n === null || n === undefined) return fail(value, `>= ${min}`, 'No numeric value');
      return n >= min ? ok(String(n)) : fail(String(n), `>= ${min}`, `Value below minimum ${min}`);
    }

    case 'NUMERIC_MAX': {
      const max = Number(expected?.max);
      const n = declaration?.numericValue;
      if (n === null || n === undefined) return fail(value, `<= ${max}`, 'No numeric value');
      return n <= max ? ok(String(n)) : fail(String(n), `<= ${max}`, `Value above maximum ${max}`);
    }

    case 'NUMERIC_RANGE': {
      const min = Number(expected?.min);
      const max = Number(expected?.max);
      const n = declaration?.numericValue;
      if (n === null || n === undefined) return fail(value, `${min}..${max}`, 'No numeric value');
      return n >= min && n <= max
        ? ok(String(n))
        : fail(String(n), `${min}..${max}`, `Value outside range ${min}..${max}`);
    }

    case 'MIN_TEXT_HEIGHT_MM': {
      const minMm = Number(expected?.minMm);
      // Measurement is separate from OCR — the engine only consumes its result.
      if (!measurement || !measurement.hasValidScale || measurement.heightMm === null) {
        return {
          pass: false,
          actual: measurement?.heightMm != null ? `${measurement.heightMm}mm` : 'unknown',
          expected: `>= ${minMm}mm`,
          // Not a definite failure: without a valid scale we cannot assert mm.
          reason: 'MANUAL_REVIEW:No valid scale reference for pixel-to-mm conversion',
        };
      }
      return measurement.heightMm >= minMm
        ? ok(`${measurement.heightMm}mm`)
        : fail(`${measurement.heightMm}mm`, `>= ${minMm}mm`, `Text height below minimum ${minMm}mm`);
    }

    case 'ENUM_IN': {
      const allowed = (expected?.values as string[]) ?? [];
      if (value === null) return fail(null, allowed.join('|'), 'No value');
      return allowed.map((s) => s.toLowerCase()).includes(value.toLowerCase())
        ? ok(value)
        : fail(value, allowed.join('|'), 'Value not in allowed set');
    }

    case 'DATE_VALID': {
      if (value === null) return fail(null, 'valid date', 'No date value');
      const d = new Date(value);
      return Number.isNaN(d.getTime()) ? fail(value, 'valid date', 'Not a valid date') : ok(value);
    }

    case 'DATE_NOT_FUTURE': {
      if (value === null) return fail(null, '<= today', 'No date value');
      const d = new Date(value);
      if (Number.isNaN(d.getTime())) return fail(value, '<= today', 'Not a valid date');
      return d.getTime() <= Date.now() ? ok(value) : fail(value, '<= today', 'Date is in the future');
    }

    case 'REQUIRED_IF_IMPORTED': {
      if (!context.isImported) return ok(value); // requirement not applicable
      return declaration && value && value.trim().length > 0
        ? ok(value)
        : fail(value, 'present (imported package)', 'Required declaration missing for imported package');
    }

    case 'CUSTOM':
    default:
      // Unknown/custom operators are not auto-decided — route to manual review.
      return {
        pass: false,
        actual: value,
        expected: null,
        reason: 'MANUAL_REVIEW:Operator requires manual evaluation',
      };
  }
}
