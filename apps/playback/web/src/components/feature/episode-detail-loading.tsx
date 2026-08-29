import type { ReactElement } from "react";
import "./episode-detail.css";

/**
 * 詳細取得 loading 表示。
 *
 * @ensure data-episode-detail-loading 付き episode-detail 容器を返す
 */
export function EpisodeDetailLoading(): ReactElement {
  return <div className="episode-detail" data-episode-detail-loading="" />;
}
