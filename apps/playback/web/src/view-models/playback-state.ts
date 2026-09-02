import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";

type ListEpisodesData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>;

export type EpisodeData = ListEpisodesData["episodes"][number];

export type CatalogStatus = { status: "loading" } | { status: "success" } | { status: "error" };

export type PlaybackPhase = "idle" | "loading" | "playing" | "paused" | "ended" | "error";

export type BlockingError =
  | { kind: "catalog-load-failed" }
  | { kind: "invalid-selected-episode"; episodeId: string }
  | { kind: "audio-load-failed"; episodeId: string };

/**
 * `playbackPhase` から再生中かを導出する。
 *
 * @ensure `playbackPhase === "playing"` のときのみ true
 */
export function deriveIsPlaying(playbackPhase: PlaybackPhase): boolean {
  return playbackPhase === "playing";
}

/**
 * 一覧 cache と選択 id から選択中 episode を導出する。
 *
 * @ensure `selectedEpisodeId` が null または一覧に無いとき null
 */
export function deriveSelectedEpisode(
  episodes: readonly EpisodeData[],
  selectedEpisodeId: string | null,
): EpisodeData | null {
  if (selectedEpisodeId === null) {
    return null;
  }
  return episodes.find((episode) => episode.episodeId === selectedEpisodeId) ?? null;
}

/**
 * 一覧 cache と再生 id から再生中 episode を導出する。
 *
 * @ensure `playedEpisodeId` が null または一覧に無いとき null
 */
export function derivePlayedEpisode(
  episodes: readonly EpisodeData[],
  playedEpisodeId: string | null,
): EpisodeData | null {
  if (playedEpisodeId === null) {
    return null;
  }
  return episodes.find((episode) => episode.episodeId === playedEpisodeId) ?? null;
}

/**
 * page 全体を止める blocking error を導出する。
 *
 * @ensure catalog load 失敗・invalid hash 由来の未知 episodeId・audio load 失敗のいずれか 1 件、
 *   または blocking なしで null
 */
export function deriveBlockingError(params: {
  catalogStatus: CatalogStatus;
  episodes: readonly EpisodeData[];
  selectedEpisodeId: string | null;
  playedEpisodeId: string | null;
  playbackPhase: PlaybackPhase;
}): BlockingError | null {
  if (params.catalogStatus.status === "error") {
    return { kind: "catalog-load-failed" };
  }
  if (params.catalogStatus.status === "loading") {
    return null;
  }
  if (params.selectedEpisodeId !== null) {
    const selectedExists = params.episodes.some(
      (episode) => episode.episodeId === params.selectedEpisodeId,
    );
    if (!selectedExists) {
      return { kind: "invalid-selected-episode", episodeId: params.selectedEpisodeId };
    }
  }
  if (params.playbackPhase === "error" && params.playedEpisodeId !== null) {
    return { kind: "audio-load-failed", episodeId: params.playedEpisodeId };
  }
  return null;
}

/**
 * 指定 episode が選択中かを導出する。
 */
export function deriveIsSelected(episodeId: string, selectedEpisodeId: string | null): boolean {
  return selectedEpisodeId === episodeId;
}

/**
 * 指定 episode が再生対象かを導出する。
 */
export function deriveIsPlayed(episodeId: string, playedEpisodeId: string | null): boolean {
  return playedEpisodeId === episodeId;
}
