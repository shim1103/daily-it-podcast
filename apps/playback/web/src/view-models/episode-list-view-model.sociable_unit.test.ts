import { act, renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { useEpisodeListViewModel } from "./episode-list-view-model.ts";

const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: [
    {
      episodeId: "ep-1",
      date: "2026-08-17",
      title: "題1",
      durationSec: 60,
      topics: [{ title: "小題1" }],
      audioRef: "/episodes/ep-1/audio",
    },
    {
      episodeId: "ep-2",
      date: "2026-08-18",
      title: "題2",
      durationSec: 90,
      topics: [{ title: "小題2" }],
      audioRef: "/episodes/ep-2/audio",
    },
  ],
};

const validGetEpisodeResponse: GetEpisodeResponse = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 30 }],
    closing: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

function createStubApiClient(overrides: Partial<PlaybackApiClient> = {}): PlaybackApiClient {
  return {
    listEpisodes: vi.fn(async () => ({ ok: true as const, data: validListEpisodesResponse })),
    getEpisode: vi.fn(async () => ({ ok: true as const, data: validGetEpisodeResponse })),
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
      selection: { kind: "none" },
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
    it("一覧 success 後に select する時、selection が確定し、詳細を loading → success で持つ", async () => {
      // Given: 一覧 load 済みの hook
      const apiClient = createStubApiClient();
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      // When: episode を select する（詳細取得は直後は未解決）
      let resolveGetEpisode:
        | ((value: Awaited<ReturnType<PlaybackApiClient["getEpisode"]>>) => void)
        | undefined;
      vi.mocked(apiClient.getEpisode).mockImplementationOnce(
        () =>
          new Promise((resolve) => {
            resolveGetEpisode = resolve;
          }),
      );

      let selectPromise: Promise<void>;
      act(() => {
        selectPromise = result.current.select("ep-1");
      });

      // Then: 即座に selection が確定し、詳細は loading
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selection: { kind: "open", episodeId: "ep-1", detail: { status: "loading" } },
      });

      await act(async () => {
        resolveGetEpisode?.({ ok: true, data: validGetEpisodeResponse });
        await selectPromise;
      });

      // Then: 詳細取得後は success
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selection: {
          kind: "open",
          episodeId: "ep-1",
          detail: { status: "success", episode: validGetEpisodeResponse },
        },
      });
    });

    it("select(episodeId) は指定した episodeId を api client へそのまま渡す", async () => {
      // Given: 呼び出し引数を記録する stub api client
      const getEpisode = vi.fn(async () => ({ ok: true as const, data: validGetEpisodeResponse }));
      const apiClient = createStubApiClient({ getEpisode });
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      // When: 特定の episodeId で select を実行する
      await act(async () => {
        await result.current.select("ep-1");
      });

      // Then: 同じ episodeId が渡る
      expect(getEpisode).toHaveBeenCalledWith("ep-1");
    });

    it("詳細取得が失敗する時、selection.detail が error になる", async () => {
      // Given: 詳細取得が失敗する stub api client
      const apiClient = createStubApiClient({
        getEpisode: vi.fn(async () => ({
          ok: false as const,
          error: "episode_not_found" as const,
        })),
      });
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      // When: episode を select する
      await act(async () => {
        await result.current.select("ep-1");
      });

      // Then: selection.detail が error
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selection: { kind: "open", episodeId: "ep-1", detail: { status: "error" } },
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
        selection: { kind: "none" },
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

    it("詳細取得中に一覧が再読込された時、古い応答で state を上書きしない", async () => {
      // Given: 詳細取得が未解決の間に一覧の再読込が未解決のまま進む hook
      let resolveGetEpisode:
        | ((value: Awaited<ReturnType<PlaybackApiClient["getEpisode"]>>) => void)
        | undefined;
      const getEpisode: PlaybackApiClient["getEpisode"] = vi.fn(
        () =>
          new Promise<Awaited<ReturnType<PlaybackApiClient["getEpisode"]>>>((resolve) => {
            resolveGetEpisode = resolve;
          }),
      );
      const listEpisodes: PlaybackApiClient["listEpisodes"] = vi
        .fn()
        .mockImplementationOnce(async () => ({
          ok: true as const,
          data: validListEpisodesResponse,
        }))
        .mockImplementationOnce(() => new Promise<never>(() => {}));
      const apiClient = createStubApiClient({ getEpisode, listEpisodes });
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      let selectPromise: Promise<void> | undefined;
      act(() => {
        selectPromise = result.current.select("ep-1");
      });
      if (selectPromise === undefined) {
        throw new Error("selectPromise が未設定");
      }
      const pendingSelect = selectPromise;

      // When: 詳細取得が未解決のまま一覧を再読込し（loading へ遷移）、その後で詳細取得が解決する
      act(() => {
        void result.current.load();
      });
      await act(async () => {
        resolveGetEpisode?.({ ok: true, data: validGetEpisodeResponse });
        await pendingSelect;
      });

      // Then: 古い詳細応答で state を上書きせず、再読込後の loading のまま
      expect(result.current.state).toEqual({ status: "loading" });
    });

    it("詳細取得中に別の episode が選択された時、古い応答で selection.detail を上書きしない", async () => {
      // Given: 詳細取得が未解決の間に別 episode が選択される hook
      let resolveFirstGetEpisode:
        | ((value: Awaited<ReturnType<PlaybackApiClient["getEpisode"]>>) => void)
        | undefined;
      const getEpisode: PlaybackApiClient["getEpisode"] = vi
        .fn()
        .mockImplementationOnce(
          () =>
            new Promise((resolve) => {
              resolveFirstGetEpisode = resolve;
            }),
        )
        .mockImplementationOnce(async () => ({ ok: true as const, data: validGetEpisodeResponse }));
      const apiClient = createStubApiClient({ getEpisode });
      const { result } = renderHook(() => useEpisodeListViewModel(apiClient));
      await act(async () => {
        await result.current.load();
      });

      let firstSelectPromise: Promise<void> | undefined;
      act(() => {
        firstSelectPromise = result.current.select("ep-1");
      });
      if (firstSelectPromise === undefined) {
        throw new Error("firstSelectPromise が未設定");
      }
      const pendingFirstSelect = firstSelectPromise;

      // When: 1件目の詳細取得が未解決のまま別 episode を選択し、その後で1件目が解決する
      await act(async () => {
        await result.current.select("ep-2");
      });
      await act(async () => {
        resolveFirstGetEpisode?.({ ok: true, data: validGetEpisodeResponse });
        await pendingFirstSelect;
      });

      // Then: 古い ep-1 の応答で state を上書きせず、ep-2 の選択が保たれる
      expect(result.current.state).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selection: {
          kind: "open",
          episodeId: "ep-2",
          detail: { status: "success", episode: validGetEpisodeResponse },
        },
      });
    });
  });
});
