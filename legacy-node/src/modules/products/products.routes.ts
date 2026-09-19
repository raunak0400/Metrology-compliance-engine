import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import * as controller from './products.controller';
import {
  createProductSchema,
  updateProductSchema,
  listProductsQuerySchema,
  idParamSchema,
} from './products.validation';

const router = Router();

router.use(authenticate);

router.post('/', requirePermission(PERMISSIONS.PRODUCTS_WRITE), validate({ body: createProductSchema }), controller.create);
router.get('/', requirePermission(PERMISSIONS.PRODUCTS_READ), validate({ query: listProductsQuerySchema }), controller.list);
router.get('/:id', requirePermission(PERMISSIONS.PRODUCTS_READ), validate({ params: idParamSchema }), controller.getById);
router.patch(
  '/:id',
  requirePermission(PERMISSIONS.PRODUCTS_WRITE),
  validate({ params: idParamSchema, body: updateProductSchema }),
  controller.update,
);

export default router;
