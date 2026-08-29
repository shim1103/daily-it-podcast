import { memo, type ReactElement, type Ref } from "react";
import type {
  EpisodeListItemData,
  EpisodeSelection,
} from "../../view-models/episode-list-view-model.ts";
import type { EpisodePlayback } from "../../view-models/use-episode-playback.ts";
import {
  EpisodeListEntry,
  type EpisodeEntryPlayback,
  type EpisodeEntrySelection,
} from "./episode-list-entry.tsx";
import "./episode-list.css";

export type EpisodeListProps = {
  episodes: EpisodeListItemData[];
  selection: EpisodeSelection;
  playback: EpisodePlayback;
  onSelect: (episodeId: string) => void;
  onPlay: (episodeId: string, audioRef: string) => void;
  onSeek: (startSec: number) => void;
  audioElementRef: Ref<HTMLAudioElement | null>;
  resolvedSrc: string | undefined;
};

/**
 * episodes 配列から EpisodeListEntry を並べる。
 *
 * @require episodes は一覧 API の episodes[]、playback / callbacks は Page 由来
 * @ensure 各 episode を EpisodeListEntry として描画する。entry 向け selection / playback union を導出する
 * @invariant ViewModel hook を呼ばない。audio を持たない
 */
export const EpisodeList = memo(function EpisodeList({
  episodes,
  selection,
  playback,
  onSelect,
  onPlay,
  onSeek,
  audioElementRef,
  resolvedSrc,
}: EpisodeListProps): ReactElement {
  const episodeCount = episodes.length;

  return (
    <div className="episode-list">
      {episodes.map((episode, episodeIndex) => {
        const entrySelection: EpisodeEntrySelection =
          selection.kind === "open" && selection.episodeId === episode.episodeId
            ? { kind: "open", detail: selection.detail }
            : { kind: "closed" };

        const entryPlayback: EpisodeEntryPlayback =
          playback.kind === "playing" && playback.episodeId === episode.episodeId
            ? {
                kind: "playing",
                positionSec: playback.positionSec,
                durationSec: playback.durationSec,
              }
            : { kind: "stopped" };

        const isPlayingEntry = entryPlayback.kind === "playing";

        return (
          <EpisodeListEntry
            key={episode.episodeId}
            episode={episode}
            episodeCount={episodeCount}
            episodeIndex={episodeIndex}
            selection={entrySelection}
            playback={entryPlayback}
            onSelect={onSelect}
            onPlay={onPlay}
            onSeek={onSeek}
            audioElementRef={isPlayingEntry ? audioElementRef : undefined}
            resolvedSrc={isPlayingEntry ? resolvedSrc : undefined}
          />
        );
      })}
    </div>
  );
});
