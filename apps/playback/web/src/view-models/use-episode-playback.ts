import { useCallback, useEffect, useRef, useState, type Ref } from "react";
import { buildRequestUrl } from "../utils/build-request-url.ts";

export type EpisodePlayback =
  | { kind: "idle" }
  | { kind: "playing"; episodeId: string; positionSec: number; durationSec: number };

export type EpisodePlaybackViewModel = {
  audioElementRef: Ref<HTMLAudioElement | null>;
  playback: EpisodePlayback;
  resolvedSrc: string | undefined;
  play(episodeId: string, audioRefPath: string): void;
  seek(startSec: number): void;
};

/**
 * 一覧 page が共有する hidden audio と再生 state・seek・timeupdate を束ねる ViewModel hook。
 *
 * @require baseUrl は audio 直結先の origin相当
 * @ensure play は URL 組立と再生開始、seek は audio の currentTime 更新と play、
 *   timeupdate で positionSec / durationSec を更新する。audio 未 mount 時の seek は pending として保持し mount 後に適用する
 * @invariant list 取得・選択 state を持たない
 */
export function useEpisodePlayback(baseUrl: string): EpisodePlaybackViewModel {
  const audioElementRef = useRef<HTMLAudioElement | null>(null);
  const pendingSeekSecRef = useRef<number | null>(null);
  const [audioMountTick, setAudioMountTick] = useState(0);
  const [playback, setPlayback] = useState<EpisodePlayback>({ kind: "idle" });
  const [activeAudioRef, setActiveAudioRef] = useState<string | null>(null);

  const registerAudioElement = useCallback((node: HTMLAudioElement | null) => {
    audioElementRef.current = node;
    if (node !== null) {
      setAudioMountTick((tick) => tick + 1);
    }
  }, []);

  const resolvedSrc =
    playback.kind === "playing" && activeAudioRef !== null
      ? buildRequestUrl(baseUrl, activeAudioRef)
      : undefined;

  const play = useCallback((episodeId: string, audioRefPath: string): void => {
    setActiveAudioRef(audioRefPath);
    setPlayback({ kind: "playing", episodeId, positionSec: 0, durationSec: 0 });
  }, []);

  const seek = useCallback((startSec: number): void => {
    const audio = audioElementRef.current;
    if (audio === null) {
      pendingSeekSecRef.current = startSec;
      setPlayback((previous) =>
        previous.kind === "playing" ? { ...previous, positionSec: startSec } : previous,
      );
      return;
    }
    audio.currentTime = startSec;
    void audio.play();
    setPlayback((previous) =>
      previous.kind === "playing" ? { ...previous, positionSec: startSec } : previous,
    );
  }, []);

  useEffect(() => {
    if (playback.kind !== "playing" || activeAudioRef === null) {
      return;
    }
    const audio = audioElementRef.current;
    if (audio === null) {
      return;
    }
    const nextSrc = buildRequestUrl(baseUrl, activeAudioRef);
    if (audio.getAttribute("src") !== nextSrc) {
      audio.src = nextSrc;
    }

    const pendingSeekSec = pendingSeekSecRef.current;
    if (pendingSeekSec !== null) {
      audio.currentTime = pendingSeekSec;
      pendingSeekSecRef.current = null;
      setPlayback((previous) => {
        /* v8 ignore next 3 -- pending seek 適用は playing effect の存続中だけ起き、解除と同じ tick で idle へ遷移しない */
        if (previous.kind !== "playing") {
          return previous;
        }
        return { ...previous, positionSec: pendingSeekSec };
      });
    }

    void audio.play();
  }, [
    activeAudioRef,
    audioMountTick,
    baseUrl,
    playback.kind,
    playback.kind === "playing" ? playback.episodeId : null,
  ]);

  useEffect(() => {
    if (playback.kind !== "playing") {
      return;
    }
    const audio = audioElementRef.current;
    if (audio === null) {
      return;
    }

    const sync = (): void => {
      setPlayback((previous) => {
        /* v8 ignore next 3 -- timeupdate listener は playing effect の存続中だけ登録され、解除と同じ tick で idle へ遷移しない */
        if (previous.kind !== "playing") {
          return previous;
        }
        return {
          ...previous,
          positionSec: audio.currentTime,
          durationSec: Number.isFinite(audio.duration) ? audio.duration : previous.durationSec,
        };
      });
    };

    sync();
    audio.addEventListener("timeupdate", sync);
    audio.addEventListener("loadedmetadata", sync);
    audio.addEventListener("durationchange", sync);

    return () => {
      audio.removeEventListener("timeupdate", sync);
      audio.removeEventListener("loadedmetadata", sync);
      audio.removeEventListener("durationchange", sync);
    };
  }, [audioMountTick, playback.kind, playback.kind === "playing" ? playback.episodeId : null]);

  return { audioElementRef: registerAudioElement, playback, resolvedSrc, play, seek };
}
