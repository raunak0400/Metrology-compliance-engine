import { Router } from 'express';
import { validate } from '../../middleware/validation.middleware';
import { authRateLimiter } from '../../middleware/rate-limit.middleware';
import * as controller from './auth.controller';
import { registerSchema, loginSchema, refreshSchema, logoutSchema } from './auth.validation';

const router = Router();

router.use(authRateLimiter);

router.post('/register', validate({ body: registerSchema }), controller.register);
router.post('/login', validate({ body: loginSchema }), controller.login);
router.post('/refresh', validate({ body: refreshSchema }), controller.refresh);
router.post('/logout', validate({ body: logoutSchema }), controller.logout);

export default router;
