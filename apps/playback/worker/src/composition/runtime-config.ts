import type { R2BucketBinding } from "../infrastructure/r2/r2-episode-repository.ts";
import type { PlaybackEnv } from "./runtime-config-bindings.ts";
import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";

/**
 * repository の選択 mode。
 *
 * `"in-memory"` は local development / unit test から明示的に選ぶ。`"r2"` は本番 Adapter。
 * mode は常に明示指定が必須で、env の中身から暗黙に解決しない。
 */
export type PlaybackRepositoryMode = "in-memory" | "r2";

/** Composition Root の repository 選択に渡す明示的な local / unit test / R2 option。 */
export type PlaybackRepositoryOptions = {
  mode?: PlaybackRepositoryMode;
};

export type ValidatedPlaybackEnv =
  | { mode: "in-memory"; env: PlaybackEnv }
  | { mode: "r2"; bucket: R2BucketBinding };

/**
 * production相当の env を検証する。設定不備は Worker 内部 Error を throw し、
 * HTTP boundary の mapping は Route Handler へ委譲する。
 *
 * @require env は Worker binding 由来。mode は呼び出し側が常に明示する
 * @ensure `options.mode === "in-memory"` または `options.mode === "r2"` かつ bucket binding が
 *   揃う時だけ正常終了する。mode 未指定は throw する（無言 fallback をしない）
 * @invariant secret 値を message に含めない
 */
export function validatePlaybackEnv(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): ValidatedPlaybackEnv {
  if (options.mode === "in-memory") {
    return { mode: "in-memory", env };
  }

  if (options.mode === "r2") {
    if (env.EPISODES === undefined) {
      throw new PlaybackRuntimeConfigError("EPISODES（R2 binding）が未設定です");
    }
    return { mode: "r2", bucket: env.EPISODES };
  }

  throw new PlaybackRuntimeConfigError(
    'Playback runtime config が不正です: mode が未指定です（"in-memory" または "r2" を明示してください）',
  );
}
