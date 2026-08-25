import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";

type ListEpisodesData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>;
export type EpisodeData = ApiSuccessData<PlaybackApiClient["getEpisode"]>;
export type EpisodeListItemData = ListEpisodesData["episodes"][number];

export type SelectedEpisodeState =
  | { status: "loading" }
  | { status: "success"; episode: EpisodeData }
  | { status: "error" };

export type EpisodeListState =
  | { status: "loading" }
  | {
      status: "success";
      episodes: ListEpisodesData["episodes"];
      selectedEpisodeId: string | null;
      selectedEpisode: SelectedEpisodeState | null;
    }
  | { status: "error" };

export type EpisodeListViewModel = {
  getState(): EpisodeListState;
  subscribe(listener: (state: EpisodeListState) => void): () => void;
  load(): Promise<void>;
  select(episodeId: string): Promise<void>;
};

/**
 * 一覧 page が使う state（loading / success / error）と選択中 episode の詳細取得を持つ ViewModel を組み立てる。
 *
 * @require apiClient は `listEpisodes()` と `getEpisode(episodeId)` を持つ
 * @ensure 初期状態は loading。`load()` 完了後、成功なら episodes と selectedEpisodeId（初期 null）・
 *   selectedEpisode（初期 null）を、失敗なら error を state に持つ。`select(episodeId)` は一覧が success の時のみ、
 *   同じ episodeId が選択中なら選択を解除し、それ以外は選択して詳細を loading → success/error の順に取得する。
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
        setState({
          status: "success",
          episodes: result.data.episodes,
          selectedEpisodeId: null,
          selectedEpisode: null,
        });
      } else {
        setState({ status: "error" });
      }
    },
    async select(episodeId: string): Promise<void> {
      // why: 一覧を読み込めていない状態（loading/error）では選択操作を受け付けない
      //   （state machineの整合性を自分で守る）
      if (state.status !== "success") {
        return;
      }

      if (state.selectedEpisodeId === episodeId) {
        setState({ ...state, selectedEpisodeId: null, selectedEpisode: null });
        return;
      }

      setState({ ...state, selectedEpisodeId: episodeId, selectedEpisode: { status: "loading" } });
      const result = await apiClient.getEpisode(episodeId);

      // why: 詳細取得中に一覧が再読込・別選択された場合、古い応答で state を上書きしない
      if (state.status !== "success" || state.selectedEpisodeId !== episodeId) {
        return;
      }

      if (result.ok) {
        setState({ ...state, selectedEpisode: { status: "success", episode: result.data } });
      } else {
        setState({ ...state, selectedEpisode: { status: "error" } });
      }
    },
  };
}
