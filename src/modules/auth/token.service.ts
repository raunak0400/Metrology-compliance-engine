/**
 * JWT issuing and verification. Stateless access tokens + rotating refresh
 * tokens (whose hashes are persisted in PostgreSQL — see auth.service).
 */
import jwt, { type SignOptions } from 'jsonwebtoken';
import { env } from '../../config/env';

export interface AccessTokenPayload {
  sub: string; // user id
  email: string;
  roles: string[];
  permissions: string[];
}

export interface RefreshTokenPayload {
  sub: string; // user id
  jti: string; // token id (matched against RefreshToken.jti in DB)
}

export function signAccessToken(payload: AccessTokenPayload): string {
  const opts: SignOptions = { expiresIn: env.JWT_ACCESS_TTL as unknown as number };
  return jwt.sign(payload, env.JWT_ACCESS_SECRET, opts);
}

export function signRefreshToken(payload: RefreshTokenPayload): string {
  const opts: SignOptions = { expiresIn: env.JWT_REFRESH_TTL as unknown as number };
  return jwt.sign(payload, env.JWT_REFRESH_SECRET, opts);
}

export function verifyAccessToken(token: string): AccessTokenPayload {
  return jwt.verify(token, env.JWT_ACCESS_SECRET) as AccessTokenPayload;
}

export function verifyRefreshToken(token: string): RefreshTokenPayload {
  return jwt.verify(token, env.JWT_REFRESH_SECRET) as RefreshTokenPayload;
}
