import { vi } from "vitest";
import type { AudioLifecyclePhase, AudioStateHandlers } from "./audio-element.ts";

/** 実 `audio-element.ts` の module 形（`importOriginal()` の戻りと `vi.mock` factory の戻りの型）。 */
type AudioElementModule = typeof import("./audio-element.ts");
type SubscribeAudioStateFn = AudioElementModule["subscribeAudioState"];
type SeekAudioElementFn = AudioElementModule["seekAudioElement"];

/**
 * `lib/audio-element.ts` を module-level で差し替えるための Fake。
 *
 * why: happy-dom の生 `<audio>` は `playing` / `error` 等の lifecycle event を発火せず、
 *   `play()` も no-op なので、phase 通知 → `deriveEpisodeRows` の `isPlaying` → DOM 強調、
 *   および audio 失敗の非 blocking error 伝播という「合成で初めて見える状態伝播」を
 *   Page 越しに検証できない（Decision 2026-09-04T18-30-00 Rejected §4）。この Fake は
 *   `subscribeAudioState` へ渡された handler を捕まえ、test 側から phase を能動発火できる形にする。
 *   `hash-selection-adapter.fake.ts` と同じく別 file に置き、`audio-element.ts` の公開 API は
 *   汚さない（`external-dependencies.md` §6-3）。
 *
 * 既定は pass-through で、`install()` を呼ぶまで実 `audio-element.ts` の関数へそのまま委譲する。
 * これにより「生 `<audio>` の src 属性だけ見る」既存 case はそのまま通り、
 * 能動発火が要る case だけ `install()` で capture mode へ切り替える。命令 API のうち
 * `setAudioSource` / `pauseAudioElement` は capture mode でも実物を使う（差し替えの必要が無い）。
 */
export type FakeAudioElementModule = {
  /** `vi.mock` の factory がそのまま返す module 形。 */
  module: AudioElementModule;
  /** test 本体から phase を発火し、再生失敗を仕込む操作面。 */
  control: {
    /** capture mode へ切り替える。以後 `subscribeAudioState` は handler を捕まえて実購読しない。 */
    install(): void;
    /** capture 済み handler へ phase 通知を流す（`onPhaseChange`）。 */
    emitPhase(phase: AudioLifecyclePhase): void;
    /** capture mode の間、`seekAudioElement({ play: true })` を reject させる。 */
    failPlayback(): void;
    /** capture mode・捕捉 handler・失敗フラグを初期状態へ戻す（`beforeEach` 用）。 */
    reset(): void;
  };
};

/**
 * Fake module を組み立てる。`original` に実 `audio-element.ts`（`importOriginal()` の戻り）を渡す。
 *
 * @ensure `install()` 前は `subscribeAudioState` / `seekAudioElement` も `original` へ委譲する。
 *   `install()` 後は `subscribeAudioState` が handler を保持して no-op unsubscribe を返し、
 *   `seekAudioElement` は `currentTime` 代入だけ行う（`failPlayback()` 済みなら reject）。
 *   `setAudioSource` / `pauseAudioElement` は常に `original` の実関数
 */
export function createFakeAudioElementModule(original: AudioElementModule): FakeAudioElementModule {
  let captureMode = false;
  let capturedHandlers: AudioStateHandlers | null = null;
  let playbackShouldFail = false;

  const subscribeAudioState = vi.fn(
    (el: HTMLAudioElement, handlers: AudioStateHandlers): (() => void) => {
      if (!captureMode) {
        return original.subscribeAudioState(el, handlers);
      }
      capturedHandlers = handlers;
      return () => {
        capturedHandlers = null;
      };
    },
  ) as unknown as SubscribeAudioStateFn;

  const seekAudioElement = vi.fn(
    (el: HTMLAudioElement, positionSec: number, opts: { play: boolean }): Promise<void> => {
      if (!captureMode) {
        return original.seekAudioElement(el, positionSec, opts);
      }
      el.currentTime = positionSec;
      if (playbackShouldFail) {
        return Promise.reject(new Error("再生失敗"));
      }
      return Promise.resolve();
    },
  ) as unknown as SeekAudioElementFn;

  return {
    module: {
      subscribeAudioState,
      seekAudioElement,
      setAudioSource: original.setAudioSource,
      pauseAudioElement: original.pauseAudioElement,
    },
    control: {
      install(): void {
        captureMode = true;
      },
      emitPhase(phase: AudioLifecyclePhase): void {
        // capture mode で購読が張られた後にだけ呼ぶ前提。未購読なら TypeError で誤用に気づける
        (capturedHandlers as AudioStateHandlers).onPhaseChange(phase);
      },
      failPlayback(): void {
        playbackShouldFail = true;
      },
      reset(): void {
        captureMode = false;
        capturedHandlers = null;
        playbackShouldFail = false;
      },
    },
  };
}
