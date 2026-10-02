import type { D1DatabaseBinding } from "../infrastructure/d1/d1-database-binding.ts";
import { EPISODE_PROGRESS_D1_BINDING } from "../infrastructure/d1/progress-d1-constants.ts";
import type { R2BucketBinding } from "../infrastructure/r2/r2-episode-repository.ts";
import type { PlaybackEnv } from "./runtime-config-bindings.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";

/**
 * episode の永続先の選択 mode。
 *
 * `"in-memory"` は local development / unit test から明示的に選ぶ。`"r2"` は本番 Adapter。
 * 進捗の永続先（{@link PlaybackProgressRepositoryMode}）とは独立した設定で、
 * mode は常に明示指定が必須で、env の中身から暗黙に解決しない。
 */
export type PlaybackRepositoryMode = "in-memory" | "r2";

/**
 * 進捗の永続先の選択 mode。
 *
 * `"in-memory"` は local development / unit test から明示的に選ぶ。`"d1"` は本番 Adapter。
 * episode の永続先（{@link PlaybackRepositoryMode}）とは独立した設定で、
 * mode は常に明示指定が必須で、D1 binding の有無から暗黙に解決しない。
 */
export type PlaybackProgressRepositoryMode = "in-memory" | "d1";

/**
 * Composition Root の repository 選択に渡す明示的な option。
 * `mode` は episode、`progressMode` は進捗の永続先を、互いに独立して選ぶ。
 */
export type PlaybackRepositoryOptions = {
  mode?: PlaybackRepositoryMode;
  progressMode?: PlaybackProgressRepositoryMode;
};

export type ValidatedEpisodeEnv = { mode: "in-memory" } | { mode: "r2"; bucket: R2BucketBinding };

export type ValidatedProgressEnv =
  | { mode: "in-memory" }
  | { mode: "d1"; progressDatabase: D1DatabaseBinding };

/**
 * episode の永続先に必要な設定だけを検証する。設定不備は Worker 内部 Error を throw し、
 * HTTP boundary の mapping は Route Handler へ委譲する。
 *
 * @require env は Worker binding 由来。`options.mode` は呼び出し側が常に明示する
 * @ensure `options.mode === "in-memory"` は env の中身を見ず正常終了する。`options.mode === "r2"` は
 *   `EPISODES`（R2 binding）が揃う時だけ正常終了し、欠落は throw する。`mode` が未指定、または
 *   既知の mode でない場合も throw する（無言 fallback をしない）
 * @invariant 進捗の設定（`options.progressMode`・`EPISODE_PROGRESS`）を見ない。secret 値を message に含めない
 */
export function validateEpisodeEnv(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): ValidatedEpisodeEnv {
  if (options.mode === "in-memory") {
    return { mode: "in-memory" };
  }

  if (options.mode === "r2") {
    if (env.EPISODES === undefined) {
      throw new PlaybackRuntimeConfigError("EPISODES（R2 binding）が未設定です");
    }
    return { mode: "r2", bucket: env.EPISODES };
  }

  throw new PlaybackRuntimeConfigError(
    'Playback runtime config が不正です: episode の mode が不正です（"in-memory" または "r2" を明示してください）',
  );
}

/**
 * 進捗の永続先に必要な設定だけを検証する。設定不備は Worker 内部 Error を throw し、
 * HTTP boundary の mapping は Route Handler へ委譲する。
 *
 * @require env は Worker binding 由来。`options.progressMode` は呼び出し側が常に明示する
 * @ensure `options.progressMode === "in-memory"` は env の中身を見ず正常終了する。
 *   `options.progressMode === "d1"` は `EPISODE_PROGRESS`（D1 binding）が揃う時だけ正常終了し、
 *   欠落は throw する。`progressMode` が未指定、または既知の mode でない場合も throw する
 *   （無言 fallback をしない）
 * @invariant episode の設定（`options.mode`・`EPISODES`）を見ない。secret 値を message に含めない
 */
export function validateProgressEnv(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): ValidatedProgressEnv {
  if (options.progressMode === "in-memory") {
    return { mode: "in-memory" };
  }

  if (options.progressMode === "d1") {
    const progressDatabase = env[EPISODE_PROGRESS_D1_BINDING];
    if (progressDatabase === undefined) {
      throw new PlaybackRuntimeConfigError(
        `${EPISODE_PROGRESS_D1_BINDING}（D1 binding）が未設定です`,
      );
    }
    return { mode: "d1", progressDatabase };
  }

  throw new PlaybackRuntimeConfigError(
    'Playback runtime config が不正です: 進捗の mode（progressMode）が不正です（"in-memory" または "d1" を明示してください）',
  );
}
