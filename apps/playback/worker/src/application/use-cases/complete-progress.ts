import type { ProgressWriteResponse } from "../../../../contracts/index.ts";
import type { ProgressRepository, ProgressWriteCommand } from "../ports/progress-repository.ts";

/**
 * 進捗 complete（HTTP POST progress/complete）。完走ゾーン突入の記録。
 *
 * A: Port へ委譲するだけ。ゾーン判定・merge は C。
 *
 * @require command は Route 側 schema で検証済み
 * @ensure ProgressWriteResponse を返す。storage 失敗は Port が Infrastructure Error を throw する
 */
export async function completeProgress(
  repository: ProgressRepository,
  command: ProgressWriteCommand,
): Promise<ProgressWriteResponse> {
  return repository.completeProgress(command);
}
