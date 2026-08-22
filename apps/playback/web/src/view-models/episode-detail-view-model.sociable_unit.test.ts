import { describe, expect, it, vi } from "vitest";
import type { GetEpisodeResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { createEpisodeDetailViewModel } from "./episode-detail-view-model.ts";

const validGetEpisodeResponse: GetEpisodeResponse = {
  episodeId: "ep-1",
  date: "2026-08-17",
  title: "題",
  durationSec: 60,
  body: {
    opening: "開始",
    topics: [{ title: "題", preface: "前置き", detail: "詳細", startSec: 0 }],
    closing: "終了",
  },
  audioRef: "/episodes/ep-1/audio",
};

function createStubApiClient(getEpisode: PlaybackApiClient["getEpisode"]): PlaybackApiClient {
  return {
    listEpisodes: vi.fn(),
    getEpisode,
    fetchAudio: vi.fn(),
  };
}

describe("createEpisodeDetailViewModel", () => {
  it("組み立て直後は loading state を持つ", () => {
    // Given: 未解決の Promise を返す stub api client
    const apiClient = createStubApiClient(() => new Promise(() => {}));

    // When: ViewModel を組み立てる
    const viewModel = createEpisodeDetailViewModel(apiClient);

    // Then: state は loading
    expect(viewModel.getState()).toEqual({ status: "loading" });
  });

  it("load(episodeId) が成功する時、state が episode を持つ success になる", async () => {
    // Given: 成功 ApiResult を返す stub api client
    const apiClient = createStubApiClient(async () => ({
      ok: true,
      data: validGetEpisodeResponse,
    }));
    const viewModel = createEpisodeDetailViewModel(apiClient);

    // When: load を実行する
    await viewModel.load("ep-1");

    // Then: state が episode を持つ success
    expect(viewModel.getState()).toEqual({
      status: "success",
      episode: validGetEpisodeResponse,
    });
  });

  it("load(episodeId) が失敗する時、state が error になる", async () => {
    // Given: 失敗 ApiResult を返す stub api client
    const apiClient = createStubApiClient(async () => ({
      ok: false,
      error: "episode_not_found",
    }));
    const viewModel = createEpisodeDetailViewModel(apiClient);

    // When: load を実行する
    await viewModel.load("missing");

    // Then: state が error
    expect(viewModel.getState()).toEqual({ status: "error" });
  });

  it("state が変化するたび、subscribe した listener を呼ぶ", async () => {
    // Given: 成功を返す stub api client と subscribe した listener
    const apiClient = createStubApiClient(async () => ({
      ok: true,
      data: validGetEpisodeResponse,
    }));
    const viewModel = createEpisodeDetailViewModel(apiClient);
    const listener = vi.fn();
    viewModel.subscribe(listener);

    // When: load を実行する
    await viewModel.load("ep-1");

    // Then: loading → success の順で listener が呼ばれる
    expect(listener).toHaveBeenNthCalledWith(1, { status: "loading" });
    expect(listener).toHaveBeenNthCalledWith(2, {
      status: "success",
      episode: validGetEpisodeResponse,
    });
  });

  it("load(episodeId) は指定した episodeId を api client へそのまま渡す", async () => {
    // Given: 呼び出し引数を記録する stub api client
    const getEpisode = vi.fn(async () => ({ ok: true as const, data: validGetEpisodeResponse }));
    const apiClient = createStubApiClient(getEpisode);
    const viewModel = createEpisodeDetailViewModel(apiClient);

    // When: 特定の episodeId で load を実行する
    await viewModel.load("ep-1");

    // Then: 同じ episodeId が渡る
    expect(getEpisode).toHaveBeenCalledWith("ep-1");
  });
});
