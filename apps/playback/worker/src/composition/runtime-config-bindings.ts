import type { R2BucketBinding } from "../infrastructure/r2/r2-episode-repository.ts";
import type { D1DatabaseBinding } from "../infrastructure/d1/d1-database-binding.ts";
import { EPISODE_PROGRESS_D1_BINDING } from "../infrastructure/d1/progress-d1-constants.ts";

/**
 * Cloudflare Workers R2 bindingから受け取るruntime config。
 *
 * @invariant `EPISODES`は`wrangler.jsonc`が定義するbinding名と一致する
 */
export type PlaybackR2Bindings = {
  EPISODES?: R2BucketBinding;
};

/**
 * Cloudflare Workers D1 bindingから受け取るruntime config。
 *
 * @invariant key 名は {@link EPISODE_PROGRESS_D1_BINDING} と一致する
 */
export type PlaybackD1Bindings = {
  [EPISODE_PROGRESS_D1_BINDING]?: D1DatabaseBinding;
};

/**
 * Playback WorkerがCloudflare Workers bindingsから受け取るruntime config。
 *
 * @require `fetch(request, env)`の`env`にWorker bindingが注入される
 * @ensure Generator / Web Clientのruntime configを含めず、Playback Workerのkeyだけを扱う
 * @invariant secretの値をError messageやlogに含めない
 */
export type PlaybackEnv = PlaybackR2Bindings & PlaybackD1Bindings;
