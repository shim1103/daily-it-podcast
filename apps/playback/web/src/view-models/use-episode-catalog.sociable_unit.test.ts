import { act, renderHook, waitFor } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { ApiSuccessData } from "../api/api-result.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { useEpisodeCatalog } from "./use-episode-catalog.ts";

// why: production（playback-state.ts）と同じく API Client の型から導出し、
//   境界共有型（contracts）を test から直接 import しない
type ListEpisodesData = ApiSuccessData<PlaybackApiClient["listEpisodes"]>;

const episodeBody = {
  opening: "開始",
  topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
  closing: "終了",
};

const validListEpisodesResponse: ListEpisodesData = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題1",
      durationSec: 60,
      body: episodeBody,
      audioRef: "/episodes/ep-1/audio",
    },
  ],
};

function createStubApiClient(overrides: Partial<PlaybackApiClient> = {}): PlaybackApiClient {
  return {
    listEpisodes: vi.fn(async () => ({ ok: true as const, data: validListEpisodesResponse })),
    ...overrides,
  };
}

describe("useEpisodeCatalog", () => {
  it("組み立て直後は loading state と空 episodes を返す", () => {
    // Given: stub api client
    const apiClient = createStubApiClient();

    // When: hook を render する
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));

    // Then: 同期時点は loading・episodes 空（auto-load の effect は render 後に走る）
    expect(result.current.catalogStatus).toEqual({ status: "loading" });
    expect(result.current.episodes).toEqual([]);
  });

  it("mount 時に自動で listEpisodes を呼び、成功なら success + episodes になる", async () => {
    // Given: 成功 ApiResult を返す stub api client
    const apiClient = createStubApiClient();

    // When: hook を render するだけ（明示 load を呼ばない）
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));

    // Then: auto-load で listEpisodes が呼ばれ success + episodes になる
    await waitFor(() => {
      expect(result.current.catalogStatus).toEqual({ status: "success" });
    });
    expect(apiClient.listEpisodes).toHaveBeenCalledTimes(1);
    expect(result.current.episodes).toEqual(validListEpisodesResponse.episodes);
  });

  it("load() が成功する時、catalogStatus が success になり episodes を保持する", async () => {
    // Given: 成功 ApiResult を返す stub api client
    const apiClient = createStubApiClient();
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));

    // When: 明示 reload として load を実行する
    await act(async () => {
      await result.current.load();
    });

    // Then: success・episodes は取得結果
    expect(result.current.catalogStatus).toEqual({ status: "success" });
    expect(result.current.episodes).toEqual(validListEpisodesResponse.episodes);
  });

  it("load() が失敗する時、例外を投げず catalogStatus が error になり episodes は空のまま", async () => {
    // Given: 失敗 ApiResult を返す stub api client
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(async () => ({ ok: false as const, error: "unavailable" as const })),
    });
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));

    // When: load を実行する（throw すれば act が reject し test は fail する）
    await act(async () => {
      await result.current.load();
    });

    // Then: error・episodes 空（view-model.md §4: 失敗は throw せず state で表現）
    expect(result.current.catalogStatus).toEqual({ status: "error" });
    expect(result.current.episodes).toEqual([]);
  });

  it("load() の再実行前に catalogStatus を loading へ戻す", async () => {
    // Given: auto-load 分と明示 1 回目は成功、明示 2 回目は未解決の stub api client
    let resolveThird: ((value: { ok: true; data: ListEpisodesData }) => void) | undefined;
    const listEpisodes = vi
      .fn<PlaybackApiClient["listEpisodes"]>()
      .mockImplementationOnce(async () => ({ ok: true as const, data: validListEpisodesResponse }))
      .mockImplementationOnce(async () => ({ ok: true as const, data: validListEpisodesResponse }))
      .mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveThird = resolve;
          }),
      );
    const apiClient = createStubApiClient({ listEpisodes });
    const { result } = renderHook(() => useEpisodeCatalog(apiClient));
    // auto-load の完了を待ってから明示 reload を検証する
    await waitFor(() => {
      expect(result.current.catalogStatus).toEqual({ status: "success" });
    });
    await act(async () => {
      await result.current.load();
    });

    // When: さらに load を起動する（完了待ちしない）
    let thirdLoad: Promise<void> = Promise.resolve();
    act(() => {
      thirdLoad = result.current.load();
    });

    // Then: 完了前は loading へ戻っている
    expect(result.current.catalogStatus).toEqual({ status: "loading" });

    // cleanup: 未解決分を解決させて hook を安定させる
    await act(async () => {
      resolveThird?.({ ok: true, data: validListEpisodesResponse });
      await thirdLoad;
    });
  });
});
