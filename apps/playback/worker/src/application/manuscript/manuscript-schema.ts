import { Ajv2020 } from "ajv/dist/2020.js";
import manuscriptJsonSchema from "../../../../../../contracts/manuscript.schema.json" with {
  type: "json",
};
import type { EpisodeManuscript } from "../ports/episode-repository.ts";

const ajv = new Ajv2020({ allErrors: true, strict: true });
const validateManuscript = ajv.compile(manuscriptJsonSchema);

/**
 * `safeParse` の戻り値。`success` を判別子とする discriminated union。
 *
 * why: `success: false` は理由（zod の `error` 相当）を持たない。呼び出し側（検証層の
 * listEpisodes / getEpisode 経路）は不適合の理由を使わず、除外または EpisodeContentError への
 * 変換をするだけなので、失敗理由は保持しない。
 */
type ManuscriptParseResult = { success: true; data: EpisodeManuscript } | { success: false };

/**
 * 原稿 JSON を検証する（repo 根 `contracts/manuscript.schema.json` が正）。
 *
 * schema 適合・stem 一致・不正 JSON の判定は Google Drive という具体 platform に依存しない
 * 純粋な判断のため、Application 層に置く。
 *
 * @ensure 適合時は `{ success: true, data }`、不適合時は `{ success: false }`（失敗理由は持たない）
 * @invariant `apps/playback/contracts`（HTTP schema）を読まない
 */
export const ManuscriptSchema = {
  safeParse(data: unknown): ManuscriptParseResult {
    if (!validateManuscript(data)) {
      return { success: false };
    }
    return { success: true, data: data as EpisodeManuscript };
  },
};
