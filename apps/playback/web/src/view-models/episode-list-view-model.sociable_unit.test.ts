import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { useEpisodeListViewModel } from "./episode-list-view-model.ts";

const episodeBody = {
  opening: "開始",
  topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
  closing: "終了",
};

const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題1",
      durationSec: 60,
      body: episodeBody,
      audioRef: "/episodes/ep-1/audio",
    },
    {
      episodeId: "ep-2",
      date: "2026-08-18",
      title: "題2",
      durationSec: 90,
      body: {
        ...episodeBody,
        topics: [{ title: "小題2", preface: "前2", detail: "詳2", startSec: 0 }],
      },
      audioRef: "/episodes/ep-2/audio",
    },
  ],
};

function createStubApiClient(overrides: Partial<PlaybackApiClient> = {}): PlaybackApiClient {
  return {
    listEpisodes: vi.fn(async () => ({ ok: true as const, data: validListEpisodesResponse })),
    ...overrides,
  };
}

describe("useEpisodeListViewModel", () => {
  it("組み立て直後は loading state を持つ", () => {
    // Given: 未解決の Promise を返す stub api client
    const listEpisodes: PlaybackApiClient["listEpisodes"] = vi.fn(
      () => new Promise<never>(() => {}),
    );
    const apiClient = createStubApiClient({ listEpisodes });

    // When: hook を組み立てる
    const { result } = renderHook(() => useEpisodeListViewModel(apiClient));

    // Then: state は loading
    expect(result.current.state).toEqual({ status: "loading" });
  });

  it("load() が成功する時、state が episodes を持つ success になり、選択中 episode は無い", async () => {
    // Given: 成功 ApiResult を返す stub api client
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListViewModel(apiClient));

    // When: load を実行する
    await act(async () => {
      await result.current.load();
    });

    // Then: state が episodes を持つ success で選択なし
    expect(result.current.state).toEqual({
      status: "success",
      episodes: validListEpisodesResponse.episodes,
      selectedEpisodeId: null,
      selectedEpisode: null,
    });
  });

  it("load() が失敗する時、state が error になる", async () => {
    // Given: 失敗 ApiResult を返す stub api client
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(async () => ({ ok: false as const, error: "network_error" as const })),
    });
    const { result } = renderHook(() => useEpisodeListViewModel(apiClient));

    // When: load を実行する
    await act(async () => {
      await result.current.load();
    });

    // Then: state が error
    expect(result.current.state).toEqual({ status: "error" });
  });

  describe("select(episodeId)", () => {
    it("一覧 success 後に select する時、一覧から lookup して selectedEpisode を success で持つ", async () => {
      // Given: 一覧 load 済みの hook
      const apiClient = createStubApiClient();
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      // When: episode を select する
      await act(async () => {
        await result.current.select("ep-1");
      });

      // Then: 2nd fetch 無しで一覧 item が詳細になる
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: "ep-1",
        selectedEpisode: {
          status: "success",
          episode: validListEpisodesResponse.episodes[0],
        },
      });
    });

    it("一覧に無い episodeId を select する時、selectedEpisode が error になる", async () => {
      // Given: 一覧 load 済みの hook
      const apiClient = createStubApiClient();
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      // When: 存在しない episodeId を select する
      await act(async () => {
        await result.current.select("missing");
      });

      // Then: selectedEpisode が error
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: "missing",
        selectedEpisode: { status: "error" },
      });
    });

    it("既に選択中の episodeId を再度 select する時、選択を解除する", async () => {
      // Given: ep-1 を選択済みの hook
      const apiClient = createStubApiClient();
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });
      await act(async () => {
        await result.current.select("ep-1");
      });

      // When: 同じ episodeId を再度 select する
      await act(async () => {
        await result.current.select("ep-1");
      });

      // Then: 選択が解除される
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: null,
        selectedEpisode: null,
      });
    });

    it("一覧が success でない時、select は状態を変えない", async () => {
      // Given: loading state のままの hook
      const listEpisodes: PlaybackApiClient["listEpisodes"] = vi.fn(
        () => new Promise<never>(() => {}),
      );
      const apiClient = createStubApiClient({ listEpisodes });
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));

      // When: select を実行する
      await act(async () => {
        await result.current.select("ep-1");
      });

      // Then: state は loading のまま
      expect(result.current.state).toEqual({ status: "loading" });
    });
  });

  it("seek は audio 要素の currentTime を startSec へ移動して再生する", () => {
    // Given: ViewModel hook と audio 要素
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
    const audio = document.createElement("audio");
    const playSpy = vi.spyOn(audio, "play").mockResolvedValue(undefined as never);
    result.current.audioElementRef.current = audio;

    // When: seek する
    act(() => {
      result.current.seek(90);
    });

    // Then: currentTime が変わり play が呼ばれる
    expect(audio.currentTime).toBe(90);
    expect(playSpy).toHaveBeenCalled();
  });

  it("seek は audio 要素が未接続の時、何もしない", () => {
    // Given: audio ref が null の ViewModel hook
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeListViewModel(apiClient));

    // When: seek する
    act(() => {
      result.current.seek(90);
    });

    // Then: 例外なく終了する（audio 未接続のため操作なし）
    expect(result.current.audioElementRef.current).toBeNull();
  });
});
