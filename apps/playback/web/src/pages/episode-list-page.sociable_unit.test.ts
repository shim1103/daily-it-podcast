import { afterEach, describe, expect, it, vi } from "vitest";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import type { EpisodeListViewModelHandle } from "./mount-episode-list-view-model.ts";

vi.mock("./mount-episode-list-view-model.ts", async (importOriginal) => {
  const actual = await importOriginal<typeof import("./mount-episode-list-view-model.ts")>();
  return {
    ...actual,
    mountEpisodeListViewModel: vi.fn(actual.mountEpisodeListViewModel),
  };
});

import { mountEpisodeListViewModel } from "./mount-episode-list-view-model.ts";
import { createEpisodeListPage } from "./episode-list-page.ts";

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

async function flushMicrotasks(): Promise<void> {
  await Promise.resolve();
  await Promise.resolve();
}

describe("createEpisodeListPage", () => {
  afterEach(() => {
    window.location.hash = "";
    vi.mocked(mountEpisodeListViewModel).mockClear();
  });

  it("mount 時に location.hash に episodeId があれば、その episode を選択した状態で描画する", async () => {
    // Given: hash に episodeId が設定済み
    window.location.hash = "#ep-1";
    const apiClient = createStubApiClient();

    // When: page を組み立てる
    const page = createEpisodeListPage(apiClient, "https://example.test");
    await flushMicrotasks();

    // Then: 選択中 episode の詳細が描画される
    expect(page.querySelector("[data-episode-title]")).not.toBeNull();
  });

  it("episode を選択すると location.hash が episodeId になる", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const page = createEpisodeListPage(apiClient, "https://example.test");
    await flushMicrotasks();

    // When: 一覧 item をクリックする
    const item = page.querySelector("article");
    item?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await flushMicrotasks();

    // Then: hash が選択した episodeId になる
    expect(window.location.hash).toBe("#ep-1");
  });

  it("hashchange で hash が変わると、対応する episode を選択する", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const page = createEpisodeListPage(apiClient, "https://example.test");
    await flushMicrotasks();

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-1";
    window.dispatchEvent(new Event("hashchange"));
    await flushMicrotasks();

    // Then: 選択中 episode の詳細が描画される
    expect(page.querySelector("[data-episode-title]")).not.toBeNull();
  });

  it("mount 時に location.hash が空の時、episode を選択せず描画する", async () => {
    // Given: hash が未設定
    const apiClient = createStubApiClient();

    // When: page を組み立てる
    const page = createEpisodeListPage(apiClient, "https://example.test");
    await flushMicrotasks();

    // Then: 選択中 episode の詳細（manuscript）は描画されない
    expect(page.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("一覧が success になる前に hashchange が起きた時、選択状態を変えない", async () => {
    // Given: 一覧取得が未解決のまま mount した page
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(() => new Promise<never>(() => {})),
    });
    const page = createEpisodeListPage(apiClient, "https://example.test");
    await flushMicrotasks();

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-1";
    window.dispatchEvent(new Event("hashchange"));
    await flushMicrotasks();

    // Then: 一覧が未取得のため選択中 episode の詳細（manuscript）は描画されない
    expect(page.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("load() 完了時点で hash が空のままにならない ViewModel の時、mount 時の hash で select する", async () => {
    // Given: load() 完了後も selectedEpisodeId を持つ success state を最初から返す ViewModel
    //   （実装の ViewModel は load() 完了直後に selectedEpisodeId を null に戻すため hash が一旦消えるが、
    //   ここでは mount 時の hash 復元ロジックだけを独立して検証する）
    window.location.hash = "#ep-1";
    const listeners = new Set<
      (state: ReturnType<EpisodeListViewModelHandle["getState"]>) => void
    >();
    const state: ReturnType<EpisodeListViewModelHandle["getState"]> = {
      status: "success",
      episodes: validListEpisodesResponse.episodes,
      selectedEpisodeId: "ep-1",
      selectedEpisode: { status: "success", episode: validGetEpisodeResponse },
    };
    const select = vi.fn(async () => {});
    const fakeViewModel: EpisodeListViewModelHandle = {
      getState: () => state,
      subscribe: (listener) => {
        listeners.add(listener);
        return () => listeners.delete(listener);
      },
      load: async () => {
        for (const listener of listeners) {
          listener(state);
        }
      },
      select,
    };
    vi.mocked(mountEpisodeListViewModel).mockReturnValueOnce(fakeViewModel);
    const apiClient = createStubApiClient();

    // When: page を組み立てる
    createEpisodeListPage(apiClient, "https://example.test");
    await flushMicrotasks();

    // Then: mount 時の hash（ep-1）で select() を呼ぶ
    expect(select).toHaveBeenCalledWith("ep-1");
  });
});
