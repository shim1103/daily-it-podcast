import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import { mountLabeledText } from "./mount-labeled-text.ts";

type Topic = EpisodeData["body"]["topics"][number];

/**
 * topics[] の1要素から title・preface・detail だけを、そのまま描画する要素として組み立てる（Contract Freeze）。
 *
 * @require topic は topics[] の1要素
 * @ensure title・preface・detail をそのまま描画する
 * @invariant startSec は受け取っても描画しない。加工・変換・分岐を持たない
 */
export function createEpisodeTopic(topic: Topic): HTMLElement {
  const section = document.createElement("section");
  section.append(
    mountLabeledText({ tag: "h3", datasetKey: "topicTitle", text: topic.title }),
    mountLabeledText({ tag: "p", datasetKey: "topicPreface", text: topic.preface }),
    mountLabeledText({ tag: "p", datasetKey: "topicDetail", text: topic.detail }),
  );
  return section;
}
