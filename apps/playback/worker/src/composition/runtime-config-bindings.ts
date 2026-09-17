import type { R2BucketBinding } from "../infrastructure/r2/r2-episode-repository.ts";

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
export type PlaybackEnv = PlaybackR2Bindings;
