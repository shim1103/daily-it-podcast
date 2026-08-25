import { PlaybackRuntimeConfigError } from "./runtime-config-error.ts";

/**
 * Playback Worker が Cloudflare Workers native secrets/vars から受け取る runtime config。
 *
 * @require `fetch(request, env)` の `env` に Worker binding が注入される
 * @ensure Generator / Web Client の runtime config を含めず、Playback Worker の key だけを扱う
 * @invariant secret の値を Error message や log に含めない
 */
export type PlaybackEnv = {
  GOOGLE_OAUTH_CLIENT_ID?: string;
  GOOGLE_OAUTH_CLIENT_SECRET?: string;
  GOOGLE_OAUTH_REFRESH_TOKEN?: string;
  DRIVE_FOLDER_ID?: string;
};

/** local development / unit test から明示的に選ぶ repository mode。 */
export type PlaybackRepositoryMode = "in-memory";

/** Composition Root の repository 選択に渡す明示的な local / unit test option。 */
export type PlaybackRepositoryOptions = {
  mode?: PlaybackRepositoryMode;
};

export type ValidatedPlaybackEnv =
  | { mode: "drive"; env: Required<PlaybackEnv> }
  | { mode: "in-memory"; env: PlaybackEnv };

function validateClientId(env: PlaybackEnv): string | undefined {
  return env.GOOGLE_OAUTH_CLIENT_ID?.trim() ? undefined : "GOOGLE_OAUTH_CLIENT_ID が未設定です";
}

function validateClientSecret(env: PlaybackEnv): string | undefined {
  return env.GOOGLE_OAUTH_CLIENT_SECRET?.trim()
    ? undefined
    : "GOOGLE_OAUTH_CLIENT_SECRET が未設定です";
}

function validateRefreshToken(env: PlaybackEnv): string | undefined {
  return env.GOOGLE_OAUTH_REFRESH_TOKEN?.trim()
    ? undefined
    : "GOOGLE_OAUTH_REFRESH_TOKEN が未設定です";
}

function validateFolderId(env: PlaybackEnv): string | undefined {
  return env.DRIVE_FOLDER_ID?.trim() ? undefined : "DRIVE_FOLDER_ID が未設定です";
}

function configuredValue(value: string | undefined, key: string): string {
  /* v8 ignore next 3 -- 呼び出し前に missingReasons で同条件を検証済みのため、この分岐は実行時に到達しない。型上 string | undefined のままの value を narrow するための防御 */
  if (!value?.trim()) {
    throw new PlaybackRuntimeConfigError(`${key} が未設定です`);
  }
  return value;
}

function isExplicitInMemoryMode(env: PlaybackEnv, options: PlaybackRepositoryOptions): boolean {
  return (
    options.mode === "in-memory" &&
    env.GOOGLE_OAUTH_CLIENT_ID === undefined &&
    env.GOOGLE_OAUTH_CLIENT_SECRET === undefined &&
    env.GOOGLE_OAUTH_REFRESH_TOKEN === undefined &&
    env.DRIVE_FOLDER_ID === undefined
  );
}

/**
 * production相当の env を検証する。設定不備は Worker 内部 Error を throw し、
 * HTTP boundary の mapping は Route Handler へ委譲する。
 *
 * @require env は Worker binding 由来。local / unit test の時だけ mode を明示する
 * @ensure 4 key が有効、または明示的 in-memory mode の時だけ正常終了する
 * @invariant `misconfigured` を返さず、secret 値を message に含めない
 */
export function validatePlaybackEnv(
  env: PlaybackEnv,
  options: PlaybackRepositoryOptions = {},
): ValidatedPlaybackEnv {
  if (isExplicitInMemoryMode(env, options)) {
    return { mode: "in-memory", env };
  }

  const missingReasons = [
    validateClientId(env),
    validateClientSecret(env),
    validateRefreshToken(env),
    validateFolderId(env),
  ].filter((reason): reason is string => reason !== undefined);

  if (missingReasons.length > 0) {
    throw new PlaybackRuntimeConfigError(
      `Playback runtime config が不正です: ${missingReasons.join("; ")}`,
    );
  }

  return {
    mode: "drive",
    env: {
      GOOGLE_OAUTH_CLIENT_ID: configuredValue(env.GOOGLE_OAUTH_CLIENT_ID, "GOOGLE_OAUTH_CLIENT_ID"),
      GOOGLE_OAUTH_CLIENT_SECRET: configuredValue(
        env.GOOGLE_OAUTH_CLIENT_SECRET,
        "GOOGLE_OAUTH_CLIENT_SECRET",
      ),
      GOOGLE_OAUTH_REFRESH_TOKEN: configuredValue(
        env.GOOGLE_OAUTH_REFRESH_TOKEN,
        "GOOGLE_OAUTH_REFRESH_TOKEN",
      ),
      DRIVE_FOLDER_ID: configuredValue(env.DRIVE_FOLDER_ID, "DRIVE_FOLDER_ID"),
    },
  };
}
