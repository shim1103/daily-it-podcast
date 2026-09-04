import type { ReactElement } from "react";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import { formatEpisodeDate } from "../../utils/format-episode-date.ts";
import { formatNumberedEpisodeTitle } from "../../utils/format-numbered-episode-title.ts";
import { formatTopicTitles } from "../../utils/format-topic-titles.ts";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";

export type EpisodeRowProps = {
  episode: EpisodeData;
  episodeCount: number;
  episodeIndex: number;
  isSelected: boolean;
  isPlaying: boolean;
  onSelect: (episodeId: string) => void;
  onPlay: (episodeId: string) => void;
  onStop: () => void;
};

/**
 * 1 episode の meta と select / play・stop affordance を描画する（契約 stub）。
 *
 * @require isSelected / isPlaying は caller が derive して渡す
 * @ensure Row 自身は derive しない。表示整形は utils に委譲する
 */
export function EpisodeRow({
  episode,
  episodeCount,
  episodeIndex,
  isSelected,
  isPlaying,
  onSelect,
  onPlay,
  onStop,
}: EpisodeRowProps): ReactElement {
  const topicTitles = formatTopicTitles(episode.body.topics);
  const numberedTitle = formatNumberedEpisodeTitle(episodeCount, episodeIndex, episode.title);

  return (
    <article
      className="episode-row"
      data-selected={isSelected ? "true" : "false"}
      data-playing={isPlaying ? "true" : "false"}
    >
      <button type="button" onClick={() => onSelect(episode.episodeId)}>
        <LabeledText tag="span" datasetKey="episodeDate" text={formatEpisodeDate(episode.date)} />
        <LabeledText tag="span" datasetKey="episodeTitle" text={numberedTitle} />
        {topicTitles !== "" && (
          <LabeledText tag="span" datasetKey="episodeTopics" text={topicTitles} />
        )}
        <LabeledText
          tag="span"
          datasetKey="episodeDurationSec"
          text={formatDurationMmSs(episode.durationSec)}
        />
      </button>
      <button
        type="button"
        aria-label={isPlaying ? "停止" : "再生"}
        onClick={() => (isPlaying ? onStop() : onPlay(episode.episodeId))}
      >
        {isPlaying ? "停止" : "再生"}
      </button>
    </article>
  );
}
