import { describe, expect, it, vi } from "vitest";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { createEpisodeListViewModel } from "./episode-list-view-model.ts";

const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: [
    { episodeId: "ep-1", date: "2026-08-17", title: "題1", durationSec: 60 },
    { episodeId: "ep-2", date: "2026-08-18", title: "題2", durationSec: 90 },
  ],
};

const validGetEpisodeResponse: GetEpisodeResponse = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題1",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
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

describe("createEpisodeListViewModel", () => {
  it("組み立て直後は loading state を持つ", () => {
    // Given: 未解決の Promise を返す stub api client
    const listEpisodes: PlaybackApiClient["listEpisodes"] = vi.fn(
      () => new Promise<never>(() => {}),
    );
    const apiClient = createStubApiClient({ listEpisodes });

    // When: ViewModel を組み立てる
    const viewModel = createEpisodeListViewModel(apiClient);

    // Then: state は loading
    expect(viewModel.getState()).toEqual({ status: "loading" });
  });

  it("load() が成功する時、state が episodes を持つ success になり、選択中 episode は無い", async () => {
    // Given: 成功 ApiResult を返す stub api client
    const apiClient = createStubApiClient();
    const viewModel = createEpisodeListViewModel(apiClient);

    // When: load を実行する
    await viewModel.load();

    // Then: state が episodes を持つ success で選択なし
    expect(viewModel.getState()).toEqual({
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
    const viewModel = createEpisodeListViewModel(apiClient);

    // When: load を実行する
    await viewModel.load();

    // Then: state が error
    expect(viewModel.getState()).toEqual({ status: "error" });
  });

  it("state が変化するたび、subscribe した listener を呼ぶ", async () => {
    // Given: 成功を返す stub api client と subscribe した listener
    const apiClient = createStubApiClient();
    const viewModel = createEpisodeListViewModel(apiClient);
    const listener = vi.fn();
    viewModel.subscribe(listener);

    // When: load を実行する
    await viewModel.load();

    // Then: loading → success の順で listener が呼ばれる
    expect(listener).toHaveBeenNthCalledWith(1, { status: "loading" });
    expect(listener).toHaveBeenNthCalledWith(2, {
      status: "success",
      episodes: validListEpisodesResponse.episodes,
      selectedEpisodeId: null,
      selectedEpisode: null,
    });
  });

  it("subscribe の戻り値を呼ぶと、以後 listener を呼ばなくなる", async () => {
    // Given: 成功を返す stub api client と unsubscribe 済みの listener
    const apiClient = createStubApiClient();
    const viewModel = createEpisodeListViewModel(apiClient);
    const listener = vi.fn();
    const unsubscribe = viewModel.subscribe(listener);
    unsubscribe();

    // When: load を実行する
    await viewModel.load();

    // Then: listener は呼ばれない
    expect(listener).not.toHaveBeenCalled();
  });

  describe("select(episodeId)", () => {
    it("一覧 success 後に select する時、selectedEpisodeId が確定し、詳細を loading → success で持つ", async () => {
      // Given: 一覧 load 済みの ViewModel
      const apiClient = createStubApiClient();
      const viewModel = createEpisodeListViewModel(apiClient);
      await viewModel.load();

      // When: episode を select する
      const selectPromise = viewModel.select("ep-1");

      // Then: 即座に selectedEpisodeId が確定し、詳細は loading
      expect(viewModel.getState()).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: "ep-1",
        selectedEpisode: { status: "loading" },
      });

      await selectPromise;

      // Then: 詳細取得後は success
      expect(viewModel.getState()).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: "ep-1",
        selectedEpisode: { status: "success", episode: validGetEpisodeResponse },
      });
    });

    it("select(episodeId) は指定した episodeId を api client へそのまま渡す", async () => {
      // Given: 呼び出し引数を記録する stub api client
      const getEpisode = vi.fn(async () => ({ ok: true as const, data: validGetEpisodeResponse }));
      const apiClient = createStubApiClient({ getEpisode });
      const viewModel = createEpisodeListViewModel(apiClient);
      await viewModel.load();

      // When: 特定の episodeId で select を実行する
      await viewModel.select("ep-1");

      // Then: 同じ episodeId が渡る
      expect(getEpisode).toHaveBeenCalledWith("ep-1");
    });

    it("詳細取得が失敗する時、selectedEpisode が error になる", async () => {
      // Given: 詳細取得が失敗する stub api client
      const apiClient = createStubApiClient({
        getEpisode: vi.fn(async () => ({
          ok: false as const,
          error: "episode_not_found" as const,
        })),
      });
      const viewModel = createEpisodeListViewModel(apiClient);
      await viewModel.load();

      // When: episode を select する
      await viewModel.select("ep-1");

      // Then: selectedEpisode が error
      expect(viewModel.getState()).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: "ep-1",
        selectedEpisode: { status: "error" },
      });
    });

    it("既に選択中の episodeId を再度 select する時、選択を解除する", async () => {
      // Given: ep-1 を選択済みの ViewModel
      const apiClient = createStubApiClient();
      const viewModel = createEpisodeListViewModel(apiClient);
      await viewModel.load();
      await viewModel.select("ep-1");

      // When: 同じ episodeId を再度 select する
      await viewModel.select("ep-1");

      // Then: 選択が解除される
      expect(viewModel.getState()).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: null,
        selectedEpisode: null,
      });
    });

    it("一覧が success でない時、select は状態を変えない", async () => {
      // Given: loading state のままの ViewModel
      const listEpisodes: PlaybackApiClient["listEpisodes"] = vi.fn(
        () => new Promise<never>(() => {}),
      );
      const apiClient = createStubApiClient({ listEpisodes });
      const viewModel = createEpisodeListViewModel(apiClient);

      // When: select を実行する
      await viewModel.select("ep-1");

      // Then: state は loading のまま
      expect(viewModel.getState()).toEqual({ status: "loading" });
    });

    it("詳細取得中に一覧が再読込された時、古い応答で state を上書きしない", async () => {
      // Given: 詳細取得が未解決の間に一覧の再読込が未解決のまま進む ViewModel
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
      const viewModel = createEpisodeListViewModel(apiClient);
      await viewModel.load();
      const selectPromise = viewModel.select("ep-1");

      // When: 詳細取得が未解決のまま一覧を再読込し（loading へ遷移）、その後で詳細取得が解決する
      void viewModel.load();
      resolveGetEpisode?.({ ok: true, data: validGetEpisodeResponse });
      await selectPromise;

      // Then: 古い詳細応答で state を上書きせず、再読込後の loading のまま
      expect(viewModel.getState()).toEqual({ status: "loading" });
    });

    it("詳細取得中に別の episode が選択された時、古い応答で selectedEpisode を上書きしない", async () => {
      // Given: 詳細取得が未解決の間に別 episode が選択される ViewModel
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
      const viewModel = createEpisodeListViewModel(apiClient);
      await viewModel.load();
      const firstSelectPromise = viewModel.select("ep-1");

      // When: 1件目の詳細取得が未解決のまま別 episode を選択し、その後で1件目が解決する
      await viewModel.select("ep-2");
      resolveFirstGetEpisode?.({ ok: true, data: validGetEpisodeResponse });
      await firstSelectPromise;

      // Then: 古い ep-1 の応答で state を上書きせず、ep-2 の選択が保たれる
      expect(viewModel.getState()).toEqual({
        status: "success",
        episodes: validListEpisodesResponse.episodes,
        selectedEpisodeId: "ep-2",
        selectedEpisode: { status: "success", episode: validGetEpisodeResponse },
      });
    });
  });
});
