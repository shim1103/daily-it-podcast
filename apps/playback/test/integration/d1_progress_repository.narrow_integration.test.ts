/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers D1 binding（実 SQLite）
 * Double: 本番 remote D1 は使わない（remoteBindings: false）
 *
 * 実 SQL の merge 意味（先勝ち / 後勝ち / 冪等 / pull 境界 / offset 混在）と、実 D1 の bind 数上限・
 * 表なし失敗を所有する。binding 呼び出しの形・Error 写像・再試行なしは sociable unit が所有する。
 *
 * @require createLocalD1Binding が実 proxy を起動し、applyEpisodeProgressMigration で表を作れる
 * @ensure D1ProgressRepository が実 SQLite に対して Port の merge 契約を満たす
 */
import { afterAll, beforeAll, beforeEach, describe, expect, it } from "vitest";
import { D1Error } from "../../worker/src/infrastructure/d1/d1-error.ts";
import { D1ProgressRepository } from "../../worker/src/infrastructure/d1/d1-progress-repository.ts";
import {
  D1_MAX_BOUND_PARAMETERS_PER_QUERY,
  EPISODE_PROGRESS_TABLE,
} from "../../worker/src/infrastructure/d1/progress-d1-constants.ts";
import { applyEpisodeProgressMigration } from "../support/apply-episode-progress-migration.ts";
import {
  type LocalD1BindingHandle,
  createLocalD1Binding,
} from "../support/create-local-d1-binding.ts";

let handle: LocalD1BindingHandle;
let repository: D1ProgressRepository;

beforeAll(async () => {
  handle = await createLocalD1Binding();
  await applyEpisodeProgressMigration(handle.database);
  repository = new D1ProgressRepository({ database: handle.database });
});

afterAll(async () => {
  await handle.dispose();
});

beforeEach(async () => {
  await handle.database.prepare(`DELETE FROM ${EPISODE_PROGRESS_TABLE}`).run();
});

async function progressOf(episodeId: string) {
  return (await repository.getByEpisodeIds([episodeId])).get(episodeId);
}

