import type { ReactElement } from "react";
import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";

type Topic = EpisodeData["body"]["topics"][number];

export type EpisodeTopicProps = {
  topic: Topic;
};

/**
 * topics[] の1要素から title・preface・detail だけを、そのまま描画する（Contract Freeze）。
 *
 * @require topic は topics[] の1要素
 * @ensure title・preface・detail をそのまま描画する
 * @invariant startSec は受け取っても描画しない。加工・変換・分岐を持たない
 */
export function EpisodeTopic({ topic }: EpisodeTopicProps): ReactElement {
  return (
    <section>
      <LabeledText tag="h3" datasetKey="topicTitle" text={topic.title} />
      <LabeledText tag="p" datasetKey="topicPreface" text={topic.preface} />
      <LabeledText tag="p" datasetKey="topicDetail" text={topic.detail} />
    </section>
  );
}
