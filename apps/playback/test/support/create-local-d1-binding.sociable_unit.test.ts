import { describe, expect, it } from "vitest";
import { createLocalD1Binding } from "./create-local-d1-binding.ts";

/**
 * A 足場: signature→zero のみ。getPlatformProxy 本実装の到達 NI は C。
 */
describe("createLocalD1Binding", () => {
  it("A stub は proxy なしで読取空・書込 success・changes 0 を返す", async () => {
    // Given / When: A stub 入口
    const handle = await createLocalD1Binding();

    // Then: zero value（本実装の prepare→実SQL は C）
    expect(await handle.database.prepare("SELECT 1 AS n").first()).toBeNull();
    expect((await handle.database.prepare("SELECT 1 AS n").all()).results).toEqual([]);
    const written = await handle.database.prepare("INSERT INTO t VALUES (1)").run();
    expect(written.success).toBe(true);
    expect(written.meta.changes).toBe(0);
    await handle.dispose();
  });
});
