import { normalize } from '../../src/modules/declarations/normalization.service';
import type { ExtractedField } from '../../src/modules/ml/ml.types';

function field(name: string, rawValue: string): ExtractedField {
  return { field: name, rawValue, confidence: 0.9 };
}

describe('declaration normalization', () => {
  it('parses MRP into a numeric value', () => {
    const n = normalize(field('MRP', 'M.R.P. Rs. 199.00'));
    expect(n.type).toBe('MRP');
    expect(n.numericValue).toBe(199);
    expect(n.normalizedValue).toBe('199.00');
  });

  it('parses net quantity value and unit', () => {
    const n = normalize(field('NET_QUANTITY', 'Net Qty: 500 g'));
    expect(n.type).toBe('NET_QUANTITY');
    expect(n.numericValue).toBe(500);
    expect(n.unit).toBe('g');
  });

  it('maps unknown fields to OTHER and preserves the raw string', () => {
    const n = normalize(field('SOMETHING_ELSE', 'hello world'));
    expect(n.type).toBe('OTHER');
    expect(n.normalizedValue).toBe('hello world');
  });
});
