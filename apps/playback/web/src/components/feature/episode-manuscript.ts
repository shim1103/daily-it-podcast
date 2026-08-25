import type { EpisodeData } from "../../view-models/episode-list-view-model.ts";
import { createLabeledText } from "../primitive/labeled-text.ts";
import { createEpisodeTopic } from "./episode-topic.ts";

type Body = EpisodeData["body"];

/**
 * GetEpisodeResponse の body（opening・topics・closing）を組み合わせて描画する要素を組み立てる。
 *
 * @require body は GetEpisodeResponse["body"]
 * @ensure opening・closing をそのまま描画し、topics[] を episode-topic として順番通りに並べる
 * @invariant opening・closing の加工・変換をしない
 */
export function createEpisodeManuscript(body: Body): HTMLElement {
  const container = document.createElement("div");

  container.appendChild(
    createLabeledText({ tag: "p", datasetKey: "manuscriptOpening", text: body.opening }),
  );

  for (const topic of body.topics) {
    container.appendChild(createEpisodeTopic(topic));
  }

  container.appendChild(
    createLabeledText({ tag: "p", datasetKey: "manuscriptClosing", text: body.closing }),
  );

  return container;
}
