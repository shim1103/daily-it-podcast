/**
 * UTC 固定幅（`YYYY-MM-DDTHH:mm:ss.sssZ`）で表した時刻。文字列の大小が時系列と一致する。
 * `normalizeUtcInstant` を通った値だけがこの型になる。
 */
export type UtcInstant = string & { readonly __brand: "UtcInstant" };

// what: offset 付き ISO-8601。秒は省略でき、小数秒は秒があるときだけ任意桁。offset は `Z` か `±hh:mm` が必須
const instantGrammar =
  /^(\d{4})-(\d{2})-(\d{2})T(\d{2}):(\d{2})(?::(\d{2})(?:\.(\d+))?)?(?:Z|([+-])(\d{2}):(\d{2}))$/;

const lowerExclusiveMs = Date.UTC(1970, 0, 1);
const upperExclusiveMs = Date.UTC(2100, 0, 1);

const MS_PER_MINUTE = 60_000;

/**
 * offset 付き ISO-8601 を、同じ instant の UTC 固定幅へ正規化する。
 *
 * 入力の文法・暦の存在・instant の範囲（1970-01-01T00:00:00Z より後、2100-01-01T00:00:00Z より前）を
 * 判定する。範囲は固定幅 ISO の前提（拡張年表記を避ける）を守るためである。
 *
 * @ensure 解釈できない・offset が無い・存在しない暦日・時刻部や offset の範囲外・instant が範囲外のとき undefined
 * @ensure 小数秒はミリ秒に切り捨てる
 * @invariant 実行環境の時刻帯に依らない（`Date.parse` の local 解釈を使わない）
 */
export function normalizeUtcInstant(input: string): UtcInstant | undefined {
  const match = instantGrammar.exec(input);
  if (match === null) {
    return undefined;
  }
  const [, year, month, day, hour, minute, second, fraction, sign, offsetHour, offsetMinute] =
    match;

  const wallMs = Date.UTC(
    Number(year),
    Number(month) - 1,
    Number(day),
    Number(hour),
    Number(minute),
    Number(second ?? "0"),
    Number((fraction ?? "").padEnd(3, "0").slice(0, 3)),
  );
  // why: `Date.UTC` は 2 月 30 日や 24 時を黙って繰り上げる。入力の日時部と往復で一致するかで弾く
  const wallText = `${year}-${month}-${day}T${hour}:${minute}:${second ?? "00"}`;
  if (new Date(wallMs).toISOString().slice(0, 19) !== wallText) {
    return undefined;
  }

  let offsetMs = 0;
  if (sign !== undefined) {
    if (Number(offsetHour) > 23 || Number(offsetMinute) > 59) {
      return undefined;
    }
    const magnitudeMs = (Number(offsetHour) * 60 + Number(offsetMinute)) * MS_PER_MINUTE;
    offsetMs = sign === "-" ? -magnitudeMs : magnitudeMs;
  }

  const instantMs = wallMs - offsetMs;
  if (instantMs <= lowerExclusiveMs || instantMs >= upperExclusiveMs) {
    return undefined;
  }
  // why: brand を付ける唯一の入口。文法・暦・範囲を上で判定した値だけが UtcInstant になる
  return new Date(instantMs).toISOString() as UtcInstant;
}

/**
 * すでに正規形（UTC 固定幅・範囲内）の `UtcInstant` か。
 * 保存先から読んだ値など、正規化済みのはずの文字列を型へ戻す境界で使う。
 */
export function isUtcInstant(value: string): value is UtcInstant {
  return normalizeUtcInstant(value) === value;
}
