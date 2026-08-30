import { app } from "./routes/app.ts";

/**
 * Cloudflare Workers の module entry。
 *
 * @require wrangler の `main` が本 file を指す
 * @ensure default export の `fetch` は `routes/app.ts` の Hono instance の `fetch` と同一参照
 * @invariant HTTP の振り分け・Error 写像・Drive 結線はここへ置かない
 */
export default { fetch: app.fetch };