describe("D1ProgressRepository on local D1", () => {
  describe("writeProgress", () => {
    it("creates_row_with_first_and_last_played_at_equal_to_client_at_when_no_row_exists", async () => {
      // Given: 行の無い episode

      // When: 初回 write する
      const got = await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 10,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // Then: first_played_at は clientAt、未完走で、行は clientAt を last_played_at に持つ
      expect(got).toEqual({ firstPlayedAt: "2026-10-01T00:00:10.000Z", firstCompletedAt: null });
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 10,
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:10.000Z",
      });
    });

    it("advances_position_and_last_played_at_and_keeps_first_played_at_when_later_client_at_arrives", async () => {
      // Given: 00:00:10 で作成済みの行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 10,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // When: より遅い clientAt で update する
      const got = await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 30,
        clientAt: "2026-10-01T00:00:30.000Z",
      });

      // Then: first_played_at は先勝ちで残り、position / last_played_at は遅い側へ進む
      expect(got.firstPlayedAt).toBe("2026-10-01T00:00:10.000Z");
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 30,
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:30.000Z",
      });
    });

    it("moves_first_played_at_earlier_and_keeps_later_position_pair_when_earlier_client_at_arrives_late", async () => {
      // Given: 00:00:30 の位置で作成済みの行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 30,
        clientAt: "2026-10-01T00:00:30.000Z",
      });

      // When: 遅れて届いた、より早い clientAt の write
      const got = await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 5,
        clientAt: "2026-10-01T00:00:05.000Z",
      });

      // Then: first_played_at は早い側へ移り、position / last_played_at は遅い側のペアのまま
      expect(got.firstPlayedAt).toBe("2026-10-01T00:00:05.000Z");
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 30,
        firstPlayedAt: "2026-10-01T00:00:05.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:00:30.000Z",
      });
    });

    it("takes_incoming_position_when_client_at_ties", async () => {
      // Given: 同じ clientAt で作成済みの行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 10,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // When: 同じ clientAt で別の位置を write する
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 20,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // Then: 後着が採用される
      expect((await progressOf("ep-1"))?.positionSec).toBe(20);
    });

    it("returns_same_winner_and_keeps_row_when_same_request_is_replayed", async () => {
      // Given: 同一内容の write を 1 度済ませた行
      const command = {
        episodeId: "ep-1",
        positionSec: 10,
        clientAt: "2026-10-01T00:00:10.000Z",
      };
      const first = await repository.writeProgress(command);
      const rowAfterFirst = await progressOf("ep-1");

      // When: 同一内容を再送する（重複 create / retry）
      const second = await repository.writeProgress(command);

      // Then: 冪等に成功し、勝ち側も行も変わらない
      expect(second).toEqual(first);
      expect(await progressOf("ep-1")).toEqual(rowAfterFirst);
    });

    it("compares_by_instant_not_by_string_when_client_at_offsets_are_mixed", async () => {
      // Given: Z 表記 00:30Z の行。文字列比較だと "…T09:00:00+09:00" の方が大きく見える
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 30,
        clientAt: "2026-10-01T00:30:00Z",
      });

      // When: 時刻としては 00:00Z（より早い）の +09:00 表記を write する
      const got = await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 1,
        clientAt: "2026-10-01T09:00:00+09:00",
      });

      // Then: 時刻として早い側が first_played_at に勝ち、position は遅い 00:30Z のまま。応答は UTC 表記
      expect(got.firstPlayedAt).toBe("2026-10-01T00:00:00.000Z");
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 30,
        firstPlayedAt: "2026-10-01T00:00:00.000Z",
        firstCompletedAt: null,
        lastPlayedAt: "2026-10-01T00:30:00.000Z",
      });
    });

    it("keeps_first_completed_at_when_update_follows_complete", async () => {
      // Given: complete 済みの行
      await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 60,
        clientAt: "2026-10-01T00:01:00.000Z",
      });

      // When: 後続の update
      const got = await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 61,
        clientAt: "2026-10-01T00:01:10.000Z",
      });

      // Then: first_completed_at には触れない
      expect(got.firstCompletedAt).toBe("2026-10-01T00:01:00.000Z");
    });
  });

  describe("completeProgress", () => {
    it("creates_row_with_all_timestamps_equal_to_client_at_when_no_row_exists", async () => {
      // Given: 行の無い episode

      // When: 初回で complete する
      const got = await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 60,
        clientAt: "2026-10-01T00:01:00.000Z",
      });

      // Then: first_played_at / first_completed_at / last_played_at がすべて clientAt
      expect(got).toEqual({
        firstPlayedAt: "2026-10-01T00:01:00.000Z",
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
      });
      expect(await progressOf("ep-1")).toEqual({
        positionSec: 60,
        firstPlayedAt: "2026-10-01T00:01:00.000Z",
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
        lastPlayedAt: "2026-10-01T00:01:00.000Z",
      });
    });

    it("fills_first_completed_at_and_keeps_first_played_at_when_row_has_no_completion", async () => {
      // Given: 未完走の行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 10,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // When: complete する
      const got = await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 60,
        clientAt: "2026-10-01T00:01:00.000Z",
      });

      // Then: first_completed_at が埋まり、first_played_at は先勝ちで残り、position は後勝ちで進む
      expect(got).toEqual({
        firstPlayedAt: "2026-10-01T00:00:10.000Z",
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
      });
      expect((await progressOf("ep-1"))?.positionSec).toBe(60);
    });

    it("keeps_earlier_first_completed_at_when_later_complete_arrives", async () => {
      // Given: 00:01:00 に完走済みの行
      await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 60,
        clientAt: "2026-10-01T00:01:00.000Z",
      });

      // When: より遅い clientAt の complete
      const got = await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 62,
        clientAt: "2026-10-01T00:02:00.000Z",
      });

      // Then: first_completed_at は先勝ちで早い側のまま、position / last_played_at は遅い側へ
      expect(got.firstCompletedAt).toBe("2026-10-01T00:01:00.000Z");
      expect(await progressOf("ep-1")).toMatchObject({
        positionSec: 62,
        lastPlayedAt: "2026-10-01T00:02:00.000Z",
      });
    });

    it("moves_first_completed_at_earlier_when_earlier_complete_arrives_late", async () => {
      // Given: 00:02:00 に完走済みの行
      await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 62,
        clientAt: "2026-10-01T00:02:00.000Z",
      });

      // When: 遅れて届いた、より早い clientAt の complete
      const got = await repository.completeProgress({
        episodeId: "ep-1",
        positionSec: 60,
        clientAt: "2026-10-01T00:01:00.000Z",
      });

      // Then: first_completed_at / first_played_at は早い側へ、position は遅い側のペアのまま
      expect(got).toEqual({
        firstPlayedAt: "2026-10-01T00:01:00.000Z",
        firstCompletedAt: "2026-10-01T00:01:00.000Z",
      });
      expect(await progressOf("ep-1")).toMatchObject({
        positionSec: 62,
        lastPlayedAt: "2026-10-01T00:02:00.000Z",
      });
    });
  });

  describe("getByEpisodeIds", () => {
    it("returns_only_existing_rows_when_some_episode_ids_have_no_row", async () => {
      // Given: ep-1 だけ行がある
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 10,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // When: 行のある ID と無い ID を一緒に取得する
      const got = await repository.getByEpisodeIds(["ep-1", "ep-unknown"]);

      // Then: 行のある episode だけが載る
      expect([...got.keys()]).toEqual(["ep-1"]);
    });

    it("finds_rows_across_chunks_when_episode_ids_exceed_real_bind_limit", async () => {
      // Given: 実 D1 の bind 数上限を超える数の ID と、チャンクをまたぐ 3 行
      const episodeIds = Array.from(
        { length: D1_MAX_BOUND_PARAMETERS_PER_QUERY * 2 + 50 },
        (_, index) => `ep-${index}`,
      );
      const seeded = ["ep-5", `ep-${D1_MAX_BOUND_PARAMETERS_PER_QUERY + 20}`, "ep-249"];
      for (const episodeId of seeded) {
        await repository.writeProgress({
          episodeId,
          positionSec: 1,
          clientAt: "2026-10-01T00:00:10.000Z",
        });
      }

      // When: 全 ID で取得する
      const got = await repository.getByEpisodeIds(episodeIds);

      // Then: 上限エラーにならず、全チャンクの行が揃う
      expect([...got.keys()].sort()).toEqual([...seeded].sort());
    });
  });

  describe("listUpdatedSince", () => {
    it("excludes_row_whose_last_played_at_equals_since", async () => {
      // Given: last_played_at が 00:00:10 / 00:00:20 の 2 行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 1,
        clientAt: "2026-10-01T00:00:10.000Z",
      });
      await repository.writeProgress({
        episodeId: "ep-2",
        positionSec: 2,
        clientAt: "2026-10-01T00:00:20.000Z",
      });

      // When: since = 00:00:10 で pull する
      const got = await repository.listUpdatedSince("2026-10-01T00:00:10.000Z");

      // Then: 厳密に後の行だけが返る
      expect(got.map((entry) => entry.episodeId)).toEqual(["ep-2"]);
    });

    it("orders_by_last_played_at_then_episode_id", async () => {
      // Given: 書込順と更新時刻順が食い違い、同時刻の行を含む 3 行
      await repository.writeProgress({
        episodeId: "ep-b",
        positionSec: 1,
        clientAt: "2026-10-01T00:00:20.000Z",
      });
      await repository.writeProgress({
        episodeId: "ep-c",
        positionSec: 1,
        clientAt: "2026-10-01T00:00:10.000Z",
      });
      await repository.writeProgress({
        episodeId: "ep-a",
        positionSec: 1,
        clientAt: "2026-10-01T00:00:20.000Z",
      });

      // When: pull する
      const got = await repository.listUpdatedSince("2026-10-01T00:00:00.000Z");

      // Then: 古い順、同時刻は episodeId 昇順
      expect(got.map((entry) => entry.episodeId)).toEqual(["ep-c", "ep-a", "ep-b"]);
    });

    it("compares_by_instant_when_since_has_offset", async () => {
      // Given: 00:00:10Z と 00:00:30Z の 2 行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 1,
        clientAt: "2026-10-01T00:00:10Z",
      });
      await repository.writeProgress({
        episodeId: "ep-2",
        positionSec: 2,
        clientAt: "2026-10-01T00:00:30Z",
      });

      // When: 時刻としては 00:00:20Z にあたる +09:00 表記の since で pull する
      const got = await repository.listUpdatedSince("2026-10-01T09:00:20+09:00");

      // Then: 時刻として後の行だけが返る（文字列比較なら両方返ってしまう）
      expect(got.map((entry) => entry.episodeId)).toEqual(["ep-2"]);
    });

    it("returns_empty_array_when_no_row_is_newer_than_since", async () => {
      // Given: 1 行
      await repository.writeProgress({
        episodeId: "ep-1",
        positionSec: 1,
        clientAt: "2026-10-01T00:00:10.000Z",
      });

      // When: 全行より後の since で pull する
      const got = await repository.listUpdatedSince("2026-10-02T00:00:00.000Z");

      // Then: 空配列
      expect(got).toEqual([]);
    });
  });

  describe("failure", () => {
    it("throws_d1_error_with_cause_when_progress_table_does_not_exist", async () => {
      // Given: migration 未適用の空の D1（実 D1 が失敗を返す）
      const bare = await createLocalD1Binding();
      try {
        const bareRepository = new D1ProgressRepository({ database: bare.database });

        // When: 書込する
        const got = await bareRepository
          .writeProgress({
            episodeId: "ep-1",
            positionSec: 1,
            clientAt: "2026-10-01T00:00:10.000Z",
          })
          .catch((error: unknown) => error);

        // Then: D1Error に畳まれ、実 D1 の失敗が cause に残る
        expect(got).toBeInstanceOf(D1Error);
        expect((got as D1Error).cause).toBeInstanceOf(Error);
      } finally {
        await bare.dispose();
      }
    });
  });
});
