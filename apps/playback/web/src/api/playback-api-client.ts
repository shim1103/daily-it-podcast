import {
  ListEpisodesResponseSchema,
  ProgressPullResponseSchema,
  ProgressWriteResponseSchema,
} from "../../../contracts/index.ts";
import type {
  ListEpisodesResponse,
  ProgressPullQuery,
  ProgressPullResponse,
  ProgressWriteRequest,
  ProgressWriteResponse,
} from "../../../contracts/index.ts";
import type { ApiResult } from "./api-result.ts";
import { resolveApiResult } from "./playback-api-response.ts";
import { createPlaybackRpcClient } from "./playback-rpc-client.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export type PlaybackApiClientDeps = {
  baseUrl: string;
  fetch: FetchLike;
};

export type PlaybackApiClient = {
  listEpisodes(): Promise<ApiResult<ListEpisodesResponse>>;
  createProgress(
    episodeId: string,
    body: ProgressWriteRequest,
  ): Promise<ApiResult<ProgressWriteResponse>>;
  updateProgress(
    episodeId: string,
    body: ProgressWriteRequest,
  ): Promise<ApiResult<ProgressWriteResponse>>;
  completeProgress(
    episodeId: string,
    body: ProgressWriteRequest,
  ): Promise<ApiResult<ProgressWriteResponse>>;
  pullProgress(query: ProgressPullQuery): Promise<ApiResult<ProgressPullResponse>>;
};

/**
 * playback worker の HTTP API を叩く client を組み立てる。
 * schema + status + network → ApiResult のみ（path / encode は持たない）。
 * Write の HTTP retry（`PROGRESS_WRITE_MAX_ATTEMPTS`）は呼び出し側（VM）が持つ。本 client は 1 試行。
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
      return resolveApiResult(() => rpc.listEpisodes(), ListEpisodesResponseSchema);
    },
    async createProgress(
      episodeId: string,
      body: ProgressWriteRequest,
    ): Promise<ApiResult<ProgressWriteResponse>> {
      return resolveApiResult(
        () => rpc.createProgress(episodeId, body),
        ProgressWriteResponseSchema,
      );
    },
    async updateProgress(
      episodeId: string,
      body: ProgressWriteRequest,
    ): Promise<ApiResult<ProgressWriteResponse>> {
      return resolveApiResult(
        () => rpc.updateProgress(episodeId, body),
        ProgressWriteResponseSchema,
      );
    },
    async completeProgress(
      episodeId: string,
      body: ProgressWriteRequest,
    ): Promise<ApiResult<ProgressWriteResponse>> {
      return resolveApiResult(
        () => rpc.completeProgress(episodeId, body),
        ProgressWriteResponseSchema,
      );
    },
    async pullProgress(query: ProgressPullQuery): Promise<ApiResult<ProgressPullResponse>> {
      return resolveApiResult(() => rpc.pullProgress(query), ProgressPullResponseSchema);
    },
  };
}
