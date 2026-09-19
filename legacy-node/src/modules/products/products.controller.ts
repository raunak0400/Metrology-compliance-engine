import type { Request, Response } from 'express';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import * as productsService from './products.service';
import { recordAudit } from '../audit/audit.service';

export const create = asyncHandler(async (req: Request, res: Response) => {
  const product = await productsService.createProduct(req.body, req.user?.id);
  await recordAudit({ action: 'CREATE', entityType: 'Product', entityId: product.id, req, newState: product });
  sendCreated(res, product);
});

export const list = asyncHandler(async (req: Request, res: Response) => {
  const result = await productsService.listProducts(req.query as Record<string, unknown>);
  sendSuccess(res, result.items, 200, {
    page: result.page,
    limit: result.limit,
    total: result.total,
    totalPages: result.totalPages,
  });
});

export const getById = asyncHandler(async (req: Request, res: Response) => {
  sendSuccess(res, await productsService.getProduct(req.params.id));
});

export const update = asyncHandler(async (req: Request, res: Response) => {
  const product = await productsService.updateProduct(req.params.id, req.body);
  await recordAudit({ action: 'UPDATE', entityType: 'Product', entityId: product.id, req, newState: product });
  sendSuccess(res, product);
});
