import type { EpisodeManuscript } from "../../entities/models/episode-manuscript.ts";
import type { EpisodeProgress } from "../../entities/models/episode-progress.ts";
import { selectValidManuscript } from "../manuscript/verify-manuscript.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";

/** 原稿に進捗を足した一覧の 1 件。進捗行がない episode は `progress: null`。 */
export type EpisodeWithProgress = EpisodeManuscript & {
  readonly progress: EpisodeProgress | null;
};

export type ListEpisodesUseCaseOutput = {
  readonly episodes: readonly EpisodeWithProgress[];
};

/**
 * 原稿一覧を取得し、schema 適合かつ stem 一致した entry だけの部分一覧を、date 降順で返す。
 * 各 episode の progress は ProgressRepository.getByEpisodeIds で embed する（行なしは null）。
 *
 * Port は取得したままの生 json 配列を返すだけなので、entry 単位の適合判定と除外は
 * この use-case が行う（`selectValidManuscript`）。Port の返却順は storage 実装依存
 * （R2 の ListObjectsV2 は key 辞書順）で日付順を保証しないため、並び替えも use-case が持つ。
 *
 * @require progressRepository は Composition が結線する
 * @ensure 個々の entry が schema 不適合・stem 不一致・不正 JSON でも throw せず、適合分だけを返す。
 *   この除外は listEpisodes 自身の仕様であり、他層 error の握りつぶしではない。
 * @ensure 返す episodes は date 降順（新しい日付が先頭）。date が同値の entry 間の順序は未規定。
 * @ensure 各 episode.progress は行ありなら EpisodeProgress、行なしなら null（Decision 2026-09-19T16:26:00 / 16:25:18）
 */
export async function listEpisodes(
  repository: EpisodeRepository,
  progressRepository: ProgressRepository,
): Promise<ListEpisodesUseCaseOutput> {
  const entries = await repository.listManuscripts();
  const manuscripts = entries
    .map((entry) => selectValidManuscript(entry.json, entry.stem))
    .filter((manuscript): manuscript is EpisodeManuscript => manuscript !== undefined)
    .sort((a, b) => (a.date < b.date ? 1 : a.date > b.date ? -1 : 0));

  const progressById = await progressRepository.getByEpisodeIds({
    episodeIds: manuscripts.map((manuscript) => manuscript.episodeId),
  });
  const episodes = manuscripts.map((manuscript) => ({
    ...manuscript,
    progress: progressById.get(manuscript.episodeId) ?? null,
  }));
  return { episodes };
}
