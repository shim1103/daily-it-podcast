import { fireEvent, render } from "@testing-library/react";
import { createElement } from "react";
import { describe, expect, it, vi } from "vitest";
import { EpisodeSeekBar } from "./episode-seek-bar.tsx";

describe("EpisodeSeekBar", () => {
  it("playing の時、position と duration で range を描画する", () => {
    // Given: playing playback
    const { container } = render(
      createElement(EpisodeSeekBar, {
        playback: { kind: "playing", positionSec: 30, durationSec: 120 },
        onSeek: vi.fn(),
      }),
    );

    // Then: range が値を持つ
    const range = container.querySelector(".episode-seek-bar") as HTMLInputElement;
    expect(range).not.toBeNull();
    expect(range.value).toBe("30");
    expect(range.max).toBe("120");
  });

  it("stopped の時、0 / 0 で range を描画する", () => {
    // Given: stopped playback
    const { container } = render(
      createElement(EpisodeSeekBar, {
        playback: { kind: "stopped" },
        onSeek: vi.fn(),
      }),
    );

    // Then: range が 0 基準
    const range = container.querySelector(".episode-seek-bar") as HTMLInputElement;
    expect(range.value).toBe("0");
    expect(range.max).toBe("0");
  });

  it("range を動かすと onSeek を呼ぶ", () => {
    // Given: playing playback と onSeek spy
    const onSeek = vi.fn();
    const { container } = render(
      createElement(EpisodeSeekBar, {
        playback: { kind: "playing", positionSec: 0, durationSec: 120 },
        onSeek,
      }),
    );

    // When: range を 45 へ動かす
    const range = container.querySelector(".episode-seek-bar") as HTMLInputElement;
    fireEvent.change(range, { target: { value: "45" } });

    // Then: onSeek が呼ばれる
    expect(onSeek).toHaveBeenCalledWith(45);
  });
});
