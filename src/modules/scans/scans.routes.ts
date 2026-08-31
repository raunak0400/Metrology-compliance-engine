import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { uploadSingleImage } from '../../middleware/upload.middleware';
import { PERMISSIONS } from '../../common/constants';
import { idParamSchema } from '../inspections/inspections.validation';
import * as controller from './scans.controller';

// Mounted at /inspections — defines the image/process/jobs sub-resources.
const router = Router();

router.use(authenticate);

router.post(
  '/:id/images',
  requirePermission(PERMISSIONS.SCANS_WRITE),
  validate({ params: idParamSchema }),
  uploadSingleImage('image'),
  controller.uploadImage,
);

router.post(
  '/:id/process',
  requirePermission(PERMISSIONS.SCANS_WRITE),
  validate({ params: idParamSchema }),
  controller.process,
);

router.get(
  '/:id/jobs',
  requirePermission(PERMISSIONS.INSPECTIONS_READ),
  validate({ params: idParamSchema }),
  controller.listJobs,
);

export default router;
