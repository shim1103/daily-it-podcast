import { useCallback, type ReactElement, type SyntheticEvent } from "react";
import type { EpisodeEntryPlayback } from "./episode-list-entry.tsx";
import "./episode-seek-bar.css";

export type EpisodeSeekBarProps = {
  playback: EpisodeEntryPlayback;
  onSeek: (startSec: number) => void;
};

/**
 * 再生位置 range。表示と onSeek のみ。自前 listener を持たない。
 *
 * @require playback は entry 向け union
 * @ensure playing 時は positionSec / durationSec を表示し、操作で onSeek を呼ぶ
 * @invariant audio 要素へ直接触れない
 */
export function EpisodeSeekBar({ playback, onSeek }: EpisodeSeekBarProps): ReactElement {
  const positionSec = playback.kind === "playing" ? playback.positionSec : 0;
  const durationSec = playback.kind === "playing" ? playback.durationSec : 0;

  const onChange = useCallback(
    (event: SyntheticEvent<HTMLInputElement>) => {
      onSeek(Number(event.currentTarget.value));
    },
    [onSeek],
  );

  return (
    <input
      type="range"
      className="episode-seek-bar"
      min={0}
      max={durationSec > 0 ? durationSec : 0}
      step={0.1}
      value={positionSec}
      onChange={onChange}
      aria-label="再生位置"
    />
  );
}
