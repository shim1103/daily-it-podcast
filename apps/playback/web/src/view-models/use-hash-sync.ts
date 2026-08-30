import { useEffect, useRef, useSyncExternalStore } from "react";
import { getLocationHash, onLocationHashChange, setLocationHash } from "../lib/location-hash.ts";

/**
 * 選択中 id ↔ `location.hash` の双方向同期を担う副作用 hook（routing library の代替）。
 *
 * @require selectedId は同期対象。文字列なら hash をその値へ、null なら hash をクリアする。
 *   undefined は「まだ同期を始めない（未初期化）」を意味し、hash は一切書き換えない
 *   （null=「選択なし」と undefined=「同期保留」を区別する。多義ではなく sync primitive の2状態）
 * @require onHashSelect は hashchange で hash が変わった時に呼ばれる。hash が空なら null、
 *   それ以外はその文字列を受け取る。何を select するかの解釈は caller が持つ
 * @ensure hash の現在値は `useSyncExternalStore` で購読し、その購読解除で unmount 時に hashchange
 *   listener を外す。`useSyncExternalStore` が報告する hash が直近同期値と異なる時に onHashSelect を
 *   呼ぶ（初回 render 時点の既存 hash と、本 hook 自身の書き込みが誘発した hashchange＝echo は流さない）。
 *   selectedId 変化のたび hash を追従させる（undefined の間は追従しない）。hashchange 由来で caller が
 *   selectedId を同じ値へ更新し hash へ書き戻しても、直近同期値と `selectedId ?? ""` が一致すれば
 *   書き込まない（無限ループ抑止）
 * @invariant throw しない。表示・JSX を持たない。select の呼び出しや toggle 解除の解釈は行わない
 */
export function useHashSync(
  selectedId: string | null | undefined,
  onHashSelect: (id: string | null) => void,
): void {
  // why: 外部ストア（location.hash）の現在値。getLocationHash は `#` を除いた文字列を返すため
  //   `useSyncExternalStore` の Object.is 比較で安定し、購読解除が unmount 時に走る
  const currentHash = useSyncExternalStore(onLocationHashChange, getLocationHash);

  // why: 直近に本 hook が同期した hash 値。読み取り差分（報告 hash が外部由来の変化か echo か）と
  //   書き込み差分（selectedId 追従が必要か）の双方をこの 1 値で判定する。初期値を現在 hash に
  //   することで、初回 render 時点の既存 hash を onHashSelect へ流さない
  const syncedHashRef = useRef<string>(currentHash);

  // why: 読み取り方向。報告 hash が直近同期値と異なる時だけ onHashSelect へ流す（一致は初回既存
  //   hash と自分の書き込み echo なので弾く）。この effect を書き込み方向より先に宣言し、両者が同一
  //   commit で発火する時に読み取り→書き込みの順を固定する。逆順だと、selectedId と onHashSelect が
  //   同じ commit で変化し currentHash 未追従の時、書き込みが先に syncedHashRef を進め、続く読み取り
  //   が旧 hash を「外部変化」と誤検知して onHashSelect(旧hash) を呼ぶ（episode-list-page の
  //   onHashSelect は [select, state] 依存なので state 変化のたびこの interleaving が起こりうる）
  useEffect(() => {
    if (currentHash === syncedHashRef.current) {
      return;
    }
    syncedHashRef.current = currentHash;
    onHashSelect(currentHash === "" ? null : currentHash);
  }, [currentHash, onHashSelect]);

  // why: 書き込み方向。selectedId 変化のたび hash を追従させる。直近同期値と一致すれば
  //   書き込まない（hashchange 由来の書き戻しによる無限ループ抑止）
  useEffect(() => {
    if (selectedId === undefined) {
      return;
    }
    const nextHash = selectedId ?? "";
    if (nextHash !== syncedHashRef.current) {
      syncedHashRef.current = nextHash;
      setLocationHash(nextHash);
    }
  }, [selectedId]);
}
