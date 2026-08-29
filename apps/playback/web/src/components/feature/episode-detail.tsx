import type { ReactElement } from "react";
import type { EpisodeDetailState } from "../../view-models/episode-list-view-model.ts";
import type { EpisodeEntryPlayback } from "./episode-list-entry.tsx";
import { EpisodeDetailError } from "./episode-detail-error.tsx";
import { EpisodeDetailLoading } from "./episode-detail-loading.tsx";
import { EpisodeDetailSuccess } from "./episode-detail-success.tsx";

export type EpisodeDetailProps = {
  detail: EpisodeDetailState;
  playback: EpisodeEntryPlayback;
  onSeek: (startSec: number) => void;
};

/**
 * 選択中 episode の詳細 state を loading / error / success に分岐して描画する。
 *
 * @require detail は ViewModel が返す EpisodeDetailState
 * @ensure status に応じて detail Feature 群の 1 つだけを描画する
 * @invariant 別 page/route へ遷移しない
 */
export function EpisodeDetail({ detail, playback, onSeek }: EpisodeDetailProps): ReactElement {
  if (detail.status === "loading") {
    return <EpisodeDetailLoading />;
  }

  if (detail.status === "error") {
    return <EpisodeDetailError />;
  }

  return <EpisodeDetailSuccess episode={detail.episode} playback={playback} onSeek={onSeek} />;
}
