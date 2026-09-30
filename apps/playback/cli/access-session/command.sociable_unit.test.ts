import { describe, expect, it } from "vitest";
import { createFakeAccessJwt } from "../../test/support/fake-access-jwt.ts";
import {
  type AccessSessionCommandIo,
  EXIT_FAILED,
  EXIT_OK,
  EXIT_USAGE,
  runAccessSessionCommand,
} from "./command.ts";
import {
  ACCESS_COOKIE_NAME,
  ACCESS_SESSION_WARNING_SEC,
  buildAccessSessionStorageState,
  SECONDS_PER_DAY,
} from "./storage-state.ts";

const ORIGIN = "https://playback.example.workers.dev";
const NOW_SEC = 1_800_000_000;

type FakeIoOptions = {
  stdin?: string;
  storageStateEnv?: string;
  readStdin?: () => Promise<string>;
};

type FakeIo = { io: AccessSessionCommandIo; stdout: string[]; stderr: string[] };

function createFakeIo(options: FakeIoOptions = {}): FakeIo {
  const stdout: string[] = [];
  const stderr: string[] = [];
  const io: AccessSessionCommandIo = {
    readStdin: options.readStdin ?? (async () => options.stdin ?? ""),
    storageStateEnv: options.storageStateEnv,
    nowSec: () => NOW_SEC,
    writeStdout: (line) => stdout.push(line),
    writeStderr: (line) => stderr.push(line),
  };
  return { io, stdout, stderr };
}

function createStorageStateJson(expiresAtSec: number): string {
  return JSON.stringify(buildAccessSessionStorageState(ORIGIN, createFakeAccessJwt(expiresAtSec)));
}

describe("runAccessSessionCommand build", () => {
  it("writes_the_storage_state_json_to_stdout_and_the_expiry_to_stderr_when_given_a_valid_jwt", async () => {
    // Given: 30 日後に失効する JWT を標準入力から受け取る
    const jwt = createFakeAccessJwt(NOW_SEC + 30 * SECONDS_PER_DAY);
    const { io, stdout, stderr } = createFakeIo({ stdin: jwt });

    // When: build を実行する
    const status = await runAccessSessionCommand(["build", ORIGIN], io);

    // Then: 成功し、標準出力は JWT の exp を期限とする storageState の JSON 1 行、標準エラーは期限の表示
    expect(status).toBe(EXIT_OK);
    expect(stdout).toEqual([JSON.stringify(buildAccessSessionStorageState(ORIGIN, jwt))]);
    expect(stderr).toHaveLength(1);
    expect(stderr[0]).toContain("期限");
  });

  it("fails_and_writes_nothing_to_stdout_when_the_jwt_is_already_expired", async () => {
    // Given: 既に失効した JWT
    const { io, stdout, stderr } = createFakeIo({ stdin: createFakeAccessJwt(NOW_SEC - 60) });

    // When: build を実行する
    const status = await runAccessSessionCommand(["build", ORIGIN], io);

    // Then: 失敗し、登録されうる JSON を標準出力へ出さない
    expect(status).toBe(EXIT_FAILED);
    expect(stdout).toEqual([]);
    expect(stderr[0]).toContain("失効");
  });

  it("returns_the_usage_status_when_the_origin_is_missing", async () => {
    // Given: origin を渡さない
    const { io, stdout, stderr } = createFakeIo({ stdin: createFakeAccessJwt(NOW_SEC + 1) });

    // When: build を実行する
    const status = await runAccessSessionCommand(["build"], io);

    // Then: 使い方の終了 status を返し、標準出力は空
    expect(status).toBe(EXIT_USAGE);
    expect(stdout).toEqual([]);
    expect(stderr[0]).toContain("使い方");
  });

  it("returns_the_usage_status_when_the_origin_is_empty", async () => {
    // Given: 空の origin
    const { io, stdout } = createFakeIo({ stdin: createFakeAccessJwt(NOW_SEC + 1) });

    // When: build を実行する
    const status = await runAccessSessionCommand(["build", ""], io);

    // Then: 使い方の終了 status を返し、標準出力は空
    expect(status).toBe(EXIT_USAGE);
    expect(stdout).toEqual([]);
  });

  it("fails_without_echoing_the_value_when_the_input_is_not_a_jwt", async () => {
    // Given: JWT の形ではない入力
    const notJwt = "not-a-jwt-value";
    const { io, stdout, stderr } = createFakeIo({ stdin: notJwt });

    // When: build を実行する
    const status = await runAccessSessionCommand(["build", ORIGIN], io);

    // Then: 失敗し、入力の値を標準出力にも標準エラーにも出さない
    expect(status).toBe(EXIT_FAILED);
    expect(stdout).toEqual([]);
    expect(stderr.join("\n")).not.toContain(notJwt);
  });
});

