import { afterEach, describe, expect, it, vi } from "vitest";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
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
});
