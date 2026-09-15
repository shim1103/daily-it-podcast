import { describe, expect, it } from "vitest";
import { createFakeLocalR2Binding } from "./create-fake-local-r2-binding.ts";

describe("createFakeLocalR2Binding", () => {
  it("returns_noop_bucket_and_dispose_without_starting_proxy", async () => {
    // Given: SU 用 Fake（実 getPlatformProxy を起動しない）
    // When: createFakeLocalR2Binding する
    const handle = await createFakeLocalR2Binding();

    // Then: bucket / dispose が使え、list は空
    expect(handle.bucket).toBeDefined();
    const listed = await handle.bucket.list();
    expect(listed.objects).toEqual([]);
    expect(await handle.bucket.get("missing")).toBeNull();
    await handle.dispose();
  });
});