describe("runAccessSessionCommand check", () => {
  it("reports_the_session_as_alive_when_the_expiry_is_beyond_the_warning_window", async () => {
    // Given: 警告の窓より 1 日先に失効する storageState
    const { io, stdout } = createFakeIo({
      storageStateEnv: createStorageStateJson(
        NOW_SEC + ACCESS_SESSION_WARNING_SEC + SECONDS_PER_DAY,
      ),
    });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 成功し、有効と表示する（失効間近の表示は出ない）
    expect(status).toBe(EXIT_OK);
    expect(stdout[0]).toContain("有効");
    expect(stdout[0]).not.toContain("失効間近");
  });

  it("reports_the_session_as_expiring_when_less_than_the_warning_window_remains", async () => {
    // Given: 3 日後に失効する storageState
    const { io, stdout } = createFakeIo({
      storageStateEnv: createStorageStateJson(NOW_SEC + 3 * SECONDS_PER_DAY),
    });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 成功するが、失効間近と表示する
    expect(status).toBe(EXIT_OK);
    expect(stdout[0]).toContain("失効間近");
  });

  it("fails_and_points_to_the_recovery_procedure_when_the_session_is_expired", async () => {
    // Given: 既に失効した storageState
    const jwt = createFakeAccessJwt(NOW_SEC - 60);
    const { io, stdout, stderr } = createFakeIo({
      storageStateEnv: JSON.stringify(buildAccessSessionStorageState(ORIGIN, jwt)),
    });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 失敗し、復旧手順の場所を示し、cookie の値は出さない
    expect(status).toBe(EXIT_FAILED);
    expect(stderr.join("\n")).toContain("失効している");
    expect(stderr.join("\n")).toContain("DEPLOY.md");
    expect(`${stdout.join("\n")}${stderr.join("\n")}`).not.toContain(jwt);
  });

  it("skips_and_succeeds_when_the_storage_state_env_is_unset", async () => {
    // Given: 環境変数が未設定
    const { io, stdout } = createFakeIo();

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 成功し、skip したと表示する
    expect(status).toBe(EXIT_OK);
    expect(stdout[0]).toContain("skip");
  });

  it("skips_and_succeeds_when_the_storage_state_env_is_blank", async () => {
    // Given: 空白だけの環境変数
    const { io, stdout } = createFakeIo({ storageStateEnv: "  \n" });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 未設定と同じく、成功して skip したと表示する
    expect(status).toBe(EXIT_OK);
    expect(stdout[0]).toContain("skip");
  });

  it("fails_and_names_the_missing_cookie_when_the_storage_state_has_no_access_cookie", async () => {
    // Given: Access cookie を持たない storageState
    const { io, stderr } = createFakeIo({
      storageStateEnv: JSON.stringify({ cookies: [], origins: [] }),
    });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 失敗し、探した cookie の名前を示す
    expect(status).toBe(EXIT_FAILED);
    expect(stderr[0]).toContain(ACCESS_COOKIE_NAME);
  });
});

describe("runAccessSessionCommand dispatch", () => {
  it("returns_the_usage_status_when_no_command_is_given", async () => {
    // Given: subcommand を渡さない
    const { io, stderr } = createFakeIo();

    // When: 実行する
    const status = await runAccessSessionCommand([], io);

    // Then: 使い方の終了 status を返す
    expect(status).toBe(EXIT_USAGE);
    expect(stderr[0]).toContain("使い方");
  });

  it("returns_the_usage_status_when_the_command_is_unknown", async () => {
    // Given: 存在しない subcommand
    const { io, stderr } = createFakeIo();

    // When: 実行する
    const status = await runAccessSessionCommand(["register"], io);

    // Then: 使い方の終了 status を返す
    expect(status).toBe(EXIT_USAGE);
    expect(stderr[0]).toContain("使い方");
  });

  it("fails_with_a_generic_message_when_a_non_error_value_is_thrown", async () => {
    // Given: 標準入力の読み取りが Error でない値で失敗する
    const { io, stderr } = createFakeIo({
      readStdin: () => Promise.reject("not an error object"),
    });

    // When: build を実行する
    const status = await runAccessSessionCommand(["build", ORIGIN], io);

    // Then: 失敗し、想定外の失敗と表示する
    expect(status).toBe(EXIT_FAILED);
    expect(stderr).toEqual(["想定外の失敗"]);
  });
});
