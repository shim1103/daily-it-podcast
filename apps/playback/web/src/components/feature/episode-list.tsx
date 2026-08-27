import { Fragment, type ReactElement, type ReactNode } from "react";
import type { EpisodeListState } from "../../view-models/episode-list-view-model.ts";
import { EpisodeListItem } from "./episode-list-item.tsx";
import { EpisodeManuscript } from "./episode-manuscript.tsx";
import { EpisodePlayer } from "./episode-player.tsx";

type SelectedEpisode = NonNullable<
  Extract<EpisodeListState, { status: "success" }>["selectedEpisode"]
>;

/**
 * 選択中 episode の詳細 state から、manuscript・player を組み合わせた要素を組み立てる。
 * title・date は EpisodeListItem が既に描画しているため、ここでは重ねて描画しない。
 * loading・error は専用の表示を返す。
 */
function SelectedEpisodeDetail({
  selectedEpisode,
  baseUrl,
}: {
  selectedEpisode: SelectedEpisode;
  baseUrl: string;
}): ReactElement {
  if (selectedEpisode.status === "loading") {
    return <div data-episode-detail-loading="" />;
  }

  if (selectedEpisode.status === "error") {
    return <div data-episode-detail-error="" />;
  }

  return (
    <div>
      <EpisodeManuscript body={selectedEpisode.episode.body} />
      <EpisodePlayer baseUrl={baseUrl} audioRef={selectedEpisode.episode.audioRef} />
    </div>
  );
}

export type EpisodeListProps = {
  state: EpisodeListState;
  baseUrl: string;
  onSelect: (episodeId: string) => void;
};

/**
 * 一覧 state から、EpisodeListItem を並べ、選択中 episode があればその直後に詳細を展開する。
 *
 * @require state は ViewModel が持つ EpisodeListState、baseUrl は playback worker の origin相当
 * @ensure success 時のみ episodes の順番通りに EpisodeListItem を並べる。selectedEpisodeId と一致する
 *   item の直後に、選択中 episode の詳細（manuscript + player、または loading/error 表示）を展開する。
 *   title・date は item が既に描画しているため詳細側では重ねて描画しない。loading / error は何も描画しない
 * @invariant item 以外の field 加工をしない
 */
export function EpisodeList({ state, baseUrl, onSelect }: EpisodeListProps): ReactElement {
  const children: ReactNode[] = [];

  if (state.status === "success") {
    for (const episode of state.episodes) {
      children.push(
        <Fragment key={episode.episodeId}>
          <EpisodeListItem episode={episode} onSelect={onSelect} />
          {state.selectedEpisodeId === episode.episodeId && state.selectedEpisode && (
            <SelectedEpisodeDetail selectedEpisode={state.selectedEpisode} baseUrl={baseUrl} />
          )}
        </Fragment>,
      );
    }
  }

  return <div>{children}</div>;
}
