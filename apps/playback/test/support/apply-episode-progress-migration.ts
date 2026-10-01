import { readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";

const MIGRATION_PATH = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../worker/migrations/0001_episode_progress.sql",
);

/**
 * 空の local D1 へ進捗表の migration を適用する。
 *
 * @require database は表が未作成の D1 binding である（persist:false の local D1 は常に空）。
 * @ensure worker/migrations の SQL をそのまま実行し、episode_progress 表を作る。SQL を test 側に複製しない。
 * @ensure 実行に失敗したら throw する。
 */
export async function applyEpisodeProgressMigration(database: D1DatabaseBinding): Promise<void> {
  // why: D1DatabaseBinding は prepare しか持たない。migration は単一文の前提で run に渡す
  const sql = await readFile(MIGRATION_PATH, "utf8");
  const result = await database.prepare(sql).run();
  if (!result.success) {
    throw new Error("episode_progress の migration 適用に失敗");
  }
}
