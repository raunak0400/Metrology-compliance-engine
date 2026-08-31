import type { Request, Response } from 'express';
import { asyncHandler, sendCreated, sendSuccess } from '../../common/utils';
import * as authService from './auth.service';
import { recordAudit } from '../audit/audit.service';

function ctxOf(req: Request) {
  return { userAgent: req.headers['user-agent'], ip: req.ip };
}

export const register = asyncHandler(async (req: Request, res: Response) => {
  const user = await authService.register(req.body);
  await recordAudit({ action: 'CREATE', entityType: 'User', entityId: user.id, req, newState: user });
  sendCreated(res, user);
});

export const login = asyncHandler(async (req: Request, res: Response) => {
  const result = await authService.login(req.body.email, req.body.password, ctxOf(req));
  await recordAudit({ action: 'LOGIN', entityType: 'User', entityId: result.user.id, req });
  sendSuccess(res, result);
});

export const refresh = asyncHandler(async (req: Request, res: Response) => {
  const result = await authService.refresh(req.body.refreshToken, ctxOf(req));
  sendSuccess(res, result);
});

export const logout = asyncHandler(async (req: Request, res: Response) => {
  await authService.logout(req.body.refreshToken);
  sendSuccess(res, { message: 'Logged out' });
});
