import { useCallback, useRef, useState } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { CatalogStatus, EpisodeData } from "./playback-state.ts";

export type EpisodeCatalogViewModel = {
  catalogStatus: CatalogStatus;
  episodes: EpisodeData[];
  load(): Promise<void>;
};

/**
 * episode 一覧の fetch と cache を担う hook。
 *
 * @require apiClient は `listEpisodes()` を持つ
 * @ensure 初期は loading・episodes は空。`load()` は `apiClient.listEpisodes()` を呼び、
 *   成功なら success + episodes、失敗なら error + episodes は空のまま。再実行前に loading へ戻す
 * @invariant throw しない。API Client の失敗は ApiResult の失敗側で受ける。hash・選択・再生は知らない
 */
export function useEpisodeCatalog(apiClient: PlaybackApiClient): EpisodeCatalogViewModel {
  const [catalogStatus, setCatalogStatus] = useState<CatalogStatus>({ status: "loading" });
  const [episodes, setEpisodes] = useState<EpisodeData[]>([]);
  const apiClientRef = useRef(apiClient);
  apiClientRef.current = apiClient;

  const load = useCallback(async (): Promise<void> => {
    setCatalogStatus({ status: "loading" });
    const result = await apiClientRef.current.listEpisodes();
    if (result.ok) {
      setEpisodes(result.data.episodes);
      setCatalogStatus({ status: "success" });
      return;
    }
    setCatalogStatus({ status: "error" });
  }, []);

  return {
    catalogStatus,
    episodes,
    load,
  };
}
