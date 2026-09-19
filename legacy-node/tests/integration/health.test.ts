/**
 * Integration test skeleton.
 *
 * This boots the real Express app (which opens Redis/DB clients), so it is
 * OPT-IN: set RUN_INTEGRATION=1 with the local stack (docker compose up) running.
 * By default it is skipped so `npm test` stays green without infrastructure.
 */
import request from 'supertest';

const run = process.env.RUN_INTEGRATION === '1';

(run ? describe : describe.skip)('GET /health', () => {
  it('returns ok', async () => {
    // Imported lazily so the heavy app graph only loads when integration runs.
    const { createApp } = await import('../../src/app');
    const app = createApp();
    const res = await request(app).get('/health');
    expect(res.status).toBe(200);
    expect(res.body.success).toBe(true);
    expect(res.body.data.status).toBe('ok');
  });
});
