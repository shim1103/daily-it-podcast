import type { ReactElement } from "react";
import "./episode-row.css";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import { formatEpisodeDate } from "../../utils/format-episode-date.ts";
import { formatNumberedEpisodeTitle } from "../../utils/format-numbered-episode-title.ts";
import { formatTopicTitles } from "../../utils/format-topic-titles.ts";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";

// why: 再生 Icon は emoji 化しない幾何記号 ▶（U+25B6）。停止 Icon は環境で emoji 化する ⏸（U+23F8）や
//   幅の揺れる ▮▮ を避け、CSS 描画の 2 本バー（`.episode-row__play-bar` ×2）で ▶ と同じ視覚サイズに
//   揃える。button は固定サイズの丸で、背景色は `data-active` で切り替える（非再生=灰 / 再生=白）
const PLAY_GLYPH = "▶";

export type EpisodeRowProps = {
  episode: EpisodeData;
  episodeCount: number;
  episodeIndex: number;
  isSelected: boolean;
  isActivePlayback: boolean;
  isPlaying: boolean;
  onSelect: (episodeId: string) => void;
  onPlay: (episodeId: string) => void;
  onStop: () => void;
};

/**
 * 1 episode の行本体。行全体が select hit（日付 / 通し番号付き title / topic 一覧）で、
 * 「再生 / 停止」だけは title・topics の横、右端に独立した兄弟 button として置き、再生を優先する。
 *
 * @require isSelected / isActivePlayback / isPlaying は caller が derive して渡す
 * @ensure select button と play button を `<div>` 直下の兄弟に置く（button の入れ子を作らない）。
 *   select button クリックで onSelect。play button は isActivePlayback（phase 不問。loading 中も
 *   含む）なら「停止」表示で onStop、そうでなければ「再生」表示で onPlay を呼ぶ。
 *   isPlaying（音が出ている phase）は `data-playing` の視覚強調にだけ使う。
 *   表示整形は utils へ委譲し、Row 自身は derive しない
 * @invariant 見た目は CSS 側の責務。ここに色・寸法を書かない。選択中の横線・行間の余白は
 *   親（EpisodeItem）の責務であり、Row は持たない
 */
export function EpisodeRow({
  episode,
  episodeCount,
  episodeIndex,
  isSelected,
  isActivePlayback,
  isPlaying,
  onSelect,
  onPlay,
  onStop,
}: EpisodeRowProps): ReactElement {
  const topicTitles = formatTopicTitles(episode.body.topics);
  const numberedTitle = formatNumberedEpisodeTitle(episodeCount, episodeIndex, episode.title);

  return (
    <div
      className="episode-row"
      data-selected={isSelected ? "true" : "false"}
      data-playing={isPlaying ? "true" : "false"}
    >
      <button
        type="button"
        className="episode-row__select"
        onClick={(event) => {
          onSelect(episode.episodeId);
          // why: click 後に残る :focus が別 row の :hover と二重に紫線を出すのを防ぐ
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
        <span className="episode-row__duration" aria-hidden="true">
          <LabeledText
            tag="span"
            datasetKey="episodeDurationSec"
            text={formatDurationMmSs(episode.durationSec)}
          />
        </span>
      </button>
      <button
        type="button"
        className="episode-row__play"
        data-active={isActivePlayback ? "true" : "false"}
        aria-label={isActivePlayback ? "停止" : "再生"}
        onClick={() => (isActivePlayback ? onStop() : onPlay(episode.episodeId))}
      >
        {/* 見た目は Icon（▶ 再生 / CSS 2 本バー 停止）。文言は aria-label が担う */}
        {isActivePlayback ? (
          <span className="episode-row__play-icon" aria-hidden="true">
            <span className="episode-row__play-bar" />
            <span className="episode-row__play-bar" />
          </span>
        ) : (
          <span className="episode-row__play-glyph" aria-hidden="true">
            {PLAY_GLYPH}
          </span>
        )}
      </button>
    </div>
  );
}
