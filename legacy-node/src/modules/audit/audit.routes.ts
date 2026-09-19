import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { PERMISSIONS } from '../../common/constants';
import * as controller from './audit.controller';

const router = Router();

router.get('/', authenticate, requirePermission(PERMISSIONS.AUDIT_READ), controller.list);

export default router;
