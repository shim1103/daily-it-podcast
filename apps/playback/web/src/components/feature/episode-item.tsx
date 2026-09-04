import type { ReactElement } from "react";
import "./episode-item.css";
import type { EpisodeData } from "../../view-models/playback-state.ts";
import { EpisodeManuscript } from "./episode-manuscript.tsx";
import { EpisodeRow } from "./episode-row.tsx";

export type EpisodeItemProps = {
  episode: EpisodeData;
  episodeCount: number;
  episodeIndex: number;
  isSelected: boolean;
  isActivePlayback: boolean;
  isPlaying: boolean;
  onSelect: (episodeId: string) => void;
  onPlay: (episodeId: string) => void;
  onStop: () => void;
  onSeek: (startSec: number) => void;
};

/**
 * 一覧 1 件分の枠。行本体（`EpisodeRow`）と、選択中だけ出す原稿（`EpisodeManuscript`）を束ねる。
 *
 * @require row 由来の isSelected / isActivePlayback / isPlaying は caller が derive して渡す。
 *   onSeek は選択中 episode の audio 位置を変える handler
 * @ensure `EpisodeRow` を常に描き、`isSelected` のときだけその直後に `EpisodeManuscript` を描く。
 *   選択中は横（左 edge）に紫線を出す（`data-selected`）
 * @invariant この枠が「選択中の紫線」と「行間の空白」の責務を持つ。`EpisodeRow` /
 *   `EpisodeManuscript` はそれらを持たない。表示整形・derive はしない
 */
export function EpisodeItem({
  episode,
  episodeCount,
  episodeIndex,
  isSelected,
  isActivePlayback,
  isPlaying,
  onSelect,
  onPlay,
  onStop,
  onSeek,
}: EpisodeItemProps): ReactElement {
  return (
    <article className="episode-item" data-selected={isSelected ? "true" : "false"}>
      <EpisodeRow
        episode={episode}
        episodeCount={episodeCount}
        episodeIndex={episodeIndex}
        isSelected={isSelected}
        isActivePlayback={isActivePlayback}
        isPlaying={isPlaying}
        onSelect={onSelect}
        onPlay={onPlay}
        onStop={onStop}
      />
      {isSelected && (
        <EpisodeManuscript body={episode.body} durationSec={episode.durationSec} onSeek={onSeek} />
      )}
    </article>
  );
}
