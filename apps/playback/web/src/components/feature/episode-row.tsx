import type { ReactElement } from "react";
import { formatEpisodeDate } from "../../utils/format-episode-date.ts";
import { formatNumberedEpisodeTitle } from "../../utils/format-numbered-episode-title.ts";
import { formatTopicTitles } from "../../utils/format-topic-titles.ts";
import type { EpisodeListItemData } from "../../view-models/episode-list-view-model.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";
import { EpisodePlayButton } from "./episode-play-button.tsx";
import "./episode-row.css";

export type EpisodeRowProps = {
  episode: EpisodeListItemData;
  episodeCount: number;
  episodeIndex: number;
  onSelect: (episodeId: string) => void;
  onPlay: () => void;
};

/**
 * 一覧 1 行。表示と play / select のみ。
 *
 * @require episode は EpisodeListItem 1件
 * @ensure date・通し番号付き title・topics 行と再生 pill を描画する
 * @invariant selection / detail / audio を知らない
 */
export function EpisodeRow({
  episode,
  episodeCount,
  episodeIndex,
  onSelect,
  onPlay,
}: EpisodeRowProps): ReactElement {
  const topicTitles = formatTopicTitles(episode.topics);
  const numberedTitle = formatNumberedEpisodeTitle(episodeCount, episodeIndex, episode.title);

  return (
    <article className="episode-row">
      <button
        type="button"
        className="episode-row__hit"
        onClick={(event) => {
          onSelect(episode.episodeId);
          // why: click 後の :focus が残り、別 item の :hover と二重に紫線が付くのを防ぐ
          event.currentTarget.blur();
        }}
      >
        <span className="episode-row__date">
          <LabeledText tag="span" datasetKey="episodeDate" text={formatEpisodeDate(episode.date)} />
        </span>
        <span className="episode-row__title">
          <LabeledText tag="span" datasetKey="episodeTitle" text={numberedTitle} />
        </span>
        {topicTitles !== "" && (
          <span className="episode-row__topics">
            <LabeledText tag="span" datasetKey="episodeTopics" text={topicTitles} />
          </span>
        )}
      </button>
      <div className="episode-row__play">
        <EpisodePlayButton durationSec={episode.durationSec} onPlay={onPlay} />
      </div>
    </article>
  );
}
