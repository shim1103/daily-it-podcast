/**
 * episode を指す不透明な対応キー。生成規則は Generator に閉じ、Reader は stem との一致だけを見る
 * （repo 根 `contracts/episode-layout.md`）。空でないことは、入口の契約（zod の `min(1)`）と
 * `manuscript.schema.json` の `minLength` が保証し、この型は名前だけを持つ。
 */
export type EpisodeId = string;
