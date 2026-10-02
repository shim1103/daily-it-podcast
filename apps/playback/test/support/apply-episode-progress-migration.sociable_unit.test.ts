import { describe, expect, it } from "vitest";
import { applyEpisodeProgressMigration } from "./apply-episode-progress-migration.ts";
import { createFakeLocalD1Binding } from "./create-fake-local-d1-binding.ts";

describe("applyEpisodeProgressMigration", () => {
  it("runs_create_table_then_create_index_in_breakpoint_order_when_applying", async () => {
    // Given: 呼び出しを記録する Fake の database
    const { database, calls } = await createFakeLocalD1Binding();

    // When: migration を適用する
    await applyEpisodeProgressMigration(database);

    // Then: 生成 migration の 2 文の開始句が、区切り行で分けた順に並ぶ
    const startingClauses = calls.map((call) => call.sql.split(/\s+/).slice(0, 2).join(" "));
    expect(startingClauses).toEqual(["CREATE TABLE", "CREATE INDEX"]);
  });

  it("excludes_statement_breakpoint_marker_from_executed_statements_when_a_file_has_multiple_statements", async () => {
    // Given: 文と文の間に区切り行を持つ migration file を読む Fake の database
    const { database, calls } = await createFakeLocalD1Binding();

    // When: migration を適用する
    await applyEpisodeProgressMigration(database);

    // Then: 実行された文に区切り行が混ざらない
    expect(calls.length).toBeGreaterThan(0);
    for (const call of calls) {
      expect(call.sql).not.toContain("statement-breakpoint");
    }
  });

  it("throws_when_a_statement_run_reports_failure", async () => {
    // Given: run が success: false を返す Fake の database
    const { database } = await createFakeLocalD1Binding({
      run: async () => ({ success: false, meta: { changes: 0 } }),
    });

    // When: migration を適用する
    const applying = applyEpisodeProgressMigration(database);

    // Then: 適用失敗として throw する
    await expect(applying).rejects.toThrow("migration 適用に失敗");
  });
});
