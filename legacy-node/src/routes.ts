/**
 * API router aggregation. All module routers are mounted here under the API
 * prefix (default /api/v1). Some inspection-scoped resources (declarations,
 * reports, violations, datasets) define their own full paths and are mounted at
 * the router root.
 */
import { Router } from 'express';

import authRoutes from './modules/auth/auth.routes';
import usersRoutes from './modules/users/users.routes';
import rolesRoutes from './modules/roles/roles.routes';
import productsRoutes from './modules/products/products.routes';
import inspectionsRoutes from './modules/inspections/inspections.routes';
import scansRoutes from './modules/scans/scans.routes';
import complianceRoutes from './modules/compliance/compliance.routes';
import evidenceRoutes from './modules/evidence/evidence.routes';
import declarationsRoutes from './modules/declarations/declarations.routes';
import violationsRoutes from './modules/violations/violations.routes';
import reportsRoutes from './modules/reports/reports.routes';
import rulesRoutes from './modules/rules/rules.routes';
import datasetsRoutes from './modules/datasets/datasets.routes';
import dashboardRoutes from './modules/dashboard/dashboard.routes';
import auditRoutes from './modules/audit/audit.routes';

export const apiRouter = Router();

apiRouter.get('/', (_req, res) => {
  res.json({ success: true, data: { name: 'Legal Metrology Compliance Checker API', version: 'v1' } });
});

apiRouter.use('/auth', authRoutes);
apiRouter.use('/users', usersRoutes);
apiRouter.use('/roles', rolesRoutes);
apiRouter.use('/products', productsRoutes);
apiRouter.use('/rules', rulesRoutes);
apiRouter.use('/dashboard', dashboardRoutes);
apiRouter.use('/audit-logs', auditRoutes);

// Inspection-scoped sub-resources (mounted at /inspections).
apiRouter.use('/inspections', inspectionsRoutes);
apiRouter.use('/inspections', scansRoutes);
apiRouter.use('/inspections', complianceRoutes);
apiRouter.use('/inspections', evidenceRoutes);

// Routers that own multiple base paths (mounted at the router root).
apiRouter.use('/', declarationsRoutes);
apiRouter.use('/', violationsRoutes);
apiRouter.use('/', reportsRoutes);
apiRouter.use('/', datasetsRoutes);

export default apiRouter;
