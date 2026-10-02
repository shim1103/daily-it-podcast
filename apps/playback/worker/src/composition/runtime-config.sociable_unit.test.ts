import { describe, expect, it } from "vitest";
import { createFakeLocalD1Binding } from "../../../test/support/create-fake-local-d1-binding.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";
import {
  validateEpisodeEnv,
  validateProgressEnv,
  type PlaybackRepositoryOptions,
} from "./runtime-config.ts";

const bucket = { get: async () => null, list: async () => ({ objects: [] }) };

// why: 型は不正な mode を弾くため、JS 呼び出し側や設定の取り違えを模して不正な文字列を渡すには assertion が要る
const invalidModeOptions = {
  mode: "sqlite",
  progressMode: "sqlite",
} as unknown as PlaybackRepositoryOptions;

describe("validateEpisodeEnv", () => {
  it("returns_in_memory_episode_config_when_mode_is_explicit_in_memory", () => {
    // Given: local / unit test 用の明示的な episode in-memory mode と空 env
    // When: episode の runtime config を検証する
    const got = validateEpisodeEnv({}, { mode: "in-memory" });

    // Then: env の中身を見ず in-memory config として返す
    expect(got).toEqual({ mode: "in-memory" });
  });

  it("returns_r2_episode_config_with_bucket_when_mode_is_r2_and_episodes_binding_exists", () => {
    // Given: R2 binding を持つ env と、本番と同じ episode r2 ＋ 進捗 d1 の明示 option
    const env = { EPISODES: bucket };

    // When: episode の runtime config を検証する
    const got = validateEpisodeEnv(env, { mode: "r2", progressMode: "d1" });

    // Then: r2 config として bucket を返す
    expect(got).toEqual({ mode: "r2", bucket });
  });

  it("throws_when_mode_is_r2_and_episodes_binding_is_missing", () => {
    // Given: EPISODES を持たない env と明示的な episode r2 mode
    const env = {};

    // When / Then: 未設定は binding 名を含む runtime config error
    expect(() => validateEpisodeEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validateEpisodeEnv(env, { mode: "r2" })).toThrow("EPISODES");
  });

  it("does_not_throw_when_mode_is_in_memory_and_episodes_binding_is_missing", () => {
    // Given: EPISODES を持たない env と明示的な episode in-memory mode
    const env = {};

    // When / Then: in-memory は env の中身を見ない
    expect(() => validateEpisodeEnv(env, { mode: "in-memory" })).not.toThrow();
  });

  it("throws_without_silent_fallback_when_mode_is_unspecified", () => {
    // Given: episode の mode を指定しない option（進捗 mode だけ明示）
    const env = { EPISODES: bucket };

    // When / Then: 暗黙の mode 解決をせず、どの設定が不正かを message に残す
    expect(() => validateEpisodeEnv(env, { progressMode: "in-memory" })).toThrow(
      PlaybackRuntimeConfigError,
    );
    expect(() => validateEpisodeEnv(env, { progressMode: "in-memory" })).toThrow("episode の mode");
  });

  it("throws_when_mode_is_neither_in_memory_nor_r2", () => {
    // Given: episode の mode が in-memory でも r2 でもない文字列
    const env = { EPISODES: bucket };

    // When / Then: 既知の mode 以外は runtime config error にする
    expect(() => validateEpisodeEnv(env, invalidModeOptions)).toThrow(PlaybackRuntimeConfigError);
  });

  it("validates_r2_episode_when_progress_mode_is_in_memory_and_d1_binding_is_missing", () => {
    // Given: EPISODES だけを持つ env と、episode r2 ＋ 進捗 in-memory の明示 option
    const env = { EPISODES: bucket };

    // When: episode の runtime config を検証する
    const got = validateEpisodeEnv(env, { mode: "r2", progressMode: "in-memory" });

    // Then: 進捗の設定に従属せず、D1 binding が無くても r2 config が成立する
    expect(got).toEqual({ mode: "r2", bucket });
  });

  it("validates_in_memory_episode_when_progress_mode_is_d1_and_episodes_binding_is_missing", async () => {
    // Given: D1 binding だけを持つ env と、episode in-memory ＋ 進捗 d1 の明示 option
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When: episode の runtime config を検証する
    const got = validateEpisodeEnv(env, { mode: "in-memory", progressMode: "d1" });

    // Then: 進捗の設定に従属せず、EPISODES が無くても in-memory config が成立する
    expect(got).toEqual({ mode: "in-memory" });
  });
});

