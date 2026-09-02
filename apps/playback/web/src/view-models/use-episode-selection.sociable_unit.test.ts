import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useEpisodeSelection } from "./use-episode-selection.ts";

describe("useEpisodeSelection", () => {
  it("組み立て直後は selectedEpisodeId=null を返す", () => {
    // Given: なし
    // When: hook を render する
    const { result } = renderHook(() => useEpisodeSelection());

    // Then: 選択なし
    expect(result.current.selectedEpisodeId).toBeNull();
  });

  it("select / deselect / toggle を呼んでも例外を投げない", () => {
    // Given: hook
    const { result } = renderHook(() => useEpisodeSelection());

    // When: 各操作を実行する
    act(() => {
      result.current.select("ep-1");
      result.current.deselect();
      result.current.toggle("ep-1");
    });

    // Then: 選択は変わらない（stub）
    expect(result.current.selectedEpisodeId).toBeNull();
  });
});
