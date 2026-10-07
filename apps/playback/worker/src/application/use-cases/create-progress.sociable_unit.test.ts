import { describe, expect, it } from "vitest";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import type { ProgressRow } from "../ports/progress-repository.ts";
import type { ProgressWriteUseCaseInput } from "../progress/progress-write-use-case-io.ts";
import { createProgress } from "./create-progress.ts";

/**
 * scope: Sociable Unit
 * real: createProgress use-case（merge・冪等）
 * double: InMemoryProgressRepository (Fake)
 */
const EARLIER = "2026-09-22T10:00:00.000Z";
const LATER = "2026-09-22T11:00:00.000Z";

const input: ProgressWriteUseCaseInput = {
  episodeId: "ep-1",
  positionSec: 12,
  clientAt: EARLIER,
};

describe("createProgress", () => {
  it("行なしの時、merge で初回行を作り first* を返す", async () => {
    // Given: 空 Fake
    const repository = new InMemoryProgressRepository();

    // When: create を実行する
    const got = await createProgress(repository, input);

    // Then: firstPlayedAt は clientAt、完走は null。永続も同値
    expect(got).toEqual({ firstPlayedAt: EARLIER, firstCompletedAt: null });
    const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.get("ep-1")).toEqual({
      positionSec: 12,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    });
  });

  it("行ありの時、Error にせず merge し勝ち側 first* を返す", async () => {
    // Given: 既存行あり（後から届くより早い初回）
    const seed: ProgressRow = {
      episodeId: "ep-1",
      positionSec: 40,
      firstPlayedAt: LATER,
      firstCompletedAt: null,
      lastPlayedAt: LATER,
    };
    const repository = new InMemoryProgressRepository([seed]);
    const lateCreate: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 5,
      clientAt: EARLIER,
    };

    // When: 既存行へ create する
    const got = await createProgress(repository, lateCreate);

    // Then: 冪等成功。firstPlayedAt は先勝ち、position/last は後勝ちで既存のまま
    expect(got).toEqual({ firstPlayedAt: EARLIER, firstCompletedAt: null });
    const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.get("ep-1")).toEqual({
      positionSec: 40,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: LATER,
    });
  });
});
