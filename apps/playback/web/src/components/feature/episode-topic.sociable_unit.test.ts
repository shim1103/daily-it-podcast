import { fireEvent, render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeTopic } from "./episode-topic.tsx";

type Topic = EpisodeData["body"]["topics"][number];

const topic: Topic = { title: "小題", preface: "前置き", detail: "詳細", startSec: 90 };

describe("EpisodeTopic", () => {
  it("root に episode-topic class を付ける", () => {
    // Given: topics[] の1要素
    // When: JSX として render する
    const { container } = render(
      createElement(EpisodeTopic, { topic, topicIndex: 0, onSeek: vi.fn() }),
    );

    // Then: topic 容器の class が root にある（見た目は CSS 側の責務）
    expect(container.firstElementChild?.className).toBe("episode-topic");
  });

  it("mm:ss と通し番号付き title を見出しに描画する", () => {
    // Given: topics[] の1要素（startSec 90）
    const { container } = render(
      createElement(EpisodeTopic, { topic, topicIndex: 1, onSeek: vi.fn() }),
    );

    // Then: seek ボタンと title が分かれて描画される
    expect(container.querySelector("[data-topic-start-sec]")?.textContent).toBe("01:30");
    expect(container.querySelector("[data-topic-title]")?.textContent).toBe("2. 小題");
    expect(container.querySelector("[data-topic-preface]")?.textContent).toBe("前置き");
    expect(container.querySelector("[data-topic-detail]")?.textContent).toBe("詳細");
  });

  it("seek ボタンを押すと onSeek が startSec 付きで呼ばれる", () => {
    // Given: onSeek の spy
    const onSeek = vi.fn();
    const { container } = render(createElement(EpisodeTopic, { topic, topicIndex: 0, onSeek }));

    // When: mm:ss ボタンをクリックする
    const button = container.querySelector("[data-topic-start-sec]");
    fireEvent.click(button as Element);

    // Then: startSec が渡される
    expect(onSeek).toHaveBeenCalledWith(90);
  });
});
