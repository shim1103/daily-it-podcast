import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { getLocationHash, onLocationHashChange, setLocationHash } from "../lib/location-hash.ts";
import { mountEpisodeList } from "./mount-episode-list.ts";
import { mountEpisodeListViewModel } from "./mount-episode-list-view-model.ts";

/**
 * 一覧 page。ViewModel と Feature Component、hash 同期用 Driven Adapter を組み立てるだけ。
 *
 * @require apiClient は `listEpisodes()` と `getEpisode(episodeId)` を持つ。baseUrl は audio 直結先の origin相当
 * @ensure ViewModel の state 変化のたび、一覧 Feature Component を再描画し、location.hash を
 *   selectedEpisodeId へ同期する。mount 時に load() を開始し、hash に episodeId があればその episode を選択する。
 *   hashchange 時は対応する episodeId を選択する。episode 選択は ViewModel の select() を呼ぶ
 * @invariant ここに表示ロジック・API 呼び出しの詳細を書かない
 */
export function createEpisodeListPage(apiClient: PlaybackApiClient, baseUrl: string): HTMLElement {
  const container = document.createElement("div");
  // why: ViewModel は hook 化済みだが Page は未 JSX。一時橋で getState/subscribe 面を保つ
  //   （削除予定: playback-web-page-jsx-mount）
  const viewModel = mountEpisodeListViewModel(apiClient);
  // why: EpisodeList は JSX 化済みだが Page は未 JSX。mount-episode-list.ts の一時橋で
  //   HTMLElement 面を保つ（削除予定: playback-web-page-jsx-mount）。root-level イベント委譲を
  //   保つため、element は container へ1回だけ挿入し、以後は update() で再描画する
  const episodeList = mountEpisodeList();
  container.appendChild(episodeList.element);

  // why: hashchange listener 由来の select() が setState を発火させ、そのまま同じ値を
  //   setLocationHash へ書き戻すと無限ループになりうる。書き込み直前の値と比較して抑止する
  let lastSyncedHash = getLocationHash();

  function render(): void {
    episodeList.update(viewModel.getState(), baseUrl, (episodeId) => {
      void viewModel.select(episodeId);
    });

    const state = viewModel.getState();
    const nextHash = state.status === "success" ? (state.selectedEpisodeId ?? "") : lastSyncedHash;
    if (nextHash !== lastSyncedHash) {
      lastSyncedHash = nextHash;
      setLocationHash(nextHash);
    }
  }

  viewModel.subscribe(render);
  render();

  onLocationHashChange(() => {
    const hash = getLocationHash();
    if (hash === lastSyncedHash) {
      return;
    }
    lastSyncedHash = hash;

    const state = viewModel.getState();
    if (state.status !== "success") {
      return;
    }

    // why: hash が空になった場合は、選択中 episode を select() で選択解除させる
    //   （select() は同じ episodeId を渡すと選択解除する契約）
    if (hash !== "") {
      void viewModel.select(hash);
    } else if (state.selectedEpisodeId !== null) {
      void viewModel.select(state.selectedEpisodeId);
    }
  });

  void viewModel.load().then(() => {
    const initialHash = getLocationHash();
    if (initialHash !== "") {
      void viewModel.select(initialHash);
    }
  });

  return container;
}
