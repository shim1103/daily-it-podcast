import { describe, expect, it } from "vitest";
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

  it("明示的 r2 mode で EPISODES binding がある時、r2 config として返す", () => {
    // Given: R2 binding だけを持つ env と明示的な r2 mode
    const bucket = { get: async () => null, list: async () => ({ objects: [] }) };
    const env = { EPISODES: bucket };

    // When: runtime config を検証する
    const got = validatePlaybackEnv(env, { mode: "r2" });

    // Then: r2 config として bucket を返す
    expect(got).toEqual({ mode: "r2", bucket });
  });

  it("明示的 r2 mode で EPISODES binding が無い時、throw する", () => {
    // Given: R2 binding が無い env と明示的な r2 mode
    const env = {};

    // When / Then: 未設定は明確な runtime config error
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validatePlaybackEnv(env, { mode: "r2" })).toThrow("EPISODES");
  });

  it("mode が未指定の時、throw する（無言 fallback をしない）", () => {
    // Given: mode を指定しない options
    const env = {};

    // When / Then: 暗黙の mode 解決をせず、明示指定を必須にする
    expect(() => validatePlaybackEnv(env)).toThrow(PlaybackRuntimeConfigError);
  });
});
