import { describe, expect, it } from "vitest";
import { applyEpisodeProgressMigration } from "./apply-episode-progress-migration.ts";
import { createFakeLocalD1Binding } from "./create-fake-local-d1-binding.ts";

describe("applyEpisodeProgressMigration", () => {
  it("runs_statements_of_0001_then_0002_in_file_name_order_when_applying", async () => {
    // Given: 呼び出しを記録する Fake の database
    const { database, calls } = await createFakeLocalD1Binding();

    // When: migration を適用する
    await applyEpisodeProgressMigration(database);

    // Then: 適用済みで変わらない 0001（1 文）と 0002（3 文）の開始句が、この順で先頭 4 件に並ぶ
    const startingClauses = calls
      .slice(0, 4)
      .map((call) => call.sql.split(/\s+/).slice(0, 2).join(" "));
    expect(startingClauses).toEqual([
      "CREATE TABLE",
      "ALTER TABLE",
      "UPDATE episode_progress",
      "CREATE INDEX",
    ]);
  });

  it("excludes_comment_lines_from_executed_statements_when_a_file_has_a_header_comment", async () => {
    // Given: header comment を持つ migration file を読む Fake の database
    const { database, calls } = await createFakeLocalD1Binding();

    // When: migration を適用する
    await applyEpisodeProgressMigration(database);

    // Then: 実行された文に comment 行が混ざらない
    expect(calls.length).toBeGreaterThan(0);
    for (const call of calls) {
      expect(call.sql).not.toContain("--");
    }
  });

  it("skips_empty_statements_when_a_file_ends_with_a_terminator", async () => {
    // Given: 末尾が `;` と改行で終わる migration file を読む Fake の database
    const { database, calls } = await createFakeLocalD1Binding();

    // When: migration を適用する
    await applyEpisodeProgressMigration(database);

    // Then: 空の文は実行されない
    for (const call of calls) {
      expect(call.sql.trim()).not.toBe("");
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
