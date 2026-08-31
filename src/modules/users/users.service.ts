import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';
import { parsePagination, toPrismaSkipTake, buildPaginated } from '../../common/utils/pagination';

const userSelect = {
  id: true,
  email: true,
  fullName: true,
  phone: true,
  isActive: true,
  lastLoginAt: true,
  createdAt: true,
  updatedAt: true,
  roles: { select: { role: { select: { id: true, name: true } } } },
} as const;

export async function listUsers(query: Record<string, unknown>) {
  const pagination = parsePagination(query);
  const { skip, take } = toPrismaSkipTake(pagination);

  const where = {
    ...(query.isActive !== undefined ? { isActive: query.isActive === 'true' } : {}),
    ...(query.search
      ? {
          OR: [
            { email: { contains: String(query.search), mode: 'insensitive' as const } },
            { fullName: { contains: String(query.search), mode: 'insensitive' as const } },
          ],
        }
      : {}),
  };

  const [items, total] = await Promise.all([
    prisma.user.findMany({ where, select: userSelect, skip, take, orderBy: { createdAt: 'desc' } }),
    prisma.user.count({ where }),
  ]);

  return buildPaginated(items, total, pagination);
}

export async function getUser(id: string) {
  const user = await prisma.user.findUnique({ where: { id }, select: userSelect });
  if (!user) throw new NotFoundError('User');
  return user;
}

export async function updateUser(
  id: string,
  data: { fullName?: string; phone?: string; isActive?: boolean; roleIds?: string[] },
) {
  const existing = await prisma.user.findUnique({ where: { id } });
  if (!existing) throw new NotFoundError('User');

  return prisma.$transaction(async (tx) => {
    if (data.roleIds) {
      await tx.userRole.deleteMany({ where: { userId: id } });
      await tx.userRole.createMany({
        data: data.roleIds.map((roleId) => ({ userId: id, roleId })),
        skipDuplicates: true,
      });
    }
    return tx.user.update({
      where: { id },
      data: {
        fullName: data.fullName,
        phone: data.phone,
        isActive: data.isActive,
      },
      select: userSelect,
    });
  });
}
