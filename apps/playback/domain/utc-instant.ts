/**
 * UTC 固定幅（`YYYY-MM-DDTHH:mm:ss.sssZ`）で表した時刻。文字列の大小が時系列と一致する。
 * `toUtcInstant` を通った値だけがこの型になる。
 */
export type UtcInstant = string & { readonly __brand: "UtcInstant" };

const lowerExclusiveMs = Date.UTC(1970, 0, 1);
const upperExclusiveMs = Date.UTC(2100, 0, 1);

/**
 * instant を UTC 固定幅へ変換する。入力が `Date` なので、文字列の文法は知らない
 * （offset 付き ISO の受理は契約の責務）。範囲は固定幅 ISO の前提（拡張年表記を避ける）を守るためである。
 *
 * @ensure 1970-01-01T00:00:00Z より後、2100-01-01T00:00:00Z より前の instant だけを返す
 * @ensure Invalid Date と範囲外のとき undefined
 */
export function toUtcInstant(date: Date): UtcInstant | undefined {
  const instantMs = date.getTime();
  // why: Invalid Date は NaN で、どちらの比較も偽になるので、この 1 判定が範囲外と無効の両方を弾く
  if (!(instantMs > lowerExclusiveMs && instantMs < upperExclusiveMs)) {
    return undefined;
  }
  // why: brand を付ける唯一の入口。範囲を上で判定した instant だけが UtcInstant になる
  return date.toISOString() as UtcInstant;
}

/**
 * すでに正規形（UTC 固定幅・範囲内）の `UtcInstant` か。
 * 保存先から読んだ値など、正規化済みのはずの文字列を型へ戻す境界で使う。
 *
 * @invariant 往復で一致した時だけ true。`Date` が繰り上げる暦日や、offset・小数秒の違う表記は false
 */
export function isUtcInstant(value: string): value is UtcInstant {
  return toUtcInstant(new Date(value)) === value;
}
