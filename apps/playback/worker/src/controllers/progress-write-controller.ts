import type { ProgressWriteRequest, ProgressWriteResponse } from "../../../contracts/index.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type ProgressWriteUseCase = (
  episodeId: string,
  body: ProgressWriteRequest,
) => Promise<ProgressWriteResponse>;

export type ProgressWriteController = (
  episodeId: string,
  body: ProgressWriteRequest,
) => Promise<ProgressWriteResponse>;

/**
 * 進捗 Write（create / update / complete）用 Controller を組み立てる。
 * 操作の違いは渡す use case が担う。本 factory は写像だけを共有する。
 *
 * @require episodeId / body は呼び出し側（Hono route の zValidator）が契約 schema で検証済み
 * @ensure 戻り関数は検証済み入力で ProgressWriteResponse を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createProgressWriteController(
  useCase: ProgressWriteUseCase,
): ProgressWriteController {
  return async function progressWriteController(
    episodeId: string,
    body: ProgressWriteRequest,
  ): Promise<ProgressWriteResponse> {
    try {
      return await useCase(episodeId, body);
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
