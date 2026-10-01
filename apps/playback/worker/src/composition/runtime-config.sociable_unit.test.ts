import { describe, expect, it } from "vitest";
import { createFakeLocalD1Binding } from "../../../test/support/create-fake-local-d1-binding.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import { validatePlaybackEnv } from "./runtime-config.ts";

describe("validatePlaybackEnv", () => {
  it("明示的 in-memory mode の時、in-memory config を返す", () => {
    // Given: local / unit test 用の明示的な in-memory mode
    // When: runtime config を検証する
    const got = validatePlaybackEnv({}, { mode: "in-memory" });

    // Then: in-memory config として返す
    expect(got).toEqual({ mode: "in-memory", env: {} });
  });

  it("明示的 r2 mode で EPISODES と EPISODE_PROGRESS の binding が揃う時、r2 config として返す", async () => {
    // Given: R2 binding と D1 binding を持つ env と明示的な r2 mode
    const bucket = { get: async () => null, list: async () => ({ objects: [] }) };
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODES: bucket, EPISODE_PROGRESS: database };

    // When: runtime config を検証する
    const got = validatePlaybackEnv(env, { mode: "r2" });

    // Then: r2 config として bucket と進捗 database を返す
    expect(got).toEqual({ mode: "r2", bucket, progressDatabase: database });
  });

  it("明示的 r2 mode で EPISODES binding が無い時、throw する", async () => {
    // Given: D1 binding だけを持つ env と明示的な r2 mode
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When / Then: 未設定は明確な runtime config error
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow("EPISODES");
  });

  it("明示的 r2 mode で EPISODE_PROGRESS binding が無い時、throw する", () => {
    // Given: R2 binding だけを持つ env と明示的な r2 mode（進捗が無言で消える構成）
    const bucket = { get: async () => null, list: async () => ({ objects: [] }) };
    const env = { EPISODES: bucket };

    // When / Then: 進捗の永続先が無いまま起動せず、未設定を明確な runtime config error にする
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow("EPISODE_PROGRESS");
  });

  it("明示的 in-memory mode の時、EPISODE_PROGRESS binding が無くても throw しない", () => {
    // Given: D1 binding が無い env と明示的な in-memory mode
    const env = {};

    // When / Then: in-memory は env の中身を見ない
    expect(() => validatePlaybackEnv(env, { mode: "in-memory" })).not.toThrow();
  });

  it("mode が未指定の時、throw する（無言 fallback をしない）", () => {
    // Given: mode を指定しない options
    const env = {};

    // When / Then: 暗黙の mode 解決をせず、明示指定を必須にする
    expect(() => validatePlaybackEnv(env)).toThrow(PlaybackRuntimeConfigError);
  });
});
