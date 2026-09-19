import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import * as controller from './inspections.controller';
import {
  createInspectionSchema,
  updateInspectionSchema,
  listInspectionsQuerySchema,
  idParamSchema,
} from './inspections.validation';

const router = Router();

router.use(authenticate);

router.post('/', requirePermission(PERMISSIONS.INSPECTIONS_WRITE), validate({ body: createInspectionSchema }), controller.create);
router.get('/', requirePermission(PERMISSIONS.INSPECTIONS_READ), validate({ query: listInspectionsQuerySchema }), controller.list);
router.get('/:id', requirePermission(PERMISSIONS.INSPECTIONS_READ), validate({ params: idParamSchema }), controller.getById);
router.patch(
  '/:id',
  requirePermission(PERMISSIONS.INSPECTIONS_WRITE),
  validate({ params: idParamSchema, body: updateInspectionSchema }),
  controller.update,
);

export default router;
