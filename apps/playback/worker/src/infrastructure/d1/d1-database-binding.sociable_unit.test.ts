import { describe, expect, it } from "vitest";
import { StubD1Database } from "./d1-database-binding.ts";

/**
 * real: StubD1Database（A 足場。zero value のみ）
 */
describe("StubD1Database", () => {
  it("prepare→first は null を返す", async () => {
    const db = new StubD1Database();

    const got = await db.prepare("SELECT 1").bind("ep-1").first();

    expect(got).toBeNull();
  });

  it("prepare→all は空 results を返す", async () => {
    const db = new StubD1Database();

    const got = await db.prepare("SELECT 1").all();

    expect(got).toEqual({ results: [] });
  });

  it("prepare→run は success と changes 0 を返す", async () => {
    const db = new StubD1Database();

    const got = await db.prepare("UPDATE x").run();

    expect(got).toEqual({ success: true, meta: { changes: 0 } });
  });
});
