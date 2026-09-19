import { Router } from 'express';
import { z } from 'zod';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { uploadSingleImage } from '../../middleware/upload.middleware';
import { PERMISSIONS } from '../../common/constants';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import { idParamSchema } from '../inspections/inspections.validation';
import * as evidenceService from './evidence.service';
import { recordAudit } from '../audit/audit.service';

// Mounted at /inspections — owns /:id/evidence.
const router = Router();
router.use(authenticate);

const addEvidenceSchema = z.object({
  type: z
    .enum(['ORIGINAL_IMAGE', 'CROPPED_PANEL', 'ANNOTATED_IMAGE', 'DOCUMENT', 'OTHER'])
    .optional(),
  violationId: z.string().uuid().optional(),
  imageId: z.string().uuid().optional(),
  bbox: z.record(z.string(), z.unknown()).optional(),
  caption: z.string().max(500).optional(),
  metadata: z.record(z.string(), z.unknown()).optional(),
});

router.post(
  '/:id/evidence',
  requirePermission(PERMISSIONS.INSPECTIONS_WRITE),
  validate({ params: idParamSchema }),
  // Optional file upload on field `file`; JSON-only references are also allowed.
  uploadSingleImage('file'),
  validate({ body: addEvidenceSchema }),
  asyncHandler(async (req, res) => {
    const file = req.file
      ? {
          buffer: req.file.buffer,
          originalname: req.file.originalname,
          mimetype: req.file.mimetype,
          size: req.file.size,
        }
      : undefined;
    const evidence = await evidenceService.addEvidence(req.params.id, req.body, file);
    await recordAudit({ action: 'CREATE', entityType: 'Evidence', entityId: evidence.id, req });
    sendCreated(res, evidence);
  }),
);

router.get(
  '/:id/evidence',
  requirePermission(PERMISSIONS.INSPECTIONS_READ),
  validate({ params: idParamSchema }),
  asyncHandler(async (req, res) => {
    sendSuccess(res, await evidenceService.listByInspection(req.params.id));
  }),
);

export default router;
