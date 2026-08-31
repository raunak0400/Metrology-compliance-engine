import { evaluate } from '../../src/modules/compliance/rule-engine/engine';
import type {
  EngineRule,
  EngineDeclaration,
  EngineContext,
} from '../../src/modules/compliance/rule-engine/types';

const mrpRule: EngineRule = {
  id: 'rule-mrp',
  code: 'MRP-PRESENCE',
  title: 'MRP present',
  actReference: 'PCR 2011 Rule 6',
  severityDefault: 'HIGH',
  requirements: [
    {
      id: 'req-mrp-exists',
      key: 'declaration.MRP.present',
      declarationType: 'MRP',
      operator: 'EXISTS',
      expected: null,
      unit: null,
      severity: null,
      message: 'MRP missing',
      isMandatory: true,
    },
  ],
};

function ctx(declarations: EngineDeclaration[]): EngineContext {
  return {
    declarations,
    measurements: [],
    isImported: false,
    manualReviewThreshold: 0.7,
  };
}

describe('rule engine', () => {
  it('returns COMPLIANT when a mandatory declaration is present with high confidence', () => {
    const out = evaluate([mrpRule], ctx([
      { declarationId: 'd1', type: 'MRP', normalizedValue: '199.00', numericValue: 199, unit: 'INR', confidence: 0.95 },
    ]));
    expect(out.outcome).toBe('COMPLIANT');
    expect(out.checks[0].status).toBe('PASS');
  });

  it('returns NON_COMPLIANT when a mandatory declaration is absent', () => {
    const out = evaluate([mrpRule], ctx([]));
    expect(out.outcome).toBe('NON_COMPLIANT');
    expect(out.checks[0].status).toBe('FAIL');
  });

  it('routes low-confidence failing checks to MANUAL_REVIEW instead of a violation', () => {
    const regexRule: EngineRule = {
      ...mrpRule,
      requirements: [
        {
          id: 'req-mrp-format',
          key: 'declaration.MRP.format',
          declarationType: 'MRP',
          operator: 'REGEX_MATCH',
          expected: { pattern: '^\\d+\\.\\d{2}$' },
          unit: null,
          severity: null,
          message: 'bad format',
          isMandatory: true,
        },
      ],
    };
    const out = evaluate([regexRule], ctx([
      { declarationId: 'd1', type: 'MRP', normalizedValue: 'abc', numericValue: null, unit: null, confidence: 0.4 },
    ]));
    expect(out.outcome).toBe('MANUAL_REVIEW');
    expect(out.checks[0].status).toBe('MANUAL_REVIEW');
    expect(out.checks[0].requiresManualReview).toBe(true);
  });

  it('routes MIN_TEXT_HEIGHT_MM to MANUAL_REVIEW when no valid scale exists', () => {
    const heightRule: EngineRule = {
      ...mrpRule,
      requirements: [
        {
          id: 'req-height',
          key: 'measurement.minHeight',
          declarationType: 'MRP',
          operator: 'MIN_TEXT_HEIGHT_MM',
          expected: { minMm: 1 },
          unit: 'mm',
          severity: null,
          message: 'too small',
          isMandatory: true,
        },
      ],
    };
    const out = evaluate([heightRule], {
      declarations: [
        { declarationId: 'd1', type: 'MRP', normalizedValue: '199', numericValue: 199, unit: 'INR', confidence: 0.99 },
      ],
      measurements: [{ declarationId: 'd1', type: 'MRP', heightMm: null, hasValidScale: false }],
      isImported: false,
      manualReviewThreshold: 0.7,
    });
    expect(out.outcome).toBe('MANUAL_REVIEW');
  });
});
