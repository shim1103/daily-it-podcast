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
 * 一覧 page。`useEpisodeListPage` を呼び、`pageStatus` で全画面を 3 分岐し、`rows` を map して
 * `EpisodeItem` を並べ、常に `AudioControls` を配置するだけ。logic も副作用も持たない。
 *
 * @require apiClient は `listEpisodes()` を持つ。baseUrl は audio 直結先の origin 相当で、
 *   そのまま `useEpisodeListPage` へ渡す（URL 組み立ては hook の責務）
 * @ensure `pageStatus.kind` が loading なら loading marker、unavailable なら全画面 Error UI、
 *   ready なら本体を描画する。本体は `rows` を map し、各 row を `EpisodeItem` 1 つへ渡す
 *   （行本体と選択中の原稿は `EpisodeItem` が束ねる）。`AudioControls` は再生中かどうかに
 *   関わらず常に描画し、`audioElementRef` と `nowPlaying`（再生中 episode の見出し。hook が投影）を
 *   渡す。音源 URL の指定は `useEpisodePlayback` が ref 経由で命令的に行うため、page は `src` を
 *   組み立てない
 * @invariant ここに表示ロジック・API 呼び出しの詳細・副作用・URL 組み立てを書かない。
 *   state machine と hash ↔ selection の同期、起動、deep-link 復元、audioRef→URL 解決は
 *   `useEpisodeListPage` とその下位 hook が持つ
 */
export function EpisodeListPage({ apiClient, baseUrl }: EpisodeListPageProps): ReactElement {
  const { rows, nowPlaying, pageStatus, toggleSelection, play, seek, stop, audioElementRef } =
    useEpisodeListPage(apiClient, baseUrl);

  if (pageStatus.kind === "loading") {
    return <p data-page-loading>読み込み中</p>;
  }

  if (pageStatus.kind === "unavailable") {
    return <div data-page-error>一覧を表示できません</div>;
  }

  return (
    <div className="episode-list">
      {rows.map((row, episodeIndex) => (
        <EpisodeItem
          key={row.episodeId}
          episode={row.episode}
          episodeCount={rows.length}
          episodeIndex={episodeIndex}
          isSelected={row.isSelected}
          isActivePlayback={row.isActivePlayback}
          isPlaying={row.isPlaying}
          onSelect={toggleSelection}
          onPlay={play}
          onStop={stop}
          onSeek={(startSec) => seek(row.episodeId, startSec)}
        />
      ))}
      <AudioControls audioRef={audioElementRef} nowPlaying={nowPlaying} />
    </div>
  );
}
