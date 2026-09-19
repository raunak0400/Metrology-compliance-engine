import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import * as controller from './declarations.controller';
import {
  inspectionIdParamSchema,
  declarationIdParamSchema,
  updateDeclarationSchema,
} from './declarations.validation';

// Mounted at the API root: owns /inspections/:id/declarations and /declarations/:id.
const router = Router();

router.use(authenticate);

router.get(
  '/inspections/:id/declarations',
  requirePermission(PERMISSIONS.INSPECTIONS_READ),
  validate({ params: inspectionIdParamSchema }),
  controller.listByInspection,
);

router.patch(
  '/declarations/:id',
  requirePermission(PERMISSIONS.DECLARATIONS_WRITE),
  validate({ params: declarationIdParamSchema, body: updateDeclarationSchema }),
  controller.update,
);

export default router;
