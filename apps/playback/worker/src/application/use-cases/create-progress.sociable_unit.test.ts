import { describe, expect, it, vi } from "vitest";
import { utc } from "../../../../test/support/utc-instant.ts";
import { PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC } from "../../entities/constants/progress.ts";
import { ProgressRuleError } from "../../entities/errors/progress-rule-error.ts";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import { FakeClock } from "../ports/clock.fake.ts";
import type { ProgressRow } from "../ports/progress-repository.ts";
import { createSpyProgressRepository } from "../ports/progress-repository.spy.ts";
import type { ProgressWriteUseCaseInput } from "../progress/progress-write-use-case-io.ts";
import { createProgress } from "./create-progress.ts";

/**
 * scope: Sociable Unit
 * real: createProgress use-case（merge・冪等・skew）
 * double: InMemoryProgressRepository (Fake) / FakeClock / Spy ProgressRepository
 */
const SERVER_AT = new Date("2026-09-22T12:00:00.000Z");
const EARLIER = utc("2026-09-22T10:00:00.000Z");
const LATER = utc("2026-09-22T11:00:00.000Z");
const FUTURE_SKEW_MS = PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC * 1000;

const input: ProgressWriteUseCaseInput = {
  episodeId: "ep-1",
  positionSec: 12,
  clientAt: EARLIER,
};

describe("createProgress", () => {
  it("行なしの時、merge で初回行を作り first* を返す", async () => {
    // Given: 空 Fake・許容内 clientAt
    const repository = new InMemoryProgressRepository();
    const clock = new FakeClock(SERVER_AT);

    // When: create を実行する
    const got = await createProgress(repository, clock, input);

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
    // Given: 既存行あり（後から届くより早い初回）・許容内 clientAt
    const seed: ProgressRow = {
      episodeId: "ep-1",
      positionSec: 40,
      firstPlayedAt: LATER,
      firstCompletedAt: null,
      lastPlayedAt: LATER,
    };
    const repository = new InMemoryProgressRepository([seed]);
    const clock = new FakeClock(SERVER_AT);
    const lateCreate: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 5,
      clientAt: EARLIER,
    };

    // When: 既存行へ create する
    const got = await createProgress(repository, clock, lateCreate);

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

  it("allows_create_when_clientAt_is_exactly_at_future_skew_limit", async () => {
    // Given: 未来側ちょうど許容幅
    const repository = new InMemoryProgressRepository();
    const clock = new FakeClock(SERVER_AT);
    const atLimit: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 1,
      clientAt: utc(new Date(SERVER_AT.getTime() + FUTURE_SKEW_MS).toISOString()),
    };

    // When: create する
    const got = await createProgress(repository, clock, atLimit);

    // Then: ちょうどは通り first* が残る
    expect(got.firstPlayedAt).toBe(atLimit.clientAt);
  });

  it("rejects_create_without_calling_repository_when_clientAt_exceeds_future_skew_by_1ms", async () => {
    // Given: 未来側 1ms 超過・Spy repository
    const { repository, callCount } = createSpyProgressRepository();
    const clock = new FakeClock(SERVER_AT);
    const overLimit: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 1,
      clientAt: utc(new Date(SERVER_AT.getTime() + FUTURE_SKEW_MS + 1).toISOString()),
    };

    // When: create する
    const act = createProgress(repository, clock, overLimit);

    // Then: ProgressRuleError。repository は一度も呼ばない
    await expect(act).rejects.toBeInstanceOf(ProgressRuleError);
    expect(callCount()).toBe(0);
  });

  it("calls_clock_now_only_once_per_create", async () => {
    // Given: FakeClock の now を Spy した空 repository
    const repository = new InMemoryProgressRepository();
    const clock = new FakeClock(SERVER_AT);
    const nowSpy = vi.spyOn(clock, "now");

    // When: create する
    await createProgress(repository, clock, input);

    // Then: serverAt 取得は 1 回だけ
    expect(nowSpy).toHaveBeenCalledTimes(1);
  });
});
