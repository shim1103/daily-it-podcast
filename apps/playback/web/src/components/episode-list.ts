import { createEpisodeListItem } from "./episode-list-item.ts";
import { createEpisodeManuscript } from "./episode-manuscript.ts";
import { createEpisodePlayer } from "./episode-player.ts";
import type { EpisodeListState } from "../view-models/episode-list-view-model.ts";

/**
 * 選択中 episode の詳細 state から、manuscript・player を組み合わせた要素を組み立てる。
 * title・date は episode-list-item が既に描画しているため、ここでは重ねて描画しない。
 * loading・error は専用の表示を返す。
 */
function createSelectedEpisodeDetail(
  selectedEpisode: NonNullable<Extract<EpisodeListState, { status: "success" }>["selectedEpisode"]>,
  baseUrl: string,
): HTMLElement {
  if (selectedEpisode.status === "loading") {
    const loading = document.createElement("div");
    loading.dataset.episodeDetailLoading = "";
    return loading;
  }

  if (selectedEpisode.status === "error") {
    const error = document.createElement("div");
    error.dataset.episodeDetailError = "";
    return error;
  }

  const detail = document.createElement("div");
  detail.append(
    createEpisodeManuscript(selectedEpisode.episode.body),
    createEpisodePlayer(baseUrl, selectedEpisode.episode.audioRef),
  );
  return detail;
}

/**
 * 一覧 state から、episode-list-item を並べ、選択中 episode があればその直後に詳細を展開する要素を組み立てる。
 *
 * @require state は ViewModel が持つ EpisodeListState、baseUrl は playback worker の origin相当
 * @ensure success 時のみ episodes の順番通りに episode-list-item を並べる。selectedEpisodeId と一致する
 *   item の直後に、選択中 episode の詳細（manuscript + player、または loading/error 表示）を展開する。
 *   title・date は item が既に描画しているため詳細側では重ねて描画しない。loading / error は何も描画しない
 * @invariant item 以外の field 加工をしない
 */
export function createEpisodeList(
  state: EpisodeListState,
  baseUrl: string,
  onSelect: (episodeId: string) => void,
): HTMLElement {
  const container = document.createElement("div");

  if (state.status === "success") {
    for (const episode of state.episodes) {
      container.appendChild(createEpisodeListItem(episode, onSelect));

      if (state.selectedEpisodeId === episode.episodeId && state.selectedEpisode) {
        container.appendChild(createSelectedEpisodeDetail(state.selectedEpisode, baseUrl));
      }
    }
  }

  return container;
}
