import { describe, expect, it } from "vitest";
import { isUtcInstant, normalizeUtcInstant } from "./utc-instant.ts";

describe("normalizeUtcInstant", () => {
  it("UTC 固定幅の入力はそのまま返す", () => {
    // Given: すでに固定幅の UTC 表記
    const input = "2026-10-01T09:00:00.000Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 同じ値
    expect(got).toBe("2026-10-01T09:00:00.000Z");
  });

  it("小数秒が無い入力は .000 を補う", () => {
    // Given: 小数秒の無い Z 表記
    const input = "2026-10-01T09:00:00Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: ミリ秒 3 桁に揃う
    expect(got).toBe("2026-10-01T09:00:00.000Z");
  });

  it("秒が無い入力は 0 秒として補う", () => {
    // Given: 時分までの入力（既存の HTTP 契約が受理する形）
    const input = "2026-10-01T00:00Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 秒・ミリ秒が 0 で補われる
    expect(got).toBe("2026-10-01T00:00:00.000Z");
  });

  it("秒が無いのに小数だけある入力は undefined を返す", () => {
    // Given: 時分に小数が付いた入力
    const input = "2026-10-01T00:00.5Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 文法に合わない
    expect(got).toBeUndefined();
  });

  it("小数秒が 6 桁の入力はミリ秒へ切り捨てて丸めない", () => {
    // Given: 丸めると桁上がりする小数部 6 桁
    const input = "2026-10-01T00:00:00.999999Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 999 ミリ秒
    expect(got).toBe("2026-10-01T00:00:00.999Z");
  });

  it("小数秒が 3 桁に満たない入力は右を 0 で埋める", () => {
    // Given: 小数秒 1 桁
    const input = "2026-10-01T09:00:00.5Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 500 ミリ秒
    expect(got).toBe("2026-10-01T09:00:00.500Z");
  });

  it("正の offset は同じ instant の UTC 表記へ換算する", () => {
    // Given: +09:00 の入力
    const input = "2026-10-01T18:00:00+09:00";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 9 時間戻した UTC
    expect(got).toBe("2026-10-01T09:00:00.000Z");
  });

  it("負の offset は同じ instant の UTC 表記へ換算する", () => {
    // Given: -05:00 の入力
    const input = "2026-10-01T04:00:00-05:00";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 5 時間進めた UTC
    expect(got).toBe("2026-10-01T09:00:00.000Z");
  });

  it("下限（1970-01-01T00:00:00Z）ちょうどは範囲外として undefined を返す", () => {
    // Given: 排他の下限
    const input = "1970-01-01T00:00:00.000Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 範囲外
    expect(got).toBeUndefined();
  });

  it("下限の 1 ミリ秒後は通す", () => {
    // Given: 下限の直後
    const input = "1970-01-01T00:00:00.001Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: そのまま返る
    expect(got).toBe("1970-01-01T00:00:00.001Z");
  });

  it("offset 換算後に下限ちょうどになる入力は undefined を返す", () => {
    // Given: 換算すると epoch ちょうどになる +09:00 表記
    const input = "1970-01-01T09:00:00+09:00";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 範囲外
    expect(got).toBeUndefined();
  });

  it("上限の直前（2099-12-31T23:59:59.999Z）は通す", () => {
    // Given: 上限の 1 ミリ秒前
    const input = "2099-12-31T23:59:59.999Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: そのまま返る
    expect(got).toBe("2099-12-31T23:59:59.999Z");
  });

  it("上限（2100-01-01T00:00:00Z）ちょうどは範囲外として undefined を返す", () => {
    // Given: 排他の上限
    const input = "2100-01-01T00:00:00.000Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 範囲外
    expect(got).toBeUndefined();
  });

  it("offset が無い入力は実行環境の時刻帯に依らず undefined を返す", () => {
    // Given: Z も offset も無い入力
    const input = "2026-10-01T09:00:00";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 解釈しない
    expect(got).toBeUndefined();
  });

  it("日付だけの入力は undefined を返す", () => {
    // Given: 時刻の無い入力
    const input = "2026-10-01";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 解釈しない
    expect(got).toBeUndefined();
  });

  it("解釈できない文字列と空文字は undefined を返す", () => {
    // Given: ISO でない入力
    const inputs = ["not-a-date", ""];

    // When: 正規化する
    const got = inputs.map(normalizeUtcInstant);

    // Then: どちらも undefined
    expect(got).toEqual([undefined, undefined]);
  });

  it("存在しない暦日は繰り上げずに undefined を返す", () => {
    // Given: 2026 年 2 月 30 日
    const input = "2026-02-30T00:00:00.000Z";

    // When: 正規化する
    const got = normalizeUtcInstant(input);

    // Then: 3 月へ繰り上げない
    expect(got).toBeUndefined();
  });

  it("24 時・60 秒など時刻部の範囲外は繰り上げずに undefined を返す", () => {
    // Given: 24:00:00 と 00:00:60
    const inputs = ["2026-10-01T24:00:00.000Z", "2026-10-01T00:00:60.000Z"];

    // When: 正規化する
    const got = inputs.map(normalizeUtcInstant);

    // Then: どちらも undefined
    expect(got).toEqual([undefined, undefined]);
  });

  it("offset の時が 23 を、分が 59 を超える入力は undefined を返す", () => {
    // Given: +24:00 と +09:60
    const inputs = ["2026-10-01T09:00:00+24:00", "2026-10-01T09:00:00+09:60"];

    // When: 正規化する
    const got = inputs.map(normalizeUtcInstant);

    // Then: どちらも undefined
    expect(got).toEqual([undefined, undefined]);
  });
});

describe("isUtcInstant", () => {
  it("正規形（UTC 固定幅・範囲内）の時 true を返す", () => {
    // Given: 正規化済みの値
    const value = "2026-10-01T09:00:00.000Z";

    // When: 判定する
    const got = isUtcInstant(value);

    // Then: true
    expect(got).toBe(true);
  });

  it("offset 付き・小数秒なしなど正規形でない表記は false を返す", () => {
    // Given: 同じ instant を指すが正規形でない表記
    const values = ["2026-10-01T18:00:00.000+09:00", "2026-10-01T09:00:00Z"];

    // When: 判定する
    const got = values.map(isUtcInstant);

    // Then: どちらも false
    expect(got).toEqual([false, false]);
  });

  it("範囲外・暦が無い・空文字は false を返す", () => {
    // Given: 範囲外、存在しない暦日、空文字
    const values = ["1970-01-01T00:00:00.000Z", "2026-02-30T00:00:00.000Z", ""];

    // When: 判定する
    const got = values.map(isUtcInstant);

    // Then: すべて false
    expect(got).toEqual([false, false, false]);
  });
});
