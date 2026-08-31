import { Router } from 'express';
import { authenticate, authorize } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { ROLES } from '../../common/constants';
import * as controller from './users.controller';
import { idParamSchema, listUsersQuerySchema, updateUserSchema } from './users.validation';

const router = Router();

router.use(authenticate);

router.get('/', authorize(ROLES.ADMIN, ROLES.SUPERVISOR), validate({ query: listUsersQuerySchema }), controller.list);
router.get('/:id', validate({ params: idParamSchema }), controller.getById);
router.patch(
  '/:id',
  authorize(ROLES.ADMIN),
  validate({ params: idParamSchema, body: updateUserSchema }),
  controller.update,
);

export default router;
