import { describe, expect, it } from "vitest";
import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import type { ProgressRow } from "../ports/progress-repository.ts";
import type { ProgressWriteUseCaseInput } from "../progress/progress-write-use-case-io.ts";
import { updateProgress } from "./update-progress.ts";

/**
 * scope: Sociable Unit
 * real: updateProgress use-case（行なし 404・merge）
 * double: InMemoryProgressRepository (Fake)
 */
const EARLIER = "2026-09-22T10:00:00.000Z";
const LATER = "2026-09-22T11:00:00.000Z";

const input: ProgressWriteUseCaseInput = {
  episodeId: "ep-1",
  positionSec: 40,
  clientAt: LATER,
};

describe("updateProgress", () => {
  it("行なしの時、ProgressNotFoundError を throw する", async () => {
    // Given: 空 Fake
    const repository = new InMemoryProgressRepository();

    // When: update を実行する
    const act = updateProgress(repository, input);

    // Then: 行不在の Domain Error（永続は触らない）
    await expect(act).rejects.toBeInstanceOf(ProgressNotFoundError);
    const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.size).toBe(0);
  });

  it("行ありの時、merge して upsert し勝ち側 first* を返す", async () => {
    // Given: 既存行あり
    const seed: ProgressRow = {
      episodeId: "ep-1",
      positionSec: 10,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    };
    const repository = new InMemoryProgressRepository([seed]);

    // When: より遅い clientAt で update する
    const got = await updateProgress(repository, input);

    // Then: position/last は後勝ち。first* は既存据え置き
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
