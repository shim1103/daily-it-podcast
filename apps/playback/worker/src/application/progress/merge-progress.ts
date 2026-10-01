import type { EpisodeProgress } from "../../../../contracts/index.ts";

/** merge に渡す incoming write（HTTP body と同形の核）。 */
export type ProgressMergeIncoming = {
  readonly positionSec: number;
  readonly clientAt: string;
};

/** complete path 向けの任意 option。ゾーン判定は呼び出し側（`isInCompleteZone`）が行う。 */
export type ProgressMergeOptions = {
  readonly markCompleted?: boolean;
};

/**
 * 既存進捗行と incoming write を `clientAt` 鍵で merge する純関数。
 *
 * firstPlayedAt / firstCompletedAt は先勝ち（より早い時刻。同時刻は既存を残す）。
 * positionSec / lastPlayedAt は後勝ち（より遅い時刻。同時刻は incoming を採る）。
 * 同時刻規則の why は Decision の先勝ち／後勝ち意味を崩さず、再送でカーソルだけ揃えられるようにするため。
 *
 * @require `incoming.clientAt` は契約どおりのタイムゾーン付き ISO-8601
 * @require `markCompleted` が true の時、呼び出し側が完走ゾーン突入を既に確認済みであること
 * @ensure 常に完全な `EpisodeProgress` を返す（existing null なら初回行を構築）
 * @ensure `markCompleted` でない限り `firstCompletedAt` を新設しない
 */
export function mergeProgress(
  existing: EpisodeProgress | null,
  incoming: ProgressMergeIncoming,
  options: ProgressMergeOptions = {},
): EpisodeProgress {
  const markCompleted = options.markCompleted === true;

  if (existing === null) {
    return {
      positionSec: incoming.positionSec,
      firstPlayedAt: incoming.clientAt,
      firstCompletedAt: markCompleted ? incoming.clientAt : null,
      lastPlayedAt: incoming.clientAt,
    };
  }

  // why: 同時刻は既存据え置き。再書込で first* を揺らさない
  const firstPlayedAt =
    incoming.clientAt < existing.firstPlayedAt ? incoming.clientAt : existing.firstPlayedAt;

  // why: 同時刻は incoming を採る。再送で共有カーソル（position）を揃えられるようにする
  const incomingWinsLast = incoming.clientAt >= existing.lastPlayedAt;
  const positionSec = incomingWinsLast ? incoming.positionSec : existing.positionSec;
  const lastPlayedAt = incomingWinsLast ? incoming.clientAt : existing.lastPlayedAt;

  let firstCompletedAt = existing.firstCompletedAt;
  if (markCompleted) {
    if (existing.firstCompletedAt === null) {
      firstCompletedAt = incoming.clientAt;
    } else {
      // why: 同時刻は既存据え置き（firstPlayedAt と同規則）
      firstCompletedAt =
        incoming.clientAt < existing.firstCompletedAt
          ? incoming.clientAt
          : existing.firstCompletedAt;
    }
  }

  return {
    positionSec,
    firstPlayedAt,
    firstCompletedAt,
    lastPlayedAt,
  };
}
