import { useCallback } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { HashSelectionAdapter } from "../lib/hash-selection-adapter.ts";
import { buildRequestUrl } from "../utils/build-request-url.ts";
import type {
  EpisodeRowViewModel,
  NowPlayingViewModel,
  PageStatus,
  PlaybackState,
} from "./playback-state.ts";
import { deriveEpisodeRows, deriveNowPlaying, derivePageStatus } from "./playback-state.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";
import { useEpisodePlayback } from "./use-episode-playback.ts";
import type { EpisodePlaybackViewModel } from "./use-episode-playback.ts";
import { useEpisodeSelection } from "./use-episode-selection.ts";
import { useHashSelectionSync } from "./use-hash-selection-sync.ts";

export type EpisodeListPageViewModel = {
  // why: EpisodeListPage は未消費だが、rows/nowPlaying が投影しない audioRef 解決結果・
  //   phase:"error" の理由・positionSec を test が直接検証するために生 union のまま残す
  playback: PlaybackState;
  rows: EpisodeRowViewModel[];
  nowPlaying: NowPlayingViewModel | null;
  pageStatus: PageStatus;
  toggleSelection(episodeId: string): void;
  play(episodeId: string, positionSec?: number): void;
  seek(episodeId: string, positionSec: number): void;
  stop(): void;
  audioElementRef: EpisodePlaybackViewModel["audioElementRef"];
};

/**
 * catalog / selection / hash-sync / playback を compose する page 用 hook。
 * 戻り値は page が使う投影とアクションだけ。生 union と select/deselect/load/episodes は
 * 内部合成材料として使い、外へは出さない。state machine は各下位 hook が持つ
 *
 * @require apiClient は `listEpisodes()` を持つ。adapter は test 用の DI で、未指定なら
 *   `useHashSelectionSync` が既定 adapter を使う
 * @invariant throw しない
 */
export function useEpisodeListPage(
  apiClient: PlaybackApiClient,
  baseUrl: string,
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

  const resolveAudioUrl = useCallback(
    (episodeId: string): string | null => {
      const audioRef = catalog.episodes.find(
        (episode) => episode.episodeId === episodeId,
      )?.audioRef;
      return audioRef === undefined ? null : buildRequestUrl(baseUrl, audioRef);
    },
    [catalog.episodes, baseUrl],
  );
  const play = useCallback(
    (episodeId: string, positionSec?: number): void => {
      const audioUrl = resolveAudioUrl(episodeId);
      if (audioUrl === null) {
        return;
      }
      playback.play(episodeId, audioUrl, positionSec);
    },
    [resolveAudioUrl, playback.play],
  );
  const seek = useCallback(
    (episodeId: string, positionSec: number): void => {
      const audioUrl = resolveAudioUrl(episodeId);
      if (audioUrl === null) {
        return;
      }
      playback.seek(episodeId, audioUrl, positionSec);
    },
    [resolveAudioUrl, playback.seek],
  );

  return {
    playback: playback.playback,
    rows: deriveEpisodeRows(catalog.episodes, {
      selection: selection.selection,
      playback: playback.playback,
    }),
    nowPlaying: deriveNowPlaying(catalog.episodes, playback.playback),
    pageStatus: derivePageStatus(catalog.catalogStatus),
    toggleSelection: selection.toggle,
    play,
    seek,
    stop: playback.stop,
    audioElementRef: playback.audioElementRef,
  };
}
