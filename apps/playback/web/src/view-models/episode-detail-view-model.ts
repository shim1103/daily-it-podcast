import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";

type EpisodeData = ApiSuccessData<PlaybackApiClient["getEpisode"]>;

export type EpisodeDetailState =
  | { status: "loading" }
  | { status: "success"; episode: EpisodeData }
  | { status: "error" };

export type EpisodeDetailViewModel = {
  getState(): EpisodeDetailState;
  subscribe(listener: (state: EpisodeDetailState) => void): () => void;
  load(episodeId: string): Promise<void>;
};

/**
 * 詳細 page が使う state（loading / success / error）を持つ ViewModel を組み立てる。
 *
 * @require apiClient は `getEpisode(episodeId)` を持つ
 * @ensure 初期状態は loading。`load(episodeId)` 完了後、成功なら episode を、失敗なら error を state に持つ。
 *   state 変化のたび subscribe した listener を呼ぶ
 * @invariant throw しない。API Client の失敗は ApiResult の失敗側として受け取る
 */
export function createEpisodeDetailViewModel(apiClient: PlaybackApiClient): EpisodeDetailViewModel {
  let state: EpisodeDetailState = { status: "loading" };
  const listeners = new Set<(state: EpisodeDetailState) => void>();

  function setState(next: EpisodeDetailState): void {
    state = next;
    for (const listener of listeners) {
      listener(state);
    }
  }

  return {
    getState(): EpisodeDetailState {
      return state;
    },
    subscribe(listener: (state: EpisodeDetailState) => void): () => void {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    async load(episodeId: string): Promise<void> {
      setState({ status: "loading" });
      const result = await apiClient.getEpisode(episodeId);
      if (result.ok) {
        setState({ status: "success", episode: result.data });
      } else {
        setState({ status: "error" });
      }
    },
  };
}
