import type { ProgressPullResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";

/**
 * 進捗 pull（HTTP GET /progress）。`since` 以降に更新された行だけを返す。
 *
 * 差分条件 `lastPlayedAt > since` は Port の `listUpdatedSince` が持つ
 * （Decision 2026-10-01T18-54-14）。本 UseCase は応答形へ包む。
 *
 * @require since は Route 側 schema で検証済みの ISO-8601（offset 付き）
 * @ensure ProgressPullResponse を返す。該当なしは空配列。storage 失敗は Port が throw する
 */
export async function pullProgress(
  repository: ProgressRepository,
  since: string,
): Promise<ProgressPullResponse> {
  const episodes = await repository.listUpdatedSince(since);
  return { episodes: [...episodes] };
}
