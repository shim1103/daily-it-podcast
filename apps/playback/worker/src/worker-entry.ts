import { fetch } from "./routes/fetch.ts";

/**
 * Cloudflare Workers の module entry。
 *
 * @require wrangler の `main` が本 file を指す
 * @ensure default export の `fetch` は `routes/fetch.ts` の HTTP 入口と同一参照
 * @invariant HTTP の振り分け・Error 写像・Drive 結線はここへ置かない
 */
export default { fetch };
