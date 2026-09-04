import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { AudioLifecyclePhase, AudioStateHandlers } from "../lib/audio-element.ts";
import {
  pauseAudioElement,
  seekAudioElement,
  setAudioSource,
  subscribeAudioState,
} from "../lib/audio-element.ts";
import { useEpisodePlayback } from "./use-episode-playback.ts";

// why: audio の命令的操作は lib/audio-element.ts が所有する。hook 側 test はその module を
//   丸ごと Test Double へ差し替え、hook が「Adapter をどう呼び、通知された phase / position /
//   duration をどう state へ写し、購読解除をどう cleanup するか」だけを検証する
vi.mock("../lib/audio-element.ts", () => ({
  subscribeAudioState: vi.fn<typeof import("../lib/audio-element.ts").subscribeAudioState>(),
  setAudioSource: vi.fn<typeof import("../lib/audio-element.ts").setAudioSource>(),
  pauseAudioElement: vi.fn<typeof import("../lib/audio-element.ts").pauseAudioElement>(),
  seekAudioElement: vi.fn<typeof import("../lib/audio-element.ts").seekAudioElement>(),
}));

const subscribeAudioStateMock = vi.mocked(subscribeAudioState);
const setAudioSourceMock = vi.mocked(setAudioSource);
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

/**
 * hook が ref に張る先の `<audio>`。命令的操作は module mock が受けるが、`moveTo` は
 * `el.src !== audioRef` を見て `setAudioSource` の要否を判断するため、`src` の入れ物だけ持つ。
 */
function createAudioRefTarget(): HTMLAudioElement {
  return { src: "" } as HTMLAudioElement;
}

