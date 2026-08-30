import { afterEach, describe, expect, it, vi } from "vitest";
import { getLocationHash, onLocationHashChange, setLocationHash } from "./location-hash.ts";

describe("location-hash", () => {
  afterEach(() => {
    window.location.hash = "";
    vi.restoreAllMocks();
  });

  it("getLocationHash は先頭の # を除いた hash 値を返す", () => {
    // Given: `#ep-1` を持つ location
    window.location.hash = "#ep-1";

    // When: hash を読む
    const got = getLocationHash();

    // Then: `#` を除いた値
    expect(got).toBe("ep-1");
  });

  it("getLocationHash は hash が無い時、空文字を返す", () => {
    // Given: hash が無い location
    window.location.hash = "";

    // When: hash を読む
    const got = getLocationHash();

    // Then: 空文字
    expect(got).toBe("");
  });

  it("setLocationHash は値を location.hash へ設定する", () => {
    // Given: episodeId
    // When: hash を設定する
    setLocationHash("ep-2");

    // Then: `#ep-2` になる
    expect(window.location.hash).toBe("#ep-2");
  });

  it("setLocationHash は空文字を渡すと hash を消す", () => {
    // Given: hash が設定済みの location
    window.location.hash = "#ep-3";

    // When: 空文字で hash を設定する
    setLocationHash("");

    // Then: hash が空になる
    expect(getLocationHash()).toBe("");
  });

  it("onLocationHashChange は hashchange event で listener を呼ぶ", () => {
    // Given: listener を登録した状態
    const listener = vi.fn();
    onLocationHashChange(listener);

    // When: hash を変更し hashchange を発火する
    window.location.hash = "#ep-4";
    window.dispatchEvent(new Event("hashchange"));

    // Then: listener が呼ばれる
    expect(listener).toHaveBeenCalledTimes(1);
  });

  it("onLocationHashChange の戻り値を呼ぶと listener の購読を解除する", () => {
    // Given: 登録解除した listener
    const listener = vi.fn();
    const unsubscribe = onLocationHashChange(listener);
    unsubscribe();

    // When: hashchange を発火する
    window.location.hash = "#ep-5";
    window.dispatchEvent(new Event("hashchange"));

    // Then: listener は呼ばれない
    expect(listener).not.toHaveBeenCalled();
  });
});
