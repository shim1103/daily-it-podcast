export type Route = { kind: "episode-list" } | { kind: "episode-detail"; episodeId: string };

const episodeDetailRoute = /^#\/episodes\/(.+)$/;

/**
 * hash 文字列を route へ変換する純粋関数。
 *
 * @require hash は `window.location.hash` 相当の文字列
 * @ensure `#/episodes/{episodeId}` 形式（episodeId が空でない）の時のみ episode-detail を返す。それ以外は episode-list を返す
 */
export function matchRoute(hash: string): Route {
  const detailMatch = hash.match(episodeDetailRoute);
  if (detailMatch) {
    const episodeId = decodeURIComponent(detailMatch[1] ?? "");
    if (episodeId.length > 0) {
      return { kind: "episode-detail", episodeId };
    }
  }

  return { kind: "episode-list" };
}
