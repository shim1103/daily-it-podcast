import { useCallback, type ReactElement, type SyntheticEvent } from "react";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import "./episode-play-button.css";

export type EpisodePlayButtonProps = {
  durationSec: number;
  onPlay: () => void;
};

/**
 * 再生 pill。表示と onPlay のみ。
 *
 * @require durationSec は一覧 API の尺
 * @ensure click で onPlay を呼ぶ。audio 操作は持たない
 * @invariant 親の行 select と競合しないよう stopPropagation する
 */
export function EpisodePlayButton({ durationSec, onPlay }: EpisodePlayButtonProps): ReactElement {
  const onClick = useCallback(
    (event: SyntheticEvent<HTMLButtonElement>) => {
      event.stopPropagation();
      onPlay();
    },
    [onPlay],
  );

  return (
    <button type="button" className="episode-play-button" onClick={onClick}>
      <span className="episode-play-button__icon" aria-hidden="true">
        ▶︎
      </span>
      <span className="episode-play-button__duration">{formatDurationMmSs(durationSec)}</span>
    </button>
  );
}
