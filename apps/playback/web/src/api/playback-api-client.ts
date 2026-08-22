import {
  episodePath,
  GetEpisodeResponseSchema,
  ListEpisodesResponseSchema,
  listEpisodesPath,
} from "../../../contracts/index.ts";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { ApiResult } from "./api-result.ts";
import { requestBlob, requestJson } from "./playback-api-response.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export type PlaybackApiClientDeps = {
  baseUrl: string;
  fetch: FetchLike;
};

export type PlaybackApiClient = {
  listEpisodes(): Promise<ApiResult<ListEpisodesResponse>>;
  getEpisode(episodeId: string): Promise<ApiResult<GetEpisodeResponse>>;
  fetchAudio(audioRef: string): Promise<ApiResult<Blob>>;
};

/**
 * baseUrl と契約 path を 1 本の request URL へ繋ぐ。
 *
 * @require path は契約由来であり `/` から始まる
 * @ensure baseUrl と path の間の `/` は 1 つだけになる
 * @invariant path 側は書き換えない
 */
export function buildRequestUrl(baseUrl: string, path: string): string {
  // why: path が必ず `/` 始まりなので、baseUrl 末尾の `/` を落とす 1 規則で足りる。URL class は
  //   baseUrl 側の path 段を捨てるため使わない
  return `${baseUrl.replace(/\/+$/, "")}${path}`;
}

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
    async fetchAudio(audioRef: string): Promise<ApiResult<Blob>> {
      // why: audioRef は GetEpisodeResponse が持つ契約 path なので、web 側で episodeId から
      //   組み直さずそのまま繋ぐ
      return requestBlob(fetch, buildRequestUrl(baseUrl, audioRef));
    },
  };
}
