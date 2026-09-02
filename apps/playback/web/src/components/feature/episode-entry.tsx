import type { ReactElement } from "react";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeManuscript } from "./episode-manuscript.tsx";

export type EpisodeEntryProps = {
  body: EpisodeData["body"];
  onSeek: (startSec: number) => void;
};

/**
 * 選択中 episode の manuscript のみを描画する（契約 stub）。
 *
 * @require title / date は Row が既に示すため描画しない
 * @ensure body を EpisodeManuscript へ委譲する
 */
export function EpisodeEntry({ body, onSeek }: EpisodeEntryProps): ReactElement {
  return (
    <section className="episode-entry">
      <EpisodeManuscript body={body} onSeek={onSeek} />
    </section>
  );
}
