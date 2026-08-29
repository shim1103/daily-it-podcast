import { useCallback, useRef, useState } from "react";
import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";

type ListEpisodesData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>;
export type EpisodeData = ApiSuccessData<PlaybackApiClient["getEpisode"]>;
export type EpisodeListItemData = ListEpisodesData["episodes"][number];

export type EpisodeDetailState =
  | { status: "loading" }
  | { status: "success"; episode: EpisodeData }
  | { status: "error" };

export type EpisodeSelection =
  | { kind: "none" }
  | { kind: "open"; episodeId: string; detail: EpisodeDetailState };

export type EpisodeListState =
  | { status: "loading" }
  | {
      status: "success";
      episodes: ListEpisodesData["episodes"];
      selection: EpisodeSelection;
    }
  | { status: "error" };

export type EpisodeListViewModel = {
  state: EpisodeListState;
  load(): Promise<void>;
  select(episodeId: string): Promise<void>;
};

/**
 * 一覧 page が使う state（loading / success / error）と選択中 episode の詳細取得を持つ ViewModel hook。
 *
 * @require apiClient は `listEpisodes()` と `getEpisode(episodeId)` を持つ
 * @ensure 初期状態は loading。`load()` 完了後、成功なら episodes と selection（初期 kind: none）を、
 *   失敗なら error を state に持つ。`select(episodeId)` は一覧が success の時のみ、
 *   同じ episodeId が選択中なら選択を解除し、それ以外は選択して詳細を loading → success/error の順に取得する
 * @invariant throw しない。API Client の失敗は ApiResult の失敗側として受け取る
 */
export function useEpisodeListViewModel(apiClient: PlaybackApiClient): EpisodeListViewModel {
  const [state, setStateReact] = useState<EpisodeListState>({ status: "loading" });
  const stateRef = useRef(state);

  const setState = useCallback((next: EpisodeListState): void => {
    // why: race 判定は ref の同期更新で足りる。flushSync は一時橋側の同期観測都合であり hook に置かない
    stateRef.current = next;
    setStateReact(next);
  }, []);

  const load = useCallback(async (): Promise<void> => {
    setState({ status: "loading" });
    const result = await apiClient.listEpisodes();
    if (result.ok) {
      setState({
        status: "success",
        episodes: result.data.episodes,
        selection: { kind: "none" },
      });
    } else {
      setState({ status: "error" });
    }
  }, [apiClient, setState]);

  const select = useCallback(
    async (episodeId: string): Promise<void> => {
      // why: 一覧を読み込めていない状態（loading/error）では選択操作を受け付けない
      //   （state machineの整合性を自分で守る）
      const current = stateRef.current;
      if (current.status !== "success") {
        return;
      }

      if (current.selection.kind === "open" && current.selection.episodeId === episodeId) {
        setState({ ...current, selection: { kind: "none" } });
        return;
      }

      setState({
        ...current,
        selection: { kind: "open", episodeId, detail: { status: "loading" } },
      });
      const result = await apiClient.getEpisode(episodeId);

      // why: 詳細取得中に一覧が再読込・別選択された場合、古い応答で state を上書きしない
      const after = stateRef.current;
      if (
        after.status !== "success" ||
        after.selection.kind !== "open" ||
        after.selection.episodeId !== episodeId
      ) {
        return;
      }

      if (result.ok) {
        setState({
          ...after,
          selection: { kind: "open", episodeId, detail: { status: "success", episode: result.data } },
        });
      } else {
        setState({
          ...after,
          selection: { kind: "open", episodeId, detail: { status: "error" } },
        });
      }
    },
    [apiClient, setState],
  );

  return { state, load, select };
}
