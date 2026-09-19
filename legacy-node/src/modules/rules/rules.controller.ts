import type { Request, Response } from 'express';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import * as rulesService from './rules.service';
import { recordAudit } from '../audit/audit.service';

export const list = asyncHandler(async (req: Request, res: Response) => {
  const result = await rulesService.listRules(req.query as Record<string, unknown>);
  sendSuccess(res, result.items, 200, {
    page: result.page,
    limit: result.limit,
    total: result.total,
    totalPages: result.totalPages,
  });
});

export const getById = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await rulesService.getRule(req.params.id));
});

export const create = asyncHandler(async (req: Request, res: Response) => {
  const rule = await rulesService.createRule(req.body, req.user?.id);
  await recordAudit({ action: 'CREATE', entityType: 'LegalRule', entityId: rule.id, req, newState: rule });
  sendCreated(res, rule);
});

export const update = asyncHandler(async (req: Request, res: Response) => {
  const rule = await rulesService.updateRule(req.params.id, req.body);
  await recordAudit({ action: 'UPDATE', entityType: 'LegalRule', entityId: rule.id, req, newState: rule });
  sendSuccess(res, rule);
});
