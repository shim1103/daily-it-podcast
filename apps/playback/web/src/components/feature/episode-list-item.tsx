import type { ReactElement } from "react";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import { formatEpisodeDate } from "../../utils/format-episode-date.ts";
import { formatNumberedEpisodeTitle } from "../../utils/format-numbered-episode-title.ts";
import { formatTopicTitles } from "../../utils/format-topic-titles.ts";
import type { EpisodeListItemData } from "../../view-models/episode-list-view-model.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";
import "./episode-list-item.css";

export type EpisodeListItemProps = {
  episode: EpisodeListItemData;
  episodeCount: number;
  episodeIndex: number;
  onSelect: (episodeId: string) => void;
};

/**
 * EpisodeListItem 1件を、表示用に整形した field で組み立てる。
 *
 * @require episode は EpisodeListItem 1件
 * @ensure date はスラッシュ形式・durationSec は mm:ss・通し番号付き title と topics 行（" / " 区切り）を描画する。episodeId は描画しない。
 *   クリックすると onSelect(episode.episodeId) を呼ぶ
 * @invariant 表示整形は utils に委譲する。Feature 内に変換ロジックを持たない。
 *   再生ボタン風の ▶︎ は装飾のみ（機能は onSelect のまま）
 */
export function EpisodeListItem({
  episode,
  episodeCount,
  episodeIndex,
  onSelect,
}: EpisodeListItemProps): ReactElement {
  const topicTitles = formatTopicTitles(episode.topics);
  const numberedTitle = formatNumberedEpisodeTitle(episodeCount, episodeIndex, episode.title);

  return (
    <article className="episode-list-item">
      <button
        type="button"
        className="episode-list-item__hit"
        onClick={(event) => {
          onSelect(episode.episodeId);
          // why: click 後の :focus が残り、別 item の :hover と二重に紫線が付くのを防ぐ
          event.currentTarget.blur();
        }}
      >
        <span className="episode-list-item__date">
          <LabeledText tag="span" datasetKey="episodeDate" text={formatEpisodeDate(episode.date)} />
        </span>
        <span className="episode-list-item__title">
          <LabeledText tag="span" datasetKey="episodeTitle" text={numberedTitle} />
        </span>
        {topicTitles !== "" && (
          <span className="episode-list-item__topics">
            <LabeledText tag="span" datasetKey="episodeTopics" text={topicTitles} />
          </span>
        )}
        <span className="episode-list-item__play" aria-hidden="true">
          <span className="episode-list-item__play-icon">▶︎</span>
          <LabeledText
            tag="span"
            datasetKey="episodeDurationSec"
            text={formatDurationMmSs(episode.durationSec)}
          />
        </span>
      </button>
    </article>
  );
}
