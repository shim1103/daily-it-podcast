import type { ReactElement } from "react";
import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";
import { EpisodeTopic } from "./episode-topic.tsx";

type Body = EpisodeData["body"];

export type EpisodeManuscriptProps = {
  body: Body;
};

/**
 * GetEpisodeResponse の body（opening・topics・closing）を組み合わせて描画する。
 *
 * @require body は GetEpisodeResponse["body"]
 * @ensure opening・closing をそのまま描画し、topics[] を EpisodeTopic として順番通りに並べる
 * @invariant opening・closing の加工・変換をしない
 */
export function EpisodeManuscript({ body }: EpisodeManuscriptProps): ReactElement {
  return (
    <div>
      <LabeledText tag="p" datasetKey="manuscriptOpening" text={body.opening} />
      {body.topics.map((topic, index) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: topic は domain 上の一意 key を持たず並び順が固定のため index を key に使う
        <EpisodeTopic key={index} topic={topic} />
      ))}
      <LabeledText tag="p" datasetKey="manuscriptClosing" text={body.closing} />
    </div>
  );
}
