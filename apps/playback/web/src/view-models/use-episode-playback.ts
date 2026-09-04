import { useCallback, useEffect, useRef, useState, type RefObject } from "react";
import type { AudioLifecyclePhase } from "../lib/audio-element.ts";
import {
  pauseAudioElement,
  seekAudioElement,
  setAudioSource,
  subscribeAudioState,
} from "../lib/audio-element.ts";
import type { ActivePlayback, PlaybackPhase, PlaybackState } from "./playback-state.ts";

export type EpisodePlaybackViewModel = {
  playback: PlaybackState;
  audioElementRef: RefObject<HTMLAudioElement | null>;
  play(episodeId: string, audioRef: string, positionSec?: number): void;
  seek(episodeId: string, audioRef: string, positionSec: number): void;
  stop(): void;
};

const IDLE: PlaybackState = { kind: "idle" };

/** Adapter の 4 値 phase を ViewModel の判別可能 union へ写す。error 枝だけ失敗理由を持つ。 */
function toPlaybackPhase(lifecyclePhase: AudioLifecyclePhase): PlaybackPhase {
  return lifecyclePhase === "error"
    ? { phase: "error", reason: "audio-load-failed" }
    : { phase: lifecyclePhase };
}

/**
 * 再生対象 episode・phase・再生位置・長さを判別可能 union（`PlaybackState`）で持つ hook。
 * `<audio>` の命令的操作と event 購読は `lib/audio-element.ts`（Driven Adapter）へ委譲する。
 *
 * @ensure 初期は idle。`play(episodeId, audioRef, positionSec?)` は指定秒から再生（同じ episode を
 *   停止後に再度 play すると直近 positionSec から再開）。`seek(episodeId, audioRef, positionSec)` は
 *   topic の sec bar 用で、押した時点の state（idle / 別 episode / 停止中）に関わらず、その episode を
 *   その位置から再生する（別 episode なら音源を張り替える）。`stop()` は頭出しせずその位置で pause し、
 *   `active/paused`（positionSec 保持）を維持する（idle には戻さない）。
 *   `audioRef` は呼び出し側が catalog から引き当て baseUrl と結合した絶対 URL を渡し、hook は
 *   それを `active` 枝へ載せ、`seek` / `play` の直前に `<audio src>` へ命令的に張る
 *   （`el.src` が既にその URL なら張り直さない）。catalog は参照しない。audio 未 mount 時は
 *   state だけ進める。Adapter の phase / position / duration 通知は `active` の間だけ state へ写す。
 *   unmount で購読解除。
 * @invariant throw しない。JSX を持たない。選択 id は変えない。catalog status を知らない。
 *   Browser API の命令的操作と event 購読を直接持たない（すべて Adapter に閉じる）
 *
 * 関連 Decision: docs/decisions/2026-09-03T16-20-00-feature-playback-web-view-models.md
 */
