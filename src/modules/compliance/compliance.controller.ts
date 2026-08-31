import type { Request, Response } from 'express';
import { asyncHandler, sendSuccess } from '../../common/utils';
import * as complianceService from './compliance.service';
import * as violationsService from '../violations/violations.service';
import { recordAudit } from '../audit/audit.service';

/** Trigger a deterministic compliance evaluation (async via the worker). */
export const check = asyncHandler(async (req: Request, res: Response) => {
  const result = await complianceService.enqueueCompliance(req.params.id, req.user?.id);
  await recordAudit({ action: 'PROCESS', entityType: 'Inspection', entityId: req.params.id, req });
  sendSuccess(res, { ...result, status: 'QUEUED' }, 202);
});

export const listChecks = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await complianceService.listChecks(req.params.id));
});

export const listViolations = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await violationsService.listByInspection(req.params.id));
});
