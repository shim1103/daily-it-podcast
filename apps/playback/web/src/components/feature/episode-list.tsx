import { Fragment, memo, type ReactElement, type ReactNode, type RefObject } from "react";
import type { EpisodeListState } from "../../view-models/episode-list-view-model.ts";
import { EpisodeListItem } from "./episode-list-item.tsx";
import { EpisodeManuscript } from "./episode-manuscript.tsx";
import { EpisodePlayer } from "./episode-player.tsx";
import "./episode-detail.css";
import "./episode-list.css";
import "./episode-selected-group.css";

type SelectedEpisode = NonNullable<
  Extract<EpisodeListState, { status: "success" }>["selectedEpisode"]
>;

function SelectedEpisodeDetail({
  selectedEpisode,
  baseUrl,
  onSeek,
  audioElementRef,
}: {
  selectedEpisode: SelectedEpisode;
  baseUrl: string;
  onSeek: (startSec: number) => void;
  audioElementRef: RefObject<HTMLAudioElement | null>;
}): ReactElement {
  if (selectedEpisode.status === "loading") {
    return <div className="episode-detail" data-episode-detail-loading="" />;
  }

  if (selectedEpisode.status === "error") {
    return <div className="episode-detail" data-episode-detail-error="" />;
  }

  return (
    <div className="episode-detail">
      <EpisodeManuscript body={selectedEpisode.episode.body} onSeek={onSeek} />
      <EpisodePlayer
        ref={audioElementRef}
        baseUrl={baseUrl}
        audioRef={selectedEpisode.episode.audioRef}
      />
    </div>
  );
}

export type EpisodeListProps = {
  state: EpisodeListState;
  baseUrl: string;
  onSelect: (episodeId: string) => void;
  audioElementRef: RefObject<HTMLAudioElement | null>;
  onSeek: (startSec: number) => void;
};

/**
 * 一覧 state から EpisodeListItem を並べ、選択中 episode があればその item のみを紫枠付きで詳細展開する。
 *
 * @require state は ViewModel が持つ EpisodeListState、baseUrl は playback worker の origin相当
 * @ensure success 時、未選択なら全 episode を並べる。選択中は当該 episode の item + detail のみを描画する。
 *   選択中は item と detail を紫枠で囲む。topic の MM:SS クリックは onSeek へ委譲する
 * @invariant item 以外の field 加工をしない
 */
export const EpisodeList = memo(function EpisodeList({
  state,
  baseUrl,
  onSelect,
  audioElementRef,
  onSeek,
}: EpisodeListProps): ReactElement {
  const children: ReactNode[] = [];

  if (state.status === "success") {
    const episodeCount = state.episodes.length;
    const isFocusMode = state.selectedEpisodeId !== null;

    state.episodes.forEach((episode, episodeIndex) => {
      if (isFocusMode && episode.episodeId !== state.selectedEpisodeId) {
        return;
      }

      const listItem = (
        <EpisodeListItem
          key={`${episode.episodeId}-item`}
          episode={episode}
          episodeCount={episodeCount}
          episodeIndex={episodeIndex}
          onSelect={onSelect}
        />
      );

      const detail =
        state.selectedEpisodeId === episode.episodeId && state.selectedEpisode ? (
          <SelectedEpisodeDetail
            key={`${episode.episodeId}-detail`}
            selectedEpisode={state.selectedEpisode}
            baseUrl={baseUrl}
            onSeek={onSeek}
            audioElementRef={audioElementRef}
          />
        ) : null;

      if (isFocusMode) {
        children.push(
          <div key={episode.episodeId} className="episode-selected-group">
            {listItem}
            {detail}
          </div>,
        );
        return;
      }

      children.push(
        <Fragment key={episode.episodeId}>
          {listItem}
          {detail}
        </Fragment>,
      );
    });
  }

  return <div className="episode-list">{children}</div>;
});
