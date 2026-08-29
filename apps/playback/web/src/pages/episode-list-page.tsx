import { type ReactElement, useCallback, useEffect, useRef } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { EpisodeList } from "../components/feature/episode-list.tsx";
import { getLocationHash } from "../lib/location-hash.ts";
import { useEpisodeListViewModel } from "../view-models/episode-list-view-model.ts";
import { useEpisodePlayback } from "../view-models/use-episode-playback.ts";
import { useHashSync } from "../view-models/use-hash-sync.ts";
import { EpisodeListError } from "./episode-list-error.tsx";
import { EpisodeListLoading } from "./episode-list-loading.tsx";

export type EpisodeListPageProps = {
  apiClient: PlaybackApiClient;
  baseUrl: string;
};

/**
 * 一覧 page。ViewModel hook・hash 同期 hook・一覧 Feature を組み立てるだけ。
 *
 * @require apiClient は `listEpisodes()` と `getEpisode(episodeId)` を持つ。baseUrl は audio 直結先の origin相当
 * @ensure mount 時に load() を開始し、完了後 hash に episodeId があればその episode を選択する。
 *   loading / error は page 層で分岐し、success 時のみ EpisodeList を描画する
 * @invariant ここに一覧 item の表示ロジックを書かない。再生 UI（pill / seek / audio）は List / Entry へ委譲する
 */
export function EpisodeListPage({ apiClient, baseUrl }: EpisodeListPageProps): ReactElement {
  const { state, load, select } = useEpisodeListViewModel(apiClient);
  const { audioElementRef, playback, resolvedSrc, play, seek } = useEpisodePlayback(baseUrl);
  const initializedRef = useRef(false);

  useEffect(() => {
    let cancelled = false;
    void load().then(() => {
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
      if (state.status === "success" && state.selection.kind === "open") {
        void select(state.selection.episodeId);
      }
    },
    [select, state],
  );

  const selectedId =
    initializedRef.current && state.status === "success"
      ? state.selection.kind === "open"
        ? state.selection.episodeId
        : null
      : undefined;
  useHashSync(selectedId, onHashSelect);

  const onSelect = useCallback(
    (episodeId: string) => {
      void select(episodeId);
    },
    [select],
  );

  const onPlay = useCallback(
    (episodeId: string, audioRef: string) => {
      play(episodeId, audioRef);
    },
    [play],
  );

  if (state.status === "loading") {
    return <EpisodeListLoading />;
  }

  if (state.status === "error") {
    return <EpisodeListError />;
  }

  return (
    <EpisodeList
      episodes={state.episodes}
      selection={state.selection}
      playback={playback}
      onSelect={onSelect}
      onPlay={onPlay}
      onSeek={seek}
      audioElementRef={audioElementRef}
      resolvedSrc={resolvedSrc}
    />
  );
}
