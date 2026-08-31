import type { Request, Response } from 'express';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import { UnauthorizedError } from '../../common/errors';
import * as inspectionsService from './inspections.service';
import { recordAudit } from '../audit/audit.service';

export const create = asyncHandler(async (req: Request, res: Response) => {
  if (!req.user) throw new UnauthorizedError();
  const inspection = await inspectionsService.createInspection(req.body, req.user.id);
  await recordAudit({
    action: 'CREATE',
    entityType: 'Inspection',
    entityId: inspection.id,
    req,
    newState: inspection,
  });
  sendCreated(res, inspection);
});

export const list = asyncHandler(async (req: Request, res: Response) => {
  const result = await inspectionsService.listInspections(req.query as Record<string, unknown>);
  sendSuccess(res, result.items, 200, {
    page: result.page,
    limit: result.limit,
    total: result.total,
    totalPages: result.totalPages,
  });
});

export const getById = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await inspectionsService.getInspection(req.params.id));
});

export const update = asyncHandler(async (req: Request, res: Response) => {
  const inspection = await inspectionsService.updateInspection(req.params.id, req.body);
  await recordAudit({
    action: 'UPDATE',
    entityType: 'Inspection',
    entityId: inspection.id,
    req,
    newState: inspection,
  });
  sendSuccess(res, inspection);
});
