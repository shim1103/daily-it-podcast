import { Ajv2020 } from "ajv/dist/2020.js";
import type { EpisodeManuscript } from "../../application/ports/episode-repository.ts";
import manuscriptJsonSchema from "../../../../../../contracts/manuscript.schema.json" with {
  type: "json",
};

const ajv = new Ajv2020({ allErrors: true, strict: true });
const validateManuscript = ajv.compile(manuscriptJsonSchema);

/**
 * Drive 上の原稿 JSON を検証する（repo 根 `contracts/manuscript.schema.json` が正）。
 *
 * @ensure 適合時は `{ success: true, data }`、不適合時は `{ success: false }`
 * @invariant `apps/playback/contracts`（HTTP）を読まない
 */
export const ManuscriptSchema = {
  safeParse(data: unknown): { success: true; data: EpisodeManuscript } | { success: false } {
    if (!validateManuscript(data)) {
      return { success: false };
    }
    return { success: true, data: data as EpisodeManuscript };
  },
};
