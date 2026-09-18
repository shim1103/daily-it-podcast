import { useCallback, useEffect, useState } from "react";
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
 * @require apiClient は `listEpisodes()` を持つ。呼び出し側は render を跨いで同一参照を渡す
 *   （変わる場合は `load` が再生成され、mount 時の auto-load effect が再発火する）
 * @ensure 初期は loading・episodes は空。mount 時に自動で `load()` を 1 回呼ぶ。以後 `load()` は
 *   明示 reload 用として残す。`load()` は `apiClient.listEpisodes()` を呼び、成功なら
 *   success + episodes、失敗なら error + episodes は空のまま。再実行前に loading へ戻す
 * @invariant throw しない。API Client の失敗は ApiResult の失敗側で受ける。hash・選択・再生は知らない
 */
export function useEpisodeCatalog(apiClient: PlaybackApiClient): EpisodeCatalogViewModel {
  const [catalogStatus, setCatalogStatus] = useState<CatalogStatus>({ status: "loading" });
  const [episodes, setEpisodes] = useState<EpisodeData[]>([]);

  const load = useCallback(async (): Promise<void> => {
    setCatalogStatus({ status: "loading" });
    const result = await apiClient.listEpisodes();
    if (result.ok) {
      setEpisodes(result.data.episodes);
      setCatalogStatus({ status: "success" });
      return;
    }
    setCatalogStatus({ status: "error" });
  }, [apiClient]);

  // why: 起動を hook の内部へ寄せる。mount 時 1 回だけ自動 load し、呼び出し側が
  //   `useEffect(() => void load())` を書かずに済む（load は明示 reload 用に残す）
  useEffect(() => {
    void load();
  }, [load]);

  return {
    catalogStatus,
    episodes,
    load,
  };
}
