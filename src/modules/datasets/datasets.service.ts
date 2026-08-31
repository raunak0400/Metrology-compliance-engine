/**
 * Dataset management for the product-label ML dataset. The backend does NOT train
 * models — it provides dataset + annotation storage/APIs so the Python ML team
 * can consume them. Binary images live in object storage; only keys are stored.
 */
import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { parsePagination, toPrismaSkipTake, buildPaginated } from '../../common/utils/pagination';

function slugify(name: string): string {
  return name
    .toLowerCase()
    .trim()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/(^-|-$)/g, '');
}

export function createDataset(
  data: { name: string; description?: string; purpose?: string },
  createdById?: string,
) {
  return prisma.dataset.create({
    data: {
      name: data.name,
      slug: slugify(data.name),
      description: data.description,
      purpose: data.purpose,
      createdById,
    },
  });
}

export async function listDatasets(query: Record<string, unknown>) {
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);
  const [items, total] = await Promise.all([
    prisma.dataset.findMany({
      skip,
      take,
      orderBy: { createdAt: 'desc' },
      include: { _count: { select: { items: true, versions: true } } },
    }),
    prisma.dataset.count(),
  ]);
  return buildPaginated(items, total, pagination);
}

async function ensureDataset(id: string) {
  const dataset = await prisma.dataset.findUnique({ where: { id } });
  if (!dataset) throw new NotFoundError('Dataset');
  return dataset;
}

export async function addItem(
  datasetId: string,
  data: {
    storageKey: string;
    sourceImageId?: string;
    mimeType?: string;
    split?: string;
    metadata?: Record<string, unknown>;
  },
) {
  await ensureDataset(datasetId);
  return prisma.datasetItem.create({
    data: {
      datasetId,
      storageKey: data.storageKey,
      sourceImageId: data.sourceImageId,
      mimeType: data.mimeType,
      split: (data.split as never) ?? 'UNASSIGNED',
      metadata: data.metadata as never,
    },
  });
}

export async function listItems(datasetId: string, query: Record<string, unknown>) {
  await ensureDataset(datasetId);
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);

  const where = {
    datasetId,
    ...(query.split ? { split: String(query.split) as never } : {}),
    ...(query.annotationStatus ? { annotationStatus: String(query.annotationStatus) as never } : {}),
  };

  const [items, total] = await Promise.all([
    prisma.datasetItem.findMany({
      where,
      skip,
      take,
      orderBy: { createdAt: 'desc' },
      include: { _count: { select: { annotations: true } } },
    }),
    prisma.datasetItem.count({ where }),
  ]);
  return buildPaginated(items, total, pagination);
}

export async function addAnnotation(
  itemId: string,
  data: {
    fieldLabel?: string;
    classificationLabel?: string;
    bbox?: Record<string, unknown>;
    value?: string;
    confidence?: number;
    status?: string;
    metadata?: Record<string, unknown>;
  },
  annotatorId?: string,
) {
  const item = await prisma.datasetItem.findUnique({ where: { id: itemId } });
  if (!item) throw new NotFoundError('Dataset item');

  return prisma.$transaction(async (tx) => {
    const annotation = await tx.datasetAnnotation.create({
      data: {
        itemId,
        annotatorId,
        fieldLabel: data.fieldLabel,
        classificationLabel: data.classificationLabel,
        bbox: data.bbox as never,
        value: data.value,
        confidence: data.confidence,
        status: (data.status as never) ?? 'ANNOTATED',
        metadata: data.metadata as never,
      },
    });
    // Reflect progress on the item.
    await tx.datasetItem.update({ where: { id: itemId }, data: { annotationStatus: 'ANNOTATED' } });
    return annotation;
  });
}
