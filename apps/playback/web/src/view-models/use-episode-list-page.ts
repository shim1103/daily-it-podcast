import { useCallback } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { HashSelectionAdapter } from "../lib/hash-selection-adapter.ts";
import type {
  EpisodeData,
  EpisodeRowViewModel,
  PageStatus,
  PlaybackState,
} from "./playback-state.ts";
import { deriveEpisodeRows, derivePageStatus } from "./playback-state.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";
import { useEpisodePlayback } from "./use-episode-playback.ts";
import type { EpisodePlaybackViewModel } from "./use-episode-playback.ts";
import { useEpisodeSelection } from "./use-episode-selection.ts";
import { useHashSelectionSync } from "./use-hash-selection-sync.ts";

export type EpisodeListPageViewModel = {
  selectedEpisode: EpisodeData | null;
  playback: PlaybackState;
  rows: EpisodeRowViewModel[];
  pageStatus: PageStatus;
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
 *   `selection.deselect` として解釈する。公開する `play(episodeId, positionSec?)` /
 *   `seek(episodeId, positionSec)` は外部 signature を保ったまま、内部で `episodes` から
 *   `audioRef` を引き当てて下位 `playback.play` / `playback.seek` へ渡す。episodeId が一覧に
 *   無ければ `audioRef` を解決できず no-op（`useEpisodeSelection.select` の実在検証と対称）。
 *   戻り値は page が使う投影とアクションだけ。生 union（`selection` / `catalogStatus`）と
 *   `select` / `deselect` / `load` / `episodes` は内部合成材料として使い、外へは出さない。
 *   catalog の起動は `useEpisodeCatalog` の auto-load、deep-link 復元は `useHashSelectionSync`
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
  const rows = deriveEpisodeRows(catalog.episodes, {
    selection: selection.selection,
    playback: playback.playback,
  });
  const pageStatus = derivePageStatus(catalog.catalogStatus);

  const episodes = catalog.episodes;
  const playbackPlay = playback.play;
  const playbackSeek = playback.seek;
  // why: play/seek の外部 signature は episodeId だけ受ける形を維持し、audioRef 解決は hook 内部へ隠す。
  //   一覧に無い episodeId は audioRef を引けないので no-op（use-episode-selection.select と対称）
  const play = useCallback(
    (episodeId: string, positionSec?: number): void => {
      const audioRef = episodes.find((episode) => episode.episodeId === episodeId)?.audioRef;
      if (audioRef === undefined) {
        return;
      }
      playbackPlay(episodeId, audioRef, positionSec);
    },
    [episodes, playbackPlay],
  );
  const seek = useCallback(
    (episodeId: string, positionSec: number): void => {
      const audioRef = episodes.find((episode) => episode.episodeId === episodeId)?.audioRef;
      if (audioRef === undefined) {
        return;
      }
      playbackSeek(episodeId, audioRef, positionSec);
    },
    [episodes, playbackSeek],
  );

  return {
    selectedEpisode,
    playback: playback.playback,
    rows,
    pageStatus,
    toggleSelection: selection.toggle,
    play,
    seek,
    stop: playback.stop,
    audioElementRef: playback.audioElementRef,
  };
}
