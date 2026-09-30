// @vitest-environment node
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { EXIT_FAILED, EXIT_OK, EXIT_USAGE } from "../../cli/access-session/command.ts";
import { ACCESS_COOKIE_NAME, SECONDS_PER_DAY } from "../../cli/access-session/storage-state.ts";

/**
 * scope: Narrow Integration
 * real: Node process（`node --experimental-strip-types` で入口 main.ts を実際に起動し、環境変数・
 *   標準出力・標準エラー・終了 status を境界として使う）
 * double: none（cookie の値は偽物。Cloudflare にも GitHub にも触れない）
 *
 * 判定の分岐は command の Sociable Unit が持つ。ここは process 境界の結線だけを見る。
 *
 * @require Node が `--experimental-strip-types` を受け付ける（apps/playback/.nvmrc）
 * @ensure env PLAYWRIGHT_STORAGE_STATE_JSON が読まれ、失効は終了 status に反映される
 * @ensure 結果は標準出力（成功）・標準エラー（失敗・使い方）へ分かれて出る
 */

const MAIN_PATH = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../cli/access-session/main.ts",
);
const COOKIE_VALUE = "fake-cookie-value";

type CliResult = { status: number | null; stdout: string; stderr: string };

function nowSec(): number {
  return Math.floor(Date.now() / 1000);
}

function createStorageStateJson(expiresAtSec: number): string {
  return JSON.stringify({
    cookies: [{ name: ACCESS_COOKIE_NAME, value: COOKIE_VALUE, expires: expiresAtSec }],
    origins: [],
  });
}

function runMain(args: string[], storageStateJson?: string): CliResult {
  const env = { ...process.env };
  delete env.PLAYWRIGHT_STORAGE_STATE_JSON;
  if (storageStateJson !== undefined) {
    env.PLAYWRIGHT_STORAGE_STATE_JSON = storageStateJson;
  }
  const result = spawnSync(
    process.execPath,
    ["--experimental-strip-types", "--no-warnings", MAIN_PATH, ...args],
    { env, encoding: "utf8" },
  );
  return { status: result.status, stdout: result.stdout, stderr: result.stderr };
}

describe("access-session main", () => {
  it("succeeds_and_reports_the_expiry_on_stdout_when_the_env_session_is_alive", () => {
    // Given: 環境変数に、30 日後に失効する storageState を渡す
    const storageStateJson = createStorageStateJson(nowSec() + 30 * SECONDS_PER_DAY);

    // When: check で入口を起動する
    const result = runMain(["check"], storageStateJson);

    // Then: 成功し、期限は標準出力へ出る。cookie の値は出ない
    expect(result.status).toBe(EXIT_OK);
    expect(result.stdout).toContain("期限");
    expect(`${result.stdout}${result.stderr}`).not.toContain(COOKIE_VALUE);
  });

  it("fails_when_the_storage_state_from_the_env_is_expired", () => {
    // Given: 環境変数に、既に失効した storageState を渡す
    const storageStateJson = createStorageStateJson(nowSec() - 60);

    // When: check で入口を起動する
    const result = runMain(["check"], storageStateJson);

    // Then: 終了 status が失敗になり、標準エラーに失効を示す
    expect(result.status).toBe(EXIT_FAILED);
    expect(result.stderr).toContain("失効している");
  });

  it("returns_the_usage_status_when_no_command_is_given", () => {
    // Given: subcommand を渡さない

    // When: 入口を起動する
    const result = runMain([]);

    // Then: 使い方の終了 status を返し、使い方は標準エラーへ出る
    expect(result.status).toBe(EXIT_USAGE);
    expect(result.stderr).toContain("使い方");
    expect(result.stdout).toBe("");
  });
});
