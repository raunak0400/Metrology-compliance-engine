import { ServiceUnavailableError } from '../errors';

export interface HttpJsonOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE';
  headers?: Record<string, string>;
  body?: unknown;
  timeoutMs?: number;
}

/**
 * Minimal JSON HTTP client built on the native `fetch`, with a timeout. Used to
 * talk to the internal Python ML service. Throws ServiceUnavailableError on
 * network/timeouts so callers get a consistent 503.
 */
export async function httpJson<T>(url: string, opts: HttpJsonOptions = {}): Promise<T> {
  const { method = 'GET', headers = {}, body, timeoutMs = 30_000 } = opts;
  const controller = new AbortController();
  const timer = setTimeout(() => controller.abort(), timeoutMs);
  try {
    const res = await fetch(url, {
      method,
      headers: { 'content-type': 'application/json', ...headers },
      body: body === undefined ? undefined : JSON.stringify(body),
      signal: controller.signal,
    });
    if (!res.ok) {
      const text = await res.text().catch(() => '');
      throw new ServiceUnavailableError(`Upstream ${url} returned ${res.status}: ${text.slice(0, 200)}`);
    }
    return (await res.json()) as T;
  } catch (err) {
    if (err instanceof ServiceUnavailableError) throw err;
    throw new ServiceUnavailableError(`Request to ${url} failed: ${(err as Error).message}`);
  } finally {
    clearTimeout(timer);
  }
}
