import { describe, expect, it } from "vitest";
import { app } from "./routes/app.ts";
import workerEntry from "./worker-entry.ts";

describe("worker-entry", () => {
  it("default export の fetch は app の Hono instance の fetch と同一参照である", () => {
    // Given: Cloudflare Workers module entry の契約
    // When: worker-entry の default.fetch を見る
    // Then: HTTP 入口本体（routes/app の Hono instance）と同じ関数を指す
    expect(workerEntry.fetch).toBe(app.fetch);
  });
});
