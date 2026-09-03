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

  it("catalog loading 中は loading marker を描画し、Row / Entry / AudioControls を出さない", () => {
    // Given: listEpisodes が未解決の stub
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(() => new Promise<never>(() => {})),
    });

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: loading marker のみ
    expect(container.querySelector("[data-page-loading]")).not.toBeNull();
    expect(container.querySelector("[data-page-error]")).toBeNull();
    expect(container.querySelector(".episode-row")).toBeNull();
    expect(container.querySelector(".episode-manuscript")).toBeNull();
    expect(container.querySelector(".audio-controls")).toBeNull();
  });

  it("catalog error 時は全画面 Error UI を描画し、Row も Entry も AudioControls も出さない", async () => {
    // Given: listEpisodes が失敗する stub
    const apiClient = createStubApiClient({
      listEpisodes: vi.fn(async () => ({ ok: false as const, error: "network_error" as const })),
    });

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: 全画面 Error UI のみ
    await waitFor(() => {
      expect(container.querySelector("[data-page-error]")).not.toBeNull();
    });
    expect(container.querySelector("[data-page-loading]")).toBeNull();
    expect(container.querySelector(".episode-row")).toBeNull();
    expect(container.querySelector(".episode-manuscript")).toBeNull();
    expect(container.querySelector(".audio-controls")).toBeNull();
  });

  it("catalog success 時は Row 一覧を描画する", async () => {
    // Given: listEpisodes が成功する stub
    const apiClient = createStubApiClient();

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: Row が episode 数だけ出る
    await waitFor(() => {
      expect(container.querySelectorAll(".episode-row")).toHaveLength(2);
    });
    expect(container.querySelector("[data-page-loading]")).toBeNull();
    expect(container.querySelector("[data-page-error]")).toBeNull();
  });

  it("Row の select ボタンを押すと Entry（manuscript）が出て、もう一度で消える", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row button")).not.toBeNull();
    });

    // When: 先頭 Row の select ボタン（1 個目の button）を押す
    const selectButton = () => container.querySelector(".episode-row button") as HTMLButtonElement;
    selectButton().dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: Entry が出る
    await waitFor(() => {
      expect(container.querySelector(".episode-manuscript")).not.toBeNull();
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });

    // When: もう一度 select ボタンを押す
    selectButton().dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: Entry が消える
    await waitFor(() => {
      expect(container.querySelector(".episode-manuscript")).toBeNull();
    });
  });

  it("Row の再生ボタンを押すと AudioControls が出る（Entry は無くてよい）", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row")).not.toBeNull();
    });

    // When: 先頭 Row の再生ボタン（2 個目の button）を押す
    const buttons = container.querySelectorAll(".episode-row button");
    buttons[1]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: AudioControls が出る。Entry は出ていない
    await waitFor(() => {
      const audio = container.querySelector(".audio-controls audio");
      expect(audio).not.toBeNull();
      expect(audio?.getAttribute("src")).toBe("https://example.test/episodes/ep-1/audio");
    });
    expect(container.querySelector(".episode-manuscript")).toBeNull();
  });

  it("再生中に deselect しても AudioControls は残る（selection と playback の直交）", async () => {
    // Given: ep-1 を select して再生した page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row")).not.toBeNull();
    });
    const buttons = () => container.querySelectorAll(".episode-row button");
    buttons()[0]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await waitFor(() => {
      expect(container.querySelector(".episode-manuscript")).not.toBeNull();
    });
    buttons()[1]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await waitFor(() => {
      expect(container.querySelector(".audio-controls")).not.toBeNull();
    });

    // When: deselect する（select ボタンをもう一度）
    buttons()[0]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: Entry は消えるが AudioControls は残る
    await waitFor(() => {
      expect(container.querySelector(".episode-manuscript")).toBeNull();
    });
    expect(container.querySelector(".audio-controls")).not.toBeNull();
  });

  it("mount 時に location.hash に実在 id があれば、その episode の Entry を描画する", async () => {
    // Given: render 前に hash へ実在 id をセット
    window.location.hash = "#ep-1";
    const apiClient = createStubApiClient();

    // When: page を render する
    const { container } = renderPage(apiClient);

    // Then: catalog success 後に初期 hash 由来で選択され manuscript が描画される
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });
  });

  it("mount 時に location.hash が空なら Entry を描画しない", async () => {
    // Given: hash 空
    const apiClient = createStubApiClient();

    // When: page を render する
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row")).not.toBeNull();
    });

    // Then: manuscript は出ない
    expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("mount 時の hash 由来 select が hash への無限書き戻しを起こさない", async () => {
    // Given: render 前に hash へ実在 id をセット
    window.location.hash = "#ep-1";
    const listEpisodes = vi.fn(async () => ({
      ok: true as const,
      data: validListEpisodesResponse,
    }));
    const apiClient = createStubApiClient({ listEpisodes });

    // When: page を render し、選択が反映されるまで待つ
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });

    // Then: hash は #ep-1 のまま・listEpisodes は 1 回のみ（再 render ループで再取得しない）
    await new Promise((resolve) => setTimeout(resolve, 50));
    expect(window.location.hash).toBe("#ep-1");
    expect(listEpisodes).toHaveBeenCalledTimes(1);
    expect(container.querySelectorAll(".episode-manuscript")).toHaveLength(1);
  });

  it("mount 時に location.hash に一覧に無い id があっても Entry を描画しない", async () => {
    // Given: render 前に hash へ実在しない id をセット
    window.location.hash = "#ghost";
    const apiClient = createStubApiClient();

    // When: page を render する
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row")).not.toBeNull();
    });

    // Then: manuscript は出ない（select は no-op）
    expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("hashchange で hash が実在 id へ変わると、その episode の Entry を描画する", async () => {
    // Given: mount 済みの page（hash 空で一覧のみ）
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row")).not.toBeNull();
    });

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-1";
    window.dispatchEvent(new Event("hashchange"));

    // Then: 選択中 episode の manuscript が描画される
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });
  });

  it("選択中に hash を空にすると、Entry が消える", async () => {
    // Given: mount 後に hashchange で ep-1 を選択した page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row")).not.toBeNull();
    });
    window.location.hash = "#ep-1";
    window.dispatchEvent(new Event("hashchange"));
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).not.toBeNull();
    });

    // When: hash を空にして hashchange を発火する
    window.location.hash = "";
    window.dispatchEvent(new Event("hashchange"));

    // Then: Entry が消える
    await waitFor(() => {
      expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
    });
  });
});
