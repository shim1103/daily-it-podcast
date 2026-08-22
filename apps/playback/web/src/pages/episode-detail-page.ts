import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { createEpisodeDetail } from "../components/episode-detail.ts";
import { createEpisodeDetailViewModel } from "../view-models/episode-detail-view-model.ts";

/**
 * 詳細 page。ViewModel と Feature Component を組み立てるだけ。
 *
 * @require apiClient は `getEpisode(episodeId)` を持つ。episodeId は route から確定済み
 * @ensure ViewModel の state 変化のたび、詳細 Feature Component を再描画する。mount 時に load(episodeId) を開始する
 * @invariant ここに表示ロジック・API 呼び出しの詳細を書かない
 */
export function createEpisodeDetailPage(
  apiClient: PlaybackApiClient,
  episodeId: string,
): HTMLElement {
  const container = document.createElement("div");
  const viewModel = createEpisodeDetailViewModel(apiClient);

  function render(): void {
    container.replaceChildren(createEpisodeDetail(viewModel.getState()));
  }

  viewModel.subscribe(render);
  render();
  void viewModel.load(episodeId);

  return container;
}
