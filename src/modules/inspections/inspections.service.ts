import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { generateCode } from '../../common/utils/crypto';
import { parsePagination, toPrismaSkipTake, buildPaginated } from '../../common/utils/pagination';

export interface CreateInspectionInput {
  productId?: string;
  categoryId?: string;
  source?: string;
  notes?: string;
  metadata?: Record<string, unknown>;
}

export async function createInspection(input: CreateInspectionInput, inspectorId: string) {
  // If a product is supplied and no category given, inherit the product category.
  let categoryId = input.categoryId;
  if (!categoryId && input.productId) {
    const product = await prisma.product.findUnique({ where: { id: input.productId } });
    categoryId = product?.categoryId;
  }

  return prisma.inspection.create({
    data: {
      code: generateCode('INSP'),
      productId: input.productId,
      categoryId,
      inspectorId,
      source: input.source,
      notes: input.notes,
      metadata: input.metadata as never,
      status: 'DRAFT',
      outcome: 'PENDING',
    },
  });
}

export async function listInspections(query: Record<string, unknown>) {
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);

  const where = {
    ...(query.status ? { status: String(query.status) as never } : {}),
    ...(query.outcome ? { outcome: String(query.outcome) as never } : {}),
    ...(query.inspectorId ? { inspectorId: String(query.inspectorId) } : {}),
    ...(query.productId ? { productId: String(query.productId) } : {}),
  };

  const [items, total] = await Promise.all([
    prisma.inspection.findMany({
      where,
      skip,
      take,
      orderBy: { createdAt: 'desc' },
      include: {
        product: { select: { id: true, name: true } },
        _count: { select: { images: true, violations: true, declarations: true } },
      },
    }),
    prisma.inspection.count({ where }),
  ]);

  return buildPaginated(items, total, pagination);
}

export async function getInspection(id: string) {
  const inspection = await prisma.inspection.findUnique({
    where: { id },
    include: {
      product: { select: { id: true, name: true, category: { select: { id: true, name: true } } } },
      inspector: { select: { id: true, fullName: true, email: true } },
      reviewer: { select: { id: true, fullName: true, email: true } },
      images: true,
      _count: { select: { declarations: true, checks: true, violations: true, reports: true } },
    },
  });
  if (!inspection) throw new NotFoundError('Inspection');
  return inspection;
}

export async function updateInspection(
  id: string,
  data: {
    status?: string;
    outcome?: string;
    reviewerId?: string;
    notes?: string;
    metadata?: Record<string, unknown>;
  },
) {
  await ensureInspection(id);
  return prisma.inspection.update({
    where: { id },
    data: {
      status: data.status as never,
      outcome: data.outcome as never,
      reviewerId: data.reviewerId,
      notes: data.notes,
      metadata: data.metadata as never,
    },
  });
}

/** Throws NotFoundError if the inspection does not exist; returns it otherwise. */
export async function ensureInspection(id: string) {
  const inspection = await prisma.inspection.findUnique({ where: { id } });
  if (!inspection) throw new NotFoundError('Inspection');
  return inspection;
}
