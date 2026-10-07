import type { ProgressRepository, ProgressUpdatedEntry } from "../ports/progress-repository.ts";

export type PullProgressUseCaseInput = { readonly since: string };

export type PullProgressUseCaseOutput = { readonly episodes: readonly ProgressUpdatedEntry[] };

// todo: C が cursor 版（`ProgressCursorPull*`・`listChangedAfter`）へ置換したら、この UseCase の since 版を削除する
/**
 * 進捗 pull（HTTP GET /progress）。`since` 以降に更新された行だけを返す。
 *
 * 差分条件 `lastPlayedAt > since` は Port の `listUpdatedSince` が持つ
 * （Decision 2026-10-01T18-54-14）。本 UseCase は出力の形へ包む。
 *
 * @require since は Controller が契約 schema の検証済み値から渡した UTC 固定幅の ISO 表記（契約境界で正規化済み）
 * @ensure 該当なしは空配列。storage 失敗は Port が throw する
 */
export async function pullProgress(
  repository: ProgressRepository,
  input: PullProgressUseCaseInput,
): Promise<PullProgressUseCaseOutput> {
  const episodes = await repository.listUpdatedSince(input.since);
  return { episodes };
}
