import { describe, expect, it } from "vitest";
import { ProgressPullResponseSchema } from "../../../../contracts/index.ts";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import type { ProgressUpsertRow } from "../ports/progress-repository.ts";
import { pullProgress } from "./pull-progress.ts";

/**
 * scope: Sociable Unit
 * real: pullProgress use-case（lastPlayedAt > since 境界）
 * double: InMemoryProgressRepository (Fake)
 */
const SINCE = "2026-09-15T00:00:00.000Z";

const beforeSince: ProgressUpsertRow = {
  episodeId: "ep-before",
  positionSec: 10,
  firstPlayedAt: "2026-09-01T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-09-10T12:00:00.000Z",
};

const afterSince: ProgressUpsertRow = {
  episodeId: "ep-after",
  positionSec: 20,
  firstPlayedAt: "2026-09-02T00:00:00.000Z",
  firstCompletedAt: "2026-09-05T00:00:00.000Z",
  lastPlayedAt: "2026-09-20T12:00:00.000Z",
};

const equalSince: ProgressUpsertRow = {
  episodeId: "ep-equal",
  positionSec: 15,
  firstPlayedAt: "2026-09-03T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: SINCE,
};

describe("pullProgress", () => {
  it("該当なしの時、空の ProgressPullResponse を返す", async () => {
    // Given: 空 Fake
    const repository = new InMemoryProgressRepository();

    // When: pull を実行する
    const got = await pullProgress(repository, SINCE);

    // Then: 契約形を満たし episodes は空
    expect(ProgressPullResponseSchema.safeParse(got).success).toBe(true);
    expect(got).toEqual({ episodes: [] });
  });

  it("lastPlayedAt が since より後の行だけを返す", async () => {
    // Given: since 前後に分かれる 2 行
    const repository = new InMemoryProgressRepository([beforeSince, afterSince]);

    // When: 間の since で pull する
    const got = await pullProgress(repository, SINCE);

    // Then: after だけ。形は pull 契約の episodes 要素
    expect(got).toEqual({
      episodes: [
        {
          episodeId: "ep-after",
          progress: {
            positionSec: 20,
            firstPlayedAt: "2026-09-02T00:00:00.000Z",
            firstCompletedAt: "2026-09-05T00:00:00.000Z",
            lastPlayedAt: "2026-09-20T12:00:00.000Z",
          },
        },
      ],
    });
  });

  it("lastPlayedAt が since と等しい行を含めない", async () => {
    // Given: lastPlayedAt === since の行
    const repository = new InMemoryProgressRepository([equalSince]);

    // When: 同一時刻の since で pull する
    const got = await pullProgress(repository, SINCE);

    // Then: 空（境界は厳密な >）
    expect(got).toEqual({ episodes: [] });
  });
});
