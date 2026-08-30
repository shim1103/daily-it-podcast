/**
 * list 行用に topics[].title を " / " 区切りの1行へ整形する。
 *
 * @require topics は listEpisode の topics 配列
 * @ensure 各 title を出現順のまま " / " で連結した文字列を返す。0 件なら空文字
 * @invariant title の加工・省略・並べ替えをしない
 */
export function formatTopicTitles(topics: readonly { title: string }[]): string {
  return topics.map((topic) => topic.title).join(" / ");
}
