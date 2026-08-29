import { Ajv2020 } from "ajv/dist/2020.js";
import type { EpisodeManuscript } from "../../application/ports/episode-repository.ts";
import manuscriptJsonSchema from "../../../../../../contracts/manuscript.schema.json" with {
  type: "json",
};

const ajv = new Ajv2020({ allErrors: true, strict: true });
const validateManuscript = ajv.compile(manuscriptJsonSchema);

/**
 * `safeParse` の戻り値。`success` を判別子とする discriminated union。
 *
 * why: `success: false` は理由（zod の `error` 相当）を持たない。呼び出し側（両 repository の
 * listEpisodes / getEpisode）は不適合の理由を使わず、除外または EpisodeNotFoundError への
 * 変換をするだけなので、失敗理由は保持しない。
 */
type ManuscriptParseResult = { success: true; data: EpisodeManuscript } | { success: false };

/**
 * Drive 上の原稿 JSON を検証する（repo 根 `contracts/manuscript.schema.json` が正）。
 *
 * @ensure 適合時は `{ success: true, data }`、不適合時は `{ success: false }`（失敗理由は持たない）
 * @invariant `apps/playback/contracts`（HTTP）を読まない
 */
export const ManuscriptSchema = {
  safeParse(data: unknown): ManuscriptParseResult {
    if (!validateManuscript(data)) {
      return { success: false };
    }
    return { success: true, data: data as EpisodeManuscript };
  },
};
