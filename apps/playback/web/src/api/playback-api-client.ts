import {
  episodePath,
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../../../contracts/index.ts";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import { buildRequestUrl } from "../utils/build-request-url.ts";
import type { ApiResult } from "./api-result.ts";
import { requestJson } from "./playback-api-response.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export type PlaybackApiClientDeps = {
  baseUrl: string;
  fetch: FetchLike;
};

export type PlaybackApiClient = {
  listEpisodes(): Promise<ApiResult<ListEpisodesResponse>>;
  getEpisode(episodeId: string): Promise<ApiResult<GetEpisodeResponse>>;
};

/**
 * playback worker の HTTP API を叩く client を組み立てる。
 *
 * @require deps.baseUrl は worker の origin。末尾の `/` は有無どちらでもよい
 * @require deps.fetch は Fetch API 互換の呼び出し
 * @ensure 戻り値の各 method は throw せず ApiResult を返す
 * @invariant baseUrl は組み立て時に 1 度だけ受け取り、各 method の引数にしない
 */
export function createPlaybackApiClient(deps: PlaybackApiClientDeps): PlaybackApiClient {
  const { baseUrl, fetch } = deps;

  return {
    async listEpisodes(): Promise<ApiResult<ListEpisodesResponse>> {
      return requestJson(
        fetch,
        buildRequestUrl(baseUrl, listEpisodesPath),
        ListEpisodesResponseSchema,
      );
    },
    async getEpisode(episodeId: string): Promise<ApiResult<GetEpisodeResponse>> {
      return requestJson(
        fetch,
        buildRequestUrl(baseUrl, episodePath(episodeId)),
        GetEpisodeResponseSchema,
      );
    },
  };
}
