import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AudioLifecyclePhase, AudioStateHandlers } from "../lib/audio-element.ts";
import {
  pauseAudioElement,
  resetAudioElement,
  seekAudioElement,
  subscribeAudioState,
} from "../lib/audio-element.ts";
import { useEpisodePlayback } from "./use-episode-playback.ts";

// why: audio の命令的操作は lib/audio-element.ts が所有する。hook 側 test はその module を
//   丸ごと Test Double へ差し替え、hook が「Adapter をどう呼び、通知された phase / position /
//   duration をどう state へ写し、購読解除をどう cleanup するか」だけを検証する
vi.mock("../lib/audio-element.ts", () => ({
  subscribeAudioState: vi.fn<typeof import("../lib/audio-element.ts").subscribeAudioState>(),
  resetAudioElement: vi.fn<typeof import("../lib/audio-element.ts").resetAudioElement>(),
  pauseAudioElement: vi.fn<typeof import("../lib/audio-element.ts").pauseAudioElement>(),
  seekAudioElement: vi.fn<typeof import("../lib/audio-element.ts").seekAudioElement>(),
}));

const subscribeAudioStateMock = vi.mocked(subscribeAudioState);
const resetAudioElementMock = vi.mocked(resetAudioElement);
const pauseAudioElementMock = vi.mocked(pauseAudioElement);
const seekAudioElementMock = vi.mocked(seekAudioElement);

/**
 * `subscribeAudioState` の Fake。最後に渡された handlers を保持し、任意の phase / position /
 * duration を発火できる。購読・解除の回数も記録する。
 */
function createSubscribeFake(): {
  emitPhase(phase: AudioLifecyclePhase): void;
  emitPosition(positionSec: number): void;
  emitDuration(durationSec: number): void;
  unsubscribeCalls: number;
  subscribeCalls: number;
} {
  const state = { handlers: null as AudioStateHandlers | null };
  const record = { unsubscribeCalls: 0, subscribeCalls: 0 };
  subscribeAudioStateMock.mockImplementation((_el, handlers) => {
    record.subscribeCalls += 1;
    state.handlers = handlers;
    return () => {
      record.unsubscribeCalls += 1;
    };
  });
  return {
    emitPhase(phase: AudioLifecyclePhase): void {
      state.handlers?.onPhaseChange(phase);
    },
    emitPosition(positionSec: number): void {
      state.handlers?.onPositionChange(positionSec);
    },
    emitDuration(durationSec: number): void {
      state.handlers?.onDurationChange(durationSec);
    },
    get unsubscribeCalls(): number {
      return record.unsubscribeCalls;
    },
    get subscribeCalls(): number {
      return record.subscribeCalls;
    },
  };
}

/** hook が ref に張る先の `<audio>`。中身は空でよい（操作は module mock が受ける）。 */
function createAudioRefTarget(): HTMLAudioElement {
  return {} as HTMLAudioElement;
}

