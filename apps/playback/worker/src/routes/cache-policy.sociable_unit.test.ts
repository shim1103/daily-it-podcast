import { describe, expect, it } from "vitest";
import {
  episodeListCacheHeaders,
  noStoreCacheHeaders,
  progressPullCacheHeaders,
  progressWriteCacheHeaders,
} from "./cache-policy.ts";

describe("cache-policy progress", () => {
  it("進捗 list / pull / Write は no-store を共有し Workers Cache も効かせない", () => {
    expect(episodeListCacheHeaders).toBe(noStoreCacheHeaders);
    expect(progressPullCacheHeaders).toBe(noStoreCacheHeaders);
    expect(progressWriteCacheHeaders).toBe(noStoreCacheHeaders);
    expect(noStoreCacheHeaders["Cache-Control"]).toBe("no-store");
    expect(noStoreCacheHeaders["Cloudflare-CDN-Cache-Control"]).toBe("no-store");
  });
});
