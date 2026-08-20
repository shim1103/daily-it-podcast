import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import { episodePath, listEpisodesPath } from "../../../contracts/index.ts";
import type { ApiResult } from "./api-result.ts";

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
      await fetch(buildRequestUrl(baseUrl, listEpisodesPath));
      // todo: 応答の status 分類・parse・schema 検証は別 issue で実装する
      return { ok: false, error: "invalid_response" };
    },
    async getEpisode(episodeId: string): Promise<ApiResult<GetEpisodeResponse>> {
      await fetch(buildRequestUrl(baseUrl, episodePath(episodeId)));
      // todo: 応答の status 分類・parse・schema 検証は別 issue で実装する
      return { ok: false, error: "invalid_response" };
    },
    async fetchAudio(audioRef: string): Promise<ApiResult<Blob>> {
      // why: audioRef は GetEpisodeResponse が持つ契約 path なので、web 側で episodeId から
      //   組み直さずそのまま繋ぐ
      await fetch(buildRequestUrl(baseUrl, audioRef));
      // todo: 応答の status 分類・blob 取得は別 issue で実装する
      return { ok: false, error: "invalid_response" };
    },
  };
}
