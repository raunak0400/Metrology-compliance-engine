import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import { idParamSchema } from '../inspections/inspections.validation';
import * as controller from './compliance.controller';

// Mounted at /inspections — owns /:id/check, /:id/checks, /:id/violations.
const router = Router();
router.use(authenticate);

router.post(
  '/:id/check',
  requirePermission(PERMISSIONS.INSPECTIONS_WRITE),
  validate({ params: idParamSchema }),
  controller.check,
);

router.get(
  '/:id/checks',
  requirePermission(PERMISSIONS.INSPECTIONS_READ),
  validate({ params: idParamSchema }),
  controller.listChecks,
);

router.get(
  '/:id/violations',
  requirePermission(PERMISSIONS.INSPECTIONS_READ),
  validate({ params: idParamSchema }),
  controller.listViolations,
);

export default router;
