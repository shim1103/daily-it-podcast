import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";

function createStubApiClient(): PlaybackApiClient {
  return { listEpisodes: vi.fn() };
}

describe("useEpisodeCatalog", () => {
  it("組み立て直後は loading state と空 episodes を返す", () => {
    // Given: stub api client
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));

    // Then: loading・episodes 空
    expect(result.current.catalogStatus).toEqual({ status: "loading" });
    expect(result.current.episodes).toEqual([]);
  });

  it("load() を呼んでも例外を投げない", async () => {
    // Given: stub api client
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));

    // When: load を実行する
    await act(async () => {
      await result.current.load();
    });

    // Then: 例外なく終了する
    expect(result.current.catalogStatus).toEqual({ status: "loading" });
  });
});
