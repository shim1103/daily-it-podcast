import { describe, expect, it } from "vitest";
import { ProgressWriteResponseSchema } from "../../../../contracts/index.ts";
import { createFakeProgressRepository } from "../ports/progress-repository.fake.ts";
import type { ProgressUpsertRow } from "../ports/progress-repository.ts";
import type { ProgressWriteCommand } from "../progress/progress-write-command.ts";
import { createProgress } from "./create-progress.ts";

/**
 * scope: Sociable Unit
 * real: createProgress use-case（merge・冪等）
 * double: createFakeProgressRepository
 */
const EARLIER = "2026-09-22T10:00:00.000Z";
const LATER = "2026-09-22T11:00:00.000Z";

const command: ProgressWriteCommand = {
  episodeId: "ep-1",
  positionSec: 12,
  clientAt: EARLIER,
};

describe("createProgress", () => {
  it("行なしの時、merge で初回行を作り first* を返す", async () => {
    // Given: 空 Fake
    const repository = createFakeProgressRepository();

    // When: create を実行する
    const got = await createProgress(repository, command);

    // Then: firstPlayedAt は clientAt、完走は null。永続も同値
    expect(ProgressWriteResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual({ firstPlayedAt: EARLIER, firstCompletedAt: null });
    const stored = await repository.getByEpisodeIds(["ep-1"]);
    expect(stored.get("ep-1")).toEqual({
      positionSec: 12,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    });
  });

  it("行ありの時、Error にせず merge し勝ち側 first* を返す", async () => {
    // Given: 既存行あり（後から届くより早い初回）
    const seed: ProgressUpsertRow = {
      episodeId: "ep-1",
      positionSec: 40,
      firstPlayedAt: LATER,
      firstCompletedAt: null,
      lastPlayedAt: LATER,
    };
    const repository = createFakeProgressRepository([seed]);
    const lateCreate: ProgressWriteCommand = {
      episodeId: "ep-1",
      positionSec: 5,
      clientAt: EARLIER,
    };

    // When: 既存行へ create する
    const got = await createProgress(repository, lateCreate);

    // Then: 冪等成功。firstPlayedAt は先勝ち、position/last は後勝ちで既存のまま
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
