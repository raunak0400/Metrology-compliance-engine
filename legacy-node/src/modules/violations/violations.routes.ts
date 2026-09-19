import { Router } from 'express';
import { z } from 'zod';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import { asyncHandler, sendSuccess } from '../../common/utils';
import * as violationsService from './violations.service';
import { recordAudit } from '../audit/audit.service';

// Mounted at the API root: owns /violations/:id.
const router = Router();
router.use(authenticate);

const idParam = z.object({ id: z.string().uuid() });
const updateSchema = z
  .object({
    status: z.enum(['OPEN', 'CONFIRMED', 'DISMISSED', 'RESOLVED']).optional(),
    severity: z.enum(['LOW', 'MEDIUM', 'HIGH', 'CRITICAL']).optional(),
    reviewNotes: z.string().max(2000).optional(),
  })
  .refine((v) => Object.keys(v).length > 0, { message: 'At least one field must be provided' });

router.get(
  '/violations/:id',
  requirePermission(PERMISSIONS.INSPECTIONS_READ),
  validate({ params: idParam }),
  asyncHandler(async (req, res) => {
    sendSuccess(res, await violationsService.getViolation(req.params.id));
  }),
);

router.patch(
  '/violations/:id',
  requirePermission(PERMISSIONS.VIOLATIONS_WRITE),
  validate({ params: idParam, body: updateSchema }),
  asyncHandler(async (req, res) => {
    const violation = await violationsService.updateViolation(req.params.id, req.body, req.user?.id);
    await recordAudit({
      action: 'REVIEW',
      entityType: 'Violation',
      entityId: req.params.id,
      req,
      newState: violation,
    });
    sendSuccess(res, violation);
  }),
);

export default router;
