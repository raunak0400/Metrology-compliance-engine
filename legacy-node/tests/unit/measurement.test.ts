import { estimateTextHeightMm, resolvePxPerMm } from '../../src/modules/compliance/measurement.service';

describe('measurement service', () => {
  it('returns hasValidScale=false and null mm when no scale reference exists', () => {
    const r = estimateTextHeightMm({ x: 0, y: 0, width: 100, height: 30 }, null);
    expect(r.hasValidScale).toBe(false);
    expect(r.heightMm).toBeNull();
  });

  it('converts px to mm using a direct pxPerMm scale', () => {
    const r = estimateTextHeightMm({ x: 0, y: 0, width: 100, height: 30 }, { pxPerMm: 10 });
    expect(r.hasValidScale).toBe(true);
    expect(r.heightMm).toBe(3);
  });

  it('derives pxPerMm from a reference object', () => {
    expect(resolvePxPerMm({ referenceObjectMm: 10, referenceObjectPx: 100 })).toBe(10);
    expect(resolvePxPerMm({})).toBeNull();
  });
});
