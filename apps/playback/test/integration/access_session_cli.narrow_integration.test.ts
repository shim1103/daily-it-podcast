// @vitest-environment node
import { spawnSync } from "node:child_process";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";
import {
  ACCESS_COOKIE_NAME,
  ACCESS_SESSION_WARNING_SEC,
  buildAccessSessionStorageState,
  SECONDS_PER_DAY,
} from "../support/access-session-storage-state.ts";
import { createFakeAccessJwt } from "../support/fake-access-jwt.ts";

/**
 * scope: Narrow Integration
 * real: Node process（`node --experimental-strip-types` で CLI を実際に起動し、標準入力・環境変数・
 *   標準出力・標準エラー・終了 status を境界として使う）
 * double: none（偽なのは署名を検証しない JWT だけ。Cloudflare にも GitHub にも触れない）
 *
 * @require Node が `--experimental-strip-types` を受け付ける（apps/playback/.nvmrc）
 * @ensure build は storageState の JSON だけを標準出力へ書き、期限を標準エラーへ書く
 * @ensure check は失効のときだけ終了 status 1 を返す
 * @invariant cookie の値を、build の JSON を除いて標準出力・標準エラーへ出さない
 */

const CLI_PATH = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  "../support/access-session-cli.ts",
);
const ORIGIN = "https://playback.example.workers.dev";

const EXIT_OK = 0;
const EXIT_FAILED = 1;
const EXIT_USAGE = 2;

type CliOptions = { stdin?: string; storageStateJson?: string };

type CliResult = { status: number | null; stdout: string; stderr: string };

function nowSec(): number {
  return Math.floor(Date.now() / 1000);
}

function runCli(args: string[], options: CliOptions = {}): CliResult {
  const env = { ...process.env };
  delete env.PLAYWRIGHT_STORAGE_STATE_JSON;
  if (options.storageStateJson !== undefined) {
    env.PLAYWRIGHT_STORAGE_STATE_JSON = options.storageStateJson;
  }
  const result = spawnSync(
    process.execPath,
    ["--experimental-strip-types", "--no-warnings", CLI_PATH, ...args],
    { input: options.stdin ?? "", env, encoding: "utf8" },
  );
  return { status: result.status, stdout: result.stdout, stderr: result.stderr };
}

function createStorageStateJson(expiresAtSec: number): string {
  return JSON.stringify(buildAccessSessionStorageState(ORIGIN, createFakeAccessJwt(expiresAtSec)));
}

describe("access-session-cli build", () => {
  it("writes_the_storage_state_json_to_stdout_and_the_expiry_to_stderr_when_given_a_valid_jwt", () => {
    // Given: 30 日後に失効する JWT
    const expiresAtSec = nowSec() + 30 * SECONDS_PER_DAY;
    const jwt = createFakeAccessJwt(expiresAtSec);

    // When: build を実行する
    const result = runCli(["build", ORIGIN], { stdin: jwt });

    // Then: 成功し、標準出力は JWT の exp を期限とする storageState の JSON、標準エラーは期限の表示
    expect(result.status).toBe(EXIT_OK);
    expect(JSON.parse(result.stdout)).toEqual(buildAccessSessionStorageState(ORIGIN, jwt));
    expect(result.stderr).toContain("期限");
  });

  it("writes_nothing_to_stdout_and_fails_when_the_jwt_is_already_expired", () => {
    // Given: 既に失効した JWT
    const jwt = createFakeAccessJwt(nowSec() - 60);

    // When: build を実行する
    const result = runCli(["build", ORIGIN], { stdin: jwt });

    // Then: 失敗し、登録されうる JSON を標準出力へ出さない
    expect(result.status).toBe(EXIT_FAILED);
    expect(result.stdout).toBe("");
    expect(result.stderr).toContain("失効");
  });

  it("fails_with_usage_status_when_the_origin_is_missing", () => {
    // Given: origin を渡さない
    const jwt = createFakeAccessJwt(nowSec() + SECONDS_PER_DAY);

    // When: build を実行する
    const result = runCli(["build"], { stdin: jwt });

    // Then: 使い方の終了 status を返し、標準出力は空
    expect(result.status).toBe(EXIT_USAGE);
    expect(result.stdout).toBe("");
  });

  it("fails_without_echoing_the_value_when_the_input_is_not_a_jwt", () => {
    // Given: JWT の形ではない入力
    const notJwt = "not-a-jwt-value";

    // When: build を実行する
    const result = runCli(["build", ORIGIN], { stdin: notJwt });

    // Then: 失敗し、入力の値を標準出力にも標準エラーにも出さない
    expect(result.status).toBe(EXIT_FAILED);
    expect(result.stdout).toBe("");
    expect(result.stderr).not.toContain(notJwt);
  });
});

