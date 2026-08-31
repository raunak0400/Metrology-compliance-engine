import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { parsePagination, toPrismaSkipTake, buildPaginated } from '../../common/utils/pagination';

export interface RequirementInput {
  key: string;
  declarationType?: string;
  operator: string;
  expected?: Record<string, unknown>;
  unit?: string;
  severity?: string;
  message?: string;
  isMandatory?: boolean;
  orderIndex?: number;
}

export interface CreateRuleInput {
  code: string;
  title: string;
  description?: string;
  actReference?: string;
  section?: string;
  severityDefault?: string;
  effectiveFrom?: string;
  effectiveTo?: string;
  isActive?: boolean;
  params?: Record<string, unknown>;
  categoryIds?: string[];
  requirements?: RequirementInput[];
}

const ruleInclude = {
  requirements: { orderBy: { orderIndex: 'asc' as const } },
  categories: { include: { category: { select: { id: true, code: true, name: true } } } },
} as const;

export async function listRules(query: Record<string, unknown>) {
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);

  const where = {
    ...(query.isActive !== undefined ? { isActive: query.isActive === 'true' } : {}),
    ...(query.search
      ? {
          OR: [
            { code: { contains: String(query.search), mode: 'insensitive' as const } },
            { title: { contains: String(query.search), mode: 'insensitive' as const } },
          ],
        }
      : {}),
  };

  const [items, total] = await Promise.all([
    prisma.legalRule.findMany({
      where,
      skip,
      take,
      orderBy: { code: 'asc' },
      include: { _count: { select: { requirements: true } } },
    }),
    prisma.legalRule.count({ where }),
  ]);

  return buildPaginated(items, total, pagination);
}

export async function getRule(id: string) {
  const rule = await prisma.legalRule.findUnique({ where: { id }, include: ruleInclude });
  if (!rule) throw new NotFoundError('Rule');
  return rule;
}

export function createRule(input: CreateRuleInput, createdById?: string) {
  return prisma.legalRule.create({
    data: {
      code: input.code,
      title: input.title,
      description: input.description,
      actReference: input.actReference,
      section: input.section,
      severityDefault: (input.severityDefault as never) ?? undefined,
      effectiveFrom: input.effectiveFrom ? new Date(input.effectiveFrom) : undefined,
      effectiveTo: input.effectiveTo ? new Date(input.effectiveTo) : undefined,
      isActive: input.isActive ?? true,
      params: input.params as never,
      createdById,
      requirements: input.requirements
        ? {
            create: input.requirements.map((r, i) => ({
              key: r.key,
              declarationType: (r.declarationType as never) ?? undefined,
              operator: r.operator as never,
              expected: r.expected as never,
              unit: r.unit,
              severity: (r.severity as never) ?? undefined,
              message: r.message,
              isMandatory: r.isMandatory ?? true,
              orderIndex: r.orderIndex ?? i,
            })),
          }
        : undefined,
      categories: input.categoryIds
        ? { create: input.categoryIds.map((categoryId) => ({ categoryId })) }
        : undefined,
    },
    include: ruleInclude,
  });
}

export async function updateRule(id: string, input: Partial<CreateRuleInput>) {
  await getRule(id);

  return prisma.$transaction(async (tx) => {
    if (input.requirements) {
      await tx.ruleRequirement.deleteMany({ where: { ruleId: id } });
      await tx.ruleRequirement.createMany({
        data: input.requirements.map((r, i) => ({
          ruleId: id,
          key: r.key,
          declarationType: (r.declarationType as never) ?? null,
          operator: r.operator as never,
          expected: (r.expected as never) ?? undefined,
          unit: r.unit ?? null,
          severity: (r.severity as never) ?? null,
          message: r.message ?? null,
          isMandatory: r.isMandatory ?? true,
          orderIndex: r.orderIndex ?? i,
        })),
      });
    }

    if (input.categoryIds) {
      await tx.ruleCategory.deleteMany({ where: { ruleId: id } });
      await tx.ruleCategory.createMany({
        data: input.categoryIds.map((categoryId) => ({ ruleId: id, categoryId })),
        skipDuplicates: true,
      });
    }

    return tx.legalRule.update({
      where: { id },
      data: {
        title: input.title,
        description: input.description,
        actReference: input.actReference,
        section: input.section,
        severityDefault: (input.severityDefault as never) ?? undefined,
        effectiveFrom: input.effectiveFrom ? new Date(input.effectiveFrom) : undefined,
        effectiveTo: input.effectiveTo ? new Date(input.effectiveTo) : undefined,
        isActive: input.isActive,
        params: input.params as never,
      },
      include: ruleInclude,
    });
  });
}
