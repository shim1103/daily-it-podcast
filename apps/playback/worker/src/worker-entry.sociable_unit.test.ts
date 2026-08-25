import { describe, expect, it } from "vitest";
import { fetch as routeFetch } from "./routes/fetch.ts";
import workerEntry from "./worker-entry.ts";

describe("worker-entry", () => {
  it("default export の fetch は routes/fetch の fetch と同一参照である", () => {
    // Given: Cloudflare Workers module entry の契約
    // When: worker-entry の default.fetch を見る
    // Then: HTTP 入口本体（routes/fetch）と同じ関数を指す
    expect(workerEntry.fetch).toBe(routeFetch);
  });
});
