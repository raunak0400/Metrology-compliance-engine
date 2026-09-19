import type { Request, Response } from 'express';
import { asyncHandler, sendSuccess } from '../../common/utils';
import * as declarationsService from './declarations.service';
import { recordAudit } from '../audit/audit.service';

export const listByInspection = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await declarationsService.listByInspection(req.params.id));
});

export const update = asyncHandler(async (req: Request, res: Response) => {
  const before = await declarationsService.getDeclaration(req.params.id);
  const declaration = await declarationsService.updateDeclaration(req.params.id, req.body, req.user?.id);
  await recordAudit({
    action: 'REVIEW',
    entityType: 'ExtractedDeclaration',
    entityId: req.params.id,
    req,
    oldState: before,
    newState: declaration,
  });
  sendSuccess(res, declaration);
});
