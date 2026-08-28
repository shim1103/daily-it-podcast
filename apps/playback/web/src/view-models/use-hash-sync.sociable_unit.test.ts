import { act, renderHook } from "@testing-library/react";
import { useRef } from "react";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { useHashSync } from "./use-hash-sync.ts";

// why: useSyncExternalStore の購読は hashchange で再 render → 副作用を走らせる。
//   render 外での hash 変更と hashchange 発火を act で括り、React の更新を flush する
function dispatchHashChange(nextHash: string): void {
  act(() => {
    window.location.hash = nextHash;
    window.dispatchEvent(new Event("hashchange"));
  });
}

describe("useHashSync", () => {
  beforeEach(() => {
    window.location.hash = "";
  });

  afterEach(() => {
    window.location.hash = "";
    vi.restoreAllMocks();
  });

  it("selectedId が非 null の時、location.hash をその値へ同期する", () => {
    // Given: hash 空、onHashSelect の spy
    const onHashSelect = vi.fn();

    // When: selectedId=ep-1 で hook を render する
    renderHook(({ id }) => useHashSync(id, onHashSelect), { initialProps: { id: "ep-1" } });

    // Then: hash が #ep-1 になる
    expect(window.location.hash).toBe("#ep-1");
  });

  it("selectedId が非 null → 別の値へ変わると、location.hash を追従させる", () => {
    // Given: ep-1 を同期済みの hook
    const onHashSelect = vi.fn();
    const { rerender } = renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: "ep-1" as string | null },
    });

    // When: selectedId を ep-2 へ変える
    rerender({ id: "ep-2" });

    // Then: hash が #ep-2 になる
    expect(window.location.hash).toBe("#ep-2");
  });

  it("selectedId が非 null → null へ変わると、location.hash をクリアする", () => {
    // Given: ep-1 を同期済みの hook
    const onHashSelect = vi.fn();
    const { rerender } = renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: "ep-1" as string | null },
    });

    // When: selectedId を null へ変える
    rerender({ id: null });

    // Then: hash がクリアされる
    expect(window.location.hash).toBe("");
  });

  it("selectedId が undefined の間、location.hash を書き換えない（未初期化）", () => {
    // Given: hash に #ep-1 が既にある
    window.location.hash = "#ep-1";
    const onHashSelect = vi.fn();

    // When: selectedId=undefined で hook を render する
    renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: undefined as string | null | undefined },
    });

    // Then: hash は元のまま
    expect(window.location.hash).toBe("#ep-1");
  });

  it("hashchange で hash が非空へ変わると、その文字列で onHashSelect を呼ぶ", () => {
    // Given: selectedId=null で同期済みの hook
    const onHashSelect = vi.fn();
    renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: null as string | null },
    });

    // When: hash を #ep-1 へ変え hashchange を発火する
    dispatchHashChange("#ep-1");

    // Then: onHashSelect("ep-1") が呼ばれる
    expect(onHashSelect).toHaveBeenCalledWith("ep-1");
  });

  it("hashchange で hash が空へ変わると、null で onHashSelect を呼ぶ", () => {
    // Given: ep-1 を同期済みの hook
    const onHashSelect = vi.fn();
    const { rerender } = renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: "ep-1" as string | null },
    });
    rerender({ id: "ep-1" });

    // When: hash を空へ変え hashchange を発火する
    dispatchHashChange("");

    // Then: onHashSelect(null) が呼ばれる
    expect(onHashSelect).toHaveBeenCalledWith(null);
  });

  it("hashchange 由来で selectedId が同じ値へ変わり hash へ書き戻しても、無限ループしない", () => {
    // Given: hashchange を受けたら selectedId をその値へ更新する caller を模したセットアップ
    let currentId: string | null = null;
    const onHashSelect = vi.fn((id: string | null) => {
      currentId = id;
    });
    const { rerender } = renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: currentId },
    });

    // When: hash が #ep-1 へ変わり hashchange → onHashSelect → caller が selectedId=ep-1 で再 render
    dispatchHashChange("#ep-1");
    rerender({ id: currentId });
    // さらに hashchange が誘発されていないか確認するため、もう一度発火機会を与える
    act(() => {
      window.dispatchEvent(new Event("hashchange"));
    });

    // Then: onHashSelect は 1 回だけ（書き戻しによる再 hashchange の連鎖が無い）。hash も #ep-1 のまま
    expect(onHashSelect).toHaveBeenCalledTimes(1);
    expect(window.location.hash).toBe("#ep-1");
  });

  it("外部 hash が不変のまま再 render しても、getSnapshot の Object.is 安定で余分な再 render が起きない", () => {
    // why: getSnapshot が毎回新参照（オブジェクト等）を返すと useSyncExternalStore の Object.is
    //   比較が破れ "Maximum update depth exceeded" の無限 render になる。文字列戻り値でそれが
    //   起きないことを render 回数で固定する
    // Given: hash に既存値、render 回数を数える wrapper 越しに hook を張る
    window.location.hash = "#ep-1";
    const onHashSelect = vi.fn();
    const rerenderCount = 3;
    let renderCount = 0;
    const useProbe = () => {
      const seen = useRef(0);
      seen.current += 1;
      renderCount = seen.current;
      useHashSync("ep-1", onHashSelect);
    };
    const { rerender } = renderHook(() => useProbe());

    // When: 外部 hash を変えずに親を rerenderCount 回 rerender する
    for (let i = 0; i < rerenderCount; i += 1) {
      rerender();
    }

    // Then: render 回数は初回 1 + rerenderCount 回ぶんだけ。onHashSelect も発火しない
    expect(renderCount).toBe(1 + rerenderCount);
    expect(onHashSelect).not.toHaveBeenCalled();
  });

  it("unmount すると hashchange listener が外れ、以後 onHashSelect を呼ばない", () => {
    // Given: 同期済みの hook
    const onHashSelect = vi.fn();
    const { unmount } = renderHook(({ id }) => useHashSync(id, onHashSelect), {
      initialProps: { id: null as string | null },
    });

    // When: unmount してから hashchange を発火する
    unmount();
    dispatchHashChange("#ep-1");

    // Then: onHashSelect は呼ばれない
    expect(onHashSelect).not.toHaveBeenCalled();
  });
});
