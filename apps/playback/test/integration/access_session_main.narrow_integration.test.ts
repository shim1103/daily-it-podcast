// @vitest-environment node
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import { EXIT_FAILED, EXIT_OK, EXIT_USAGE } from "../../cli/access-session/command.ts";
import {
  buildAccessSessionStorageState,
  SECONDS_PER_DAY,
} from "../../cli/access-session/storage-state.ts";
import { createFakeAccessJwt } from "../support/fake-access-jwt.ts";

/**
 * scope: Narrow Integration
 * real: Node process（`node --experimental-strip-types` で入口 main.ts を実際に起動し、標準入力・環境変数・
 *   標準出力・標準エラー・終了 status を境界として使う）
 * double: none（偽なのは署名を検証しない JWT だけ。Cloudflare にも GitHub にも触れない）
 *
 * 判定の分岐は command の Sociable Unit が持つ。ここは process 境界の結線だけを見る。
 *
 * @require Node が `--experimental-strip-types` を受け付ける（apps/playback/.nvmrc）
 * @ensure build の JSON は標準出力へ、期限の表示は標準エラーへ出る（channel が分かれる）
 * @ensure env PLAYWRIGHT_STORAGE_STATE_JSON が読まれ、失効は終了 status に反映される
 */

const MAIN_PATH = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../../cli/access-session/main.ts",
);
const ORIGIN = "https://playback.example.workers.dev";

type CliOptions = { stdin?: string; storageStateJson?: string };

type CliResult = { status: number | null; stdout: string; stderr: string };

function nowSec(): number {
  return Math.floor(Date.now() / 1000);
}

function runMain(args: string[], options: CliOptions = {}): CliResult {
  const env = { ...process.env };
  delete env.PLAYWRIGHT_STORAGE_STATE_JSON;
  if (options.storageStateJson !== undefined) {
    env.PLAYWRIGHT_STORAGE_STATE_JSON = options.storageStateJson;
  }
  const result = spawnSync(
    process.execPath,
    ["--experimental-strip-types", "--no-warnings", MAIN_PATH, ...args],
    { input: options.stdin ?? "", env, encoding: "utf8" },
  );
  return { status: result.status, stdout: result.stdout, stderr: result.stderr };
}

describe("access-session main", () => {
  it("writes_the_json_to_stdout_and_the_expiry_to_stderr_when_given_a_jwt_on_stdin", () => {
    // Given: 30 日後に失効する JWT を標準入力へ渡す
    const jwt = createFakeAccessJwt(nowSec() + 30 * SECONDS_PER_DAY);

    // When: build で入口を起動する
    const result = runMain(["build", ORIGIN], { stdin: `${jwt}\n` });

    // Then: 成功し、標準出力は JSON だけ、標準エラーは期限の表示だけ（標準入力末尾の改行は無視される）
    expect(result.status).toBe(EXIT_OK);
    expect(JSON.parse(result.stdout)).toEqual(buildAccessSessionStorageState(ORIGIN, jwt));
    expect(result.stderr).toContain("期限");
    expect(result.stderr).not.toContain(jwt);
  });

  it("fails_when_the_storage_state_from_the_env_is_expired", () => {
    // Given: 環境変数に、既に失効した storageState を渡す
    const jwt = createFakeAccessJwt(nowSec() - 60);
    const storageStateJson = JSON.stringify(buildAccessSessionStorageState(ORIGIN, jwt));

    // When: check で入口を起動する
    const result = runMain(["check"], { storageStateJson });

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
