import { describe, expect, it } from "vitest";
import { ProgressWriteResponseSchema } from "../../../../contracts/index.ts";
import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import { createFakeProgressRepository } from "../ports/progress-repository.fake.ts";
import type { ProgressUpsertRow } from "../ports/progress-repository.ts";
import type { ProgressWriteCommand } from "../progress/progress-write-command.ts";
import { updateProgress } from "./update-progress.ts";

/**
 * scope: Sociable Unit
 * real: updateProgress use-case（行なし 404・merge）
 * double: createFakeProgressRepository
 */
const EARLIER = "2026-09-22T10:00:00.000Z";
const LATER = "2026-09-22T11:00:00.000Z";

const command: ProgressWriteCommand = {
  episodeId: "ep-1",
  positionSec: 40,
  clientAt: LATER,
};

describe("updateProgress", () => {
  it("行なしの時、ProgressNotFoundError を throw する", async () => {
    // Given: 空 Fake
    const repository = createFakeProgressRepository();

    // When: update を実行する
    const act = updateProgress(repository, command);

    // Then: 行不在の Domain Error（永続は触らない）
    await expect(act).rejects.toBeInstanceOf(ProgressNotFoundError);
    const stored = await repository.getByEpisodeIds(["ep-1"]);
    expect(stored.size).toBe(0);
  });

  it("行ありの時、merge して upsert し勝ち側 first* を返す", async () => {
    // Given: 既存行あり
    const seed: ProgressUpsertRow = {
      episodeId: "ep-1",
      positionSec: 10,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    };
    const repository = createFakeProgressRepository([seed]);

    // When: より遅い clientAt で update する
    const got = await updateProgress(repository, command);

    // Then: position/last は後勝ち。first* は既存据え置き
    expect(ProgressWriteResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual({ firstPlayedAt: EARLIER, firstCompletedAt: null });
    const stored = await repository.getByEpisodeIds(["ep-1"]);
    expect(stored.get("ep-1")).toEqual({
      positionSec: 40,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: LATER,
    });
  });
});
