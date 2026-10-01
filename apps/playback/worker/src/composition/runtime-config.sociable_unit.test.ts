import { describe, expect, it } from "vitest";
import { createFakeLocalD1Binding } from "../../../test/support/create-fake-local-d1-binding.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import { validatePlaybackEnv } from "./runtime-config.ts";

describe("validatePlaybackEnv", () => {
  it("returns_in_memory_config_when_mode_is_explicit_in_memory", () => {
    // Given: local / unit test 用の明示的な in-memory mode
    // When: runtime config を検証する
    const got = validatePlaybackEnv({}, { mode: "in-memory" });

    // Then: in-memory config として返す
    expect(got).toEqual({ mode: "in-memory", env: {} });
  });

  it("returns_r2_config_when_mode_is_r2_and_episodes_and_episode_progress_bindings_exist", async () => {
    // Given: R2 binding と D1 binding を持つ env と明示的な r2 mode
    const bucket = { get: async () => null, list: async () => ({ objects: [] }) };
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODES: bucket, EPISODE_PROGRESS: database };

    // When: runtime config を検証する
    const got = validatePlaybackEnv(env, { mode: "r2" });

    // Then: r2 config として bucket と進捗 database を返す
    expect(got).toEqual({ mode: "r2", bucket, progressDatabase: database });
  });

  it("throws_when_mode_is_r2_and_episodes_binding_is_missing", async () => {
    // Given: D1 binding だけを持つ env と明示的な r2 mode
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When / Then: 未設定は明確な runtime config error
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow("EPISODES");
  });

  it("throws_when_mode_is_r2_and_episode_progress_binding_is_missing", () => {
    // Given: R2 binding だけを持つ env と明示的な r2 mode（進捗が無言で消える構成）
    const bucket = { get: async () => null, list: async () => ({ objects: [] }) };
    const env = { EPISODES: bucket };

    // When / Then: 進捗の永続先が無いまま起動せず、未設定を明確な runtime config error にする
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow("EPISODE_PROGRESS");
  });

  it("does_not_throw_when_mode_is_in_memory_and_episode_progress_binding_is_missing", () => {
    // Given: D1 binding が無い env と明示的な in-memory mode
    const env = {};

    // When / Then: in-memory は env の中身を見ない
    expect(() => validatePlaybackEnv(env, { mode: "in-memory" })).not.toThrow();
  });

  it("throws_without_silent_fallback_when_mode_is_unspecified", () => {
    // Given: mode を指定しない options
    const env = {};

    // When / Then: 暗黙の mode 解決をせず、明示指定を必須にする
    expect(() => validatePlaybackEnv(env)).toThrow(PlaybackRuntimeConfigError);
  });
});
