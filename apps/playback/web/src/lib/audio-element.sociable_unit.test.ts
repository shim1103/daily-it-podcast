import { describe, expect, it, vi } from "vitest";
import {
  type AudioLifecyclePhase,
  pauseAudioElement,
  seekAudioElement,
  setAudioSource,
  subscribeAudioState,
} from "./audio-element.ts";

/**
 * `<audio>` を模した Fake。listener 登録・解除を実体で持ち、event を任意発火できる。
 * pause / load / play の呼び出し回数と currentTime・duration を制御・記録する。
 */
function createFakeAudioElement(): HTMLAudioElement & {
  emit(type: string): void;
  listenerCount(type: string): number;
  pauseCalls: number;
  loadCalls: number;
  playCalls: number;
  duration: number;
} {
  const listeners = new Map<string, Set<EventListener>>();
  const fake = {
    src: "",
    currentTime: 0,
    duration: Number.NaN,
    pauseCalls: 0,
    loadCalls: 0,
    playCalls: 0,
    addEventListener(type: string, listener: EventListener): void {
      const set = listeners.get(type) ?? new Set<EventListener>();
      set.add(listener);
      listeners.set(type, set);
    },
    removeEventListener(type: string, listener: EventListener): void {
      listeners.get(type)?.delete(listener);
    },
    pause(): void {
      fake.pauseCalls += 1;
    },
    load(): void {
      fake.loadCalls += 1;
    },
    play(): Promise<void> {
      fake.playCalls += 1;
      return Promise.resolve();
    },
    emit(type: string): void {
      for (const listener of listeners.get(type) ?? []) {
        listener(new Event(type));
      }
    },
    listenerCount(type: string): number {
      return listeners.get(type)?.size ?? 0;
    },
  };
  return fake as unknown as HTMLAudioElement & {
    emit(type: string): void;
    listenerCount(type: string): number;
    src: string;
    pauseCalls: number;
    loadCalls: number;
    playCalls: number;
    duration: number;
  };
}

function createStateHandlers(): {
  onPhaseChange: ReturnType<typeof vi.fn<(phase: AudioLifecyclePhase) => void>>;
  onPositionChange: ReturnType<typeof vi.fn<(positionSec: number) => void>>;
  onDurationChange: ReturnType<typeof vi.fn<(durationSec: number) => void>>;
} {
  return {
    onPhaseChange: vi.fn<(phase: AudioLifecyclePhase) => void>(),
    onPositionChange: vi.fn<(positionSec: number) => void>(),
    onDurationChange: vi.fn<(durationSec: number) => void>(),
  };
}

describe("subscribeAudioState", () => {
  it("playing event で onPhaseChange を phase=playing で呼ぶ", () => {
    // Given: 購読済みの audio
    const audio = createFakeAudioElement();
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: playing event が発火する
    audio.emit("playing");

    // Then: phase=playing で呼ばれる
    expect(handlers.onPhaseChange).toHaveBeenCalledExactlyOnceWith("playing");
  });

  it("pause event で onPhaseChange を phase=paused で呼ぶ", () => {
    // Given: 購読済みの audio
    const audio = createFakeAudioElement();
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: pause event が発火する
    audio.emit("pause");

    // Then: phase=paused で呼ばれる
    expect(handlers.onPhaseChange).toHaveBeenCalledExactlyOnceWith("paused");
  });

  it("ended event で onPhaseChange を phase=ended で呼ぶ", () => {
    // Given: 購読済みの audio
    const audio = createFakeAudioElement();
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: ended event が発火する
    audio.emit("ended");

    // Then: phase=ended で呼ばれる
    expect(handlers.onPhaseChange).toHaveBeenCalledExactlyOnceWith("ended");
  });

  it("error event で onPhaseChange を phase=error で呼ぶ", () => {
    // Given: 購読済みの audio
    const audio = createFakeAudioElement();
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: error event が発火する
    audio.emit("error");

    // Then: phase=error で呼ばれる
    expect(handlers.onPhaseChange).toHaveBeenCalledExactlyOnceWith("error");
  });

  it("timeupdate event で onPositionChange を currentTime で呼ぶ", () => {
    // Given: 再生位置が進んだ audio
    const audio = createFakeAudioElement();
    audio.currentTime = 73.5;
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: timeupdate event が発火する
    audio.emit("timeupdate");

    // Then: onPositionChange が currentTime で呼ばれる
    expect(handlers.onPositionChange).toHaveBeenCalledExactlyOnceWith(73.5);
  });

  it("loadedmetadata event で onDurationChange を duration で呼ぶ", () => {
    // Given: duration が確定した audio
    const audio = createFakeAudioElement();
    audio.duration = 1800;
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: loadedmetadata event が発火する
    audio.emit("loadedmetadata");

    // Then: onDurationChange が duration で呼ばれる
    expect(handlers.onDurationChange).toHaveBeenCalledExactlyOnceWith(1800);
  });

  it("loadedmetadata で duration が NaN の時は onDurationChange を呼ばない", () => {
    // Given: duration が未確定（NaN）の audio
    const audio = createFakeAudioElement();
    audio.duration = Number.NaN;
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: loadedmetadata event が発火する
    audio.emit("loadedmetadata");

    // Then: onDurationChange は呼ばれない（有限でない長さは state へ写さない）
    expect(handlers.onDurationChange).not.toHaveBeenCalled();
  });

  it("loadedmetadata で duration が Infinity の時は onDurationChange を呼ばない", () => {
    // Given: duration が Infinity（ライブ配信等）の audio
    const audio = createFakeAudioElement();
    audio.duration = Number.POSITIVE_INFINITY;
    const handlers = createStateHandlers();
    subscribeAudioState(audio, handlers);

    // When: loadedmetadata event が発火する
    audio.emit("loadedmetadata");

    // Then: onDurationChange は呼ばれない
    expect(handlers.onDurationChange).not.toHaveBeenCalled();
  });

  it("戻り値の関数を呼ぶと全 event の listener を解除する", () => {
    // Given: 購読済みの audio
    const audio = createFakeAudioElement();
    const handlers = createStateHandlers();
    const unsubscribe = subscribeAudioState(audio, handlers);

    // When: 購読解除してから各 event を発火する
    unsubscribe();
    audio.emit("playing");
    audio.emit("pause");
    audio.emit("ended");
    audio.emit("error");
    audio.emit("timeupdate");
    audio.emit("loadedmetadata");

    // Then: どの handler も呼ばれず、listener も残らない
    expect(handlers.onPhaseChange).not.toHaveBeenCalled();
    expect(handlers.onPositionChange).not.toHaveBeenCalled();
    expect(handlers.onDurationChange).not.toHaveBeenCalled();
    for (const type of ["playing", "pause", "ended", "error", "timeupdate", "loadedmetadata"]) {
      expect(audio.listenerCount(type)).toBe(0);
    }
  });
});

