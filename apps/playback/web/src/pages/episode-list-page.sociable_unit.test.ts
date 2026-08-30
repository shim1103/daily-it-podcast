import { cleanup, render, waitFor } from "@testing-library/react";
import { createElement } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { EpisodeListPage } from "./episode-list-page.tsx";

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

function renderPage(apiClient: PlaybackApiClient) {
  return render(createElement(EpisodeListPage, { apiClient, baseUrl: "https://example.test" }));
}

describe("EpisodeListPage", () => {
  beforeEach(() => {
    window.location.hash = "";
  });

  afterEach(() => {
    // why: page を unmount して hashchange listener を外し、次テストへ持ち越さない
    cleanup();
  });

  it("mount 時に location.hash に episodeId があれば、その episode を選択した状態で描画する", async () => {
    // Given: hash に episodeId が設定済み
    window.location.hash = "#ep-1";
    const apiClient = createStubApiClient();

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: 選択中 episode の manuscript（原稿）が描画される
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });
  });

  it("episode を選択すると location.hash が episodeId になる", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("article button")).not.toBeNull();
    });

    // When: 一覧 item をクリックする
    const item = container.querySelector("article button");
    item?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: hash が選択した episodeId になる
    await waitFor(() => {
      expect(window.location.hash).toBe("#ep-1");
    });
  });

  it("hashchange で hash が変わると、対応する episode を選択する", async () => {
    // Given: mount 済みの page（hash 空で一覧のみ）
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("article button")).not.toBeNull();
    });

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-1";
    window.dispatchEvent(new Event("hashchange"));

    // Then: 選択中 episode の manuscript（原稿）が描画される
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });
  });

  it("mount 時に location.hash が空の時、episode を選択せず描画する", async () => {
    // Given: hash が未設定
    const apiClient = createStubApiClient();

    // When: page を render する
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("article button")).not.toBeNull();
    });

    // Then: 選択中 episode の詳細（manuscript）は描画されない
    expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("一覧が success になる前に hashchange が起きた時、選択状態を変えない", async () => {
    // Given: 一覧取得が未解決のまま mount した page
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(() => new Promise<never>(() => {})),
    });
    const { container } = renderPage(apiClient);

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-1";
    window.dispatchEvent(new Event("hashchange"));
    await Promise.resolve();

    // Then: 一覧が未取得のため選択中 episode の詳細（manuscript）は描画されない
    expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("mount 時の hash から load() 完了後に select() が呼ばれ、一覧 lookup で詳細を表示する", async () => {
    // Given: hash に ep-2
    window.location.hash = "#ep-2";
    const listEpisodes = vi.fn(async () => ({
      ok: true as const,
      data: validListEpisodesResponse,
    }));
    const apiClient = createStubApiClient({ listEpisodes });

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: 2nd fetch 無しで ep-2 の manuscript が描画される
    await waitFor(() => {
      expect(listEpisodes).toHaveBeenCalled();
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });
  });

  it("load() 解決前に unmount された時、完了後の hash 由来 select を実行しない", async () => {
    // Given: 一覧取得を後から解決できる stub と hash 付き mount
    window.location.hash = "#ep-1";
    let resolveList: ((value: { ok: true; data: ListEpisodesResponse }) => void) | undefined;
    const listEpisodes = vi.fn(
      () =>
        new Promise<{ ok: true; data: ListEpisodesResponse }>((resolve) => {
          resolveList = resolve;
        }),
    );
    const apiClient = createStubApiClient({ listEpisodes });
    const { unmount } = renderPage(apiClient);

    // When: load() 未解決のまま unmount し、その後で load() を解決させる
    unmount();
    resolveList?.({ ok: true, data: validListEpisodesResponse });
    await Promise.resolve();
    await Promise.resolve();

    // Then: unmount 済みのため hash 由来 select は走らない（listEpisodes は1回のみ）
    expect(listEpisodes).toHaveBeenCalledTimes(1);
  });

  it("選択中に hash を空にすると、選択が解除され manuscript が消える", async () => {
    // Given: ep-1 を選択済みの page
    window.location.hash = "#ep-1";
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });

    // When: hash を空にして hashchange を発火する
    window.location.hash = "";
    window.dispatchEvent(new Event("hashchange"));

    // Then: 選択が解除され manuscript が消える
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
    });
  });
});
