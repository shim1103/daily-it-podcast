import { cleanup, render, waitFor } from "@testing-library/react";
import { createElement } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../contracts/index.ts";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { EpisodeListPage } from "./episode-list-page.tsx";

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

function renderPage(apiClient: PlaybackApiClient) {
  return render(createElement(EpisodeListPage, { apiClient, baseUrl: "https://example.test" }));
}

describe("EpisodeListPage", () => {
  it("一覧取得 loading の時、Loading を描画する", () => {
    // Given: 一覧取得が未解決の api client
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(() => new Promise<never>(() => {})),
    });

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: Loading だけが出る
    expect(container.textContent).toBe("Loading");
    expect(container.querySelector("article")).toBeNull();
  });

  it("一覧取得 error の時、Error を描画する", async () => {
    // Given: 一覧取得が失敗する api client
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(async () => ({ ok: false as const, error: "unavailable" as const })),
    });

    // When: page を render する
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.textContent).toBe("Error");
    });

    // Then: 一覧 item は出ない
    expect(container.querySelector("article")).toBeNull();
  });

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

  it("mount 時の hash から load() 完了後に select() が呼ばれ、その episodeId で詳細を取得する", async () => {
    // Given: hash に episodeId、詳細取得の呼び出し引数を記録する stub
    window.location.hash = "#ep-2";
    const getEpisode = vi.fn(async () => ({
      ok: true as const,
      data: { ...validGetEpisodeResponse, episodeId: "ep-2" },
    }));
    const apiClient = createStubApiClient({ getEpisode });

    // When: page を render する
    renderPage(apiClient);

    // Then: mount 時の hash（ep-2）で詳細取得が呼ばれる
    await waitFor(() => {
      expect(getEpisode).toHaveBeenCalledWith("ep-2");
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
    const getEpisode = vi.fn(async () => ({ ok: true as const, data: validGetEpisodeResponse }));
    const apiClient = createStubApiClient({ listEpisodes, getEpisode });
    const { unmount } = renderPage(apiClient);

    // When: load() 未解決のまま unmount し、その後で load() を解決させる
    unmount();
    resolveList?.({ ok: true, data: validListEpisodesResponse });
    await Promise.resolve();
    await Promise.resolve();

    // Then: unmount 済みのため hash 由来の詳細取得は呼ばれない
    expect(getEpisode).not.toHaveBeenCalled();
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

  it("再生前は page 直下に list の上へ audio を置かない", async () => {
    // Given: hash 空で mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("article button")).not.toBeNull();
    });

    // Then: list の直前に audio は無い
    const list = container.querySelector(".episode-list");
    expect(list).not.toBeNull();
    expect(list?.previousElementSibling).toBeNull();
    expect(container.querySelector(".episode-list-entry audio")).toBeNull();
  });

  it("再生開始後は audio が再生中 entry 内にだけある", async () => {
    // Given: hash 空で mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("article button")).not.toBeNull();
    });

    // When: play pill を click する
    const playButton = container.querySelector(".episode-play-button");
    playButton?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: list の直前に audio は無く、再生中 entry 内に audio がある
    await waitFor(() => {
      const list = container.querySelector(".episode-list");
      expect(list?.previousElementSibling).toBeNull();
      const playingEntry = container.querySelector(".episode-list-entry audio")?.closest(
        ".episode-list-entry",
      );
      expect(playingEntry).not.toBeNull();
      expect(playingEntry?.querySelector("audio")?.getAttribute("src")).toBe(
        "https://example.test/episodes/ep-1/audio",
      );
    });
  });

  it("再生は選択なしでも開始でき、manuscript は開かない", async () => {
    // Given: hash 空で mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("article button")).not.toBeNull();
    });

    // When: play pill を click する（行 select ではない）
    const playButton = container.querySelector(".episode-play-button");
    playButton?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: manuscript は開かず、audio が再生用 src を持つ
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
      const audio = container.querySelector("audio");
      expect(audio?.getAttribute("src")).toBe("https://example.test/episodes/ep-1/audio");
    });
  });
});
