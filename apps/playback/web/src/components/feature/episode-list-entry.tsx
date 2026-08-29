import type { ReactElement, Ref } from "react";
import type {
  EpisodeDetailState,
  EpisodeListItemData,
} from "../../view-models/episode-list-view-model.ts";
import { EpisodeAudio } from "./episode-audio.tsx";
import { EpisodeDetail } from "./episode-detail.tsx";
import { EpisodeRow } from "./episode-row.tsx";
import "./episode-list-entry.css";

export type EpisodeEntrySelection =
  | { kind: "closed" }
  | { kind: "open"; detail: EpisodeDetailState };

export type EpisodeEntryPlayback =
  | { kind: "stopped" }
  | { kind: "playing"; positionSec: number; durationSec: number };

export type EpisodeListEntryProps = {
  episode: EpisodeListItemData;
  episodeCount: number;
  episodeIndex: number;
  selection: EpisodeEntrySelection;
  playback: EpisodeEntryPlayback;
  onSelect: (episodeId: string) => void;
  onPlay: (episodeId: string, audioRef: string) => void;
  onSeek: (startSec: number) => void;
  audioElementRef?: Ref<HTMLAudioElement | null>;
  resolvedSrc?: string;
};

/**
 * 一覧 1 entry。Row + 任意 Detail + 選択 modifier。
 *
 * @require episode は 1 件、selection / playback は entry 向け union
 * @ensure selection open 時は --selected modifier と Detail を描画する。
 *   playback playing 時は entry 内先頭に EpisodeAudio を mount する
 * @invariant 一覧全体・field 整形を知らない
 */
export function EpisodeListEntry({
  episode,
  episodeCount,
  episodeIndex,
  selection,
  playback,
  onSelect,
  onPlay,
  onSeek,
  audioElementRef,
  resolvedSrc,
}: EpisodeListEntryProps): ReactElement {
  const selected = selection.kind === "open";

  return (
    <div className={`episode-list-entry${selected ? " episode-list-entry--selected" : ""}`}>
      {playback.kind === "playing" && audioElementRef !== undefined ? (
        <EpisodeAudio ref={audioElementRef} src={resolvedSrc} />
      ) : null}
      <EpisodeRow
        episode={episode}
        episodeCount={episodeCount}
        episodeIndex={episodeIndex}
        onSelect={onSelect}
        onPlay={() => {
          onPlay(episode.episodeId, episode.audioRef);
        }}
      />
      {selection.kind === "open" ? (
        <EpisodeDetail detail={selection.detail} playback={playback} onSeek={onSeek} />
      ) : null}
    </div>
  );
}
