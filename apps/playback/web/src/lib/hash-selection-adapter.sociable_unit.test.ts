import { afterEach, describe, expect, it, vi } from "vitest";
import { createHashSelectionAdapter } from "./hash-selection-adapter.ts";

describe("createHashSelectionAdapter", () => {
  afterEach(() => {
    window.location.hash = "";
    vi.restoreAllMocks();
  });

  it("getEpisodeId は先頭の # を除いた hash 値を返す", () => {
    // Given: `#ep-1` を持つ location と adapter
    window.location.hash = "#ep-1";
    const adapter = createHashSelectionAdapter();

    // When: episodeId を読む
    const got = adapter.getEpisodeId();

    // Then: `#` を除いた値
    expect(got).toBe("ep-1");
  });

  it("getEpisodeId は hash が無い時、null を返す", () => {
    // Given: hash が無い location と adapter
    window.location.hash = "";
    const adapter = createHashSelectionAdapter();

    // When: episodeId を読む
    const got = adapter.getEpisodeId();

    // Then: null
    expect(got).toBeNull();
  });

  it("setEpisodeId は値を location.hash へ設定する", () => {
    // Given: adapter
    const adapter = createHashSelectionAdapter();

    // When: episodeId を設定する
    adapter.setEpisodeId("ep-2");

    // Then: `#ep-2` になる
    expect(window.location.hash).toBe("#ep-2");
  });

  it("setEpisodeId(null) は hash を消す", () => {
    // Given: hash が設定済みの location と adapter
    window.location.hash = "#ep-3";
    const adapter = createHashSelectionAdapter();

    // When: null で episodeId を設定する
    adapter.setEpisodeId(null);

    // Then: hash が空になる
    expect(adapter.getEpisodeId()).toBeNull();
  });

  it("subscribe は hashchange event で listener を呼ぶ", () => {
    // Given: listener を登録した adapter
    const listener = vi.fn();
    const adapter = createHashSelectionAdapter();
    adapter.subscribe(listener);

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-4";
    window.dispatchEvent(new Event("hashchange"));

    // Then: listener が呼ばれる
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("subscribe の戻り値を呼ぶと listener の購読を解除する", () => {
    // Given: 登録解除した listener
    const listener = vi.fn();
    const adapter = createHashSelectionAdapter();
    const unsubscribe = adapter.subscribe(listener);
    unsubscribe();

    // When: hashchange を発火する
    window.location.hash = "#ep-5";
    window.dispatchEvent(new Event("hashchange"));

    // Then: listener は呼ばれない
    expect(listener).not.toHaveBeenCalled();
  });
});
