import { cleanup, render, waitFor } from "@testing-library/react";
import { createElement } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import type { ListEpisodesResponse } from "../../contracts/index.ts";
import type { PlaybackApiClient } from "../../web/src/api/playback-api-client.ts";
import { EpisodeListPage } from "../../web/src/pages/episode-list-page.tsx";

// why: happy-dom の生 <audio> は lifecycle event を発火せず play() も no-op なので、
//   phase 通知 → derive → DOM の合成経路と audio 失敗の非 blocking 伝播を検証できない。
//   `audio-element.ts` を module-level で Fake へ差し替える（既定 pass-through、install() で capture mode）
const audioFake = vi.hoisted(async () => {
  const { createFakeAudioElementModule } = await import("../../web/src/lib/audio-element.fake.ts");
  const original = await vi.importActual<typeof import("../../web/src/lib/audio-element.ts")>(
    "../../web/src/lib/audio-element.ts",
  );
  return createFakeAudioElementModule(original);
});

vi.mock("../../web/src/lib/audio-element.ts", async () => (await audioFake).module);

/**
 * scope: Broad Integration
 * real: EpisodeListPage・useEpisodeListPage 配下の catalog/selection/hash-sync/playback・
 *   全 feature/primitive Component・hash-selection-adapter（実 location.hash / hashchange）・
 *   playback-state.ts の derive 関数
 * double: PlaybackApiClient（listEpisodes を ApiResult で返す Stub）、
 *   lib/audio-element.ts（<audio> 命令 API：play/pause/currentTime/event 購読を Fake）
 * precondition: apiClient.listEpisodes が成功/失敗を返す。location.hash を各 case で操作
 * postcondition: 合成入口（Page）から見た配線・状態伝播・error 伝播・直交性が observable
 * invariant: 下位 hook の内部分岐を再 assert しない。代表 failure のみ
 */

