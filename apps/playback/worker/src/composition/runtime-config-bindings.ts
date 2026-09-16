import type { R2BucketBinding } from "../infrastructure/r2/r2-episode-repository.ts";

/** Cloudflare Workers Variablesから受け取る非secret runtime config。 */
export type PlaybackVariables = {
  GOOGLE_OAUTH_CLIENT_ID?: string;
  DRIVE_FOLDER_ID?: string;
};

/** Cloudflare Workers Secretsから受け取るcredential。 */
export type PlaybackSecrets = {
  GOOGLE_OAUTH_CLIENT_SECRET?: string;
  GOOGLE_OAUTH_REFRESH_TOKEN?: string;
};

/**
 * Cloudflare Workers R2 bindingから受け取るruntime config。
 *
 * @invariant `EPISODES`は`wrangler.jsonc`が定義するbinding名と一致する
 */
export type PlaybackR2Bindings = {
  EPISODES?: R2BucketBinding;
};

/**
 * Playback WorkerがCloudflare Workers bindingsから受け取るruntime config。
 *
 * @require `fetch(request, env)`の`env`にWorker bindingが注入される
 * @ensure Generator / Web Clientのruntime configを含めず、Playback Workerのkeyだけを扱う
 * @invariant secretの値をError messageやlogに含めない
 */
export type PlaybackEnv = PlaybackVariables & PlaybackSecrets & PlaybackR2Bindings;
