import type { ReactElement } from "react";
import type { EpisodeListItemData } from "../../view-models/episode-list-view-model.ts";
import { LabeledText } from "../primitive/labeled-text.tsx";

export type EpisodeListItemProps = {
  episode: EpisodeListItemData;
  onSelect: (episodeId: string) => void;
};

/**
 * EpisodeListItem 1件を、field をそのまま描画する要素として組み立てる（Contract Freeze）。
 *
 * @require episode は EpisodeListItem 1件
 * @ensure episodeId・date・title・durationSec をそのまま描画する。クリックすると onSelect(episode.episodeId) を呼ぶ
 * @invariant 加工・変換・分岐を持たない
 */
export function EpisodeListItem({ episode, onSelect }: EpisodeListItemProps): ReactElement {
  return (
    <article>
      <button type="button" onClick={() => onSelect(episode.episodeId)}>
        <LabeledText tag="span" datasetKey="episodeId" text={episode.episodeId} />
        <LabeledText tag="span" datasetKey="episodeDate" text={episode.date} />
        <LabeledText tag="span" datasetKey="episodeTitle" text={episode.title} />
        <LabeledText
          tag="span"
          datasetKey="episodeDurationSec"
          text={String(episode.durationSec)}
        />
      </button>
    </article>
  );
}
