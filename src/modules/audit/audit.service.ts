/**
 * Audit logging. Important mutations must call `recordAudit` to persist who did
 * what, when, from where, and the before/after state. Failures to write an audit
 * record are logged but never break the primary request.
 */
import type { Request } from 'express';
import { prisma } from '../../config/database';
import { logger } from '../../config/logger';
import { parsePagination, toPrismaSkipTake, buildPaginated } from '../../common/utils/pagination';

export type AuditActionName =
  | 'CREATE'
  | 'UPDATE'
  | 'DELETE'
  | 'LOGIN'
  | 'LOGOUT'
  | 'PROCESS'
  | 'REVIEW'
  | 'FINALIZE'
  | 'EXPORT';

export interface RecordAuditParams {
  action: AuditActionName;
  entityType: string;
  entityId?: string;
  req?: Request;
  actorId?: string;
  oldState?: unknown;
  newState?: unknown;
}

export async function recordAudit(params: RecordAuditParams): Promise<void> {
  try {
    await prisma.auditLog.create({
      data: {
        action: params.action as never,
        entityType: params.entityType,
        entityId: params.entityId,
        actorId: params.actorId ?? params.req?.user?.id,
        oldState: params.oldState === undefined ? undefined : (params.oldState as never),
        newState: params.newState === undefined ? undefined : (params.newState as never),
        ip: params.req?.ip,
        userAgent: params.req?.headers['user-agent'],
        requestId: params.req?.id !== undefined ? String(params.req.id) : undefined,
      },
    });
  } catch (err) {
    logger.error({ err, action: params.action, entityType: params.entityType }, 'Failed to write audit log');
  }
}

export async function listAuditLogs(query: Record<string, unknown>) {
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);

  const where = {
    ...(query.entityType ? { entityType: String(query.entityType) } : {}),
    ...(query.entityId ? { entityId: String(query.entityId) } : {}),
    ...(query.actorId ? { actorId: String(query.actorId) } : {}),
  };

  const [items, total] = await Promise.all([
    prisma.auditLog.findMany({ where, skip, take, orderBy: { createdAt: 'desc' } }),
    prisma.auditLog.count({ where }),
  ]);

  return buildPaginated(items, total, pagination);
}
