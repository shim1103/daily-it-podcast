import { act, renderHook } from "@testing-library/react";
import type { Ref } from "react";
import { describe, expect, it, vi } from "vitest";
import { useEpisodePlayback } from "./use-episode-playback.ts";

const baseUrl = "https://example.test";

function connectAudio(audioElementRef: Ref<HTMLAudioElement | null>, audio: HTMLAudioElement): void {
  if (typeof audioElementRef === "function") {
    audioElementRef(audio);
    return;
  }
  if (audioElementRef !== null) {
    audioElementRef.current = audio;
  }
}

describe("useEpisodePlayback", () => {
  it("audioElementRef と idle playback を返す", () => {
    // Given: hook を render する
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));

    // Then: callback ref と idle playback が返る
    expect(typeof result.current.audioElementRef).toBe("function");
    expect(result.current.playback).toEqual({ kind: "idle" });
    expect(result.current.resolvedSrc).toBeUndefined();
  });

  it("play が playback を playing にし resolvedSrc を組み立てる", () => {
    // Given: hook を render する
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));

    // When: play を呼ぶ
    act(() => {
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // Then: playback が playing になり src が解決される
    expect(result.current.playback).toEqual({
      kind: "playing",
      episodeId: "ep-1",
      positionSec: 0,
      durationSec: 0,
    });
    expect(result.current.resolvedSrc).toBe("https://example.test/episodes/ep-1/audio");
  });

  it("別 episode を play すると再生対象が乗り換わる", () => {
    // Given: ep-1 を再生中の hook
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));
    act(() => {
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // When: ep-2 を play する
    act(() => {
      result.current.play("ep-2", "/episodes/ep-2/audio");
    });

    // Then: playback が ep-2 になる
    expect(result.current.playback).toEqual({
      kind: "playing",
      episodeId: "ep-2",
      positionSec: 0,
      durationSec: 0,
    });
    expect(result.current.resolvedSrc).toBe("https://example.test/episodes/ep-2/audio");
  });

  it("play 後に audio ref が無い時、再生 effect は no-op する", () => {
    // Given: audio 未接続の hook
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));

    // When: play を呼ぶ
    act(() => {
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // Then: playback は playing、resolvedSrc は解決される
    expect(result.current.playback.kind).toBe("playing");
    expect(result.current.resolvedSrc).toBe("https://example.test/episodes/ep-1/audio");
  });

  it("play 時に audio の src が既に一致していれば src を差し替えない", () => {
    // Given: audio 要素を ref に差し込んだ hook
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));
    const audio = document.createElement("audio");
    const nextSrc = "https://example.test/episodes/ep-1/audio";
    audio.setAttribute("src", nextSrc);
    const play = vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      connectAudio(result.current.audioElementRef, audio);
    });

    // When: play を呼ぶ
    act(() => {
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // Then: src 属性は変わらず play だけ呼ばれる
    expect(audio.getAttribute("src")).toBe(nextSrc);
    expect(play).toHaveBeenCalled();
  });

  it("timeupdate で positionSec / durationSec を更新する", () => {
    // Given: 再生中の hook と audio
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));
    const audio = document.createElement("audio");
    Object.defineProperty(audio, "currentTime", { configurable: true, value: 12 });
    Object.defineProperty(audio, "duration", { configurable: true, value: 90 });
    vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      connectAudio(result.current.audioElementRef, audio);
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // When: timeupdate を発火する
    act(() => {
      audio.dispatchEvent(new Event("timeupdate"));
    });

    // Then: playback の位置と尺が更新される
    expect(result.current.playback).toEqual({
      kind: "playing",
      episodeId: "ep-1",
      positionSec: 12,
      durationSec: 90,
    });
  });

  it("timeupdate で duration が非有限の時は直前の durationSec を維持する", () => {
    // Given: 再生中の hook と duration が NaN の audio
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));
    const audio = document.createElement("audio");
    Object.defineProperty(audio, "currentTime", { configurable: true, value: 5 });
    Object.defineProperty(audio, "duration", { configurable: true, value: Number.NaN });
    vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      connectAudio(result.current.audioElementRef, audio);
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // When: timeupdate を発火する
    act(() => {
      audio.dispatchEvent(new Event("timeupdate"));
    });

    // Then: position は更新され duration は 0 のまま
    expect(result.current.playback).toEqual({
      kind: "playing",
      episodeId: "ep-1",
      positionSec: 5,
      durationSec: 0,
    });
  });

  it("playing 中に seek すると positionSec を更新する", () => {
    // Given: 再生中の hook
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));
    const audio = document.createElement("audio");
    vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      connectAudio(result.current.audioElementRef, audio);
      result.current.play("ep-1", "/episodes/ep-1/audio");
    });

    // When: seek を呼ぶ
    act(() => {
      result.current.seek(42);
    });

    // Then: playback の position が更新される
    expect(result.current.playback).toEqual({
      kind: "playing",
      episodeId: "ep-1",
      positionSec: 42,
      durationSec: 0,
    });
  });

  it("idle の時、seek は playback state を変えない", () => {
    // Given: idle の hook と audio
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));
    const audio = document.createElement("audio");
    vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      connectAudio(result.current.audioElementRef, audio);
    });

    // When: seek を呼ぶ
    act(() => {
      result.current.seek(42);
    });

    // Then: audio は動くが playback は idle のまま
    expect(audio.currentTime).toBe(42);
    expect(result.current.playback).toEqual({ kind: "idle" });
  });

  it("audio 要素が無い時、seek は pending として保持し mount 後に適用する", () => {
    // Given: ref が null の hook
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));

    // When: 再生開始前に seek を呼ぶ
    act(() => {
      result.current.play("ep-1", "/episodes/ep-1/audio");
      result.current.seek(30);
    });

    // Then: playback の位置は更新され、audio 未接続のまま
    expect(result.current.playback).toEqual({
      kind: "playing",
      episodeId: "ep-1",
      positionSec: 30,
      durationSec: 0,
    });

    const audio = document.createElement("audio");
    const play = vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      connectAudio(result.current.audioElementRef, audio);
    });

    // Then: pending seek が audio に適用される
    expect(audio.currentTime).toBe(30);
    expect(play).toHaveBeenCalled();
  });

  it("再生前の topic seek は play 後の mount で適用される", () => {
    // Given: idle の hook
    const { result } = renderHook(() => useEpisodePlayback(baseUrl));

    // When: play 前に seek を呼ぶ
    act(() => {
      result.current.seek(42);
    });

    // Then: idle のまま
    expect(result.current.playback).toEqual({ kind: "idle" });

    // When: play して audio を接続する
    const audio = document.createElement("audio");
    vi.spyOn(audio, "play").mockResolvedValue(undefined);
    act(() => {
      result.current.play("ep-1", "/episodes/ep-1/audio");
      connectAudio(result.current.audioElementRef, audio);
    });

    // Then: mount 後に pending seek が適用される
    expect(audio.currentTime).toBe(42);
  });
});
