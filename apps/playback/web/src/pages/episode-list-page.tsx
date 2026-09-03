import { type ReactElement, useEffect, useRef } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { AudioControls } from "../components/feature/audio-controls.tsx";
import { EpisodeEntry } from "../components/feature/episode-entry.tsx";
import { EpisodeRow } from "../components/feature/episode-row.tsx";
import { getLocationHash } from "../lib/location-hash.ts";
import { buildRequestUrl } from "../utils/build-request-url.ts";
import { useEpisodeListPage } from "../view-models/use-episode-list-page.ts";

export type EpisodeListPageProps = {
  apiClient: PlaybackApiClient;
  baseUrl: string;
};

/**
 * 一覧 page。`useEpisodeListPage` を組み立て、`pageStatus` で全画面を分岐し、
 * Row + 条件付き Entry + AudioControls を並べるだけ。
 *
 * @require apiClient は `listEpisodes()` を持つ。baseUrl は audio 直結先の origin相当
 * @ensure mount 時に load() を開始する。`pageStatus.kind` が loading なら loading marker、
 *   unavailable なら全画面 Error UI、ready なら本体を描画する。Row は `rows`（各要素が episode 実体を
 *   持つ）を map し、Entry は選択中 episode の直後、AudioControls は再生対象 episode があれば selection と
 *   独立に描画する。catalog が初めて ready になった時点で 1 回だけ、mount 時点の location.hash に
 *   episodeId があれば select() して deep-link を復元する（select 失敗時の retry 経路は現状 scope 外）
 * @invariant ここに表示ロジック・API 呼び出しの詳細を書かない。row と episode の結合は
 *   `deriveEpisodeRows` が持ち、page は別配列の index 引きをしない。hash ↔ selection の継続同期は
 *   useEpisodeListPage が持つ。deep-link 復元は初回のみで、以後は hook 側の同期に委ねる
 */
export function EpisodeListPage({ apiClient, baseUrl }: EpisodeListPageProps): ReactElement {
  const {
    selectedEpisode,
    playedEpisode,
    rows,
    pageStatus,
    load,
    select,
    toggleSelection,
    play,
    stop,
    seek,
    audioElementRef,
  } = useEpisodeListPage(apiClient);

  useEffect(() => {
    void load();
  }, [load]);

  // why: hook の hash 同期は echo 抑止で初期 hash を流さないため、deep-link で開いた時の初回展開だけ
  //   page が担う。mount 時点の hash を控え、catalog が初めて ready になった時に 1 回だけ select する
  const initialHashRef = useRef<string | null>(null);
  if (initialHashRef.current === null) {
    initialHashRef.current = getLocationHash();
  }
  const deepLinkRestoredRef = useRef(false);
  useEffect(() => {
    if (pageStatus.kind !== "ready" || deepLinkRestoredRef.current) {
      return;
    }
    deepLinkRestoredRef.current = true;
    const initialEpisodeId = initialHashRef.current;
    if (initialEpisodeId !== null && initialEpisodeId !== "") {
      select(initialEpisodeId);
    }
  }, [pageStatus.kind, select]);

  if (pageStatus.kind === "loading") {
    return <p data-page-loading>読み込み中</p>;
  }

  if (pageStatus.kind === "unavailable") {
    return <div data-page-error>一覧を表示できません</div>;
  }

  return (
    <div className="episode-list">
      {rows.map((row, episodeIndex) => (
        <div key={row.episodeId} className="episode-list__row">
          <EpisodeRow
            episode={row.episode}
            episodeCount={rows.length}
            episodeIndex={episodeIndex}
            isSelected={row.isSelected}
            isPlayed={row.isPlayed}
            isPlaying={row.isPlaying}
            onSelect={toggleSelection}
            onPlay={play}
            onStop={stop}
          />
          {row.isSelected && selectedEpisode !== null && (
            <EpisodeEntry
              body={selectedEpisode.body}
              onSeek={(startSec) => seek(selectedEpisode.episodeId, startSec)}
            />
          )}
        </div>
      ))}
      {playedEpisode !== null && (
        <AudioControls
          audioRef={audioElementRef}
          audioSrc={buildRequestUrl(baseUrl, playedEpisode.audioRef)}
        />
      )}
    </div>
  );
}
