import { describe, expect, it, vi } from "vitest";
import type { ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { createEpisodeListViewModel } from "./episode-list-view-model.ts";

const validListEpisodesResponse: ListEpisodesResponse = {
  episodes: [{ episodeId: "ep-1", date: "2026-08-17", title: "題", durationSec: 60 }],
};

function createStubApiClient(listEpisodes: PlaybackApiClient["listEpisodes"]): PlaybackApiClient {
  return {
    listEpisodes,
    getEpisode: vi.fn(),
    fetchAudio: vi.fn(),
  };
}

describe("createEpisodeListViewModel", () => {
  it("組み立て直後は loading state を持つ", () => {
    // Given: 未解決の Promise を返す stub api client
    const apiClient = createStubApiClient(() => new Promise(() => {}));

    // When: ViewModel を組み立てる
    const viewModel = createEpisodeListViewModel(apiClient);

    // Then: state は loading
    expect(viewModel.getState()).toEqual({ status: "loading" });
  });

  it("load() が成功する時、state が episodes を持つ success になる", async () => {
    // Given: 成功 ApiResult を返す stub api client
    const apiClient = createStubApiClient(async () => ({
      ok: true,
      data: validListEpisodesResponse,
    }));
    const viewModel = createEpisodeListViewModel(apiClient);

    // When: load を実行する
    await viewModel.load();

    // Then: state が episodes を持つ success
    expect(viewModel.getState()).toEqual({
      status: "success",
      episodes: validListEpisodesResponse.episodes,
    });
  });

  it("load() が失敗する時、state が error になる", async () => {
    // Given: 失敗 ApiResult を返す stub api client
    const apiClient = createStubApiClient(async () => ({
      ok: false,
      error: "network_error",
    }));
    const viewModel = createEpisodeListViewModel(apiClient);

    // When: load を実行する
    await viewModel.load();

    // Then: state が error
    expect(viewModel.getState()).toEqual({ status: "error" });
  });

  it("state が変化するたび、subscribe した listener を呼ぶ", async () => {
    // Given: 成功を返す stub api client と subscribe した listener
    const apiClient = createStubApiClient(async () => ({
      ok: true,
      data: validListEpisodesResponse,
    }));
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
    });
  });

  it("subscribe の戻り値を呼ぶと、以後 listener を呼ばなくなる", async () => {
    // Given: 成功を返す stub api client と unsubscribe 済みの listener
    const apiClient = createStubApiClient(async () => ({
      ok: true,
      data: validListEpisodesResponse,
    }));
    const viewModel = createEpisodeListViewModel(apiClient);
    const listener = vi.fn();
    const unsubscribe = viewModel.subscribe(listener);
    unsubscribe();

    // When: load を実行する
    await viewModel.load();

    // Then: listener は呼ばれない
    expect(listener).not.toHaveBeenCalled();
  });
});
