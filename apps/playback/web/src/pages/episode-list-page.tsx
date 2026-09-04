import type { ReactElement } from "react";
import type { PlaybackApiClient } from "../api/playback-api-client.ts";
import { AudioControls } from "../components/feature/audio-controls.tsx";
import { EpisodeManuscript } from "../components/feature/episode-manuscript.tsx";
import { EpisodeRow } from "../components/feature/episode-row.tsx";
import { buildRequestUrl } from "../utils/build-request-url.ts";
import { useEpisodeListPage } from "../view-models/use-episode-list-page.ts";

export type EpisodeListPageProps = {
  apiClient: PlaybackApiClient;
  baseUrl: string;
};

/**
 * 一覧 page。`useEpisodeListPage` を呼び、`pageStatus` で全画面を 3 分岐し、`rows` を map して
 * Row + 条件付き原稿 + AudioControls を配置するだけ。副作用は持たない（一覧取得の起動は
 * `useEpisodeCatalog` の auto-load、hash 同期と deep-link 復元は `useHashSelectionSync`）。
 *
 * @require apiClient は `listEpisodes()` を持つ。baseUrl は audio 直結先の origin 相当
 * @ensure `pageStatus.kind` が loading なら loading marker、unavailable なら全画面 Error UI、
 *   ready なら本体を描画する。本体は `rows` を map し、各 row の `episode` 実体をそのまま
 *   `EpisodeRow` へ渡す。`selectedEpisode?.episodeId === row.episodeId` の row 直後にのみ、
 *   選択中 episode の原稿（`EpisodeManuscript`）を配置する。AudioControls は
 *   `playback.kind === "active"` のとき selection と独立に描画し、
 *   `audioSrc` は `buildRequestUrl(baseUrl, playback.audioRef)` で page が組む
 * @invariant ここに表示ロジック・API 呼び出しの詳細・副作用を書かない。state machine と
 *   hash ↔ selection の同期、起動、deep-link 復元は `useEpisodeListPage` とその下位 hook が持つ
 */
export function EpisodeListPage({ apiClient, baseUrl }: EpisodeListPageProps): ReactElement {
  const {
    selectedEpisode,
    playback,
    rows,
    pageStatus,
    toggleSelection,
    play,
    seek,
    stop,
    audioElementRef,
  } = useEpisodeListPage(apiClient);

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
            isPlaying={row.isPlaying}
            onSelect={toggleSelection}
            onPlay={play}
            onStop={stop}
          />
          {selectedEpisode?.episodeId === row.episodeId && (
            <EpisodeManuscript
              body={selectedEpisode.body}
              onSeek={(startSec) => seek(selectedEpisode.episodeId, startSec)}
            />
          )}
        </div>
      ))}
      {playback.kind === "active" && (
        <AudioControls
          audioRef={audioElementRef}
          audioSrc={buildRequestUrl(baseUrl, playback.audioRef)}
        />
      )}
    </div>
  );
}
