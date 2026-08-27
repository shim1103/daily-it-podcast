import { useEffect, useRef } from "react";
import { getLocationHash, onLocationHashChange, setLocationHash } from "../lib/location-hash.ts";

/**
 * 選択中 id ↔ `location.hash` の双方向同期を担う副作用 hook（routing library の代替）。
 *
 * @require selectedId は同期対象。文字列なら hash をその値へ、null なら hash をクリアする。
 *   undefined は「まだ同期を始めない（未初期化）」を意味し、hash は一切書き換えない
 *   （null=「選択なし」と undefined=「同期保留」を区別する。多義ではなく sync primitive の2状態）
 * @require onHashSelect は hashchange で hash が変わった時に呼ばれる。hash が空なら null、
 *   それ以外はその文字列を受け取る。何を select するかの解釈は caller が持つ
 * @ensure selectedId 変化のたび hash を追従させる（undefined の間は追従しない）。hashchange 購読を
 *   mount 時 1 回張り、unmount で外す。hashchange 由来で caller が selectedId を同じ値へ更新し
 *   hash へ書き戻しても、直前に同期した値と一致すれば書き込まない（無限ループ抑止）
 * @invariant throw しない。表示・JSX を持たない。select の呼び出しや toggle 解除の解釈は行わない
 */
export function useHashSync(
  selectedId: string | null | undefined,
  onHashSelect: (id: string | null) => void,
): void {
  // why: hashchange 由来の onHashSelect が selectedId を更新し、そのまま同じ値を hash へ書き戻すと
  //   無限ループになりうる。書き込み直前の値と比較して抑止する
  const lastSyncedHashRef = useRef<string>(getLocationHash());

  // why: onHashSelect は caller 側で毎 render 生成されうる。listener を貼り替えず最新を呼ぶための可変参照
  const onHashSelectRef = useRef(onHashSelect);
  onHashSelectRef.current = onHashSelect;

  useEffect(() => {
    if (selectedId === undefined) {
      return;
    }
    const nextHash = selectedId ?? "";
    if (nextHash !== lastSyncedHashRef.current) {
      lastSyncedHashRef.current = nextHash;
      setLocationHash(nextHash);
    }
  }, [selectedId]);

  useEffect(() => {
    return onLocationHashChange(() => {
      const hash = getLocationHash();
      if (hash === lastSyncedHashRef.current) {
        return;
      }
      lastSyncedHashRef.current = hash;
      onHashSelectRef.current(hash === "" ? null : hash);
    });
  }, []);
}