describe("useEpisodePlayback", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    subscribeAudioStateMock.mockReturnValue(() => {});
    seekAudioElementMock.mockResolvedValue(undefined);
  });

  it("組み立て直後は playback={kind:'idle'}・audioElementRef 空を返す", () => {
    // Given: なし
    // When: hook を render する
    const { result } = renderHook(() => useEpisodePlayback());

    // Then: idle・ref 空
    expect(result.current.playback).toEqual({ kind: "idle" });
    expect(result.current.audioElementRef.current).toBeNull();
  });

  describe("play", () => {
    it("play('ep-1') は idle から kind:active・phase:loading・positionSec:0・durationSec:null にし、seekAudioElement(audio, 0, {play:true}) を呼ぶ", () => {
      // Given: audio を張った hook
      const { result } = renderHook(() => useEpisodePlayback());
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });

      // When: ep-1 を play する
      act(() => {
        result.current.play("ep-1");
      });

      // Then: active/loading/0/null・再生付き seek を Adapter へ委譲
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "loading" },
        positionSec: 0,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 0, { play: true });
    });

    it("play('ep-1', 90) は positionSec:90 にし、seekAudioElement(audio, 90, {play:true}) を呼ぶ", () => {
      // Given: audio を張った hook
      const { result } = renderHook(() => useEpisodePlayback());
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });

      // When: 90 秒指定で play する
      act(() => {
        result.current.play("ep-1", 90);
      });

      // Then: positionSec=90・(audio, 90, {play:true})
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        positionSec: 90,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 90, { play: true });
    });

    it("同じ episode の再生中に play('ep-1')（positionSec 省略）は prev.positionSec から再開する", () => {
      // Given: ep-1 を再生し timeupdate で 55 秒まで進んだ hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        subscribe.emitPosition(55);
      });
      seekAudioElementMock.mockClear();

      // When: positionSec 省略で同じ ep-1 を play する
      act(() => {
        result.current.play("ep-1");
      });

      // Then: prev.positionSec=55 から再開
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        positionSec: 55,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 55, { play: true });
    });

    it("play('ep-2') を ep-1 active から呼ぶと resetAudioElement を呼び positionSec:0・durationSec:null にする", () => {
      // Given: ep-1 を再生中で duration も入った hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        subscribe.emitDuration(600);
        subscribe.emitPosition(30);
      });
      resetAudioElementMock.mockClear();
      seekAudioElementMock.mockClear();

      // When: 別 episode ep-2 を play する
      act(() => {
        result.current.play("ep-2");
      });

      // Then: 直前 audio を reset・ep-2/loading/0/null（duration は引き継がない）
      expect(resetAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio);
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-2",
        phase: { phase: "loading" },
        positionSec: 0,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 0, { play: true });
    });

    it("同じ episode へ再度 play しても resetAudioElement は呼ばない", () => {
      // Given: ep-1 を再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });
      resetAudioElementMock.mockClear();

      // When: 同じ ep-1 を play する
      act(() => {
        result.current.play("ep-1");
      });

      // Then: reset は起きない
      expect(resetAudioElementMock).not.toHaveBeenCalled();
    });

    it("audioElementRef 未設定でも play は例外を投げず、state だけ active/loading/引数の positionSec へ倒す", () => {
      // Given: ref 未設定の hook
      const { result } = renderHook(() => useEpisodePlayback());

      // When: 45 秒指定で play する
      act(() => {
        result.current.play("ep-1", 45);
      });

      // Then: 例外なし・state だけ倒る・audio 操作は呼ばれない
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "loading" },
        positionSec: 45,
        durationSec: null,
      });
      expect(seekAudioElementMock).not.toHaveBeenCalled();
      expect(resetAudioElementMock).not.toHaveBeenCalled();
    });
  });

  describe("seek", () => {
    it("seek('ep-1', 120) を idle から呼ぶと active/phase:paused/positionSec:120 にし、seekAudioElement(audio, 120, {play:false}) を呼ぶ", () => {
      // Given: audio を張った idle の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });

      // When: ep-1 を 120 秒へ seek する
      act(() => {
        result.current.seek("ep-1", 120);
      });

      // Then: active/paused/120/null・play なし seek
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "paused" },
        positionSec: 120,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: false });
    });

    it("seek('ep-1', 120) を同じ episode の再生中（phase:playing）から呼ぶと phase:playing 継続で seekAudioElement(audio, 120, {play:true}) を呼ぶ", () => {
      // Given: ep-1 を再生し playing になっている hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        subscribe.emitPhase("playing");
      });
      seekAudioElementMock.mockClear();

      // When: 120 秒へ seek する
      act(() => {
        result.current.seek("ep-1", 120);
      });

      // Then: playing 継続・(audio, 120, {play:true})
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "playing" },
        positionSec: 120,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: true });
    });

    it("seek('ep-1', 120) を同じ episode の停止中（phase:paused）から呼ぶと phase:paused で seekAudioElement(audio, 120, {play:false}) を呼ぶ", () => {
      // Given: ep-1 を再生後 paused になっている hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        subscribe.emitPhase("paused");
      });
      seekAudioElementMock.mockClear();

      // When: 120 秒へ seek する
      act(() => {
        result.current.seek("ep-1", 120);
      });

      // Then: paused 継続・(audio, 120, {play:false})
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "paused" },
        positionSec: 120,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: false });
    });

    it("seek('ep-2', 30) を ep-1 再生中から呼ぶと resetAudioElement を呼び ep-2/paused/30/null にし、seekAudioElement(audio, 30, {play:false}) を呼ぶ", () => {
      // Given: ep-1 を再生し playing になっている hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        subscribe.emitPhase("playing");
        subscribe.emitDuration(600);
      });
      resetAudioElementMock.mockClear();
      seekAudioElementMock.mockClear();

      // When: 別 episode ep-2 へ seek する
      act(() => {
        result.current.seek("ep-2", 30);
      });

      // Then: reset・ep-2/paused/30/null（違う episode なので play 継続せず duration も引き継がない）
      expect(resetAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio);
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-2",
        phase: { phase: "paused" },
        positionSec: 30,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 30, { play: false });
    });

    it("audioElementRef 未設定でも seek は例外を投げず、state だけ引数の positionSec へ倒す", () => {
      // Given: ref 未設定の hook
      const { result } = renderHook(() => useEpisodePlayback());

      // When: seek する
      act(() => {
        result.current.seek("ep-1", 42);
      });

      // Then: 例外なし・state だけ倒る・audio 操作は呼ばれない
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "paused" },
        positionSec: 42,
        durationSec: null,
      });
      expect(seekAudioElementMock).not.toHaveBeenCalled();
    });
  });

  describe("Adapter からの通知を state へ写す", () => {
    it("phase=playing 通知で phase が playing になる", () => {
      // Given: ep-1 を play 済みの hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });

      // When: playing phase を通知する
      act(() => {
        subscribe.emitPhase("playing");
      });

      // Then: playing
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "playing" },
      });
    });

    it("phase=error 通知で phase が error/audio-load-failed になる", () => {
      // Given: 再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });

      // When: error phase を通知する
      act(() => {
        subscribe.emitPhase("error");
      });

      // Then: error/audio-load-failed
      expect(result.current.playback).toMatchObject({
        kind: "active",
        phase: { phase: "error", reason: "audio-load-failed" },
      });
    });

    it("timeupdate 通知で active 枝の positionSec が差分更新される", () => {
      // Given: ep-1 を再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });

      // When: 位置通知が来る
      act(() => {
        subscribe.emitPosition(37.2);
      });

      // Then: positionSec だけ更新・他は維持
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "loading" },
        positionSec: 37.2,
        durationSec: null,
      });
    });

    it("kind=idle 中の timeupdate 通知は無視される", () => {
      // Given: ep-1 を play 後 stop した hook（購読 Fake は handlers を保持したまま）
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        result.current.stop();
      });

      // When: idle 中に位置通知が来る
      act(() => {
        subscribe.emitPosition(99);
      });

      // Then: idle のまま
      expect(result.current.playback).toEqual({ kind: "idle" });
    });

    it("loadedmetadata 通知で active 枝の durationSec が差分更新される", () => {
      // Given: ep-1 を再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });

      // When: 長さ通知が来る
      act(() => {
        subscribe.emitDuration(1234);
      });

      // Then: durationSec だけ更新・他は維持
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "loading" },
        positionSec: 0,
        durationSec: 1234,
      });
    });

    it("kind=idle 中の loadedmetadata 通知は無視される", () => {
      // Given: ep-1 を play 後 stop した hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        result.current.stop();
      });

      // When: idle 中に長さ通知が来る
      act(() => {
        subscribe.emitDuration(500);
      });

      // Then: idle のまま
      expect(result.current.playback).toEqual({ kind: "idle" });
    });

    it("stop() 後に phase 通知が来ても playback は idle のまま（不正遷移拒否）", () => {
      // Given: ep-1 を play 後 stop 済みの hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });
      act(() => {
        result.current.stop();
      });

      // When: idle の状態で phase を通知する
      act(() => {
        subscribe.emitPhase("playing");
      });

      // Then: 不正遷移は拒否され idle のまま
      expect(result.current.playback).toEqual({ kind: "idle" });
    });

    it("seekAudioElement の rejection で phase が error/audio-load-failed になる", async () => {
      // Given: seek が reject する Adapter を張った hook
      const { result } = renderHook(() => useEpisodePlayback());
      seekAudioElementMock.mockRejectedValue(new Error("再生失敗"));
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });

      // When: ep-1 を play する
      await act(async () => {
        result.current.play("ep-1");
        await Promise.resolve();
      });

      // Then: error/audio-load-failed
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "error", reason: "audio-load-failed" },
      });
    });
  });

  describe("stop / 購読 / 参照安定", () => {
    it("stop() すると直前 audio を pauseAudioElement で止め（load は呼ばない）、playback を idle にする", () => {
      // Given: ep-1 を再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1");
      });
      pauseAudioElementMock.mockClear();
      resetAudioElementMock.mockClear();

      // When: stop する
      act(() => {
        result.current.stop();
      });

      // Then: pause + idle。別 episode 切替専用の resetAudioElement は呼ばない
      expect(pauseAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio);
      expect(resetAudioElementMock).not.toHaveBeenCalled();
      expect(result.current.playback).toEqual({ kind: "idle" });
    });

    it("audioElementRef 未設定でも stop() は例外を投げず playback を idle へ戻す", () => {
      // Given: play 済みだが ref 未設定の hook
      const { result } = renderHook(() => useEpisodePlayback());
      act(() => {
        result.current.play("ep-1");
      });

      // When: stop する
      act(() => {
        result.current.stop();
      });

      // Then: 例外なし・idle・Adapter は呼ばれない
      expect(pauseAudioElementMock).not.toHaveBeenCalled();
      expect(result.current.playback).toEqual({ kind: "idle" });
    });

    it("別 episode へ play すると直前の購読を解除してから新しく購読する", () => {
      // Given: ep-1 を再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });

      // When: ep-2 を play する
      act(() => {
        result.current.play("ep-2");
      });

      // Then: 購読は 2 回、解除は 1 回以上
      expect(subscribe.subscribeCalls).toBe(2);
      expect(subscribe.unsubscribeCalls).toBeGreaterThanOrEqual(1);
    });

    it("unmount すると購読解除関数が呼ばれ、以後の phase 通知で state が変わらない", () => {
      // Given: 再生中の hook
      const { result, unmount } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1");
      });
      const playbackBefore = result.current.playback;

      // When: unmount してから phase 通知する
      unmount();
      act(() => {
        subscribe.emitPhase("playing");
      });

      // Then: unsubscribe が呼ばれ、state は変わらない
      expect(subscribe.unsubscribeCalls).toBeGreaterThanOrEqual(1);
      expect(result.current.playback).toBe(playbackBefore);
    });

    it("play / seek / stop は再 render をまたいで同一参照を保つ", () => {
      // Given: hook
      const { result, rerender } = renderHook(() => useEpisodePlayback());
      const firstPlay = result.current.play;
      const firstSeek = result.current.seek;
      const firstStop = result.current.stop;

      // When: state を変えて再 render する
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
        result.current.play("ep-1");
      });
      rerender();

      // Then: 参照不変
      expect(result.current.play).toBe(firstPlay);
      expect(result.current.seek).toBe(firstSeek);
      expect(result.current.stop).toBe(firstStop);
    });
  });
});
