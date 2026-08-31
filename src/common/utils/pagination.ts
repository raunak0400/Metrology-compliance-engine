import { PAGINATION } from '../constants';
import type { Paginated, PaginationQuery } from '../types';

/** Normalize raw query params into a safe pagination object. */
export function parsePagination(query: Record<string, unknown>): PaginationQuery {
  const page = Math.max(1, Number(query.page) || PAGINATION.DEFAULT_PAGE);
  const rawLimit = Number(query.limit) || PAGINATION.DEFAULT_LIMIT;
  const limit = Math.min(PAGINATION.MAX_LIMIT, Math.max(1, rawLimit));
  const sortOrder = query.sortOrder === 'asc' ? 'asc' : 'desc';
  const sortBy = typeof query.sortBy === 'string' ? query.sortBy : undefined;
  return { page, limit, sortBy, sortOrder };
}

/** Translate a pagination query into Prisma skip/take. */
export function toPrismaSkipTake(p: PaginationQuery): { skip: number; take: number } {
  return { skip: (p.page - 1) * p.limit, take: p.limit };
}

/** Wrap a page of results into the standard Paginated envelope. */
export function buildPaginated<T>(items: T[], total: number, p: PaginationQuery): Paginated<T> {
  return {
    items,
    page: p.page,
    limit: p.limit,
    total,
    totalPages: Math.max(1, Math.ceil(total / p.limit)),
  };
}
