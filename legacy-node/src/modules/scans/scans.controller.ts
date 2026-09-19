import type { Request, Response } from 'express';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import { BadRequestError } from '../../common/errors';
import * as scansService from './scans.service';
import { recordAudit } from '../audit/audit.service';

export const uploadImage = asyncHandler(async (req: Request, res: Response) => {
  const file = req.file;
  if (!file) throw new BadRequestError('No image file provided (field name: "image")');

  const result = await scansService.addImage(req.params.id, {
    buffer: file.buffer,
    originalname: file.originalname,
    mimetype: file.mimetype,
    size: file.size,
  });

  await recordAudit({
    action: 'CREATE',
    entityType: 'InspectionImage',
    entityId: result.image.id,
    req,
    newState: { storageKey: result.image.storageKey },
  });

  // 202-style semantics: work has been accepted and queued.
  sendCreated(res, {
    imageId: result.image.id,
    storageKey: result.image.storageKey,
    jobId: result.jobId,
    status: 'QUEUED',
  });
});

export const process = asyncHandler(async (req: Request, res: Response) => {
  const result = await scansService.processInspection(req.params.id);
  await recordAudit({ action: 'PROCESS', entityType: 'Inspection', entityId: req.params.id, req });
  sendSuccess(res, result, 202);
});

export const listJobs = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await scansService.listJobs(req.params.id));
});
