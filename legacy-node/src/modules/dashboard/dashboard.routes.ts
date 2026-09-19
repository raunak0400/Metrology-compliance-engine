import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { PERMISSIONS } from '../../common/constants';
import { asyncHandler, sendSuccess } from '../../common/utils';
import * as dashboardService from './dashboard.service';

const router = Router();
router.use(authenticate, requirePermission(PERMISSIONS.DASHBOARD_READ));

router.get(
  '/summary',
  asyncHandler(async (_req, res) => sendSuccess(res, await dashboardService.getSummary())),
);

router.get(
  '/violations',
  asyncHandler(async (_req, res) => sendSuccess(res, await dashboardService.getViolationStats())),
);

router.get(
  '/inspections',
  asyncHandler(async (_req, res) => sendSuccess(res, await dashboardService.getInspectionStats())),
);

export default router;
