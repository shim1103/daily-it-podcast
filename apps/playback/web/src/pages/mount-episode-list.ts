import { createElement } from "react";
import { flushSync } from "react-dom";
import { createRoot } from "react-dom/client";
import { EpisodeList } from "../components/feature/episode-list.tsx";
import type { EpisodeListState } from "../view-models/episode-list-view-model.ts";

export type EpisodeListMountHandle = {
  element: HTMLElement;
  update(state: EpisodeListState, baseUrl: string, onSelect: (episodeId: string) => void): void;
};

/**
 * EpisodeList を、page が持つ container 内へ1回だけ挿入し続ける root として mount する
 * （既存 DOM 組み立て page 向けの橋渡し）。
 *
 * @require なし
 * @ensure `element` は createRoot した根 HTMLElement で、page 側の DOM tree へ1回だけ挿入する。
 *   `update()` を呼ぶたび、その props で EpisodeList を同期再描画する
 * @invariant Page の本格 JSX 化はしない。橋渡しのみ。`mount-episode-list-view-model.ts` と同型の一時橋であり、
 *   `playback-web-page-jsx-mount` Issue で Page を JSX 化する時に削除する。要素を取り出して他の DOM へ
 *   移し替えない（root-level イベント委譲が切れ、onClick 等の合成イベントが機能しなくなるため）
 */
export function mountEpisodeList(): EpisodeListMountHandle {
  const container = document.createElement("div");
  const root = createRoot(container);

  return {
    element: container,
    update(state, baseUrl, onSelect): void {
      flushSync(() => {
        root.render(createElement(EpisodeList, { state, baseUrl, onSelect }));
      });
    },
  };
}
