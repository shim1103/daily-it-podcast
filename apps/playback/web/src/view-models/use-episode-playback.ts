import { useCallback, useEffect, useRef, useState, type RefObject } from "react";
import type { AudioLifecyclePhase } from "../lib/audio-element.ts";
import {
  pauseAudioElement,
  resetAudioElement,
  seekAudioElement,
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

/** 同じ episode が今まさに再生中（`playing`）か。seek で再生を継続するかの判定に使う。 */
function isPlayingSameEpisode(prev: PlaybackState, episodeId: string): boolean {
  return prev.kind === "active" && prev.episodeId === episodeId && prev.phase.phase === "playing";
}

/**
 * 再生対象 episode・phase・再生位置・長さを判別可能 union（`PlaybackState`）で持つ hook。
 * `<audio>` の命令的操作と event 購読は `lib/audio-element.ts`（Driven Adapter）へ委譲する。
 *
 * @ensure 初期は idle。`play(episodeId, audioRef, positionSec?)` は指定秒から再生、
 *   `seek(episodeId, audioRef, positionSec)` は位置移動のみ（同じ episode が再生中なら再生継続）。
 *   `audioRef` は呼び出し側が catalog から引き当てて渡し、hook はそれを `active` 枝へ載せるだけで
 *   catalog を参照しない。違う episode へ切り替える時は直前 audio を reset。audio 未 mount 時は
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

  const releaseSubscription = useCallback((): void => {
    unsubscribeRef.current?.();
    unsubscribeRef.current = null;
  }, []);

  const commitPlayback = useCallback((next: PlaybackState): void => {
    playbackRef.current = next;
    setPlayback(next);
  }, []);

  // why: idle 中に届いた Adapter 通知は不正遷移として無視する（防御責務 §7-1）。
  //   active のときだけ patch を当て、state と ref を同時に書く
  const updateActive = useCallback((patch: (state: ActivePlayback) => ActivePlayback): void => {
    setPlayback((current) => {
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

      if (prev.kind === "active" && prev.episodeId !== episodeId && audio !== null) {
        // why: 違う episode へ切り替える時、直前 audio を止めて位置を戻し読み直す（B Decision §1-3）
        resetAudioElement(audio);
      }
      commitPlayback({
        kind: "active",
        episodeId,
        audioRef,
        phase: transition.phase,
        positionSec,
        durationSec,
      });
      if (audio === null) {
        // todo: 表示側 Issue で `<audio src>` を ViewModel 管理へ寄せ、要素 mount 後に
        //   この positionSec へ飛ばす。本 Issue は state 遷移だけ（B Decision §1-6）
        return;
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
      const keepPlaying = isPlayingSameEpisode(playbackRef.current, episodeId);
      const phase: PlaybackPhase = keepPlaying ? { phase: "playing" } : { phase: "paused" };
      moveTo(episodeId, audioRef, positionSec, { phase, shouldPlay: keepPlaying });
    },
    [moveTo],
  );

  const stop = useCallback((): void => {
    const audio = audioElementRef.current;
    if (audio !== null) {
      pauseAudioElement(audio);
    }
    releaseSubscription();
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
