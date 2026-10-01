import { describe, expect, it } from "vitest";
import { createFakeProgressRepository } from "./progress-repository.fake.ts";
import type { ProgressUpsertRow } from "./progress-repository.ts";

/**
 * scope: Sociable Unit
 * real: createFakeProgressRepository（C の test support。in-memory Port）
 * double: なし
 */
const rowA: ProgressUpsertRow = {
  episodeId: "ep-a",
  positionSec: 10,
  firstPlayedAt: "2026-09-01T00:00:00.000Z",
  firstCompletedAt: null,
  lastPlayedAt: "2026-09-10T12:00:00.000Z",
};

const rowB: ProgressUpsertRow = {
  episodeId: "ep-b",
  positionSec: 20,
  firstPlayedAt: "2026-09-02T00:00:00.000Z",
  firstCompletedAt: "2026-09-05T00:00:00.000Z",
  lastPlayedAt: "2026-09-20T12:00:00.000Z",
};

describe("createFakeProgressRepository", () => {
  it("seed した行を getByEpisodeIds で返す。無い id は Map に載せない", async () => {
    // Given: seed 済み Fake
    const fake = createFakeProgressRepository([rowA, rowB]);

    // When: 存在する id と存在しない id を混ぜて取得する
    const got = await fake.getByEpisodeIds(["ep-a", "ep-missing", "ep-b"]);

    // Then: seed 行だけが入り、欠けた id は載らない
    expect(got.size).toBe(2);
    expect(got.get("ep-a")).toEqual({
      positionSec: 10,
      firstPlayedAt: "2026-09-01T00:00:00.000Z",
      firstCompletedAt: null,
      lastPlayedAt: "2026-09-10T12:00:00.000Z",
    });
    expect(got.get("ep-b")).toEqual({
      positionSec: 20,
      firstPlayedAt: "2026-09-02T00:00:00.000Z",
      firstCompletedAt: "2026-09-05T00:00:00.000Z",
      lastPlayedAt: "2026-09-20T12:00:00.000Z",
    });
    expect(got.has("ep-missing")).toBe(false);
  });

  it("upsertProgress は行をそのまま永続し、再 upsert は merge せず置き換える", async () => {
    // Given: 空 Fake
    const fake = createFakeProgressRepository();

    // When: 同じ episodeId を 2 回 upsert する（2 回目は全 field 差し替え）
    await fake.upsertProgress(rowA);
    const replaced: ProgressUpsertRow = {
      episodeId: "ep-a",
      positionSec: 99,
      firstPlayedAt: "2026-09-15T00:00:00.000Z",
      firstCompletedAt: "2026-09-16T00:00:00.000Z",
      lastPlayedAt: "2026-09-17T00:00:00.000Z",
    };
    await fake.upsertProgress(replaced);
    const got = await fake.getByEpisodeIds(["ep-a"]);

    // Then: 2 回目の行がそのまま残り、1 回目の値は混ざらない
    expect(got.get("ep-a")).toEqual({
      positionSec: 99,
      firstPlayedAt: "2026-09-15T00:00:00.000Z",
      firstCompletedAt: "2026-09-16T00:00:00.000Z",
      lastPlayedAt: "2026-09-17T00:00:00.000Z",
    });
  });

  it("listUpdatedSince は lastPlayedAt が since より後の行だけを {episodeId, progress} で返す", async () => {
    // Given: lastPlayedAt が前後で分かれる 2 行
    const fake = createFakeProgressRepository([rowA, rowB]);

    // When: rowA と rowB の間の since で差分取得する
    const got = await fake.listUpdatedSince("2026-09-15T00:00:00.000Z");

    // Then: rowB のみ。形は pull 契約の episodes 要素と同形
    expect(got).toEqual([
      {
        episodeId: "ep-b",
        progress: {
          positionSec: 20,
          firstPlayedAt: "2026-09-02T00:00:00.000Z",
          firstCompletedAt: "2026-09-05T00:00:00.000Z",
          lastPlayedAt: "2026-09-20T12:00:00.000Z",
        },
      },
    ]);
  });

  it("listUpdatedSince は lastPlayedAt が since と等しい行を含めない", async () => {
    // Given: lastPlayedAt が since と同一の行
    const fake = createFakeProgressRepository([rowA]);

    // When: rowA.lastPlayedAt と同じ since で差分取得する
    const got = await fake.listUpdatedSince(rowA.lastPlayedAt);

    // Then: 空（境界は厳密な >）
    expect(got).toEqual([]);
  });
});
