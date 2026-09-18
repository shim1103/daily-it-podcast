import type { ReactElement } from "react";
import "./episode-list.css";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { AudioControls } from "../components/feature/audio-controls.tsx";
import { EpisodeItem } from "../components/feature/episode-item.tsx";
import { useEpisodeListPage } from "../view-models/use-episode-list-page.ts";

export type EpisodeListPageProps = {
  apiClient: PlaybackApiClient;
  baseUrl: string;
};

/**
 * 一覧 page。`useEpisodeListPage` の出力を分岐・map するだけで、logic も副作用も持たない。
 *
 * @require apiClient は `listEpisodes()` を持つ。baseUrl は audio 直結先の origin 相当で、
 *   そのまま `useEpisodeListPage` へ渡す（URL 組み立ては hook の責務）
 * @ensure `pageStatus.kind` が loading なら loading marker、unavailable なら全画面 Error UI
 *   （`retryable` が true なら message 下に retry button も出す）、ready なら `rows` を map して
 *   `EpisodeItem` を並べる。`AudioControls` は常に描画する
 * @invariant ここに表示ロジック・API 呼び出しの詳細・副作用・URL 組み立てを書かない。
 *   state machine と hash ↔ selection の同期、起動、deep-link 復元、audioRef→URL 解決は
 *   `useEpisodeListPage` とその下位 hook が持つ
 */
export function EpisodeListPage({ apiClient, baseUrl }: EpisodeListPageProps): ReactElement {
  const {
    rows,
    nowPlaying,
    pageStatus,
    toggleSelection,
    play,
    seek,
    stop,
    retry,
    audioElementRef,
  } = useEpisodeListPage(apiClient, baseUrl);

  if (pageStatus.kind === "loading") {
    return <p data-page-loading>読み込み中</p>;
  }

  if (pageStatus.kind === "unavailable") {
    return (
      <div data-page-error>
        <p>{pageStatus.message}</p>
        {pageStatus.retryable ? (
          <button type="button" className="page-error-retry" onClick={() => void retry()}>
            リトライ
          </button>
        ) : null}
      </div>
    );
  }

  return (
    <div className="episode-list">
      {rows.map((row, episodeIndex) => (
        <EpisodeItem
          key={row.episode.episodeId}
          episode={row.episode}
          episodeCount={rows.length}
          episodeIndex={episodeIndex}
          isSelected={row.isSelected}
          isActivePlayback={row.isActivePlayback}
          isPlaying={row.isPlaying}
          onSelect={toggleSelection}
          onPlay={play}
          onStop={stop}
          onSeek={(startSec) => seek(row.episode.episodeId, startSec)}
        />
      ))}
      <AudioControls audioRef={audioElementRef} nowPlaying={nowPlaying} />
    </div>
  );
}
