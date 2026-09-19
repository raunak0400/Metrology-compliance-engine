import type { Request, Response } from 'express';
import { asyncHandler, sendSuccess } from '../../common/utils';
import * as usersService from './users.service';
import { recordAudit } from '../audit/audit.service';

export const list = asyncHandler(async (req: Request, res: Response) => {
  const result = await usersService.listUsers(req.query as Record<string, unknown>);
  sendSuccess(res, result.items, 200, {
    page: result.page,
    limit: result.limit,
    total: result.total,
    totalPages: result.totalPages,
  });
});

export const getById = asyncHandler(async (req: Request, res: Response) => {
  const user = await usersService.getUser(req.params.id);
  sendSuccess(res, user);
});

export const update = asyncHandler(async (req: Request, res: Response) => {
  const before = await usersService.getUser(req.params.id);
  const user = await usersService.updateUser(req.params.id, req.body);
  await recordAudit({
    action: 'UPDATE',
    entityType: 'User',
    entityId: req.params.id,
    req,
    oldState: before,
    newState: user,
  });
  sendSuccess(res, user);
});
