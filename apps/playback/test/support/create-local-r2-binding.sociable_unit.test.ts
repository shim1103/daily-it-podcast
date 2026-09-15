import { describe, expect, it } from "vitest";
import { createLocalR2Binding } from "./create-local-r2-binding.ts";

describe("createLocalR2Binding", () => {
  it("returns_noop_bucket_and_dispose_when_stub", async () => {
    // Given: A stub
    // When: createLocalR2Binding する
    const handle = await createLocalR2Binding();

    // Then: bucket / dispose が使え、list は空
    expect(handle.bucket).toBeDefined();
    const listed = await handle.bucket.list();
    expect(listed.objects).toEqual([]);
    await handle.dispose();
  });
});
