import type { ZodType } from "zod";
import type { ApiResult } from "./api-result.ts";
import { mapHttpStatusToApiError } from "./playback-api-error.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

async function request<T>(
  fetch: FetchLike,
  url: string,
  readSuccessBody: (response: Response) => Promise<T>,
): Promise<ApiResult<T>> {
  let response: Response;
  try {
    response = await fetch(url);
  } catch {
    return { ok: false, error: "network_error" };
  }

  if (!response.ok) {
    return { ok: false, error: mapHttpStatusToApiError(response.status) };
  }

  try {
    return { ok: true, data: await readSuccessBody(response) };
  } catch {
    return { ok: false, error: "invalid_response" };
  }
}

/**
 * JSON response を成功 body として読み、契約 schema を検証する。
 *
 * @require fetch は Fetch API 互換、schema は期待する response の schema
 * @ensure fetch reject・非成功 response・body/schema failure を throw せず ApiResult へ変換する
 * @invariant 非成功 response の body は読まない
 */
export function requestJson<T>(
  fetch: FetchLike,
  url: string,
  schema: ZodType<T>,
): Promise<ApiResult<T>> {
  return request(fetch, url, async (response) => schema.parse(await response.json()));
}
