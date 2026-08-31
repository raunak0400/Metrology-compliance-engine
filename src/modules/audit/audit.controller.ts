import type { Request, Response } from 'express';
import { asyncHandler, sendSuccess } from '../../common/utils';
import { listAuditLogs } from './audit.service';

export const list = asyncHandler(async (req: Request, res: Response) => {
  const result = await listAuditLogs(req.query as Record<string, unknown>);
  sendSuccess(res, result.items, 200, {
    page: result.page,
    limit: result.limit,
    total: result.total,
    totalPages: result.totalPages,
  });
});
