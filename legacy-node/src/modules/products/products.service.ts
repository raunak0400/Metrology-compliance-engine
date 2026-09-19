import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { parsePagination, toPrismaSkipTake, buildPaginated } from '../../common/utils/pagination';

export interface CreateProductInput {
  name: string;
  sku?: string;
  brand?: string;
  barcode?: string;
  categoryId: string;
  manufacturerId?: string;
  netQuantityDeclared?: string;
  mrpDeclared?: number;
  attributes?: Record<string, unknown>;
}

const productInclude = {
  category: { select: { id: true, name: true, code: true } },
  manufacturer: { select: { id: true, name: true, country: true } },
  media: true,
} as const;

export function createProduct(input: CreateProductInput, createdById?: string) {
  return prisma.product.create({
    data: {
      name: input.name,
      sku: input.sku,
      brand: input.brand,
      barcode: input.barcode,
      categoryId: input.categoryId,
      manufacturerId: input.manufacturerId,
      netQuantityDeclared: input.netQuantityDeclared,
      mrpDeclared: input.mrpDeclared,
      attributes: input.attributes as never,
      createdById,
    },
    include: productInclude,
  });
}

export async function listProducts(query: Record<string, unknown>) {
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);

  const where = {
    ...(query.categoryId ? { categoryId: String(query.categoryId) } : {}),
    ...(query.manufacturerId ? { manufacturerId: String(query.manufacturerId) } : {}),
    ...(query.search
      ? {
          OR: [
            { name: { contains: String(query.search), mode: 'insensitive' as const } },
            { brand: { contains: String(query.search), mode: 'insensitive' as const } },
            { sku: { contains: String(query.search), mode: 'insensitive' as const } },
          ],
        }
      : {}),
  };

  const [items, total] = await Promise.all([
    prisma.product.findMany({ where, include: productInclude, skip, take, orderBy: { createdAt: 'desc' } }),
    prisma.product.count({ where }),
  ]);

  return buildPaginated(items, total, pagination);
}

export async function getProduct(id: string) {
  const product = await prisma.product.findUnique({ where: { id }, include: productInclude });
  if (!product) throw new NotFoundError('Product');
  return product;
}

export async function updateProduct(id: string, data: Partial<CreateProductInput>) {
  await getProduct(id);
  return prisma.product.update({
    where: { id },
    data: { ...data, attributes: data.attributes as never },
    include: productInclude,
  });
}
