import { type ReactElement, useCallback, useEffect, useRef } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { EpisodeList } from "../components/feature/episode-list.tsx";
import { getLocationHash } from "../lib/location-hash.ts";
import { useEpisodeListViewModel } from "../view-models/episode-list-view-model.ts";
import { useHashSync } from "../view-models/use-hash-sync.ts";

export type EpisodeListPageProps = {
  apiClient: PlaybackApiClient;
  baseUrl: string;
};

/**
 * 一覧 page。ViewModel hook・hash 同期 hook・一覧 Feature Component を組み立てるだけ。
 *
 * @require apiClient は `listEpisodes()` と `getEpisode(episodeId)` を持つ。baseUrl は audio 直結先の origin相当
 * @ensure mount 時に load() を開始し、完了後 hash に episodeId があればその episode を選択する。
 *   以後は useHashSync が選択中 episodeId と location.hash を双方向同期する。hash が空へ変わった時は
 *   選択中 episode を select() へ渡して選択解除する。一覧 Feature Component を state / baseUrl / onSelect で描画する
 * @invariant ここに表示ロジック・API 呼び出しの詳細を書かない。hash 同期の機構は useHashSync が持つ。
 *   load() 完了までは selectedId に undefined を渡し、hash→state 同期を保留させる（完了前に
 *   selectedEpisodeId=null で mount 時の hash を消す race を防ぐ）
 */
export function EpisodeListPage({ apiClient, baseUrl }: EpisodeListPageProps): ReactElement {
  const { state, load, select, audioElementRef, seek } = useEpisodeListViewModel(apiClient);

  // why: load() 完了前は hash 同期を保留する（ViewModel の load ライフサイクルとの結合であり、
  //   hash 同期の関心事ではないため useHashSync ではなく page が持つ）
  const initializedRef = useRef(false);

  // mount 時: load() を開始し、完了後 hash に episodeId があればその episode を選択する
  useEffect(() => {
    let cancelled = false;
    void load().then(() => {
      // why: in-flight 中に unmount / 再 mount された場合、完了後の hash 復元 select を打ち切る
      if (cancelled) {
        return;
      }
      initializedRef.current = true;
      const initialHash = getLocationHash();
      if (initialHash !== "") {
        void select(initialHash);
      }
    });
    return () => {
      cancelled = true;
    };
  }, [load, select]);

  const onHashSelect = useCallback(
    (id: string | null) => {
      if (id !== null) {
        void select(id);
        return;
      }
      // why: hash が空になった時は、選択中 episode を select() へ渡して選択解除させる
      //   （select() は同じ episodeId を渡すと選択解除する ViewModel 契約）
      if (state.status === "success" && state.selectedEpisodeId !== null) {
        void select(state.selectedEpisodeId);
      }
    },
    [select, state],
  );

  const selectedId =
    initializedRef.current && state.status === "success" ? state.selectedEpisodeId : undefined;
  useHashSync(selectedId, onHashSelect);

  const onSelect = useCallback(
    (episodeId: string) => {
      void select(episodeId);
    },
    [select],
  );

  return (
    <EpisodeList
      state={state}
      baseUrl={baseUrl}
      onSelect={onSelect}
      audioElementRef={audioElementRef}
      onSeek={seek}
    />
  );
}
