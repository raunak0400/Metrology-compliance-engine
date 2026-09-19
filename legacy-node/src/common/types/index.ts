/** Shared application types and the Express Request augmentation. */

/** Authenticated principal attached to the request by auth.middleware. */
export interface AuthUser {
  id: string;
  email: string;
  roles: string[];
  permissions: string[];
}

/** Standard success envelope. */
export interface ApiSuccess<T> {
  success: true;
  data: T;
  meta?: Record<string, unknown>;
}

/** Standard error envelope (see error.middleware). */
export interface ApiError {
  success: false;
  error: {
    code: string;
    message: string;
    details?: unknown[];
    requestId?: string;
  };
}

export interface PaginationQuery {
  page: number;
  limit: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

export interface Paginated<T> {
  items: T[];
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

// ---- Express request augmentation -------------------------------------------
declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace Express {
    interface Request {
      // `id` is provided by pino-http (type ReqId = string | number).
      user?: AuthUser;
    }
  }
}

export {};
