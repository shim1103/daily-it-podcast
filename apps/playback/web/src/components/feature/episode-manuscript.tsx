import type { ReactElement } from "react";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";
import { EpisodeTopic } from "./episode-topic.tsx";
import "./episode-manuscript.css";

type Body = EpisodeData["body"];

export type EpisodeManuscriptProps = {
  body: Body;
  onSeek: (startSec: number) => void;
};

/**
 * EpisodeItem の body（opening・topics・ending）を組み合わせて描画する。
 *
 * @require body は EpisodeItem["body"]、onSeek は audio の currentTime を変える handler
 * @ensure opening を `opening.startSec` の seek bar 付き、ending を `ending.startSec` の seek bar 付きで描き、
 *   間に topics[] を EpisodeTopic として順番通りに並べる。見出しの文言（「導入」「まとめ」）は
 *   持たず seek bar だけを見出しに置く。opening / ending 本文の加工・変換はしない
 * @invariant 導入は contract の `opening.startSec`、まとめは contract の `ending.startSec` を使う
 */
export function EpisodeManuscript({ body, onSeek }: EpisodeManuscriptProps): ReactElement {
  return (
    <div className="episode-manuscript">
      <section className="episode-topic episode-topic--bookend">
        <div className="episode-topic__heading">
          <button
            type="button"
            className="episode-topic__seek"
            data-manuscript-opening-start-sec=""
            onClick={() => {
              onSeek(body.opening.startSec);
            }}
          >
            {formatDurationMmSs(body.opening.startSec)}
          </button>
        </div>
        <LabeledText tag="p" datasetKey="manuscriptOpening" text={body.opening.text} />
      </section>
      {body.topics.map((topic, index) => (
        // biome-ignore lint/suspicious/noArrayIndexKey: topic は domain 上の一意 key を持たず並び順が固定のため index を key に使う
        <EpisodeTopic key={index} topic={topic} topicIndex={index} onSeek={onSeek} />
      ))}
      <section className="episode-topic episode-topic--bookend">
        <div className="episode-topic__heading">
          <button
            type="button"
            className="episode-topic__seek"
            data-manuscript-ending-start-sec=""
            onClick={() => {
              onSeek(body.ending.startSec);
            }}
          >
            {formatDurationMmSs(body.ending.startSec)}
          </button>
        </div>
        <LabeledText tag="p" datasetKey="manuscriptEnding" text={body.ending.text} />
      </section>
    </div>
  );
}
