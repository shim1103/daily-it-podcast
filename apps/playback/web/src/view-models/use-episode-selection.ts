import { useCallback } from "react";

export type EpisodeSelectionViewModel = {
  selectedEpisodeId: string | null;
  select(episodeId: string): void;
  deselect(): void;
  toggle(episodeId: string): void;
};

/**
 * 選択 id の toggle / select / deselect のみを担う hook（契約 stub）。
 *
 * @ensure 初期は selectedEpisodeId=null。各操作は no-op（実装は C で置換）
 * @invariant hash・再生・catalog は知らない
 */
export function useEpisodeSelection(): EpisodeSelectionViewModel {
  const select = useCallback((_episodeId: string): void => {
    // stub: 実装は C で置換
  }, []);

  const deselect = useCallback((): void => {
    // stub: 実装は C で置換
  }, []);

  const toggle = useCallback((_episodeId: string): void => {
    // stub: 実装は C で置換
  }, []);

  return {
    selectedEpisodeId: null,
    select,
    deselect,
    toggle,
  };
}
