/**
 * Scope: Narrow Integration（local binding infra）
 * 実物境界: getPlatformProxy が返す Workers R2 binding
 * Double: 本番 remote R2 は使わない（remoteBindings: false）
 *
 * @require createLocalR2Binding が実 proxy を起動する
 * @ensure put→get で binding が観測可能（列 5 が刺せる入口）
 */
import { describe, expect, it } from "vitest";
import { createLocalR2Binding } from "../support/create-local-r2-binding.ts";

type R2BucketWithPut = {
  put(key: string, value: string): Promise<unknown>;
  get(key: string): Promise<{ text(): Promise<string> } | null>;
  list(): Promise<{ objects: Array<{ key: string }> }>;
};

describe("createLocalR2Binding", () => {
  it("exposes_put_get_list_on_episodes_binding_when_platform_proxy_starts", async () => {
    // Given: getPlatformProxy local binding infra
    const handle = await createLocalR2Binding();

    try {
      const bucket = handle.bucket as unknown as R2BucketWithPut;

      // When: probe object を put して get / list する
      await bucket.put("c1-infra-probe.json", '{"probe":true}');
      const got = await bucket.get("c1-infra-probe.json");
      const listed = await bucket.list();

      // Then: 実 binding として読み書きできる
      expect(got).not.toBeNull();
      if (got === null) {
        throw new Error("get が null");
      }
      expect(await got.text()).toBe('{"probe":true}');
      expect(listed.objects.some((o) => o.key === "c1-infra-probe.json")).toBe(true);
    } finally {
      await handle.dispose();
    }
  });
});
