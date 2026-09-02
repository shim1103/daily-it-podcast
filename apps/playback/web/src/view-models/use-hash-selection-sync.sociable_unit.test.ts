import { renderHook } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { useHashSelectionSync } from "./use-hash-selection-sync.ts";

describe("useHashSelectionSync", () => {
  it("selectedEpisodeId=null で render しても例外を投げない", () => {
    // Given: onHashEpisodeIdChange の spy
    const onHashEpisodeIdChange = vi.fn();

    // When: hook を render する
    renderHook(() => useHashSelectionSync(null, onHashEpisodeIdChange));

    // Then: 例外なく render できる
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
  });

  it("selectedEpisodeId=undefined で render しても例外を投げない", () => {
    // Given: onHashEpisodeIdChange の spy
    const onHashEpisodeIdChange = vi.fn();

    // When: hook を render する
    renderHook(() => useHashSelectionSync(undefined, onHashEpisodeIdChange));

    // Then: 例外なく render できる
    expect(onHashEpisodeIdChange).not.toHaveBeenCalled();
  });
});
