import { useEffect, useMemo, useRef, useSyncExternalStore } from "react";
import {
  type HashSelectionAdapter,
  createHashSelectionAdapter,
} from "../lib/hash-selection-adapter.ts";
import type { SelectionState } from "./playback-state.ts";

/**
 * 選択 ↔ hash の双方向同期のみを担う副作用 hook。「catalog 完了まで同期しない」保留判断も
 * この hook の内部に閉じる（page の関心事にしない）。
 *
 * @require context.catalogReady が false の間は hash を一切書き換えず、外部 hash 変化も
 *   caller へ流さない（同期保留）。catalog が完了して初めて同期を開始する
 * @require context.selection は選択 union。選択中なら episode の episodeId が hash の書き込み値、
 *   選択なしなら hash クリア
 * @require onHashEpisodeIdChange は hash 変化時に episodeId（空 hash は null）を受け取る。
 *   catalog 完了後の初回に既存 hash（非 null）を 1 回流す（deep-link 復元）。それ以降は
 *   本 hook 自身の書き込みが誘発した通知を流さない（echo 抑止）
 * @require adapter は render をまたいで安定した参照であること。未指定時は内部で `useMemo` により安定化する
 * @ensure adapter 未指定時は `createHashSelectionAdapter()` を使う。adapter の subscribe を
 *   `useSyncExternalStore` で購読し、その購読解除で unmount 時に listener を外す。
 *   catalog 完了後の初回 read effect で、既存 hash が非 null なら `onHashEpisodeIdChange` へ
 *   1 回流し（deep-link 復元）、`syncedEpisodeIdRef` を更新するので書き込み方向 effect は
 *   同値で no-op になり無限書き戻しは起きない。既存 hash が空なら流れない。
 *   catalog 完了後は selection 変化のたび hash を追従させる。外部変化由来で caller が selection を
 *   同じ episode へ更新し hash へ書き戻しても、直近同期値と一致すれば書き込まず無限ループしない
 * @invariant throw しない。JSX を持たない。select の解釈は caller が持つ
 */
export function useHashSelectionSync(
  context: { catalogReady: boolean; selection: SelectionState },
  onHashEpisodeIdChange: (episodeId: string | null) => void,
  adapter?: HashSelectionAdapter,
): void {
  // why: 購読先の identity を固定する
  const fallbackAdapter = useMemo(() => createHashSelectionAdapter(), []);
  const activeAdapter = adapter ?? fallbackAdapter;

  // why: 保留判断を hook 内へ閉じる。catalog 未完了なら undefined（＝同期停止）、完了後は
  //   selection union から hash の書き込み値（string | null）を決める
  const selectedEpisodeId: string | null | undefined = !context.catalogReady
    ? undefined
    : context.selection.selected
      ? context.selection.episode.episodeId
      : null;

  // why: getEpisodeId は string | null を返し新規オブジェクトを毎回作らないため getSnapshot に直接使える（新規 object を返すと無限 render）
  const currentEpisodeId = useSyncExternalStore(
    activeAdapter.subscribe,
    activeAdapter.getEpisodeId,
  );

  // why: 読み取り差分と書き込み差分をこの 1 値で判定する。初期値を null にすることで、catalog
  //   完了後の初回 read effect が既存 hash（非 null）を「差分あり」と見なして 1 回流す（deep-link
  //   復元）。空 hash なら null 同士で流れない。流した後 ref を進めるので書き込み方向は no-op になる
  const syncedEpisodeIdRef = useRef<string | null>(null);

  // why: 読み取り方向。報告値が直近同期値と異なる時だけ onHashEpisodeIdChange へ流す。この effect を
  //   書き込み方向より先に宣言し、両者が同一 commit で発火する時に読み取り→書き込みの順を固定する
  //   （逆順だと書き込みが先に syncedEpisodeIdRef を進め、続く読み取りが旧 hash を外部変化と誤検知する）
  useEffect(() => {
    if (selectedEpisodeId === undefined) {
      // why: catalog 未完了は同期保留。保留中の外部 hash 変化は caller へ流さず、解除後の読み取りで初めて反映する
      return;
    }
    if (currentEpisodeId === syncedEpisodeIdRef.current) {
      return;
    }
    syncedEpisodeIdRef.current = currentEpisodeId;
    onHashEpisodeIdChange(currentEpisodeId);
  }, [selectedEpisodeId, currentEpisodeId, onHashEpisodeIdChange]);

  // why: 書き込み方向。undefined は同期保留。直近同期値と一致すれば書き込まない（書き戻しによる無限ループ抑止）
  useEffect(() => {
    if (selectedEpisodeId === undefined) {
      return;
    }
    const nextEpisodeId = selectedEpisodeId;
    if (nextEpisodeId !== syncedEpisodeIdRef.current) {
      syncedEpisodeIdRef.current = nextEpisodeId;
      activeAdapter.setEpisodeId(nextEpisodeId);
    }
  }, [selectedEpisodeId, activeAdapter]);
}
