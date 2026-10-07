import { describe, expect, it } from "vitest";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import { ProgressNotFoundError } from "../../entities/errors/progress-not-found-error.ts";
import { ProgressRuleError } from "../../entities/errors/progress-rule-error.ts";
import { PROGRESS_COMPLETE_ZONE_SEC } from "../../entities/constants/progress.ts";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import type { ProgressRow } from "../ports/progress-repository.ts";
import type { ProgressWriteUseCaseInput } from "../progress/progress-write-use-case-io.ts";
import { completeProgress } from "./complete-progress.ts";

/**
 * scope: Sociable Unit
 * real: completeProgress use-case（原稿 duration・完走ゾーン・行必須・merge）
 * double: EpisodeRepository Fake / InMemoryProgressRepository (Fake)
 */
const EARLIER = "2026-09-22T10:00:00.000Z";
const DURATION_SEC = 60;
const IN_ZONE_POSITION = DURATION_SEC - PROGRESS_COMPLETE_ZONE_SEC;
const OUT_ZONE_POSITION = IN_ZONE_POSITION - 1;

const validManuscript = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: DURATION_SEC,
  body: {
    opening: { text: "開始", startSec: 0 },
    topics: [
      { title: "第一", preface: "前1", detail: "詳1", startSec: 0 },
      { title: "第二", preface: "前2", detail: "詳2", startSec: 30 },
    ],
    ending: { text: "終了", startSec: 55 },
  },
};

const inZoneInput: ProgressWriteUseCaseInput = {
  episodeId: "ep-1",
  positionSec: IN_ZONE_POSITION,
  clientAt: EARLIER,
};

const existingRow: ProgressRow = {
  episodeId: "ep-1",
  positionSec: 10,
  firstPlayedAt: EARLIER,
  firstCompletedAt: null,
  lastPlayedAt: EARLIER,
};

function createFakeEpisodeRepository(
  overrides: Partial<EpisodeRepository> = {},
): EpisodeRepository {
  return {
    listManuscripts: async () => {
      throw new Error("not used");
    },
    getManuscript: async ({ episodeId }) =>
      episodeId === "ep-1" ? { stem: "ep-1", json: validManuscript } : undefined,
    getAudio: async () => {
      throw new Error("not used");
    },
    ...overrides,
  };
}

describe("completeProgress", () => {
  it("行なしの時、ProgressNotFoundError を throw し永続しない", async () => {
    // Given: 原稿あり・進捗行なし・ゾーン内
    const episodes = createFakeEpisodeRepository();
    const progress = new InMemoryProgressRepository();

    // When: complete を実行する
    const act = completeProgress(episodes, progress, inZoneInput);

    // Then: 行不在。bootstrap しない
    await expect(act).rejects.toBeInstanceOf(ProgressNotFoundError);
    const stored = await progress.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.size).toBe(0);
  });

  it("既存行がありゾーン内の時、markCompleted merge して firstCompletedAt を返す", async () => {
    // Given: 未完走行あり・ゾーン内
    const episodes = createFakeEpisodeRepository();
    const progress = new InMemoryProgressRepository([existingRow]);

    // When: complete を実行する
    const got = await completeProgress(episodes, progress, inZoneInput);

    // Then: firstPlayedAt は既存、firstCompletedAt は今回 clientAt
    expect(got).toEqual({ firstPlayedAt: EARLIER, firstCompletedAt: EARLIER });
    const stored = await progress.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.get("ep-1")).toEqual({
      positionSec: IN_ZONE_POSITION,
      firstPlayedAt: EARLIER,
      firstCompletedAt: EARLIER,
      lastPlayedAt: EARLIER,
    });
  });

  it("既存行がありゾーン内の後続 complete の時、先勝ち firstCompletedAt を保ち merge する", async () => {
    // Given: 未完走行あり・ゾーン内の後続 complete
    const episodes = createFakeEpisodeRepository();
    const progress = new InMemoryProgressRepository([existingRow]);
    const later: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: IN_ZONE_POSITION,
      clientAt: "2026-09-22T11:00:00.000Z",
    };

    // When: complete を実行する
    const got = await completeProgress(episodes, progress, later);

    // Then: firstPlayedAt は既存、firstCompletedAt は今回 clientAt
    expect(got).toEqual({
      firstPlayedAt: EARLIER,
      firstCompletedAt: "2026-09-22T11:00:00.000Z",
    });
  });

  it("ゾーン外の時、ProgressRuleError を throw し永続しない", async () => {
    // Given: 行あり・ゾーン外 position
    const episodes = createFakeEpisodeRepository();
    const progress = new InMemoryProgressRepository([existingRow]);
    const outOfZone: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: OUT_ZONE_POSITION,
      clientAt: EARLIER,
    };

    // When: complete を実行する
    const act = completeProgress(episodes, progress, outOfZone);

    // Then: 完走規則違反。既存行は完走にならない
    await expect(act).rejects.toBeInstanceOf(ProgressRuleError);
    const stored = await progress.getByEpisodeIds({ episodeIds: ["ep-1"] });
    expect(stored.get("ep-1")?.firstCompletedAt).toBeNull();
  });

  it("原稿が無い時、EpisodeContentError を throw する", async () => {
    // Given: getManuscript が undefined・行あり
    const episodes = createFakeEpisodeRepository({
      getManuscript: async () => undefined,
    });
    const progress = new InMemoryProgressRepository([existingRow]);

    // When: complete を実行する
    const act = completeProgress(episodes, progress, inZoneInput);

    // Then: 既存の原稿不在 path
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
  });

  it("原稿が schema 不適合の時、EpisodeContentError を throw する", async () => {
    // Given: schema 不適合 json・行あり
    const episodes = createFakeEpisodeRepository({
      getManuscript: async () => ({ stem: "ep-1", json: { episodeId: "ep-1" } }),
    });
    const progress = new InMemoryProgressRepository([existingRow]);

    // When: complete を実行する
    const act = completeProgress(episodes, progress, inZoneInput);

    // Then: verifyManuscript の schema 不適合 path
    await expect(act).rejects.toBeInstanceOf(EpisodeContentError);
  });

  it("短尺（durationSec がゾーン幅未満）でも同じ不等式でゾーン内なら完走できる", async () => {
    // Given: durationSec=2（ゾーン幅 3 未満）・行あり・positionSec=0
    const shortManuscript = { ...validManuscript, durationSec: 2 };
    const episodes = createFakeEpisodeRepository({
      getManuscript: async () => ({ stem: "ep-1", json: shortManuscript }),
    });
    const progress = new InMemoryProgressRepository([existingRow]);
    const shortInput: ProgressWriteUseCaseInput = {
      episodeId: "ep-1",
      positionSec: 0,
      clientAt: EARLIER,
    };

    // When: complete を実行する
    const got = await completeProgress(episodes, progress, shortInput);

    // Then: 短尺専用枝なしで完走記録される
    expect(got.firstCompletedAt).toBe(EARLIER);
  });
});
