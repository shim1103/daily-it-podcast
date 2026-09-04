import { describe, expect, it } from "vitest";
import { ManuscriptSchema } from "./manuscript-schema.ts";

const validManuscript = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: {
      text: "開始",
      startSec: 0,
    },
    topics: [
      {
        title: "題",
        preface: "前置き",
        detail: "詳細",
        startSec: 0,
      },
    ],
    ending: {
      text: "終了",
      startSec: 30,
    },
  },
};

describe("ManuscriptSchema", () => {
  it("Drive 原稿の必須 field が揃っている時、success と data を返す", () => {
    // Given: repo 根 manuscript schema に適合する原稿
    // When: 検証する
    const got = ManuscriptSchema.safeParse(validManuscript);

    // Then: 適合し、入力と同形の data を返す
    expect(got.success).toBe(true);
    if (!got.success) {
      return;
    }
    expect(got.data).toEqual(validManuscript);
  });

  it("必須 field が欠けている時、success: false を返す", () => {
    // Given: episodeId だけの不完全 JSON
    // When: 検証する
    const got = ManuscriptSchema.safeParse({ episodeId: "ep-1" });

    // Then: 不適合
    expect(got.success).toBe(false);
  });

  it("HTTP 専用 field の audioRef がある時、success: false を返す", () => {
    // Given: Drive 原稿に無い audioRef を付けた JSON（HTTP EpisodeItem 形）
    // When: 検証する
    const got = ManuscriptSchema.safeParse({
      ...validManuscript,
      audioRef: "/episodes/ep-1/audio",
    });

    // Then: additionalProperties で落ち、HTTP schema の omit ではないことを示す
    expect(got.success).toBe(false);
  });

  it("opening が旧来の文字列の時、success: false を返す", () => {
    // Given: opening が { text, startSec } でなく文字列（拡張前の形）
    // When: 検証する
    const got = ManuscriptSchema.safeParse({
      ...validManuscript,
      body: { ...validManuscript.body, opening: "開始" },
    });

    // Then: 不適合
    expect(got.success).toBe(false);
  });

  it("opening.startSec が欠けている時、success: false を返す", () => {
    // Given: opening.text だけの原稿
    // When: 検証する
    const got = ManuscriptSchema.safeParse({
      ...validManuscript,
      body: { ...validManuscript.body, opening: { text: "開始" } },
    });

    // Then: 不適合
    expect(got.success).toBe(false);
  });

  it("ending が旧来の文字列の時、success: false を返す", () => {
    // Given: ending が { text, startSec } でなく文字列（拡張前の形）
    // When: 検証する
    const got = ManuscriptSchema.safeParse({
      ...validManuscript,
      body: { ...validManuscript.body, ending: "終了" },
    });

    // Then: 不適合
    expect(got.success).toBe(false);
  });

  it("ending.startSec が欠けている時、success: false を返す", () => {
    // Given: ending.text だけの原稿
    // When: 検証する
    const got = ManuscriptSchema.safeParse({
      ...validManuscript,
      body: { ...validManuscript.body, ending: { text: "終了" } },
    });

    // Then: 不適合
    expect(got.success).toBe(false);
  });

  it("null や非 object は success: false を返す", () => {
    // Given: schema 外の入力
    // When / Then
    expect(ManuscriptSchema.safeParse(null).success).toBe(false);
    expect(ManuscriptSchema.safeParse("ep-1").success).toBe(false);
  });
});
