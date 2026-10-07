import { describe, expect, it } from "vitest";
import { InMemoryProgressRepository } from "../../infrastructure/in-memory/in-memory-progress-repository.ts";
import type { ProgressRow } from "../ports/progress-repository.ts";
import { pullProgress } from "./pull-progress.ts";

/**
 * scope: Sociable Unit
 * real: pullProgress use-case（lastPlayedAt > since 境界）
 * double: InMemoryProgressRepository (Fake)
 */
const SINCE = "2026-09-15T00:00:00.000Z";

const beforeSince: ProgressRow = {
  episodeId: "ep-before",
  positionSec: 10,
  firstPlayedAt: "2026-09-01T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-09-10T12:00:00.000Z",
};

const afterSince: ProgressRow = {
  episodeId: "ep-after",
  positionSec: 20,
  firstPlayedAt: "2026-09-02T00:00:00.000Z",
  firstCompletedAt: "2026-09-05T00:00:00.000Z",
  lastPlayedAt: "2026-09-20T12:00:00.000Z",
};

const equalSince: ProgressRow = {
  episodeId: "ep-equal",
  positionSec: 15,
  firstPlayedAt: "2026-09-03T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: SINCE,
};

describe("pullProgress", () => {
  it("該当なしの時、空の episodes を返す", async () => {
    // Given: 空 Fake
    const repository = new InMemoryProgressRepository();

    // When: pull を実行する
    const got = await pullProgress(repository, { since: SINCE });

    // Then: episodes は空
    expect(got).toEqual({ episodes: [] });
  });

  it("lastPlayedAt が since より後の行だけを返す", async () => {
    // Given: since 前後に分かれる 2 行
    const repository = new InMemoryProgressRepository([beforeSince, afterSince]);

    // When: 間の since で pull する
    const got = await pullProgress(repository, { since: SINCE });

    // Then: after だけ。要素は episodeId と progress 本体
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
    const got = await pullProgress(repository, { since: SINCE });

    // Then: 空（境界は厳密な >）
    expect(got).toEqual({ episodes: [] });
  });
});