describe("validateProgressEnv", () => {
  it("returns_in_memory_progress_config_when_progress_mode_is_explicit_in_memory", () => {
    // Given: local / unit test 用の明示的な進捗 in-memory mode と空 env
    // When: 進捗の runtime config を検証する
    const got = validateProgressEnv({}, { progressMode: "in-memory" });

    // Then: env の中身を見ず in-memory config として返す
    expect(got).toEqual({ mode: "in-memory" });
  });

  it("returns_d1_progress_config_with_database_when_progress_mode_is_d1_and_binding_exists", async () => {
    // Given: D1 binding を持つ env と、本番と同じ episode r2 ＋ 進捗 d1 の明示 option
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When: 進捗の runtime config を検証する
    const got = validateProgressEnv(env, { mode: "r2", progressMode: "d1" });

    // Then: d1 config として進捗 database を返す
    expect(got).toEqual({ mode: "d1", progressDatabase: database });
  });

  it("throws_when_progress_mode_is_d1_and_episode_progress_binding_is_missing", () => {
    // Given: EPISODE_PROGRESS を持たない env と明示的な進捗 d1 mode（進捗が無言で消える構成）
    const env = { EPISODES: bucket };

    // When / Then: 進捗の永続先が無いまま起動せず、binding 名を含む runtime config error にする
    expect(() => validateProgressEnv(env, { progressMode: "d1" })).toThrow(
      PlaybackRuntimeConfigError,
    );
    expect(() => validateProgressEnv(env, { progressMode: "d1" })).toThrow("EPISODE_PROGRESS");
  });

  it("does_not_throw_when_progress_mode_is_in_memory_and_episode_progress_binding_is_missing", () => {
    // Given: D1 binding を持たない env と明示的な進捗 in-memory mode
    const env = {};

    // When / Then: in-memory は env の中身を見ない
    expect(() => validateProgressEnv(env, { progressMode: "in-memory" })).not.toThrow();
  });

  it("throws_without_silent_fallback_when_progress_mode_is_unspecified", async () => {
    // Given: D1 binding があり、episode の mode だけ明示した option（binding の有無から進捗 mode を選ばない）
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When / Then: 暗黙の mode 解決をせず、どの設定が不正かを message に残す
    expect(() => validateProgressEnv(env, { mode: "r2" })).toThrow(PlaybackRuntimeConfigError);
    expect(() => validateProgressEnv(env, { mode: "r2" })).toThrow("progressMode");
  });

  it("throws_when_progress_mode_is_neither_in_memory_nor_d1", async () => {
    // Given: 進捗の mode が in-memory でも d1 でもない文字列
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When / Then: 既知の mode 以外は runtime config error にする
    expect(() => validateProgressEnv(env, invalidModeOptions)).toThrow(PlaybackRuntimeConfigError);
  });

  it("validates_in_memory_progress_when_mode_is_r2_and_d1_binding_is_missing", () => {
    // Given: EPISODES だけを持つ env と、episode r2 ＋ 進捗 in-memory の明示 option
    const env = { EPISODES: bucket };

    // When: 進捗の runtime config を検証する
    const got = validateProgressEnv(env, { mode: "r2", progressMode: "in-memory" });

    // Then: episode の設定に従属せず、D1 binding が無くても in-memory config が成立する
    expect(got).toEqual({ mode: "in-memory" });
  });

  it("validates_d1_progress_when_mode_is_in_memory_and_episodes_binding_is_missing", async () => {
    // Given: D1 binding だけを持つ env と、episode in-memory ＋ 進捗 d1 の明示 option
    const { database } = await createFakeLocalD1Binding();
    const env = { EPISODE_PROGRESS: database };

    // When: 進捗の runtime config を検証する
    const got = validateProgressEnv(env, { mode: "in-memory", progressMode: "d1" });

    // Then: episode の設定に従属せず、EPISODES が無くても d1 config が成立する
    expect(got).toEqual({ mode: "d1", progressDatabase: database });
  });
});
