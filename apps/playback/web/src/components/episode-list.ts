import type { EpisodeListState } from "../view-models/episode-list-view-model.ts";

/**
 * 一覧 state から、episode 毎の title（詳細 page への hash link）だけを描画する要素を組み立てる。
 *
 * @require state は ViewModel が持つ EpisodeListState
 * @ensure success 時のみ episodes の順番通りに title 要素を並べる。各要素の href は `#/episodes/{episodeId}`。
 *   loading / error は title を描画しない
 * @invariant title 以外の body 内容（date・durationSec 等）を描画しない
 */
export function createEpisodeList(state: EpisodeListState): HTMLElement {
  const container = document.createElement("div");

  if (state.status === "success") {
    for (const episode of state.episodes) {
      const title = document.createElement("a");
      title.dataset.episodeTitle = "";
      title.href = `#/episodes/${encodeURIComponent(episode.episodeId)}`;
      title.textContent = episode.title;
      container.appendChild(title);
    }
  }

  return container;
}
