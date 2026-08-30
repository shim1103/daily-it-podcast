import { ListEpisodesResponseSchema } from "../../../contracts/index.ts";
import type { ListEpisodesResponse } from "../../../contracts/index.ts";
import type { ApiResult } from "./api-result.ts";
import { readJsonResult } from "./playback-api-response.ts";
import { createPlaybackRpcClient } from "./playback-rpc-client.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export type PlaybackApiClientDeps = {
  baseUrl: string;
  fetch: FetchLike;
};

export type PlaybackApiClient = {
  listEpisodes(): Promise<ApiResult<ListEpisodesResponse>>;
};

/**
 * playback worker の HTTP API を叩く client を組み立てる。
 * schema + status + network → ApiResult のみ（path / encode は持たない）。
 *
 * @require deps.baseUrl は worker の origin。末尾の `/` は有無どちらでもよい
 * @require deps.fetch は Fetch API 互換の呼び出し
 * @ensure 戻り値の各 method は throw せず ApiResult を返す
 * @invariant baseUrl は組み立て時に 1 度だけ受け取り、各 method の引数にしない
 */
export function createPlaybackApiClient(deps: PlaybackApiClientDeps): PlaybackApiClient {
  const rpc = createPlaybackRpcClient(deps);

  return {
    async listEpisodes(): Promise<ApiResult<ListEpisodesResponse>> {
      return readJsonResult(() => rpc.listEpisodes(), ListEpisodesResponseSchema);
    },
  };
}