describe("access-session-cli check", () => {
  it("reports_the_session_as_alive_when_the_expiry_is_beyond_the_warning_window", () => {
    // Given: 警告の窓より十分先に失効する storageState
    const storageStateJson = createStorageStateJson(
      nowSec() + ACCESS_SESSION_WARNING_SEC + SECONDS_PER_DAY,
    );

    // When: check を実行する
    const result = runCli(["check"], { storageStateJson });

    // Then: 成功し、有効と表示する（失効間近の表示は出ない）
    expect(result.status).toBe(EXIT_OK);
    expect(result.stdout).toContain("有効");
    expect(result.stdout).not.toContain("失効間近");
  });

  it("reports_the_session_as_expiring_when_less_than_the_warning_window_remains", () => {
    // Given: 3 日後に失効する storageState
    const storageStateJson = createStorageStateJson(nowSec() + 3 * SECONDS_PER_DAY);

    // When: check を実行する
    const result = runCli(["check"], { storageStateJson });

    // Then: 成功するが、失効間近と表示する
    expect(result.status).toBe(EXIT_OK);
    expect(result.stdout).toContain("失効間近");
  });

  it("fails_and_points_to_the_recovery_procedure_when_the_session_is_expired", () => {
    // Given: 既に失効した storageState
    const jwt = createFakeAccessJwt(nowSec() - 60);
    const storageStateJson = JSON.stringify(buildAccessSessionStorageState(ORIGIN, jwt));

    // When: check を実行する
    const result = runCli(["check"], { storageStateJson });

    // Then: 失敗し、復旧手順の場所を示し、cookie の値は出さない
    expect(result.status).toBe(EXIT_FAILED);
    expect(result.stderr).toContain("失効している");
    expect(result.stderr).toContain("DEPLOY.md");
    expect(`${result.stdout}${result.stderr}`).not.toContain(jwt);
  });

  it("skips_and_succeeds_when_the_storage_state_env_is_unset", () => {
    // Given: PLAYWRIGHT_STORAGE_STATE_JSON が未設定

    // When: check を実行する
    const result = runCli(["check"]);

    // Then: 成功し、skip したと表示する
    expect(result.status).toBe(EXIT_OK);
    expect(result.stdout).toContain("skip");
  });

  it("fails_when_the_storage_state_has_no_access_cookie", () => {
    // Given: Access cookie を持たない storageState
    const storageStateJson = JSON.stringify({ cookies: [], origins: [] });

    // When: check を実行する
    const result = runCli(["check"], { storageStateJson });

    // Then: 失敗し、探した cookie の名前を示す
    expect(result.status).toBe(EXIT_FAILED);
    expect(result.stderr).toContain(ACCESS_COOKIE_NAME);
  });
});

describe("access-session-cli usage", () => {
  it("fails_with_usage_status_when_no_command_is_given", () => {
    // Given: subcommand を渡さない

    // When: CLI を実行する
    const result = runCli([]);

    // Then: 使い方の終了 status を返す
    expect(result.status).toBe(EXIT_USAGE);
    expect(result.stderr).toContain("使い方");
  });
});
