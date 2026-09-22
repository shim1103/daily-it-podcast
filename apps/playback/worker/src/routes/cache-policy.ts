export const audioCacheHeaders = {
  "Cache-Control": "public, max-age=86400",
  "Cloudflare-CDN-Cache-Control": "public, max-age=604800, stale-if-error=86400",
} as const;

/**
 * 進捗関連・error など Workers Cache / browser に載せない応答。
 * `Cache-Control` と edge 向け header の両方を no-store にし、Workers Cache に入れない。
 */
export const noStoreCacheHeaders = {
  "Cache-Control": "no-store",
  "Cloudflare-CDN-Cache-Control": "no-store",
} as const;

/**
 * listEpisodes（progress embed 済み）の成功応答。
 * 進捗を含むため catalog 向け TTL は使わず no-store（Decision 2026-09-19T19-12-01）。
 */
export const episodeListCacheHeaders = noStoreCacheHeaders;

/**
 * session 中の進捗 pull 成功応答。同様に no-store。
 */
export const progressPullCacheHeaders = noStoreCacheHeaders;

/**
 * 進捗 Write（create / update / complete）成功応答。同様に no-store。
 */
export const progressWriteCacheHeaders = noStoreCacheHeaders;
