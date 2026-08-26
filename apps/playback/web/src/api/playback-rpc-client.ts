import { hc } from "hono/client";
import type { AppType } from "../../../worker/src/routes/app.ts";

type FetchLike = (input: string, init?: RequestInit) => Promise<Response>;

export type PlaybackRpcClientDeps = {
  baseUrl: string;
  fetch: FetchLike;
};

export type PlaybackRpcClient = {
  listEpisodes(): Promise<Response>;
  getEpisode(episodeId: string): Promise<Response>;
};

/**
 * playback worker 向け Hono RPC client を組み立てる。
 * request 面（URL / method / path param の encode）を持つ。
 *
 * @require deps.baseUrl は worker の origin。末尾の `/` は有無どちらでもよい
 * @require deps.fetch は Fetch API 互換の呼び出し
 * @ensure listEpisodes / getEpisode は throw しうる Response 取得を返す
 */
export function createPlaybackRpcClient(deps: PlaybackRpcClientDeps): PlaybackRpcClient {
  const client = hc<AppType>(deps.baseUrl, {
    fetch: deps.fetch as typeof globalThis.fetch,
  });

  return {
    listEpisodes() {
      return client.episodes.$get();
    },
    getEpisode(episodeId: string) {
      // why: Hono RPC は param を encode しない。wire は contracts `episodePath` と一致させる
      return client.episodes[":episodeId"].$get({
        param: { episodeId: encodeURIComponent(episodeId) },
      });
    },
  };
}
