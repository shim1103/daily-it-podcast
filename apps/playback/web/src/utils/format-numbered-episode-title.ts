/**
 * list 行用に episode title へ通し番号を付ける。先頭（新しい方）ほど大きい番号、最古が 1。
 *
 * @require episodeCount は一覧件数、episodeIndex は 0 始まりの出現位置、title は episode.title
 * @ensure `${episodeCount - episodeIndex}.　${title}` を返す
 * @invariant title の加工・省略をしない
 */
export function formatNumberedEpisodeTitle(
  episodeCount: number,
  episodeIndex: number,
  title: string,
): string {
  return `${episodeCount - episodeIndex}.　${title}`;
}
