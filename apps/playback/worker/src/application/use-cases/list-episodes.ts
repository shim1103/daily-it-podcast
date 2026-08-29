import type { ListEpisodesResponse } from "../../../../contracts/index.ts";
import { selectValidListItem } from "../manuscript/verify-manuscript.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";

/**
 * 原稿一覧を取得し、schema 適合かつ stem 一致した entry だけの部分一覧を返す。
 *
 * Port は取得したままの生 json 配列を返すだけなので、entry 単位の適合判定と除外は
 * この use-case が行う（`selectValidListItem`）。
 *
 * @ensure 個々の entry が schema 不適合・stem 不一致・不正 JSON でも throw せず、適合分だけを返す。
 *   この除外は listEpisodes 自身の仕様であり、他層 error の握りつぶしではない。
 */
export async function listEpisodes(repository: EpisodeRepository): Promise<ListEpisodesResponse> {
  const entries = await repository.listManuscripts();
  const episodes = entries
    .map((entry) => selectValidListItem(entry.json, entry.stem))
    .filter((item): item is NonNullable<typeof item> => item !== undefined);
  return { episodes };
}
