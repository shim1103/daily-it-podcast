import { describe, expect, it } from "vitest";
import { isUtcInstant, toUtcInstant } from "./utc-instant.ts";

describe("toUtcInstant", () => {
  it("範囲内の instant を UTC 固定幅の文字列にする", () => {
    // Given: ミリ秒つきの instant
    const date = new Date("2026-10-01T09:00:00.123Z");

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: ミリ秒まで固定幅
    expect(got).toBe("2026-10-01T09:00:00.123Z");
  });

  it("offset 付きで作った Date も同じ instant の UTC 表記にする", () => {
    // Given: +09:00 で作った Date
    const date = new Date("2026-10-01T18:00:00+09:00");

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: UTC へ換算される
    expect(got).toBe("2026-10-01T09:00:00.000Z");
  });

  it("下限（1970-01-01T00:00:00Z）ちょうどは範囲外として undefined を返す", () => {
    // Given: 排他の下限
    const date = new Date(0);

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: 範囲外
    expect(got).toBeUndefined();
  });

  it("下限の 1 ミリ秒後は通す", () => {
    // Given: 下限の直後
    const date = new Date(1);

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: 固定幅で返る
    expect(got).toBe("1970-01-01T00:00:00.001Z");
  });

  it("上限の直前（2099-12-31T23:59:59.999Z）は通す", () => {
    // Given: 上限の 1 ミリ秒前
    const date = new Date("2099-12-31T23:59:59.999Z");

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: そのまま返る
    expect(got).toBe("2099-12-31T23:59:59.999Z");
  });

  it("上限（2100-01-01T00:00:00Z）ちょうどは範囲外として undefined を返す", () => {
    // Given: 排他の上限
    const date = new Date("2100-01-01T00:00:00.000Z");

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: 範囲外
    expect(got).toBeUndefined();
  });

  it("下限より前の instant は undefined を返す", () => {
    // Given: 1969 年の instant
    const date = new Date("1969-12-31T23:59:59.999Z");

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: 範囲外
    expect(got).toBeUndefined();
  });

  it("Invalid Date は undefined を返す", () => {
    // Given: 解釈できなかった Date
    const date = new Date("not-a-date");

    // When: 変換する
    const got = toUtcInstant(date);

    // Then: 時刻にならない
    expect(got).toBeUndefined();
  });
});

describe("isUtcInstant", () => {
  it("正規形（UTC 固定幅・範囲内）の時 true を返す", () => {
    // Given: 正規形の値
    const value = "2026-10-01T09:00:00.000Z";

    // When: 判定する
    const got = isUtcInstant(value);

    // Then: true
    expect(got).toBe(true);
  });

  it("同じ instant でも正規形でない表記は false を返す", () => {
    // Given: offset 付き、小数秒なし、小数秒が 3 桁でない表記
    const values = [
      "2026-10-01T18:00:00.000+09:00",
      "2026-10-01T09:00:00Z",
      "2026-10-01T09:00:00.5Z",
      "2026-10-01T09:00:00.1234Z",
    ];

    // When: 判定する
    const got = values.map(isUtcInstant);

    // Then: すべて false
    expect(got).toEqual([false, false, false, false]);
  });

  it("offset が無い・日付だけの表記は実行環境の時刻帯に依らず false を返す", () => {
    // Given: Z も offset も無い表記と、日付だけの表記
    const values = ["2026-10-01T09:00:00.000", "2026-10-01"];

    // When: 判定する
    const got = values.map(isUtcInstant);

    // Then: どちらも false
    expect(got).toEqual([false, false]);
  });

  it("範囲外・存在しない暦日・解釈できない文字列・空文字は false を返す", () => {
    // Given: 下限ちょうど、2 月 30 日、ISO でない文字列、空文字
    const values = ["1970-01-01T00:00:00.000Z", "2026-02-30T00:00:00.000Z", "not-a-date", ""];

    // When: 判定する
    const got = values.map(isUtcInstant);

    // Then: すべて false（2 月 30 日は 3 月へ繰り上がるので往復で一致しない）
    expect(got).toEqual([false, false, false, false]);
  });
});
