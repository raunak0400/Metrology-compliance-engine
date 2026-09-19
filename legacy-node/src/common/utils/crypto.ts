import { createHash, randomUUID } from 'node:crypto';

/** SHA-256 hex digest — used for token hashing (refresh tokens) and checksums. */
export function sha256(input: string | Buffer): string {
  return createHash('sha256').update(input).digest('hex');
}

/** Generate a random human-friendly reference code, e.g. INSP-2026-A1B2C3D4. */
export function generateCode(prefix: string): string {
  const year = new Date().getFullYear();
  const rand = randomUUID().replace(/-/g, '').slice(0, 8).toUpperCase();
  return `${prefix}-${year}-${rand}`;
}

export { randomUUID };
