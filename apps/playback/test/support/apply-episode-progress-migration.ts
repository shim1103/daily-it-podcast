import { readFile, readdir } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import type { D1DatabaseBinding } from "../../worker/src/infrastructure/d1/d1-database-binding.ts";

const MIGRATIONS_DIR = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../worker/migrations",
);

// what: drizzle-kit が migration.sql の中で文と文の間に置く区切り行
const STATEMENT_BREAKPOINT = "--> statement-breakpoint";

// why: drizzle-kit の命名（`<時刻>_<名前>`）は時刻が先頭に来るため、名前の昇順が適用順になる
async function listMigrationFiles(): Promise<string[]> {
  const entries = await readdir(MIGRATIONS_DIR, { withFileTypes: true });
  return entries
    .filter((entry) => entry.isDirectory())
    .map((entry) => entry.name)
    .sort()
    .map((name) => path.join(MIGRATIONS_DIR, name, "migration.sql"));
}

// why: D1DatabaseBinding は prepare しか持たず、1 回の prepare で流せるのは 1 文だけ。複数文の file は区切り行で文ごとに分ける
function splitStatements(sql: string): string[] {
  return sql.split(STATEMENT_BREAKPOINT).map((statement) => statement.trim());
}

/**
 * 空の local D1 へ進捗表の migration を適用する。
 *
 * @require database は表が未作成の D1 binding である（persist:false の local D1 は常に空）。
 * @require worker/migrations/<名前>/migration.sql は drizzle-kit の生成物で、文字列 literal に `--> statement-breakpoint` を含めない（文の区切りをその行で判定するため）。
 * @ensure worker/migrations 直下の各 migration を名前の昇順に読み、file 内の文の順に 1 文ずつ実行する。SQL を test 側に複製しない。
 * @ensure 区切り行は実行しない。
 * @ensure 実行に失敗したら throw する。
 */
export async function applyEpisodeProgressMigration(database: D1DatabaseBinding): Promise<void> {
  for (const file of await listMigrationFiles()) {
    const sql = await readFile(file, "utf8");
    for (const statement of splitStatements(sql)) {
      const result = await database.prepare(statement).run();
      if (!result.success) {
        throw new Error("migration 適用に失敗");
      }
    }
  }
}
