import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";

type ListEpisodesData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>;

export type EpisodeListState =
  | { status: "loading" }
  | { status: "success"; episodes: ListEpisodesData["episodes"] }
  | { status: "error" };

export type EpisodeListViewModel = {
  getState(): EpisodeListState;
  subscribe(listener: (state: EpisodeListState) => void): () => void;
  load(): Promise<void>;
};

/**
 * 一覧 page が使う state（loading / success / error）を持つ ViewModel を組み立てる。
 *
 * @require apiClient は `listEpisodes()` を持つ
 * @ensure 初期状態は loading。`load()` 完了後、成功なら episodes を、失敗なら error を state に持つ。
 *   state 変化のたび subscribe した listener を呼ぶ
 * @invariant throw しない。API Client の失敗は ApiResult の失敗側として受け取る
 */
export function createEpisodeListViewModel(apiClient: PlaybackApiClient): EpisodeListViewModel {
  let state: EpisodeListState = { status: "loading" };
  const listeners = new Set<(state: EpisodeListState) => void>();

  function setState(next: EpisodeListState): void {
    state = next;
    for (const listener of listeners) {
      listener(state);
    }
  }

  return {
    getState(): EpisodeListState {
      return state;
    },
    subscribe(listener: (state: EpisodeListState) => void): () => void {
      listeners.add(listener);
      return () => listeners.delete(listener);
    },
    async load(): Promise<void> {
      setState({ status: "loading" });
      const result = await apiClient.listEpisodes();
      if (result.ok) {
        setState({ status: "success", episodes: result.data.episodes });
      } else {
        setState({ status: "error" });
      }
    },
  };
}
