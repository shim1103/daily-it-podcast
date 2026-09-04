import type { ReactElement } from "react";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import { formatNumberedTopicTitle } from "../../utils/format-numbered-topic-title.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";
import "./episode-topic.css";

type Topic = EpisodeData["body"]["topics"][number];

export type EpisodeTopicProps = {
  topic: Topic;
  topicIndex: number;
  onSeek: (startSec: number) => void;
};

/**
 * topics[] の1要素から seek 可能な見出しと preface・detail を描画する。
 *
 * @require topic は topics[] の1要素、topicIndex は 0 始まりの出現位置、onSeek は startSec へシークする handler
 * @ensure 見出しは mm:ss（seek bar）を上、通し番号付き title を下の 2 行に分けて描画する。
 *   preface・detail はそのまま描画する
 * @invariant preface・detail の加工・変換・分岐を持たない
 */
export function EpisodeTopic({ topic, topicIndex, onSeek }: EpisodeTopicProps): ReactElement {
  const numberedTitle = formatNumberedTopicTitle(topicIndex, topic.title);
  const startLabel = formatDurationMmSs(topic.startSec);

  return (
    <section className="episode-topic">
      <div className="episode-topic__heading">
        <button
          type="button"
          className="episode-topic__seek"
          data-topic-start-sec=""
          onClick={() => {
            onSeek(topic.startSec);
          }}
        >
          {startLabel}
        </button>
        <h3 className="episode-topic__heading-title" data-topic-title="">
          {numberedTitle}
        </h3>
      </div>
      <LabeledText tag="p" datasetKey="topicPreface" text={topic.preface} />
      <LabeledText tag="p" datasetKey="topicDetail" text={topic.detail} />
    </section>
  );
}
