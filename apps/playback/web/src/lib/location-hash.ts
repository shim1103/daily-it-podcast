/**
 * `window.location.hash` から先頭の `#` を除いた値を返す（Driven Adapter）。
 *
 * @ensure hash が無い時は空文字を返す
 */
export function getLocationHash(): string {
  return window.location.hash.replace(/^#/, "");
}

/**
 * `window.location.hash` へ値を設定する（Driven Adapter）。
 *
 * @require value は `#` を含まない
 * @ensure value が空文字の時、hash を消す
 */
export function setLocationHash(value: string): void {
  window.location.hash = value;
}

/**
 * `hashchange` event を購読する（Driven Adapter）。
 *
 * @ensure hash 変化のたび listener を呼ぶ戻り値の関数を呼ぶと購読を解除する
 */
export function onLocationHashChange(listener: () => void): () => void {
  window.addEventListener("hashchange", listener);
  return () => window.removeEventListener("hashchange", listener);
}
