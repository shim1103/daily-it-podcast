import type { ClientAt } from "../../../../domain/index.ts";
import {
  PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC,
  PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC,
} from "../../entities/constants/progress.ts";
import { ProgressRuleError } from "../../entities/errors/progress-rule-error.ts";

/**
 * `clientAt` が `serverAt` に対する許容 skew 内かを検査する純関数。
 *
 * @require `clientAt` は契約 schema 通過済みの UTC 固定幅。`serverAt` は Clock が返した受信時刻
 * @ensure ちょうど許容幅は通し、超過（`>`）だけ `ProgressRuleError` を throw する
 */
export function assertClientAtWithinSkew(clientAt: ClientAt, serverAt: Date): void {
  const clientAtMs = new Date(clientAt).getTime();
  const serverAtMs = serverAt.getTime();
  const deltaMs = clientAtMs - serverAtMs;

  // why: 拒否は `>`。ちょうど定数幅は通し、汚染上界を定数と一致させる（Decision 2026-10-01T23-34-00）
  if (deltaMs > PROGRESS_CLIENT_AT_MAX_FUTURE_SKEW_SEC * 1000) {
    throw new ProgressRuleError("clientAt が serverAt より未来側の許容を超える");
  }
  if (-deltaMs > PROGRESS_CLIENT_AT_MAX_PAST_SKEW_SEC * 1000) {
    throw new ProgressRuleError("clientAt が serverAt より過去側の許容を超える");
  }
}
