import type { HashSelectionAdapter } from "../lib/hash-selection-adapter.ts";

/**
 * 選択 id ↔ hash の双方向同期のみを担う副作用 hook（契約 stub）。
 *
 * @require selectedEpisodeId が undefined の間は hash を書き換えない（同期保留）
 * @require onHashEpisodeIdChange は hash 変化時に episodeId（空 hash は null）を受け取る
 * @ensure adapter 未指定時は createHashSelectionAdapter 相当を使う想定だが、stub は no-op
 * @invariant select の解釈は caller が持つ
 */
export function useHashSelectionSync(
  _selectedEpisodeId: string | null | undefined,
  _onHashEpisodeIdChange: (episodeId: string | null) => void,
  _adapter?: HashSelectionAdapter,
): void {
  // stub: 実装は C で置換
}
