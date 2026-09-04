import type { ReactElement } from "react";
import { formatDurationMmSs } from "../../utils/format-duration-mm-ss.ts";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";
import { EpisodeTopic } from "./episode-topic.tsx";
import "./episode-manuscript.css";

type Body = EpisodeData["body"];

export type EpisodeManuscriptProps = {
  body: Body;
  durationSec: number;
  onSeek: (startSec: number) => void;
};

/**
 * EpisodeItem の body（opening・topics・ending）を組み合わせて描画する。
 *
 * @require body は EpisodeItem["body"]、durationSec は総尺、onSeek は audio の currentTime を変える handler
 * @ensure opening を「導入」（`00:00` の seek bar 付き）、ending を「まとめ」（総尺 `durationSec` の
 *   seek bar 付き）として描き、間に topics[] を EpisodeTopic として順番通りに並べる。
 *   opening / ending 本文の加工・変換はしない
 * @invariant sec の正は contract に無いため、導入は 0、まとめは総尺（`durationSec`）を使う
 */
export function EpisodeManuscript({
  body,
  durationSec,
  onSeek,
}: EpisodeManuscriptProps): ReactElement {
  return (
    <div className="episode-manuscript">
      <section className="episode-topic episode-topic--bookend">
        <div className="episode-topic__heading">
          <button
            type="button"
            className="episode-topic__seek"
            data-manuscript-opening-start-sec=""
            onClick={() => {
              onSeek(0);
            }}
          >
            {formatDurationMmSs(0)}
          </button>
          <h3 className="episode-topic__heading-title">導入</h3>
        </div>
        <LabeledText tag="p" datasetKey="manuscriptOpening" text={body.opening} />
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
              onSeek(durationSec);
            }}
          >
            {formatDurationMmSs(durationSec)}
          </button>
          <h3 className="episode-topic__heading-title">まとめ</h3>
        </div>
        <LabeledText tag="p" datasetKey="manuscriptEnding" text={body.ending} />
      </section>
    </div>
  );
}
