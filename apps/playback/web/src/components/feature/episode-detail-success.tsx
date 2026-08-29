import type { ReactElement } from "react";
import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import type { EpisodeEntryPlayback } from "./episode-list-entry.tsx";
import { EpisodeManuscript } from "./episode-manuscript.tsx";
import { EpisodeSeekBar } from "./episode-seek-bar.tsx";
import "./episode-detail.css";

export type EpisodeDetailSuccessProps = {
  episode: EpisodeData;
  playback: EpisodeEntryPlayback;
  onSeek: (startSec: number) => void;
};

/**
 * 詳細取得 success 時の seek bar と原稿全文。
 *
 * @require episode は GetEpisodeResponse、playback / onSeek は Page 由来
 * @ensure seek bar を先頭に、body を EpisodeManuscript で描画する
 * @invariant audio 要素へ直接触れない
 */
export function EpisodeDetailSuccess({
  episode,
  playback,
  onSeek,
}: EpisodeDetailSuccessProps): ReactElement {
  return (
    <div className="episode-detail">
      <EpisodeSeekBar playback={playback} onSeek={onSeek} />
      <EpisodeManuscript body={episode.body} onSeek={onSeek} />
    </div>
  );
}