describe("setAudioSource", () => {
  it("src を渡した URL にし、load で読み込ませる（頭出しを兼ねる）", () => {
    // Given: 別 URL を指した audio
    const audio = createFakeAudioElement();
    audio.src = "https://example.test/episodes/old/audio";
    audio.currentTime = 42;

    // When: 新しい URL をセットする
    setAudioSource(audio, "https://example.test/episodes/new/audio");

    // Then: src が入れ替わり load される（load が position を 0 へ戻す）
    expect(audio.src).toBe("https://example.test/episodes/new/audio");
    expect(audio.loadCalls).toBe(1);
  });

  it("同じ URL でも load を呼ぶ（呼び出し側が差分判定する前提）", () => {
    // Given: 既に同じ URL を指す audio
    const audio = createFakeAudioElement();
    audio.src = "https://example.test/episodes/ep-1/audio";

    // When: 同じ URL をセットする
    setAudioSource(audio, "https://example.test/episodes/ep-1/audio");

    // Then: load は起きる（重複回避は moveTo 側の責務）
    expect(audio.loadCalls).toBe(1);
  });
});

describe("pauseAudioElement", () => {
  it("stop 用に pause だけを呼び、currentTime はその位置のまま残す", () => {
    // Given: 再生位置が進んだ audio
    const audio = createFakeAudioElement();
    audio.currentTime = 30;

    // When: pause する
    pauseAudioElement(audio);

    // Then: pause は呼ぶが currentTime は 30 のまま（頭出ししない）
    expect(audio.pauseCalls).toBe(1);
    expect(audio.currentTime).toBe(30);
  });

  it("load は呼ばない（source 読み直しは別 episode 切替専用）", () => {
    // Given: audio
    const audio = createFakeAudioElement();

    // When: pause する
    pauseAudioElement(audio);

    // Then: load は起きない
    expect(audio.loadCalls).toBe(0);
  });
});

describe("seekAudioElement", () => {
  it("渡した秒を currentTime にセットする", () => {
    // Given: 再生位置が頭の audio
    const audio = createFakeAudioElement();

    // When: 120 秒へ seek する
    void seekAudioElement(audio, 120, { play: false });

    // Then: currentTime が 120 になる
    expect(audio.currentTime).toBe(120);
  });

  it("play:true のとき seek 後に再生を開始し、戻り Promise が解決する", async () => {
    // Given: audio
    const audio = createFakeAudioElement();

    // When: play:true で seek する
    const result = seekAudioElement(audio, 45, { play: true });

    // Then: currentTime=45・play が呼ばれる・戻り Promise は解決する
    expect(audio.currentTime).toBe(45);
    expect(audio.playCalls).toBe(1);
    await expect(result).resolves.toBeUndefined();
  });

  it("play:false のとき play は呼ばず、戻り Promise が解決する", async () => {
    // Given: audio
    const audio = createFakeAudioElement();

    // When: play:false で seek する
    const result = seekAudioElement(audio, 45, { play: false });

    // Then: currentTime=45・play は呼ばれない・戻り Promise は解決する
    expect(audio.currentTime).toBe(45);
    expect(audio.playCalls).toBe(0);
    await expect(result).resolves.toBeUndefined();
  });

  it("play が undefined を返しても戻り Promise へ包んで返す（古いブラウザ実装）", async () => {
    // Given: play が undefined を返す audio
    const audio = createFakeAudioElement();
    vi.spyOn(audio, "play").mockReturnValue(undefined as unknown as Promise<void>);

    // When: play:true で seek する
    const result = seekAudioElement(audio, 10, { play: true });

    // Then: Promise として解決する
    await expect(result).resolves.toBeUndefined();
  });

  it("play:true で play の rejection を戻り Promise で伝播する", async () => {
    // Given: play が reject する audio
    const audio = createFakeAudioElement();
    vi.spyOn(audio, "play").mockRejectedValue(new Error("再生失敗"));

    // When: play:true で seek する
    const result = seekAudioElement(audio, 10, { play: true });

    // Then: 同じ rejection が戻り Promise で伝わる
    await expect(result).rejects.toThrow("再生失敗");
  });

  it("load は呼ばない（source 読み直しは別 episode 切替専用）", () => {
    // Given: audio
    const audio = createFakeAudioElement();

    // When: seek する
    void seekAudioElement(audio, 10, { play: false });

    // Then: load は起きない
    expect(audio.loadCalls).toBe(0);
  });
});
