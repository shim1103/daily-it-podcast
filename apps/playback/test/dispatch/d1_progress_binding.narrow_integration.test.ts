/**
 * Scope: Narrow Integration（進捗 D1 の一過性疎通。dispatch 専用）
 * 実物境界: getPlatformProxy（remoteBindings）が返す TEST D1 の EPISODE_PROGRESS binding と、
 *   binding とは独立な remote 読み取り（wrangler CLI の `--remote`）
 * Double: なし。本番 D1 は使わない（共有 test/support/wrangler.smoke.jsonc の TEST D1 だけ）
 *
 * @require 表 episode_progress が TEST D1 に適用済み（scripts/playback/test-d1-smoke.sh が migration を適用する）
 * @ensure adapter が呼ぶ面（prepare / bind / run / first）で、書く・読む・消す・消えたことを読むが成立する
 * @ensure binding で書いた行が、binding とは独立な remote 経路でも見える（binding が local の模擬に落ちていない）
 * @ensure probe 行は各 case の後に削除される
 */
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";
import {
  createRemoteTestD1Binding,
  type RemoteTestD1BindingHandle,
} from "./create-remote-test-d1-binding.ts";
import {
  createProbeRow,
  deleteProbeRow,
  readProbeRow,
  writeProbeRow,
} from "./d1-progress-probe.ts";
import { countRemoteProgressRows } from "./remote-read.ts";

const probeEpisodeId = `d1-smoke-probe-${process.env.GITHUB_RUN_ID ?? "manual"}`;
const probeRow = createProbeRow(probeEpisodeId);

describe("TEST D1 EPISODE_PROGRESS binding", () => {
  let handle: RemoteTestD1BindingHandle;

  beforeAll(async () => {
    handle = await createRemoteTestD1Binding();
  });

  afterEach(async () => {
    await deleteProbeRow(handle.database, probeEpisodeId);
  });

  afterAll(async () => {
    await handle.dispose();
  });

  it("writes_one_row_when_inserting_a_progress_row", async () => {
    // Given: 一意な episode_id の probe 行（未書込）

    // When: prepare→bind→run で書く
    const changes = await writeProbeRow(handle.database, probeRow);

    // Then: 1 行書けた
    expect(changes).toBe(1);
  });

  it("returns_the_written_row_unchanged_when_selecting_by_episode_id", async () => {
    // Given: binding で書き済みの probe 行
    await writeProbeRow(handle.database, probeRow);

    // When: episode_id で読む
    const stored = await readProbeRow(handle.database, probeEpisodeId);

    // Then: 5 列（NULL・REAL を含む）が書いた値のまま返る
    expect(stored).toEqual(probeRow);
  });

  it("deletes_one_row_when_deleting_by_episode_id", async () => {
    // Given: binding で書き済みの probe 行
    await writeProbeRow(handle.database, probeRow);

    // When: episode_id で消す
    const changes = await deleteProbeRow(handle.database, probeEpisodeId);

    // Then: 1 行消えた
    expect(changes).toBe(1);
  });

  it("returns_null_when_selecting_after_the_row_is_deleted", async () => {
    // Given: 書いて消した probe 行
    await writeProbeRow(handle.database, probeRow);
    await deleteProbeRow(handle.database, probeEpisodeId);

    // When: episode_id で読む
    const stored = await readProbeRow(handle.database, probeEpisodeId);

    // Then: 行が残っていない
    expect(stored).toBeNull();
  });

  it("shows_the_row_to_the_remote_database_when_written_through_the_binding", async () => {
    // Given: binding で書き済みの probe 行
    await writeProbeRow(handle.database, probeRow);

    // When: binding とは独立な経路（wrangler CLI の --remote）で件数を数える
    const remoteCount = await countRemoteProgressRows(probeEpisodeId);

    // Then: 実 TEST D1 に 1 行ある（local の模擬へ書いていれば 0 になる）
    expect(remoteCount).toBe(1);
  });
});
