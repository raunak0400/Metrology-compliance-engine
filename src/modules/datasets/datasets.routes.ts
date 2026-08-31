import { Router } from 'express';
import { authenticate, requirePermission } from '../../middleware/auth.middleware';
import { validate } from '../../middleware/validation.middleware';
import { PERMISSIONS } from '../../common/constants';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import * as datasetsService from './datasets.service';
import { recordAudit } from '../audit/audit.service';
import {
  createDatasetSchema,
  addItemSchema,
  addAnnotationSchema,
  listQuerySchema,
  datasetIdParam,
  itemIdParam,
} from './datasets.validation';

// Mounted at the API root: owns /datasets, /datasets/:id/items and
// /dataset-items/:id/annotations.
const router = Router();
router.use(authenticate);

router.post(
  '/datasets',
  requirePermission(PERMISSIONS.DATASETS_WRITE),
  validate({ body: createDatasetSchema }),
  asyncHandler(async (req, res) => {
    const dataset = await datasetsService.createDataset(req.body, req.user?.id);
    await recordAudit({ action: 'CREATE', entityType: 'Dataset', entityId: dataset.id, req });
    sendCreated(res, dataset);
  }),
);

router.get(
  '/datasets',
  validate({ query: listQuerySchema }),
  asyncHandler(async (req, res) => {
    const result = await datasetsService.listDatasets(req.query as Record<string, unknown>);
    sendSuccess(res, result.items, 200, {
      page: result.page,
      limit: result.limit,
      total: result.total,
      totalPages: result.totalPages,
    });
  }),
);

router.post(
  '/datasets/:id/items',
  requirePermission(PERMISSIONS.DATASETS_WRITE),
  validate({ params: datasetIdParam, body: addItemSchema }),
  asyncHandler(async (req, res) => {
    const item = await datasetsService.addItem(req.params.id, req.body);
    sendCreated(res, item);
  }),
);

router.get(
  '/datasets/:id/items',
  validate({ params: datasetIdParam, query: listQuerySchema }),
  asyncHandler(async (req, res) => {
    const result = await datasetsService.listItems(req.params.id, req.query as Record<string, unknown>);
    sendSuccess(res, result.items, 200, {
      page: result.page,
      limit: result.limit,
      total: result.total,
      totalPages: result.totalPages,
    });
  }),
);

router.post(
  '/dataset-items/:id/annotations',
  requirePermission(PERMISSIONS.DATASETS_WRITE),
  validate({ params: itemIdParam, body: addAnnotationSchema }),
  asyncHandler(async (req, res) => {
    const annotation = await datasetsService.addAnnotation(req.params.id, req.body, req.user?.id);
    sendCreated(res, annotation);
  }),
);

export default router;
