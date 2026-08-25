import type { EpisodeListItemData } from "../view-models/episode-list-view-model.ts";
import { createLabeledText } from "./labeled-text.ts";

/**
 * EpisodeListItem 1件を、field をそのまま描画する要素として組み立てる（Contract Freeze）。
 *
 * @require episode は EpisodeListItem 1件
 * @ensure episodeId・date・title・durationSec をそのまま描画する。クリックすると onSelect(episode.episodeId) を呼ぶ
 * @invariant 加工・変換・分岐を持たない
 */
export function createEpisodeListItem(
  episode: EpisodeListItemData,
  onSelect: (episodeId: string) => void,
): HTMLElement {
  const item = document.createElement("article");
  item.addEventListener("click", () => onSelect(episode.episodeId));

  item.append(
    createLabeledText({ tag: "span", datasetKey: "episodeId", text: episode.episodeId }),
    createLabeledText({ tag: "span", datasetKey: "episodeDate", text: episode.date }),
    createLabeledText({ tag: "span", datasetKey: "episodeTitle", text: episode.title }),
    createLabeledText({
      tag: "span",
      datasetKey: "episodeDurationSec",
      text: String(episode.durationSec),
    }),
  );
  return item;
}
