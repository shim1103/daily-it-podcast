import { act, renderHook } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useEpisodePlayback } from "./use-episode-playback.ts";

describe("useEpisodePlayback", () => {
  it("組み立て直後は idle・playedEpisodeId=null を返す", () => {
    // Given: なし
    // When: hook を render する
    const { result } = renderHook(() => useEpisodePlayback());

    // Then: idle・再生なし
    expect(result.current.playbackPhase).toBe("idle");
    expect(result.current.playedEpisodeId).toBeNull();
    expect(result.current.audioElementRef.current).toBeNull();
  });

  it("play / stop を呼んでも例外を投げない", () => {
    // Given: hook
    const { result } = renderHook(() => useEpisodePlayback());

    // When: play / stop を実行する
    act(() => {
      result.current.play("ep-1");
      result.current.stop();
    });

    // Then: 状態は変わらない（stub）
    expect(result.current.playbackPhase).toBe("idle");
    expect(result.current.playedEpisodeId).toBeNull();
  });
});
