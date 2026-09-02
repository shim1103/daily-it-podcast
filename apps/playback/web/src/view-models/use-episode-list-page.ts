import { useCallback } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { BlockingError, CatalogStatus, EpisodeData } from "./playback-state.ts";
import {
  deriveBlockingError,
  deriveIsPlayed,
  deriveIsPlaying,
  deriveIsSelected,
  derivePlayedEpisode,
  deriveSelectedEpisode,
} from "./playback-state.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";
import { useEpisodePlayback } from "./use-episode-playback.ts";
import { useEpisodeSelection } from "./use-episode-selection.ts";
import { useHashSelectionSync } from "./use-hash-selection-sync.ts";
import type { EpisodePlaybackViewModel } from "./use-episode-playback.ts";

export type EpisodeListPageViewModel = {
  catalogStatus: CatalogStatus;
  episodes: EpisodeData[];
  selectedEpisodeId: string | null;
  selectedEpisode: EpisodeData | null;
  playedEpisodeId: string | null;
  playedEpisode: EpisodeData | null;
  isPlaying: boolean;
  blockingError: BlockingError | null;
  isSelected(episodeId: string): boolean;
  isPlayed(episodeId: string): boolean;
  load(): Promise<void>;
  select(episodeId: string): void;
  deselect(): void;
  toggleSelection(episodeId: string): void;
  play(episodeId: string): void;
  stop(): void;
  audioElementRef: EpisodePlaybackViewModel["audioElementRef"];
};

/**
 * catalog / selection / hash-sync / playback を compose する page 用 hook（契約 stub）。
 *
 * @require apiClient は `listEpisodes()` を持つ
 * @ensure 下位 hook stub を束ね、derive 関数で投影を返す。hash 同期は catalog 完了後に開始する想定だが stub は no-op
 * @invariant page が直接持つのは compose のみ
 */
export function useEpisodeListPage(apiClient: PlaybackApiClient): EpisodeListPageViewModel {
  const catalog = useEpisodeCatalog(apiClient);
  const selection = useEpisodeSelection();
  const playback = useEpisodePlayback();

  const hashSyncSelectedId =
    catalog.catalogStatus.status === "success" ? selection.selectedEpisodeId : undefined;
  useHashSelectionSync(hashSyncSelectedId, () => {
    // stub: hash 変化の解釈は C で page compose に実装
  });

  const selectedEpisode = deriveSelectedEpisode(catalog.episodes, selection.selectedEpisodeId);
  const playedEpisode = derivePlayedEpisode(catalog.episodes, playback.playedEpisodeId);
  const isPlaying = deriveIsPlaying(playback.playbackPhase);
  const blockingError = deriveBlockingError({
    catalogStatus: catalog.catalogStatus,
    episodes: catalog.episodes,
    selectedEpisodeId: selection.selectedEpisodeId,
    playedEpisodeId: playback.playedEpisodeId,
    playbackPhase: playback.playbackPhase,
  });

  const isSelected = useCallback(
    (episodeId: string): boolean => deriveIsSelected(episodeId, selection.selectedEpisodeId),
    [selection.selectedEpisodeId],
  );

  const isPlayed = useCallback(
    (episodeId: string): boolean => deriveIsPlayed(episodeId, playback.playedEpisodeId),
    [playback.playedEpisodeId],
  );

  return {
    catalogStatus: catalog.catalogStatus,
    episodes: catalog.episodes,
    selectedEpisodeId: selection.selectedEpisodeId,
    selectedEpisode,
    playedEpisodeId: playback.playedEpisodeId,
    playedEpisode,
    isPlaying,
    blockingError,
    isSelected,
    isPlayed,
    load: catalog.load,
    select: selection.select,
    deselect: selection.deselect,
    toggleSelection: selection.toggle,
    play: playback.play,
    stop: playback.stop,
    audioElementRef: playback.audioElementRef,
  };
}
