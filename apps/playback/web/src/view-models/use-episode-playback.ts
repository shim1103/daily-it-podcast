import { useCallback, useRef, type RefObject } from "react";
import type { PlaybackPhase } from "./playback-state.ts";

export type EpisodePlaybackViewModel = {
  playedEpisodeId: string | null;
  playbackPhase: PlaybackPhase;
  audioElementRef: RefObject<HTMLAudioElement | null>;
  play(episodeId: string): void;
  stop(): void;
};

/**
 * 再生 id・playback phase・`<audio>` lifecycle を担う hook（契約 stub）。
 *
 * @ensure 初期は playedEpisodeId=null・playbackPhase=idle。play/stop は no-op（実装は C で置換）
 * @invariant 選択 id は変えない
 */
export function useEpisodePlayback(): EpisodePlaybackViewModel {
  const audioElementRef = useRef<HTMLAudioElement>(null);

  const play = useCallback((_episodeId: string): void => {
    // stub: 実装は C で置換
  }, []);

  const stop = useCallback((): void => {
    // stub: 実装は C で置換
  }, []);

  return {
    playedEpisodeId: null,
    playbackPhase: "idle",
    audioElementRef,
    play,
    stop,
  };
}
