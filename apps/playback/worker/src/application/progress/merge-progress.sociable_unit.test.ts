import { describe, expect, it } from "vitest";
import type { EpisodeProgress } from "../../../../contracts/index.ts";
import { mergeProgress, mergeProgressAsCompleted } from "./merge-progress.ts";

const EARLIER = "2026-09-22T10:00:00.000Z";
const LATER = "2026-09-22T11:00:00.000Z";
const SAME = "2026-09-22T10:30:00.000Z";

function progress(overrides: Partial<EpisodeProgress> = {}): EpisodeProgress {
  return {
    positionSec: 10,
    firstPlayedAt: EARLIER,
    firstCompletedAt: null,
    lastPlayedAt: EARLIER,
    ...overrides,
  };
}

describe("mergeProgress", () => {
  it("existing が null の時、初回行を clientAt と positionSec で作る", () => {
    // Given: 行なし
    // When: merge する
    const got = mergeProgress(null, { positionSec: 12, clientAt: EARLIER });

    // Then: first* / last* は clientAt、完走は null
    expect(got).toEqual({
      positionSec: 12,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    });
  });

  it("incoming clientAt が既存 lastPlayedAt より遅い時、positionSec と lastPlayedAt を後勝ちする", () => {
    // Given: 既存より遅い write
    const existing = progress({ positionSec: 10, lastPlayedAt: EARLIER, firstPlayedAt: EARLIER });

    // When: merge する
    const got = mergeProgress(existing, { positionSec: 40, clientAt: LATER });

    // Then: 後勝ち字段だけ差し替え
    expect(got.positionSec).toBe(40);
    expect(got.lastPlayedAt).toBe(LATER);
    expect(got.firstPlayedAt).toBe(EARLIER);
    expect(got.firstCompletedAt).toBeNull();
  });

  it("incoming clientAt が既存 lastPlayedAt より早い時、positionSec と lastPlayedAt を既存のまま残す", () => {
    // Given: 遅延到着の早い write
    const existing = progress({ positionSec: 40, lastPlayedAt: LATER, firstPlayedAt: EARLIER });

    // When: merge する
    const got = mergeProgress(existing, { positionSec: 5, clientAt: EARLIER });

    // Then: 後勝ち字段は負けて既存のまま
    expect(got.positionSec).toBe(40);
    expect(got.lastPlayedAt).toBe(LATER);
  });

  it("incoming clientAt が既存 firstPlayedAt より早い時、firstPlayedAt を先勝ちで置き換える", () => {
    // Given: 後から届いたより早い初回
    const existing = progress({
      positionSec: 40,
      firstPlayedAt: LATER,
      lastPlayedAt: LATER,
    });

    // When: merge する
    const got = mergeProgress(existing, { positionSec: 5, clientAt: EARLIER });

    // Then: firstPlayedAt だけ早い側へ。後勝ち字段は既存のまま
    expect(got.firstPlayedAt).toBe(EARLIER);
    expect(got.positionSec).toBe(40);
    expect(got.lastPlayedAt).toBe(LATER);
  });

  it("既存 firstCompletedAt がある時、完走時刻を変えない", () => {
    // Given: 既完走行への通常 update
    const existing = progress({ firstCompletedAt: EARLIER, lastPlayedAt: EARLIER });

    // When: merge する
    const got = mergeProgress(existing, { positionSec: 20, clientAt: LATER });

    // Then: 完走時刻は触らない
    expect(got.firstCompletedAt).toBe(EARLIER);
  });

  it("clientAt が既存 firstPlayedAt と同じ時、firstPlayedAt は既存を残す", () => {
    // why: 同時刻は先勝ちの「既に勝っている側」を保つ（再書込で揺らさない）
    // Given: first* 同時刻
    const existing = progress({
      firstPlayedAt: SAME,
      lastPlayedAt: SAME,
      positionSec: 10,
    });

    // When: merge する
    const got = mergeProgress(existing, { positionSec: 99, clientAt: SAME });

    // Then: first* は既存据え置き
    expect(got.firstPlayedAt).toBe(SAME);
  });

  it("clientAt が既存 lastPlayedAt と同じ時、positionSec と lastPlayedAt は incoming を採る", () => {
    // why: 同時刻の後勝ちは「今届いた共有カーソル」を優先し、再送で位置を揃えられるようにする
    // Given: last-wins 同時刻
    const existing = progress({
      firstPlayedAt: EARLIER,
      lastPlayedAt: SAME,
      positionSec: 10,
    });

    // When: merge する
    const got = mergeProgress(existing, { positionSec: 99, clientAt: SAME });

    // Then: 後勝ち字段は incoming
    expect(got.positionSec).toBe(99);
    expect(got.lastPlayedAt).toBe(SAME);
  });
});

describe("mergeProgressAsCompleted", () => {
  it("既存 firstCompletedAt が null の時、clientAt を入れる", () => {
    // Given: 未完走行への完走 write
    const existing = progress({ firstCompletedAt: null, lastPlayedAt: EARLIER });

    // When: 完走 merge する
    const got = mergeProgressAsCompleted(existing, { positionSec: 57, clientAt: LATER });

    // Then: 初回完走を記録し、後勝ち字段も更新
    expect(got.firstCompletedAt).toBe(LATER);
    expect(got.positionSec).toBe(57);
    expect(got.lastPlayedAt).toBe(LATER);
  });

  it("既存 firstCompletedAt より早い clientAt の時、firstCompletedAt を先勝ちで置き換える", () => {
    // Given: 既完走だがより早い完走時刻が遅延到着
    const existing = progress({
      firstCompletedAt: LATER,
      firstPlayedAt: EARLIER,
      lastPlayedAt: LATER,
      positionSec: 58,
    });

    // When: 完走 merge する
    const got = mergeProgressAsCompleted(existing, { positionSec: 57, clientAt: EARLIER });

    // Then: firstCompletedAt は早い側。後勝ち字段は既存のまま
    expect(got.firstCompletedAt).toBe(EARLIER);
    expect(got.positionSec).toBe(58);
    expect(got.lastPlayedAt).toBe(LATER);
  });

  it("clientAt が既存 firstCompletedAt と同じ時、firstCompletedAt は既存を残す", () => {
    // why: 同時刻の first* は既存据え置き（firstPlayedAt と同規則）
    // Given: 完走同時刻
    const existing = progress({
      firstCompletedAt: SAME,
      firstPlayedAt: EARLIER,
      lastPlayedAt: SAME,
      positionSec: 50,
    });

    // When: 完走 merge する
    const got = mergeProgressAsCompleted(existing, { positionSec: 57, clientAt: SAME });

    // Then: firstCompletedAt は既存据え置き
    expect(got.firstCompletedAt).toBe(SAME);
  });
});
