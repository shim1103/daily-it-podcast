/**
 * wire の日付（YYYY-MM-DD）を表示用（YYYY/MM/DD）へ変換する。
 *
 * @require date は YYYY-MM-DD
 * @ensure ハイフンをスラッシュへ置き換えた文字列を返す
 * @invariant 副作用なし。外部依存なし
 */
export function formatEpisodeDate(date: string): string {
  return date.replaceAll("-", "/");
}
