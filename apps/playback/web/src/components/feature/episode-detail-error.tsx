import type { ReactElement } from "react";
import "./episode-detail.css";

/**
 * 詳細取得 error 表示。
 *
 * @ensure data-episode-detail-error 付き episode-detail 容器を返す
 */
export function EpisodeDetailError(): ReactElement {
  return <div className="episode-detail" data-episode-detail-error="" />;
}
