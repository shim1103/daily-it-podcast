import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";

const MIGRATIONS_DIR = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../worker/migrations",
);

// why: D1DatabaseBinding は prepare しか持たず、1 回の prepare で流せるのは 1 文だけ。複数文の file は文ごとに分ける
function splitStatements(sql: string): string[] {
  const withoutComments = sql
    .split("\n")
    .filter((line) => !line.trimStart().startsWith("--"))
    .join("\n");
  return withoutComments
    .split(";")
    .map((statement) => statement.trim())
    .filter((statement) => statement !== "");
}

/**
 * 空の local D1 へ進捗表の migration を全部適用する。
 *
 * @require database は表が未作成の D1 binding である（persist:false の local D1 は常に空）。
 * @require worker/migrations の SQL は、文字列 literal に `;` を含めない（文の区切りを `;` で判定するため）。
 * @ensure worker/migrations の *.sql を名前の昇順で全部、file 内の文の順に実行する。SQL を test 側に複製しない。
 * @ensure `--` で始まる行は実行しない。空の文は実行しない。
 * @ensure 実行に失敗したら throw する。
 */
export async function applyEpisodeProgressMigration(database: D1DatabaseBinding): Promise<void> {
  const files = (await readdir(MIGRATIONS_DIR)).filter((name) => name.endsWith(".sql")).sort();
  for (const file of files) {
    const sql = await readFile(path.join(MIGRATIONS_DIR, file), "utf8");
    for (const statement of splitStatements(sql)) {
      const result = await database.prepare(statement).run();
      if (!result.success) {
        throw new Error(`${file} の migration 適用に失敗`);
      }
    }
  }
}
