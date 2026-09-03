import { useCallback } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { HashSelectionAdapter } from "../lib/hash-selection-adapter.ts";
import type {
  CatalogStatus,
  EpisodeData,
  EpisodeRowViewModel,
  PageStatus,
  PlaybackState,
  SelectionState,
} from "./playback-state.ts";
import { deriveEpisodeRows, derivePageStatus, derivePlayedEpisode } from "./playback-state.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";
import { useEpisodePlayback } from "./use-episode-playback.ts";
import type { EpisodePlaybackViewModel } from "./use-episode-playback.ts";
import { useEpisodeSelection } from "./use-episode-selection.ts";
import { useHashSelectionSync } from "./use-hash-selection-sync.ts";

export type EpisodeListPageViewModel = {
  catalogStatus: CatalogStatus;
  episodes: EpisodeData[];
  selection: SelectionState;
  selectedEpisode: EpisodeData | null;
  playback: PlaybackState;
  playedEpisode: EpisodeData | null;
  rows: EpisodeRowViewModel[];
  pageStatus: PageStatus;
  load(): Promise<void>;
  select(episodeId: string): void;
  deselect(): void;
  toggleSelection(episodeId: string): void;
  play(episodeId: string, positionSec?: number): void;
  seek(episodeId: string, positionSec: number): void;
  stop(): void;
  audioElementRef: EpisodePlaybackViewModel["audioElementRef"];
};

/**
 * catalog / selection / hash-sync / playback を compose する page 用 hook。
 *
 * @require apiClient は `listEpisodes()` を持つ。adapter は test 用の DI で、未指定なら
 *   `useHashSelectionSync` が既定 adapter を使う
 * @ensure 下位 hook を束ね、`playback-state.ts` の derive 関数で投影を返す。page が全画面の
 *   振る舞いを決めるために見るのは `pageStatus` 1 型（`derivePageStatus(catalog.catalogStatus)`）。
 *   hash 同期の保留判断（catalog 完了まで同期しない）は `useHashSelectionSync` の内部に閉じる。
 *   hash 変化は episodeId があれば `selection.select`（一覧に無ければ no-op）、null なら
 *   `selection.deselect` として解釈する
 * @invariant throw しない。page が直接持つのは compose のみ。state machine は各下位 hook
 */
export function useEpisodeListPage(
  apiClient: PlaybackApiClient,
  adapter?: HashSelectionAdapter,
): EpisodeListPageViewModel {
  const catalog = useEpisodeCatalog(apiClient);
  const selection = useEpisodeSelection(catalog.episodes);
  const playback = useEpisodePlayback();

  const onHashEpisodeIdChange = useCallback(
    (episodeId: string | null): void => {
      if (episodeId === null) {
        selection.deselect();
        return;
      }
      selection.select(episodeId);
    },
    [selection.select, selection.deselect],
  );

  useHashSelectionSync(
    { catalogReady: catalog.catalogStatus.status === "success", selection: selection.selection },
    onHashEpisodeIdChange,
    adapter,
  );

  const selectedEpisode = selection.selection.selected ? selection.selection.episode : null;
  const playedEpisode = derivePlayedEpisode(catalog.episodes, playback.playback);
  const rows = deriveEpisodeRows(catalog.episodes, {
    selection: selection.selection,
    playback: playback.playback,
  });
  const pageStatus = derivePageStatus(catalog.catalogStatus);

  return {
    catalogStatus: catalog.catalogStatus,
    episodes: catalog.episodes,
    selection: selection.selection,
    selectedEpisode,
    playback: playback.playback,
    playedEpisode,
    rows,
    pageStatus,
    load: catalog.load,
    select: selection.select,
    deselect: selection.deselect,
    toggleSelection: selection.toggle,
    play: playback.play,
    seek: playback.seek,
    stop: playback.stop,
    audioElementRef: playback.audioElementRef,
  };
}
