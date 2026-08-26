import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { LabeledText } from "./labeled-text.tsx";

function renderLabeledText(props: {
  tag: keyof HTMLElementTagNameMap;
  datasetKey: string;
  text: string;
}): HTMLElement {
  const { container } = render(createElement(LabeledText, props));
  const element = container.firstElementChild;
  if (!(element instanceof HTMLElement)) {
    throw new Error("根要素が HTMLElement ではない");
  }
  return element;
}

describe("LabeledText", () => {
  it("指定した tag に dataset 属性と textContent をそのまま設定する", () => {
    // Given: tag・dataset属性名・text
    // When: JSX として render する
    const element = renderLabeledText({
      tag: "h1",
      datasetKey: "episodeTitle",
      text: "題1",
    });

    // Then: 指定した tag・dataset・textContent を持つ
    expect(element.tagName).toBe("H1");
    expect(element.dataset.episodeTitle).toBe("");
    expect(element.textContent).toBe("題1");
  });

  it("text が空文字でもそのまま載せる", () => {
    // Given: 空の text
    // When: JSX として render する
    const element = renderLabeledText({
      tag: "span",
      datasetKey: "episodeTitle",
      text: "",
    });

    // Then: textContent は空のまま（加工しない）
    expect(element.textContent).toBe("");
  });

  it("複数 hump の datasetKey でも browser dataset と同値の data-* を載せる", () => {
    // Given: 複数 camelCase hump の datasetKey
    // When: JSX として render する
    const element = renderLabeledText({
      tag: "span",
      datasetKey: "episodeDurationSec",
      text: "12",
    });

    // Then: dataset 経由で読める（data-episode-duration-sec と同等）
    expect(element.dataset.episodeDurationSec).toBe("");
    expect(element.textContent).toBe("12");
  });
});