describe("useEpisodePlayback", () => {
  beforeEach(() => {
    vi.clearAllMocks();
    subscribeAudioStateMock.mockReturnValue(() => {});
    seekAudioElementMock.mockResolvedValue(undefined);
    // why: 本物の setAudioSource は el.src を書き換える。moveTo の「src が既に一致なら張り直さない」
    //   分岐を test でも再現するため、mock でも src を反映する
    setAudioSourceMock.mockImplementation((el, src) => {
      el.src = src;
    });
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // Then: active/loading/0/null・再生付き seek を Adapter へ委譲
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
        positionSec: 0,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 0, { play: true });
    });

    it("play は渡された audioRef をそのまま active 枝へ載せる（catalog から引き当てない）", () => {
      // Given: audio を張った hook
      const { result } = renderHook(() => useEpisodePlayback());
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });

      // When: 任意の audioRef を渡して play する
      act(() => {
        result.current.play("ep-9", "/custom/path/ep-9.mp3");
      });

      // Then: active 枝の audioRef は引数の値
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-9",
        audioRef: "/custom/path/ep-9.mp3",
      });
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
        result.current.play("ep-1", "/episodes/ep-1/audio", 90);
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitPosition(55);
      });
      seekAudioElementMock.mockClear();

      // When: positionSec 省略で同じ ep-1 を play する
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // Then: prev.positionSec=55 から再開
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        positionSec: 55,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 55, { play: true });
    });

    it("play('ep-2') を ep-1 active から呼ぶと新 audioRef を setAudioSource で張り positionSec:0・durationSec:null にする", () => {
      // Given: ep-1 を再生中で duration も入った hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitDuration(600);
        subscribe.emitPosition(30);
      });
      setAudioSourceMock.mockClear();
      seekAudioElementMock.mockClear();

      // When: 別 episode ep-2 を play する
      act(() => {
        result.current.play("ep-2", "/episodes/ep-2/audio");
      });

      // Then: 新しい音源を張り直し・ep-2/loading/0/null（duration は引き継がない）
      expect(setAudioSourceMock).toHaveBeenCalledExactlyOnceWith(audio, "/episodes/ep-2/audio");
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-2",
        audioRef: "/episodes/ep-2/audio",
        phase: { phase: "loading" },
        positionSec: 0,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 0, { play: true });
    });

    it("同じ episode へ再度 play しても setAudioSource は呼ばない（el.src が既に一致）", () => {
      // Given: ep-1 を再生中の hook
      const { result } = renderHook(() => useEpisodePlayback());
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      setAudioSourceMock.mockClear();

      // When: 同じ ep-1 を play する
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // Then: 音源の張り直しは起きない
      expect(setAudioSourceMock).not.toHaveBeenCalled();
    });

    it("play('ep-1') を idle から呼ぶと seek/play の前に setAudioSource で音源を張る", () => {
      // Given: audio を張った idle の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });

      // When: ep-1 を play する
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // Then: 空だった src に音源が張られてから seek/play が委譲される
      expect(setAudioSourceMock).toHaveBeenCalledExactlyOnceWith(audio, "/episodes/ep-1/audio");
      expect(setAudioSourceMock.mock.invocationCallOrder[0]).toBeLessThan(
        seekAudioElementMock.mock.invocationCallOrder[0] ?? Number.POSITIVE_INFINITY,
      );
    });

    it("audioElementRef 未設定でも play は例外を投げず、state だけ active/loading/引数の positionSec へ倒す", () => {
      // Given: ref 未設定の hook
      const { result } = renderHook(() => useEpisodePlayback());

      // When: 45 秒指定で play する
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio", 45);
      });

      // Then: 例外なし・state だけ倒る・audio 操作は呼ばれない
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
        positionSec: 45,
        durationSec: null,
      });
      expect(seekAudioElementMock).not.toHaveBeenCalled();
      expect(setAudioSourceMock).not.toHaveBeenCalled();
    });
  });

  describe("seek（topic の sec bar：そこから再生）", () => {
    it("seek('ep-1', 120) を idle から呼ぶと active/phase:loading/positionSec:120 にし、seekAudioElement(audio, 120, {play:true}) を呼ぶ", () => {
      // Given: audio を張った idle の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });

      // When: ep-1 を 120 秒へ seek する
      act(() => {
        result.current.seek("ep-1", "/episodes/ep-1/audio", 120);
      });

      // Then: idle からでもその位置から再生開始（play:true）
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
        positionSec: 120,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: true });
    });

    it("seek('ep-1', 120) を同じ episode の再生中（phase:playing）から呼ぶと その位置から再生継続する", () => {
      // Given: ep-1 を再生し playing になっている hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitPhase("playing");
      });
      seekAudioElementMock.mockClear();

      // When: 120 秒へ seek する
      act(() => {
        result.current.seek("ep-1", "/episodes/ep-1/audio", 120);
      });

      // Then: 再生継続・(audio, 120, {play:true})
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        positionSec: 120,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: true });
    });

    it("seek('ep-1', 120) を同じ episode の停止中（phase:paused）から呼んでも その位置から再生を始める", () => {
      // Given: ep-1 を再生後 paused になっている hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitPhase("paused");
      });
      seekAudioElementMock.mockClear();

      // When: 120 秒へ seek する（停止中でも topic bar は「そこから聴く」）
      act(() => {
        result.current.seek("ep-1", "/episodes/ep-1/audio", 120);
      });

      // Then: 停止中でも play:true で再生開始
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
        positionSec: 120,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: true });
    });

    it("seek('ep-2', 30) を ep-1 再生中から呼ぶと新 audioRef を setAudioSource で張り、ep-2 をその位置から再生する", () => {
      // Given: ep-1 を再生し playing になっている hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitPhase("playing");
        subscribe.emitDuration(600);
      });
      setAudioSourceMock.mockClear();
      seekAudioElementMock.mockClear();

      // When: 別 episode ep-2 の topic sec へ seek する
      act(() => {
        result.current.seek("ep-2", "/episodes/ep-2/audio", 30);
      });

      // Then: 音源張り直し・ep-2 をその位置から再生（play:true）。duration は引き継がない
      expect(setAudioSourceMock).toHaveBeenCalledExactlyOnceWith(audio, "/episodes/ep-2/audio");
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-2",
        audioRef: "/episodes/ep-2/audio",
        phase: { phase: "loading" },
        positionSec: 30,
        durationSec: null,
      });
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 30, { play: true });
    });

    it("seek は渡された audioRef をそのまま active 枝へ載せる", () => {
      // Given: audio を張った idle の hook
      const { result } = renderHook(() => useEpisodePlayback());
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });

      // When: 任意の audioRef を渡して seek する
      act(() => {
        result.current.seek("ep-9", "/custom/path/ep-9.mp3", 12);
      });

      // Then: active 枝の audioRef は引数の値
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-9",
        audioRef: "/custom/path/ep-9.mp3",
        positionSec: 12,
      });
    });

    it("audioElementRef 未設定でも seek は例外を投げず、state だけ引数の positionSec へ倒す", () => {
      // Given: ref 未設定の hook
      const { result } = renderHook(() => useEpisodePlayback());

      // When: seek する
      act(() => {
        result.current.seek("ep-1", "/episodes/ep-1/audio", 42);
      });

      // Then: 例外なし・state だけ倒る（loading）・audio 操作は呼ばれない
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // When: playing phase を通知する
      act(() => {
        subscribe.emitPhase("playing");
      });

      // Then: playing
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // When: 位置通知が来る
      act(() => {
        subscribe.emitPosition(37.2);
      });

      // Then: positionSec だけ更新・他は維持
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
        positionSec: 37.2,
        durationSec: null,
      });
    });

    it("kind=idle 中の timeupdate 通知は無視される", () => {
      // Given: 何も play していない idle の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });

      // When: idle のまま位置通知が来る（updateActive は active のときだけ写す）
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // When: 長さ通知が来る
      act(() => {
        subscribe.emitDuration(1234);
      });

      // Then: durationSec だけ更新・他は維持
      expect(result.current.playback).toEqual({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "loading" },
        positionSec: 0,
        durationSec: 1234,
      });
    });

    it("kind=idle 中の loadedmetadata 通知は無視される", () => {
      // Given: 何も play していない idle の hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      act(() => {
        result.current.audioElementRef.current = createAudioRefTarget();
      });

      // When: idle のまま長さ通知が来る
      act(() => {
        subscribe.emitDuration(500);
      });

      // Then: idle のまま
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
        await Promise.resolve();
      });

      // Then: error/audio-load-failed
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "error", reason: "audio-load-failed" },
      });
    });
  });

  describe("stop / 購読 / 参照安定", () => {
    it("stop() すると直前 audio を pauseAudioElement で止め（load は呼ばない）、その位置で active/paused を維持する", () => {
      // Given: ep-1 を再生中で position が進んだ hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitPhase("playing");
        subscribe.emitPosition(73.5);
      });
      pauseAudioElementMock.mockClear();
      setAudioSourceMock.mockClear();

      // When: stop する
      act(() => {
        result.current.stop();
      });

      // Then: pause は呼ぶが頭出しはしない。idle にも戻さず active/paused・positionSec 保持。setAudioSource なし
      expect(pauseAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio);
      expect(setAudioSourceMock).not.toHaveBeenCalled();
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        audioRef: "/episodes/ep-1/audio",
        phase: { phase: "paused" },
        positionSec: 73.5,
      });
    });

    it("stop() 後に同じ episode を play すると、保持した positionSec の続きから再生する", () => {
      // Given: ep-1 を再生→position 120→stop した hook
      const { result } = renderHook(() => useEpisodePlayback());
      const subscribe = createSubscribeFake();
      const audio = createAudioRefTarget();
      act(() => {
        result.current.audioElementRef.current = audio;
      });
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      act(() => {
        subscribe.emitPhase("playing");
        subscribe.emitPosition(120);
      });
      act(() => {
        result.current.stop();
      });
      seekAudioElementMock.mockClear();

      // When: 位置指定なしで同じ ep-1 を play する
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // Then: resume 位置 120 から再生（seekAudioElement(audio, 120, {play:true})）
      expect(seekAudioElementMock).toHaveBeenCalledExactlyOnceWith(audio, 120, { play: true });
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "loading" },
        positionSec: 120,
      });
    });

    it("audioElementRef 未設定でも stop() は例外を投げず、active/paused へ倒す（play 済みなら）", () => {
      // Given: play 済みだが ref 未設定の hook
      const { result } = renderHook(() => useEpisodePlayback());
      act(() => {
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // When: stop する
      act(() => {
        result.current.stop();
      });

      // Then: 例外なし・Adapter は呼ばれない・active/paused（positionSec は play 時の 0）
      expect(pauseAudioElementMock).not.toHaveBeenCalled();
      expect(result.current.playback).toMatchObject({
        kind: "active",
        episodeId: "ep-1",
        phase: { phase: "paused" },
      });
    });

    it("何も play していない idle で stop() を呼んでも idle のまま", () => {
      // Given: 何も play していない hook
      const { result } = renderHook(() => useEpisodePlayback());

      // When: stop する
      act(() => {
        result.current.stop();
      });

      // Then: idle のまま
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });

      // When: ep-2 を play する
      act(() => {
        result.current.play("ep-2", "/episodes/ep-2/audio");
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
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
        result.current.play("ep-1", "/episodes/ep-1/audio");
      });
      rerender();

      // Then: 参照不変
      expect(result.current.play).toBe(firstPlay);
      expect(result.current.seek).toBe(firstSeek);
      expect(result.current.stop).toBe(firstStop);
    });
  });
});
