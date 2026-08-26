import { createElement } from "react";
import { flushSync } from "react-dom";
import { createRoot } from "react-dom/client";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import {
  useEpisodeListViewModel,
  type EpisodeListState,
  type EpisodeListViewModel,
} from "../view-models/episode-list-view-model.ts";

/**
 * 未 JSX の page 向け ViewModel 橋の公開面（getState / subscribe）。
 * `playback-web-page-jsx-mount` で hook を直接使う時に削除する。
 */
export type EpisodeListViewModelHandle = {
  getState(): EpisodeListState;
  subscribe(listener: (state: EpisodeListState) => void): () => void;
  load(): Promise<void>;
  select(episodeId: string): Promise<void>;
};

type HostProps = {
  apiClient: PlaybackApiClient;
  methodsRef: { current: Pick<EpisodeListViewModel, "load" | "select"> };
  stateRef: { current: EpisodeListState };
};

function EpisodeListViewModelHost({ apiClient, methodsRef, stateRef }: HostProps) {
  const vm = useEpisodeListViewModel(apiClient);
  methodsRef.current = { load: vm.load, select: vm.select };
  stateRef.current = vm.state;
  return null;
}

/**
 * useEpisodeListViewModel を HTMLElement 組み立て page 向けの getState/subscribe 面へ橋渡しする。
 *
 * @require apiClient は `listEpisodes()` と `getEpisode(episodeId)` を持つ
 * @ensure 戻り値は旧 createEpisodeListViewModel と同型の getState/subscribe/load/select を持つ
 * @invariant Page の本格 JSX 化はしない。橋渡しのみ。`playback-web-page-jsx-mount` で削除する
 */
export function mountEpisodeListViewModel(
  apiClient: PlaybackApiClient,
): EpisodeListViewModelHandle {
  const listeners = new Set<(state: EpisodeListState) => void>();
  const stateRef: { current: EpisodeListState } = { current: { status: "loading" } };
  const methodsRef: { current: Pick<EpisodeListViewModel, "load" | "select"> } = {
    current: {
      load: async () => {},
      select: async () => {},
    },
  };

  const mountNode = document.createElement("div");
  const root = createRoot(mountNode);

  function renderHost(): void {
    root.render(
      createElement(EpisodeListViewModelHost, {
        apiClient,
        methodsRef,
        stateRef,
      }),
    );
  }

  function notifyListeners(): void {
    for (const listener of listeners) {
      listener(stateRef.current);
    }
  }

  async function runAndFlush(task: () => Promise<void>): Promise<void> {
    await task();
    // why: hook は通常 setState。一時橋の getState/subscribe 同期観測のため、await 後に Host を flush する
    flushSync(() => {
      renderHost();
    });
    notifyListeners();
  }

  const handle: EpisodeListViewModelHandle = {
    getState(): EpisodeListState {
      return stateRef.current;
    },
    subscribe(listener: (state: EpisodeListState) => void): () => void {
      listeners.add(listener);
      return () => {
        listeners.delete(listener);
      };
    },
    load(): Promise<void> {
      return runAndFlush(() => methodsRef.current.load());
    },
    select(episodeId: string): Promise<void> {
      return runAndFlush(() => methodsRef.current.select(episodeId));
    },
  };

  // why: React 18 の createRoot.render は非同期。返り値直後に load/select を叩けるよう同期 commit する
  flushSync(() => {
    renderHost();
  });

  return handle;
}
