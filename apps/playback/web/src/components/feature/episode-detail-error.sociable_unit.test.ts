import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { EpisodeDetailError } from "./episode-detail-error.tsx";

describe("EpisodeDetailError", () => {
  it("error 用 data 属性付き episode-detail を描画する", () => {
    // Given / When: JSX として render する
    const { container } = render(createElement(EpisodeDetailError));

    // Then: error 表示の容器がある
    expect(container.querySelector(".episode-detail[data-episode-detail-error]")).not.toBeNull();
  });
});
