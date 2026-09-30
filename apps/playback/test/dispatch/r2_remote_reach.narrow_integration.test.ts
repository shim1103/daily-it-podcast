/**
 * Scope: Narrow Integration（R2 binding が remote へ届くかの疎通。dispatch 専用）
 * 実物境界: getPlatformProxy（remoteBindings）が返す TEST R2 の EPISODES binding と、
 *   binding とは独立な remote 読み取り（wrangler CLI の `--remote`）
 * Double: なし。本番 bucket は使わない（共有 test/support/wrangler.smoke.jsonc の TEST bucket だけ）
 *
 * @require 共有 test/support/wrangler.smoke.jsonc の EPISODES binding（TEST bucket）と、実 Cloudflare 認証がある
 * @ensure binding で put した object が、binding とは独立な remote 経路でも読める（binding が local の模擬に落ちていない）
 * @ensure probe object は各 case の後に削除される
 */
import { afterAll, afterEach, beforeAll, describe, expect, it } from "vitest";
import {
  createRemoteTestR2Binding,
  type RemoteTestR2BindingHandle,
} from "../support/create-remote-test-r2-binding.ts";
import { readRemoteEpisodesObject } from "./remote-read.ts";

type R2BucketWithWrite = {
  put(key: string, value: string): Promise<unknown>;
  delete(key: string): Promise<void>;
};

const probeKey = `r2-remote-reach-probe-${process.env.GITHUB_RUN_ID ?? "manual"}.txt`;
const probeBody = "remote-reach-probe";

describe("TEST R2 EPISODES binding", () => {
  let handle: RemoteTestR2BindingHandle;
  let bucket: R2BucketWithWrite;

  beforeAll(async () => {
    handle = await createRemoteTestR2Binding();
    // why: R2BucketBinding は playback が使う読取面（get / list）だけ。疎通は put / delete も要る
    bucket = handle.bucket as unknown as R2BucketWithWrite;
  });

  afterEach(async () => {
    await bucket.delete(probeKey);
  });

  afterAll(async () => {
    await handle.dispose();
  });

  it("shows_the_object_to_the_remote_bucket_when_put_through_the_binding", async () => {
    // Given: binding で put 済みの probe object
    await bucket.put(probeKey, probeBody);

    // When: binding とは独立な経路（wrangler CLI の --remote）で本文を読む
    const remoteBody = await readRemoteEpisodesObject(probeKey);

    // Then: 実 TEST bucket から同じ本文が読める（local の模擬へ put していれば読めず失敗する）
    expect(remoteBody).toBe(probeBody);
  });
});
