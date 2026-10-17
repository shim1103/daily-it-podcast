import { describe, expect, it, vi } from "vitest";
import { utc } from "../../../../test/support/utc-instant.ts";
import { PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC } from "../../entities/constants/progress.ts";
import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import { ProgressRuleError } from "../../entities/errors/progress-rule-error.ts";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import { FakeClock } from "../ports/clock.fake.ts";
import type { ProgressRow } from "../ports/progress-repository.ts";
import { createSpyProgressRepository } from "../ports/progress-repository.spy.ts";
import type { ProgressWriteUseCaseInput } from "../progress/progress-write-use-case-io.ts";
import { updateProgress } from "./update-progress.ts";

/**
 * scope: Sociable Unit
 * real: updateProgress use-case（行なし 404・merge・skew）
 * double: InMemoryProgressRepository (Fake) / FakeClock / Spy ProgressRepository
 */
const SERVER_AT = new Date("2026-09-22T12:00:00.000Z");
const EARLIER = utc("2026-09-22T10:00:00.000Z");
const LATER = utc("2026-09-22T11:00:00.000Z");
const PAST_SKEW_MS = PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC * 1000;

const input: ProgressWriteUseCaseInput = {
  episodeId: "ep-1",
  positionSec: 40,
  clientAt: LATER,
};

describe("updateProgress", () => {
  it("行なしの時、ProgressNotFoundError を throw する", async () => {
    // Given: 空 Fake・許容内 clientAt
    const repository = new InMemoryProgressRepository();
    const clock = new FakeClock(SERVER_AT);

    // When: update を実行する
    const act = updateProgress(repository, clock, input);

    // Then: 行不在の Domain Error（永続は触らない）
    await expect(act).rejects.toBeInstanceOf(ProgressNotFoundError);
    const stored = await repository.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.size).toBe(0);
  });

  it("行ありの時、merge して upsert し勝ち側 first* を返す", async () => {
    // Given: 既存行あり・許容内 clientAt
    const seed: ProgressRow = {
      episodeId: "ep-1",
      positionSec: 10,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    };
    const repository = new InMemoryProgressRepository([seed]);
    const clock = new FakeClock(SERVER_AT);

    // When: より遅い clientAt で update する
    const got = await updateProgress(repository, clock, input);

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

  it("allows_update_when_clientAt_is_exactly_at_past_skew_limit", async () => {
    // Given: 既存行あり・過去側ちょうど許容幅
    const seed: ProgressRow = {
      episodeId: "ep-1",
      positionSec: 10,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    };
    const repository = new InMemoryProgressRepository([seed]);
    const clock = new FakeClock(SERVER_AT);
    const atLimit: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 20,
      clientAt: utc(new Date(SERVER_AT.getTime() - PAST_SKEW_MS).toISOString()),
    };

    // When: update する
    const got = await updateProgress(repository, clock, atLimit);

    // Then: ちょうどは通り first* は先勝ちで更新側
    expect(got.firstPlayedAt).toBe(atLimit.clientAt);
  });

  it("rejects_update_without_calling_repository_when_clientAt_exceeds_past_skew_by_1ms", async () => {
    // Given: 過去側 1ms 超過・Spy repository
    const { repository, callCount } = createSpyProgressRepository();
    const clock = new FakeClock(SERVER_AT);
    const overLimit: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 20,
      clientAt: utc(new Date(SERVER_AT.getTime() - PAST_SKEW_MS - 1).toISOString()),
    };

    // When: update する
    const act = updateProgress(repository, clock, overLimit);

    // Then: ProgressRuleError。repository は一度も呼ばない
    await expect(act).rejects.toBeInstanceOf(ProgressRuleError);
    expect(callCount()).toBe(0);
  });

  it("calls_clock_now_only_once_per_update", async () => {
    // Given: 既存行あり・FakeClock の now を Spy
    const seed: ProgressRow = {
      episodeId: "ep-1",
      positionSec: 10,
      firstPlayedAt: EARLIER,
      firstCompletedAt: null,
      lastPlayedAt: EARLIER,
    };
    const repository = new InMemoryProgressRepository([seed]);
    const clock = new FakeClock(SERVER_AT);
    const nowSpy = vi.spyOn(clock, "now");

    // When: update する
    await updateProgress(repository, clock, input);

    // Then: serverAt 取得は 1 回だけ
    expect(nowSpy).toHaveBeenCalledTimes(1);
  });
});
