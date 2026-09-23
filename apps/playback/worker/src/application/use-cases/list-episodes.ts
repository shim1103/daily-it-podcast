import type { ListEpisodesResponse } from "../../../../contracts/index.ts";
import { selectValidListItem } from "../manuscript/verify-manuscript.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";

/**
 * 原稿一覧を取得し、schema 適合かつ stem 一致した entry だけの部分一覧を、date 降順で返す。
 *
 * Port は取得したままの生 json 配列を返すだけなので、entry 単位の適合判定と除外は
 * この use-case が行う（`selectValidListItem`）。Port の返却順は storage 実装依存
 * （R2 の ListObjectsV2 は key 辞書順）で日付順を保証しないため、並び替えも use-case が持つ。
 *
 * @require progressRepository は Composition が結線する（A: embed join は未実装で progress は null のまま）
 * @ensure 個々の entry が schema 不適合・stem 不一致・不正 JSON でも throw せず、適合分だけを返す。
 *   この除外は listEpisodes 自身の仕様であり、他層 error の握りつぶしではない。
 * @ensure 返す episodes は date 降順（新しい日付が先頭）。date が同値の entry 間の順序は未規定。
 */
export async function listEpisodes(
  repository: EpisodeRepository,
  progressRepository: ProgressRepository,
): Promise<ListEpisodesResponse> {
  // todo: ProgressRepository で embed join し progress を埋める（C）。結線のため引数は残す
  void progressRepository;
  const entries = await repository.listManuscripts();
  const episodes = entries
    .map((entry) => selectValidListItem(entry.json, entry.stem))
    .filter((item): item is NonNullable<typeof item> => item !== undefined)
    .sort((a, b) => (a.date < b.date ? 1 : a.date > b.date ? -1 : 0));
  return { episodes };
}
