import { useCallback } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { CatalogStatus, EpisodeData } from "./playback-state.ts";

export type EpisodeCatalogViewModel = {
  catalogStatus: CatalogStatus;
  episodes: EpisodeData[];
  load(): Promise<void>;
};

/**
 * episode 一覧の fetch と cache を担う hook（契約 stub）。
 *
 * @require apiClient は `listEpisodes()` を持つ
 * @ensure 初期は loading・episodes は空。`load()` は no-op（実装は C で置換）
 * @invariant hash・選択・再生は知らない
 */
export function useEpisodeCatalog(_apiClient: PlaybackApiClient): EpisodeCatalogViewModel {
  const load = useCallback(async (): Promise<void> => {
    // stub: 実装は C で置換
  }, []);

  return {
    catalogStatus: { status: "loading" },
    episodes: [],
    load,
  };
}
