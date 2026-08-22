import type { EpisodeDetailState } from "../view-models/episode-detail-view-model.ts";

/**
 * 詳細 state から、episode の title だけを描画する要素を組み立てる。
 *
 * @require state は ViewModel が持つ EpisodeDetailState
 * @ensure success 時のみ title 要素を描画する。loading / error は title を描画しない
 * @invariant title 以外の body 内容（opening・topics・closing 等）を描画しない
 */
export function createEpisodeDetail(state: EpisodeDetailState): HTMLElement {
  const container = document.createElement("div");

  if (state.status === "success") {
    const title = document.createElement("h1");
    title.dataset.episodeTitle = "";
    title.textContent = state.episode.title;
    container.appendChild(title);
  }

  return container;
}
