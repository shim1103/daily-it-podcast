import type { ZodType } from "zod";
import type { ApiResult } from "./api-result.ts";
import { mapHttpStatusToApiError } from "./playback-api-error.ts";

type ResponseLike = {
  ok: boolean;
  status: number;
  json(): Promise<unknown>;
};

/**
 * response 取得 → status → schema 検証を ApiResult へ落とす（API Client 3段責務の中核）。
 *
 * @require getResponse は throw しうる（network failure）。schema は成功 body の契約
 * @ensure reject・非成功 response・body/schema failure を throw せず ApiResult へ変換する
 * @invariant 非成功 response の body は読まない
 */
export async function readJsonResult<T>(
  getResponse: () => Promise<ResponseLike>,
  schema: ZodType<T>,
): Promise<ApiResult<T>> {
  let response: ResponseLike;
  try {
    response = await getResponse();
  } catch {
    return { ok: false, error: "network_error" };
  }

  if (!response.ok) {
    return { ok: false, error: mapHttpStatusToApiError(response.status) };
  }

  try {
    return { ok: true, data: schema.parse(await response.json()) };
  } catch {
    return { ok: false, error: "invalid_response" };
  }
}
