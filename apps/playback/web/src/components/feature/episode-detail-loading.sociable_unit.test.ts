import { render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it } from "vitest";
import { EpisodeDetailLoading } from "./episode-detail-loading.tsx";

describe("EpisodeDetailLoading", () => {
  it("loading 用 data 属性付き episode-detail を描画する", () => {
    // Given / When: JSX として render する
    const { container } = render(createElement(EpisodeDetailLoading));

    // Then: loading 表示の容器がある
    expect(container.querySelector(".episode-detail[data-episode-detail-loading]")).not.toBeNull();
  });
});
