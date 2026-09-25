import { hc } from "hono/client";
import type { ProgressPullQuery, ProgressWriteRequest } from "../../../contracts/index.ts";
import type { AppType } from "../../../worker/src/routes/app.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export type PlaybackRpcClientDeps = {
  baseUrl: string;
  fetch: FetchLike;
};

export type PlaybackRpcClient = {
  listEpisodes(): Promise<Response>;
  createProgress(episodeId: string, body: ProgressWriteRequest): Promise<Response>;
  updateProgress(episodeId: string, body: ProgressWriteRequest): Promise<Response>;
  completeProgress(episodeId: string, body: ProgressWriteRequest): Promise<Response>;
  pullProgress(query: ProgressPullQuery): Promise<Response>;
};

/**
 * playback worker 向け Hono RPC client を組み立てる。
 * request 面（URL / method / path param の encode）を持つ。
 *
 * @require deps.baseUrl は worker の origin。末尾の `/` は有無どちらでもよい
 * @require deps.fetch は Fetch API 互換の呼び出し
 * @ensure 各 method は throw しうる Response 取得を返す
 */
export function createPlaybackRpcClient(deps: PlaybackRpcClientDeps): PlaybackRpcClient {
  const client = hc<AppType>(deps.baseUrl, {
    fetch: deps.fetch as typeof globalThis.fetch,
  });

  return {
    listEpisodes() {
      return client.episodes.$get();
    },
    createProgress(episodeId, body) {
      return client.episodes[":episodeId"].progress.$post({
        param: { episodeId },
        json: body,
      });
    },
    updateProgress(episodeId, body) {
      return client.episodes[":episodeId"].progress.$patch({
        param: { episodeId },
        json: body,
      });
    },
    completeProgress(episodeId, body) {
      return client.episodes[":episodeId"].progress.complete.$post({
        param: { episodeId },
        json: body,
      });
    },
    pullProgress(query) {
      return client.progress.$get({ query });
    },
  };
}
