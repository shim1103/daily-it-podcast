import type { ProgressPullResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository } from "../ports/progress-repository.ts";

/**
 * 進捗 pull（HTTP GET /progress）。`since` 以降に更新された行だけを返す。
 *
 * A: Port へ委譲し応答形に包むだけ。差分条件の詳細は C。
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
