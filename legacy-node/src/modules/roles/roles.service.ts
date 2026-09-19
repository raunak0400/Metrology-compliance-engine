import { prisma } from '../../config/database';
import { NotFoundError } from '../../common/errors';

export function listRoles() {
  return prisma.role.findMany({ orderBy: { name: 'asc' } });
}

export async function getRole(id: string) {
  const role = await prisma.role.findUnique({ where: { id } });
  if (!role) throw new NotFoundError('Role');
  return role;
}

export function createRole(data: { name: string; description?: string; permissions?: string[] }) {
  return prisma.role.create({
    data: {
      name: data.name.toUpperCase(),
      description: data.description,
      permissions: data.permissions ?? [],
      isSystem: false,
    },
  });
}

export async function updateRole(
  id: string,
  data: { description?: string; permissions?: string[] },
) {
  await getRole(id);
  return prisma.role.update({ where: { id }, data });
}
