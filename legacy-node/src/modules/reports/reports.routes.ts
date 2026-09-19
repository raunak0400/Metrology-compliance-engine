import { Router } from 'express';
import { z } from 'zod';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import { idParamSchema } from '../inspections/inspections.validation';
import * as reportsService from './reports.service';
import { recordAudit } from '../audit/audit.service';

// Mounted at the API root: owns /inspections/:id/reports and /reports/:id.
const router = Router();
router.use(authenticate);

const createReportSchema = z.object({ type: z.enum(['PDF', 'DOCX']).default('PDF') });
const reportIdParam = z.object({ id: z.string().uuid() });

router.post(
  '/inspections/:id/reports',
  requirePermission(PERMISSIONS.REPORTS_READ),
  validate({ params: idParamSchema, body: createReportSchema }),
  asyncHandler(async (req, res) => {
    const result = await reportsService.requestReport(req.params.id, req.body.type, req.user?.id);
    await recordAudit({ action: 'EXPORT', entityType: 'Report', entityId: result.reportId, req });
    sendCreated(res, result);
  }),
);

router.get(
  '/inspections/:id/reports',
  requirePermission(PERMISSIONS.REPORTS_READ),
  validate({ params: idParamSchema }),
  asyncHandler(async (req, res) => {
    sendSuccess(res, await reportsService.listByInspection(req.params.id));
  }),
);

router.get(
  '/reports/:id',
  requirePermission(PERMISSIONS.REPORTS_READ),
  validate({ params: reportIdParam }),
  asyncHandler(async (req, res) => {
    sendSuccess(res, await reportsService.getReport(req.params.id));
  }),
);

export default router;
