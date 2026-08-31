import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import * as controller from './rules.controller';
import {
  createRuleSchema,
  updateRuleSchema,
  listRulesQuerySchema,
  idParamSchema,
} from './rules.validation';

const router = Router();
router.use(authenticate);

router.get('/', requirePermission(PERMISSIONS.RULES_READ), validate({ query: listRulesQuerySchema }), controller.list);
router.get('/:id', requirePermission(PERMISSIONS.RULES_READ), validate({ params: idParamSchema }), controller.getById);
router.post('/', requirePermission(PERMISSIONS.RULES_WRITE), validate({ body: createRuleSchema }), controller.create);
router.patch(
  '/:id',
  requirePermission(PERMISSIONS.RULES_WRITE),
  validate({ params: idParamSchema, body: updateRuleSchema }),
  controller.update,
);

export default router;