const episodeBody = {
  opening: { text: "開始", startSec: 0 },
  topics: [{ title: "小題", preface: "前置き", detail: "詳細", startSec: 0 }],
  ending: { text: "終了", startSec: 55 },
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
  beforeEach(async () => {
    window.location.hash = "";
    // why: audio Fake は module singleton。前 case の capture mode / 失敗フラグを持ち越さない
    (await audioFake).control.reset();
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
    expect(container.querySelector(".episode-item")).toBeNull();
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
    expect(container.querySelector(".episode-item")).toBeNull();
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
      expect(container.querySelectorAll(".episode-item")).toHaveLength(2);
    });
    expect(container.querySelector("[data-page-loading]")).toBeNull();
    expect(container.querySelector("[data-page-error]")).toBeNull();
  });

  it("Row の select ボタンを押すと Entry（manuscript）が出て、もう一度で消える", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row__select")).not.toBeNull();
    });

    // When: 先頭 Row の select ボタン（1 個目の button）を押す
    const selectButton = () => container.querySelector(".episode-row__select") as HTMLButtonElement;
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

  it("選択中 Row の manuscript の topic sec bar を押すと、その episode をその位置から再生する", async () => {
    // Given: ep-1 を選択して manuscript を開いた page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-row__select")).not.toBeNull();
    });
    (container.querySelector(".episode-row__select") as HTMLButtonElement).dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    await waitFor(() => {
      expect(container.querySelector(".episode-manuscript")).not.toBeNull();
    });

    // When: topic（bookend でない最初の一件）の sec bar を押す
    const topicBar = container.querySelector(
      ".episode-manuscript .episode-topic:not(.episode-topic--bookend) [data-topic-start-sec]",
    ) as HTMLButtonElement;
    topicBar.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: fixture の topic startSec 0 で ep-1 の音源が張られる（page の onSeek 配線を通す）
    await waitFor(() => {
      const audio = container.querySelector(".audio-controls audio");
      expect(audio?.getAttribute("src")).toBe("https://example.test/episodes/ep-1/audio");
    });
  });

  it("AudioControls は catalog success 時に常に在り、再生前は src を持たない", async () => {
    // Given: mount 済みの page（未再生）
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);

    // Then: 再生していなくても AudioControls の <audio> は在る。src は付かない
    await waitFor(() => {
      expect(container.querySelector(".audio-controls audio")).not.toBeNull();
    });
    expect(container.querySelector(".audio-controls audio")?.hasAttribute("src")).toBe(false);
  });

  it("Row の再生ボタンを押すと AudioControls の <audio> に src が付く（Entry は無くてよい）", async () => {
    // Given: mount 済みの page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-item")).not.toBeNull();
    });

    // When: 先頭 Row の再生ボタンを押す
    const playButtons = container.querySelectorAll(".episode-row__play");
    playButtons[0]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

    // Then: 同じ <audio> に src が入る。Entry は出ていない
    await waitFor(() => {
      const audio = container.querySelector(".audio-controls audio");
      expect(audio?.getAttribute("src")).toBe("https://example.test/episodes/ep-1/audio");
    });
    expect(container.querySelector(".episode-manuscript")).toBeNull();
  });

  it("再生中に deselect しても AudioControls は残る（selection と playback の直交）", async () => {
    // Given: ep-1 を select して再生した page
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-item")).not.toBeNull();
    });
    const selectButtons = () => container.querySelectorAll(".episode-row__select");
    const playButtons = () => container.querySelectorAll(".episode-row__play");
    selectButtons()[0]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await waitFor(() => {
      expect(container.querySelector(".episode-manuscript")).not.toBeNull();
    });
    playButtons()[0]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));
    await waitFor(() => {
      expect(container.querySelector(".audio-controls")).not.toBeNull();
    });

    // When: deselect する（select ボタンをもう一度）
    selectButtons()[0]?.dispatchEvent(new MouseEvent("click", { bubbles: true }));

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
      expect(container.querySelector(".episode-item")).not.toBeNull();
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
      expect(container.querySelector(".episode-item")).not.toBeNull();
    });

    // Then: manuscript は出ない（select は no-op）
    expect(container.querySelector("[data-manuscript-opening]")).toBeNull();
  });

  it("hashchange で hash が実在 id へ変わると、その episode の Entry を描画する", async () => {
    // Given: mount 済みの page（hash 空で一覧のみ）
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-item")).not.toBeNull();
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
      expect(container.querySelector(".episode-item")).not.toBeNull();
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

  it("renders the playing episode's row emphasis when the audio adapter reports the playing phase", async () => {
    // Given: catalog success + 先頭 Row の再生ボタン押下済み + audio Fake が capture mode
    const fake = (await audioFake).control;
    fake.install();
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-item")).not.toBeNull();
    });
    (container.querySelectorAll(".episode-row__play")[0] as HTMLButtonElement).dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );

    // When: audio Fake が playing phase を通知する
    fake.emitPhase("playing");

    // Then: 先頭 Row に isPlaying 由来の DOM 強調（data-playing="true"）が出る。他 Row は出ない
    await waitFor(() => {
      const rows = container.querySelectorAll(".episode-row");
      expect(rows[0]?.getAttribute("data-playing")).toBe("true");
      expect(rows[1]?.getAttribute("data-playing")).toBe("false");
    });
  });

  it("keeps the catalog list visible when audio playback fails, surfacing the error only on the mini-player", async () => {
    // Given: catalog success + audio Fake が capture mode で再生失敗を仕込み済み + 再生ボタン押下済み
    const fake = (await audioFake).control;
    fake.install();
    fake.failPlayback();
    const apiClient = createStubApiClient();
    const { container } = renderPage(apiClient);
    await waitFor(() => {
      expect(container.querySelector(".episode-item")).not.toBeNull();
    });

    // When: 先頭 Row の再生ボタンを押す（seekAudioElement({play:true}) が reject する）→ phase:error
    (container.querySelectorAll(".episode-row__play")[0] as HTMLButtonElement).dispatchEvent(
      new MouseEvent("click", { bubbles: true }),
    );
    fake.emitPhase("error");

    // Then: 本体一覧（.episode-list）は残り、全画面 Error（[data-page-error]）は出ない。
    //   audio 失敗は non-blocking で derivePageStatus を汚さない（blocking と non-blocking の分離が
    //   合成で保たれる）。playing 強調も出ない（error phase は isPlaying=false）
    await waitFor(() => {
      const rows = container.querySelectorAll(".episode-row");
      expect(rows[0]?.getAttribute("data-playing")).toBe("false");
    });
    expect(container.querySelector(".episode-list")).not.toBeNull();
    expect(container.querySelectorAll(".episode-item")).toHaveLength(2);
    expect(container.querySelector("[data-page-error]")).toBeNull();
    expect(container.querySelector("[data-page-loading]")).toBeNull();
    // mini-player は再生対象 episode を指したまま残る（失敗は full-screen へ昇格しない）
    expect(container.querySelector(".audio-controls [data-now-playing]")).not.toBeNull();
  });
});
