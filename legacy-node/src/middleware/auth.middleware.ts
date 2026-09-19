/**
 * Authentication + RBAC middleware.
 *
 * `authenticate` verifies the JWT access token and attaches `req.user`.
 * Authorization is expressed via reusable policy middleware (`authorize`,
 * `requirePermission`) — controllers must NOT hard-code role checks inline.
 */
import type { Request, Response, NextFunction } from 'express';
import { ForbiddenError, UnauthorizedError } from '../common/errors';
import { verifyAccessToken } from '../modules/auth/token.service';
import type { AuthUser } from '../common/types';

function extractBearer(req: Request): string | null {
  const header = req.headers.authorization;
  if (!header || !header.startsWith('Bearer ')) return null;
  return header.slice('Bearer '.length).trim() || null;
}

/** Require a valid access token; populates req.user. */
export function authenticate(req: Request, _res: Response, next: NextFunction): void {
  const token = extractBearer(req);
  if (!token) return next(new UnauthorizedError('Missing bearer token'));
  try {
    const payload = verifyAccessToken(token);
    const user: AuthUser = {
      id: payload.sub,
      email: payload.email,
      roles: payload.roles ?? [],
      permissions: payload.permissions ?? [],
    };
    req.user = user;
    next();
  } catch {
    next(new UnauthorizedError('Invalid or expired token'));
  }
}

/** Attach req.user if a valid token is present, but do not require it. */
export function optionalAuth(req: Request, _res: Response, next: NextFunction): void {
  const token = extractBearer(req);
  if (!token) return next();
  try {
    const payload = verifyAccessToken(token);
    req.user = {
      id: payload.sub,
      email: payload.email,
      roles: payload.roles ?? [],
      permissions: payload.permissions ?? [],
    };
  } catch {
    /* ignore invalid token in optional mode */
  }
  next();
}

/** Policy: user must hold at least one of the given roles (ADMIN always passes). */
export function authorize(...roles: string[]) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    if (!req.user) return next(new UnauthorizedError());
    if (req.user.roles.includes('ADMIN')) return next();
    const ok = roles.some((r) => req.user!.roles.includes(r));
    if (!ok) return next(new ForbiddenError(`Requires one of roles: ${roles.join(', ')}`));
    next();
  };
}

/** Policy: user must hold all listed permission scopes (`*` grants everything). */
export function requirePermission(...permissions: string[]) {
  return (req: Request, _res: Response, next: NextFunction): void => {
    if (!req.user) return next(new UnauthorizedError());
    const held = req.user.permissions;
    if (held.includes('*') || req.user.roles.includes('ADMIN')) return next();
    const missing = permissions.filter((p) => !held.includes(p));
    if (missing.length > 0) return next(new ForbiddenError(`Missing permission(s): ${missing.join(', ')}`));
    next();
  };
}
