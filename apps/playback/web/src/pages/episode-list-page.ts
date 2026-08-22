import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { createEpisodeList } from "../components/episode-list.ts";
import { createEpisodeListViewModel } from "../view-models/episode-list-view-model.ts";

/**
 * 一覧 page。ViewModel と Feature Component を組み立てるだけ。
 *
 * @require apiClient は `listEpisodes()` を持つ
 * @ensure ViewModel の state 変化のたび、一覧 Feature Component を再描画する。mount 時に load() を開始する
 * @invariant ここに表示ロジック・API 呼び出しの詳細を書かない
 */
export function createEpisodeListPage(apiClient: PlaybackApiClient): HTMLElement {
  const container = document.createElement("div");
  const viewModel = createEpisodeListViewModel(apiClient);

  function render(): void {
    container.replaceChildren(createEpisodeList(viewModel.getState()));
  }

  viewModel.subscribe(render);
  render();
  void viewModel.load();

  return container;
}
