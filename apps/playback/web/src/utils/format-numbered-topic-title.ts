/**
 * detail 用に topic.title へ 1 始まりの通し番号を付ける。
 *
 * @require topicIndex は 0 始まりの出現位置、title は topic.title
 * @ensure `${topicIndex + 1}. ${title}` を返す
 * @invariant title の加工・省略をしない
 */
export function formatNumberedTopicTitle(topicIndex: number, title: string): string {
  return `${topicIndex + 1}. ${title}`;
}
