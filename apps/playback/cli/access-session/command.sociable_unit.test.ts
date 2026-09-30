import { describe, expect, it } from "vitest";
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
  SECONDS_PER_DAY,
} from "./storage-state.ts";

const NOW_SEC = 1_800_000_000;
const COOKIE_VALUE = "fake-cookie-value";

type FakeIoOptions = {
  storageStateEnv?: string;
  nowSec?: () => number;
};

type FakeIo = { io: AccessSessionCommandIo; stdout: string[]; stderr: string[] };

function createFakeIo(options: FakeIoOptions = {}): FakeIo {
  const stdout: string[] = [];
  const stderr: string[] = [];
  const io: AccessSessionCommandIo = {
    storageStateEnv: options.storageStateEnv,
    nowSec: options.nowSec ?? (() => NOW_SEC),
    writeStdout: (line) => stdout.push(line),
    writeStderr: (line) => stderr.push(line),
  };
  return { io, stdout, stderr };
}

function createStorageStateJson(expiresAtSec: number): string {
  return JSON.stringify({
    cookies: [{ name: ACCESS_COOKIE_NAME, value: COOKIE_VALUE, expires: expiresAtSec }],
    origins: [],
  });
}

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

    // Then: 成功し、有効と表示する（失効間近の表示は出ず、期限だけの確認だと添える）
    expect(status).toBe(EXIT_OK);
    expect(stdout[0]).toContain("有効");
    expect(stdout[0]).not.toContain("失効間近");
    expect(stdout[0]).toContain("Access が受理するかは E2E");
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
    const { io, stdout, stderr } = createFakeIo({
      storageStateEnv: createStorageStateJson(NOW_SEC - 60),
    });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 失敗し、復旧手順の場所を示し、cookie の値は出さない
    expect(status).toBe(EXIT_FAILED);
    expect(stderr.join("\n")).toContain("失効している");
    expect(stderr.join("\n")).toContain("DEPLOY.md");
    expect(`${stdout.join("\n")}${stderr.join("\n")}`).not.toContain(COOKIE_VALUE);
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
    // Given: 存在しない subcommand（登録用の build は持たない）
    const { io, stderr } = createFakeIo();

    // When: 実行する
    const status = await runAccessSessionCommand(["build"], io);

    // Then: 使い方の終了 status を返す
    expect(status).toBe(EXIT_USAGE);
    expect(stderr[0]).toContain("使い方");
  });

  it("fails_with_a_generic_message_when_a_non_error_value_is_thrown", async () => {
    // Given: 時刻の取得が、Error ではない値で失敗する
    const { io, stderr } = createFakeIo({
      storageStateEnv: createStorageStateJson(NOW_SEC + SECONDS_PER_DAY),
      nowSec: () => {
        throw "not an error object";
      },
    });

    // When: check を実行する
    const status = await runAccessSessionCommand(["check"], io);

    // Then: 失敗し、想定外の失敗と表示する
    expect(status).toBe(EXIT_FAILED);
    expect(stderr).toEqual(["想定外の失敗"]);
  });
});
