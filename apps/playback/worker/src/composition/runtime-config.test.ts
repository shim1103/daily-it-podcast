import { describe, expect, it } from "vitest";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import { validatePlaybackEnv } from "./runtime-config.ts";

describe("validatePlaybackEnv", () => {
  it("4 key が全て揃う時、検証を通す", () => {
    // Given: Drive 接続に必要な env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
      DRIVE_FOLDER_ID: "folder-id",
    };

    // When: runtime config を検証する
    const got = validatePlaybackEnv(env);

    // Then: 検証済みの値を Drive config として返す
    expect(got).toEqual({ mode: "drive", env });
  });

  it("4 key の欠落原因を個別に含む message を throw する", () => {
    // Given: 4 key が全て欠落した production 相当 env
    const env = {};

    // When: runtime config を検証する
    const act = () => validatePlaybackEnv(env);

    // Then: 1つの Error に全原因が含まれる
    expect(act).toThrow(PlaybackRuntimeConfigError);
    expect(act).toThrow(
      "GOOGLE_OAUTH_CLIENT_ID が未設定です; GOOGLE_OAUTH_CLIENT_SECRET が未設定です; GOOGLE_OAUTH_REFRESH_TOKEN が未設定です; DRIVE_FOLDER_ID が未設定です",
    );
  });

  it.each([
    ["GOOGLE_OAUTH_CLIENT_ID", "client id"],
    ["GOOGLE_OAUTH_CLIENT_SECRET", "client secret"],
    ["GOOGLE_OAUTH_REFRESH_TOKEN", "refresh token"],
    ["DRIVE_FOLDER_ID", "folder id"],
  ] as const)("空白の %s は %s の欠落として扱う", (key, _label) => {
    // Given: 1 key だけ空白の env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token",
      DRIVE_FOLDER_ID: "folder-id",
      [key]: "  ",
    };

    // When: runtime config を検証する
    const act = () => validatePlaybackEnv(env);

    // Then: 欠落 key の原因が message に出る
    expect(act).toThrow(`${key} が未設定です`);
  });

  it("4 key が全て undefined の明示的 in-memory mode だけを許可する", () => {
    // Given: 4 key が全て undefined の local / unit test 用 env
    // When: 明示的 in-memory mode で検証する
    const got = validatePlaybackEnv({}, { mode: "in-memory" });

    // Then: in-memory config として返す
    expect(got).toEqual({ mode: "in-memory", env: {} });
  });

  it("一部の env がある時、明示的 in-memory mode でも throw する", () => {
    // Given: 本番相当の不完全な env と local option
    const env = { GOOGLE_OAUTH_CLIENT_ID: "client-id" };

    // When: runtime config を検証する
    const act = () => validatePlaybackEnv(env, { mode: "in-memory" });

    // Then: 設定漏れを local mode に読み替えない
    expect(act).toThrow(PlaybackRuntimeConfigError);
  });

  it("診断 message に secret の値を含めない", () => {
    // Given: secret 値を含む不完全な env
    const env = {
      GOOGLE_OAUTH_CLIENT_ID: "client-id-secret",
      GOOGLE_OAUTH_CLIENT_SECRET: "client-secret-value",
      GOOGLE_OAUTH_REFRESH_TOKEN: "refresh-token-value",
    };

    // When: runtime config を検証する
    const act = () => validatePlaybackEnv(env);

    // Then: Error は内部型で、secret 値を診断へ出さない
    expect(act).toThrow(PlaybackRuntimeConfigError);
    expect(act).toThrowError(expect.not.stringContaining("client-id-secret"));
    expect(act).toThrowError(expect.not.stringContaining("client-secret-value"));
    expect(act).toThrowError(expect.not.stringContaining("refresh-token-value"));
  });
});