export function useEpisodePlayback(): EpisodePlaybackViewModel {
  const [playback, setPlayback] = useState<PlaybackState>(IDLE);
  const audioElementRef = useRef<HTMLAudioElement>(null);
  // why: play/seek/stop 内で「今どの再生 state か」を setState の非同期性に依存せず読むため、
  //   state と同型の ref を持ち、更新は commitPlayback / updateActive の 1 経路に通す（B Decision §1-4）
  const playbackRef = useRef<PlaybackState>(IDLE);
  const unsubscribeRef = useRef<(() => void) | null>(null);
  // why: 「今 <audio> にどの音源 URL を張ったか」を hook 側で覚える。`el.src` の getter は常に
  //   絶対 URL を返すため、baseUrl が空（相対 audioRef）だと `el.src !== audioRef` が毎回真になり、
  //   同じ episode の seek のたび load() で再生が途切れる。渡された文字列そのものを基準に差分判定する
  const currentSourceRef = useRef<string | null>(null);

  const releaseSubscription = useCallback((): void => {
    unsubscribeRef.current?.();
    unsubscribeRef.current = null;
  }, []);

  const commitPlayback = useCallback((next: PlaybackState): void => {
    playbackRef.current = next;
    setPlayback(next);
  }, []);

  // why: active でない間に届いた Adapter 通知は不正遷移として無視する（防御責務 §7-1）。
  //   active のときだけ patch を当て、state と ref を同時に書く
  const updateActive = useCallback((patch: (state: ActivePlayback) => ActivePlayback): void => {
    setPlayback((current) => {
      // why: 購読は play で active になった後に張り stop で外すため、通知到達時は常に active。
      //   この guard は購読解除と listener detach の間に queue 済み event が発火する race への
      //   防御で、通常経路からは到達しない（defensive-design.md §7-1）
      /* v8 ignore next 3 */
      if (current.kind !== "active") {
        return current;
      }
      const next = patch(current);
      playbackRef.current = next;
      return next;
    });
  }, []);

  const applyPhase = useCallback(
    (lifecyclePhase: AudioLifecyclePhase): void => {
      const phase = toPlaybackPhase(lifecyclePhase);
      updateActive((state) => ({ ...state, phase }));
    },
    [updateActive],
  );

  const applyPosition = useCallback(
    (positionSec: number): void => {
      updateActive((state) => ({ ...state, positionSec }));
    },
    [updateActive],
  );

  const applyDuration = useCallback(
    (durationSec: number): void => {
      updateActive((state) => ({ ...state, durationSec }));
    },
    [updateActive],
  );

  const moveTo = useCallback(
    (
      episodeId: string,
      audioRef: string,
      positionSec: number,
      transition: { phase: PlaybackPhase; shouldPlay: boolean },
    ): void => {
      const audio = audioElementRef.current;
      const prev = playbackRef.current;
      const isSameEpisode = prev.kind === "active" && prev.episodeId === episodeId;
      const durationSec = isSameEpisode ? prev.durationSec : null;

      commitPlayback({
        kind: "active",
        episodeId,
        audioRef,
        phase: transition.phase,
        positionSec,
        durationSec,
      });
      if (audio === null) {
        // why: 要素未 mount 時は state だけ進め、次に mount された要素へ moveTo が呼ばれるのを待つ。
        //   `<audio>` は常時 mount される想定なので通常ここは通らない（防御）
        return;
      }
      // why: `<audio src>` を Component の props で controlled にすると、この直後の
      //   seekAudioElement / play() が新 src の反映前に走って空要素を再生する。src 指定を
      //   Adapter 経由の命令的操作へ寄せ、seek/play の直前に必ず正しい音源を張る。
      //   同じ audioRef を張り直すと load() で再生が途切れるので、hook が覚えた前回値と比較する
      if (currentSourceRef.current !== audioRef) {
        setAudioSource(audio, audioRef);
        currentSourceRef.current = audioRef;
      }
      releaseSubscription();
      unsubscribeRef.current = subscribeAudioState(audio, {
        onPhaseChange: applyPhase,
        onPositionChange: applyPosition,
        onDurationChange: applyDuration,
      });
      void seekAudioElement(audio, positionSec, { play: transition.shouldPlay }).catch(() => {
        applyPhase("error");
      });
    },
    [applyDuration, applyPhase, applyPosition, commitPlayback, releaseSubscription],
  );

  const play = useCallback(
    (episodeId: string, audioRef: string, positionSec?: number): void => {
      const prev = playbackRef.current;
      const isResumingSameEpisode = prev.kind === "active" && prev.episodeId === episodeId;
      const resumeSec = isResumingSameEpisode ? prev.positionSec : 0;
      // why: positionSec が明示 0 のときはその 0 を使う。省略時だけ resume 位置へ倒す
      const startSec = positionSec ?? resumeSec;
      moveTo(episodeId, audioRef, startSec, { phase: { phase: "loading" }, shouldPlay: true });
    },
    [moveTo],
  );

  const seek = useCallback(
    (episodeId: string, audioRef: string, positionSec: number): void => {
      // why: topic の sec bar は「そこから聴く」ための操作。押した時点の state（idle / 別 episode /
      //   同じ episode の再生中・停止中）に関わらず、その episode をその位置から再生する。
      //   別 episode を指していれば moveTo が音源を張り替える
      moveTo(episodeId, audioRef, positionSec, { phase: { phase: "loading" }, shouldPlay: true });
    },
    [moveTo],
  );

  const stop = useCallback((): void => {
    const audio = audioElementRef.current;
    if (audio !== null) {
      pauseAudioElement(audio);
    }
    releaseSubscription();
    // why: 「停止」は頭出し（reset）ではなく、その位置で pause。idle には戻さず active/paused を
    //   維持し、直近の positionSec を残す（B Decision §1-1/§1-2）。同じ episode を再び「再生」すると
    //   `play` の `isResumingSameEpisode` がこの positionSec を resume 位置に使い、続きから再生する
    const prev = playbackRef.current;
    if (prev.kind === "active") {
      commitPlayback({ ...prev, phase: { phase: "paused" } });
      return;
    }
    commitPlayback(IDLE);
  }, [commitPlayback, releaseSubscription]);

  useEffect(() => {
    return () => {
      releaseSubscription();
    };
  }, [releaseSubscription]);

  return {
    playback,
    audioElementRef,
    play,
    seek,
    stop,
  };
}
