import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";
import { useEpisodeListPage } from "./use-episode-list-page.ts";

vi.mock("./use-episode-catalog.ts", () => ({
  useEpisodeCatalog: vi.fn(),
}));

function createStubApiClient(): PlaybackApiClient {
  return { listEpisodes: vi.fn() };
}

describe("useEpisodeListPage", () => {
  it("catalog loading 時は loading state の compose ViewModel を返す", () => {
    // Given: loading catalog stub
    vi.mocked(useEpisodeCatalog).mockReturnValue({
      catalogStatus: { status: "loading" },
      episodes: [],
      load: vi.fn(async (): Promise<void> => {}),
    });
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // Then: loading・選択なし・blocking なし
    expect(result.current.catalogStatus).toEqual({ status: "loading" });
    expect(result.current.selectedEpisodeId).toBeNull();
    expect(result.current.blockingError).toBeNull();
  });

  it("catalog success 時は success state の compose ViewModel を返す", () => {
    // Given: success catalog stub
    vi.mocked(useEpisodeCatalog).mockReturnValue({
      catalogStatus: { status: "success" },
      episodes: [],
      load: vi.fn(async (): Promise<void> => {}),
    });
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeListPage(apiClient));

    // Then: success・選択なし・blocking なし
    expect(result.current.catalogStatus).toEqual({ status: "success" });
    expect(result.current.selectedEpisodeId).toBeNull();
    expect(result.current.selectedEpisode).toBeNull();
    expect(result.current.playedEpisodeId).toBeNull();
    expect(result.current.playedEpisode).toBeNull();
    expect(result.current.isPlaying).toBe(false);
    expect(result.current.blockingError).toBeNull();
    expect(result.current.episodes).toEqual([]);
    expect(result.current.isSelected("ep-1")).toBe(false);
    expect(result.current.isPlayed("ep-1")).toBe(false);
  });
});
