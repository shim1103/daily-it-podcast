import { useCallback, useRef, useState } from "react";
import type { EpisodeData, SelectionState } from "./playback-state.ts";

export type EpisodeSelectionViewModel = {
  selection: SelectionState;
  select(episodeId: string): void;
  deselect(): void;
  toggle(episodeId: string): void;
};

/**
 * 選択 union の toggle / select / deselect を担う hook。選択確定時に catalog の `episodes` から
 * 実在を検証し、`SelectionState` の「選択中」枝には episode の実体を入れる。
 *
 * @require episodes は選択候補となる catalog の一覧。参照は render ごとに変わってよい
 * @ensure 初期は selection={selected:false}
 * @ensure `select(episodeId)` は `episodes` に実在すれば {selected:true, episode} にする。
 *   実在しなければ selection を変えない（no-op）。無効な選択を state に入れない
 * @ensure `deselect()` は {selected:false} にする
 * @ensure `toggle(episodeId)` は、その id が選択中なら deselect、それ以外は select と同じ実在検証を行う
 * @invariant throw しない。hash・再生・catalog status は知らない。playback state を一切触らない
 */
export function useEpisodeSelection(episodes: readonly EpisodeData[]): EpisodeSelectionViewModel {
  const [selection, setSelection] = useState<SelectionState>({ selected: false });
  // why: select/toggle を安定参照にしたまま最新の一覧で実在検証する。episodes 参照が毎 render 変わっても callback を貼り替えない
  const episodesRef = useRef(episodes);
  episodesRef.current = episodes;

  const select = useCallback((episodeId: string): void => {
    const episode = episodesRef.current.find((candidate) => candidate.episodeId === episodeId);
    if (episode === undefined) {
      // why: 無効な選択を state に入れない（Decision §1-2）。既存の選択も維持する
      return;
    }
    setSelection({ selected: true, episode });
  }, []);

  const deselect = useCallback((): void => {
    setSelection({ selected: false });
  }, []);

  const toggle = useCallback((episodeId: string): void => {
    setSelection((current) => {
      if (current.selected && current.episode.episodeId === episodeId) {
        return { selected: false };
      }
      const episode = episodesRef.current.find((candidate) => candidate.episodeId === episodeId);
      if (episode === undefined) {
        return current;
      }
      return { selected: true, episode };
    });
  }, []);

  return {
    selection,
    select,
    deselect,
    toggle,
  };
}
