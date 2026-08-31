import { Router } from 'express';
import { z } from 'zod';
import { authenticate, authorize } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { ROLES } from '../../common/constants';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import * as rolesService from './roles.service';
import { recordAudit } from '../audit/audit.service';

const router = Router();
router.use(authenticate);

const createRoleSchema = z.object({
  name: z.string().min(2).max(40),
  description: z.string().max(200).optional(),
  permissions: z.array(z.string()).optional(),
});

router.get(
  '/',
  asyncHandler(async (_req, res) => {
    sendSuccess(res, await rolesService.listRoles());
  }),
);

router.post(
  '/',
  authorize(ROLES.ADMIN),
  validate({ body: createRoleSchema }),
  asyncHandler(async (req, res) => {
    const role = await rolesService.createRole(req.body);
    await recordAudit({ action: 'CREATE', entityType: 'Role', entityId: role.id, req, newState: role });
    sendCreated(res, role);
  }),
);

export default router;
