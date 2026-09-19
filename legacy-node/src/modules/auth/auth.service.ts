/**
 * Authentication service: registration, login, refresh-token rotation, logout.
 * Refresh tokens are persisted as SHA-256 hashes with a jti so they can be
 * revoked and rotated. Access tokens remain stateless.
 */
import jwt from 'jsonwebtoken';
import { prisma } from '../../config/database';
import { ConflictError, UnauthorizedError } from '../../common/errors';
import { sha256, randomUUID } from '../../common/utils/crypto';
import { ROLES } from '../../common/constants';
import { hashPassword, verifyPassword } from './password.service';
import {
  signAccessToken,
  signRefreshToken,
  verifyRefreshToken,
  type AccessTokenPayload,
} from './token.service';

interface UserWithRoles {
  id: string;
  email: string;
  fullName: string;
  passwordHash: string;
  isActive: boolean;
  roles: { role: { name: string; permissions: string[] } }[];
}

function buildPayload(user: UserWithRoles): AccessTokenPayload {
  const roles = user.roles.map((r) => r.role.name);
  const permissions = Array.from(new Set(user.roles.flatMap((r) => r.role.permissions)));
  return { sub: user.id, email: user.email, roles, permissions };
}

async function issueTokens(user: UserWithRoles, ctx: { userAgent?: string; ip?: string }) {
  const payload = buildPayload(user);
  const accessToken = signAccessToken(payload);

  const jti = randomUUID();
  const refreshToken = signRefreshToken({ sub: user.id, jti });
  const decoded = jwt.decode(refreshToken) as { exp?: number } | null;
  const expiresAt = new Date((decoded?.exp ?? Math.floor(Date.now() / 1000) + 604800) * 1000);

  await prisma.refreshToken.create({
    data: {
      userId: user.id,
      jti,
      tokenHash: sha256(refreshToken),
      expiresAt,
      userAgent: ctx.userAgent,
      ip: ctx.ip,
    },
  });

  return { accessToken, refreshToken, roles: payload.roles, permissions: payload.permissions };
}

const userInclude = { roles: { include: { role: true } } } as const;

export async function register(input: {
  email: string;
  password: string;
  fullName: string;
  phone?: string;
}) {
  const existing = await prisma.user.findUnique({ where: { email: input.email } });
  if (existing) throw new ConflictError('A user with this email already exists');

  const passwordHash = await hashPassword(input.password);

  // New users default to the INSPECTOR role. Admins can adjust roles afterwards.
  const inspectorRole = await prisma.role.findUnique({ where: { name: ROLES.INSPECTOR } });

  const user = await prisma.user.create({
    data: {
      email: input.email,
      passwordHash,
      fullName: input.fullName,
      phone: input.phone,
      roles: inspectorRole ? { create: { roleId: inspectorRole.id } } : undefined,
    },
    include: userInclude,
  });

  return { id: user.id, email: user.email, fullName: user.fullName };
}

export async function login(
  email: string,
  password: string,
  ctx: { userAgent?: string; ip?: string },
) {
  const user = await prisma.user.findUnique({ where: { email }, include: userInclude });
  if (!user || !user.isActive) throw new UnauthorizedError('Invalid credentials');

  const ok = await verifyPassword(user.passwordHash, password);
  if (!ok) throw new UnauthorizedError('Invalid credentials');

  await prisma.user.update({ where: { id: user.id }, data: { lastLoginAt: new Date() } });
  const tokens = await issueTokens(user as UserWithRoles, ctx);

  return {
    user: { id: user.id, email: user.email, fullName: user.fullName, roles: tokens.roles },
    ...tokens,
  };
}

export async function refresh(token: string, ctx: { userAgent?: string; ip?: string }) {
  let payload;
  try {
    payload = verifyRefreshToken(token);
  } catch {
    throw new UnauthorizedError('Invalid refresh token');
  }

  const stored = await prisma.refreshToken.findUnique({ where: { jti: payload.jti } });
  if (!stored || stored.revokedAt || stored.expiresAt < new Date()) {
    throw new UnauthorizedError('Refresh token is no longer valid');
  }
  if (stored.tokenHash !== sha256(token)) {
    // Token reuse / mismatch — revoke defensively.
    await prisma.refreshToken.update({ where: { jti: payload.jti }, data: { revokedAt: new Date() } });
    throw new UnauthorizedError('Refresh token mismatch');
  }

  const user = await prisma.user.findUnique({ where: { id: payload.sub }, include: userInclude });
  if (!user || !user.isActive) throw new UnauthorizedError('User no longer active');

  // Rotate: revoke old, issue new.
  await prisma.refreshToken.update({ where: { jti: payload.jti }, data: { revokedAt: new Date() } });
  return issueTokens(user as UserWithRoles, ctx);
}

export async function logout(token: string): Promise<void> {
  try {
    const payload = verifyRefreshToken(token);
    await prisma.refreshToken.updateMany({
      where: { jti: payload.jti, revokedAt: null },
      data: { revokedAt: new Date() },
    });
  } catch {
    // Idempotent logout: silently ignore invalid tokens.
  }
}
