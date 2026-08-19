import type { GetEpisodeResponse } from "../../../contracts/index.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";
import { parseGetEpisodeRequest } from "./parse-get-episode-request.ts";

export type GetEpisodeUseCase = (episodeId: string) => Promise<GetEpisodeResponse>;

export type GetEpisodeController = (body: unknown) => Promise<GetEpisodeResponse>;

/**
 * 1件 JSON を返す Controller を組み立てる。
 *
 * @require useCase は確定した episodeId で GetEpisodeResponse を返す
 * @ensure 戻り関数は unknown を schema 検証し、契約の GetEpisodeResponse を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createGetEpisodeController(useCase: GetEpisodeUseCase): GetEpisodeController {
  return async function getEpisodeController(body: unknown): Promise<GetEpisodeResponse> {
    const request = parseGetEpisodeRequest(body);
    try {
      return await useCase(request.episodeId);
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
