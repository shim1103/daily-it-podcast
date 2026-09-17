export const audioCacheHeaders = {
  "Cache-Control": "public, max-age=86400",
  "Cloudflare-CDN-Cache-Control": "public, max-age=604800, stale-if-error=86400",
} as const;

export const episodeListCacheHeaders = {
  "Cache-Control": "public, max-age=60",
  "Cloudflare-CDN-Cache-Control": "public, max-age=300, stale-while-revalidate=60",
} as const;

export const noStoreCacheHeaders = {
  "Cache-Control": "no-store",
} as const;
