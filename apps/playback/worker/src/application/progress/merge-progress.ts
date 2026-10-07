import type { EpisodeProgress } from "../../entities/models/episode-progress.ts";

/** merge に渡す incoming write（HTTP body と同形の核）。 */
export type ProgressMergeIncoming = {
  readonly positionSec: number;
  readonly clientAt: string;
};

/**
 * 既存進捗行と incoming write を `clientAt` 鍵で merge する純関数（create／update 用）。
 * 完走記録は行わない。完走は {@link mergeProgressAsCompleted}。
 *
 * firstPlayedAt / firstCompletedAt は先勝ち（より早い時刻。同時刻は既存を残す）。
 * positionSec / lastPlayedAt は後勝ち（より遅い時刻。同時刻は incoming を採る）。
 * 同時刻規則の why は Decision の先勝ち／後勝ち意味を崩さず、再送でカーソルだけ揃えられるようにするため。
 *
 * @require `incoming.clientAt` は UTC 固定幅の ISO 表記（`EpisodeProgress` と同じ。契約境界で正規化済み）
 * @ensure 常に完全な `EpisodeProgress` を返す（existing null なら初回行を構築）
 * @ensure `firstCompletedAt` を新設しない（existing null なら null、既存ありなら据え置きまたは先勝ちの既存完走のみ）
 */
export function mergeProgress(
  existing: EpisodeProgress | null,
  incoming: ProgressMergeIncoming,
): EpisodeProgress {
  if (existing === null) {
    return {
      positionSec: incoming.positionSec,
      firstPlayedAt: incoming.clientAt,
      firstCompletedAt: null,
      lastPlayedAt: incoming.clientAt,
    };
  }

  return mergeFields(existing, incoming, existing.firstCompletedAt);
}

/**
 * 既存行への完走 merge。行なしは呼び出し側（UseCase）が拒む。
 *
 * @require `existing` は永続済み進捗行（null 不可）
 * @require 呼び出し側が完走ゾーン突入を既に確認済みであること
 * @ensure `firstCompletedAt` を先勝ちで更新する（未完走なら `incoming.clientAt`）
 */
export function mergeProgressAsCompleted(
  existing: EpisodeProgress,
  incoming: ProgressMergeIncoming,
): EpisodeProgress {
  let firstCompletedAt: string;
  if (existing.firstCompletedAt === null) {
    firstCompletedAt = incoming.clientAt;
  } else {
    // why: 同時刻は既存据え置き（firstPlayedAt と同規則）
    firstCompletedAt =
      incoming.clientAt < existing.firstCompletedAt ? incoming.clientAt : existing.firstCompletedAt;
  }

  return mergeFields(existing, incoming, firstCompletedAt);
}

function mergeFields(
  existing: EpisodeProgress,
  incoming: ProgressMergeIncoming,
  firstCompletedAt: string | null,
): EpisodeProgress {
  // why: 同時刻は既存据え置き。再書込で first* を揺らさない
  const firstPlayedAt =
    incoming.clientAt < existing.firstPlayedAt ? incoming.clientAt : existing.firstPlayedAt;

  // why: 同時刻は incoming を採る。再送で共有カーソル（position）を揃えられるようにする
  const incomingWinsLast = incoming.clientAt >= existing.lastPlayedAt;
  const positionSec = incomingWinsLast ? incoming.positionSec : existing.positionSec;
  const lastPlayedAt = incomingWinsLast ? incoming.clientAt : existing.lastPlayedAt;

  return {
    positionSec,
    firstPlayedAt,
    firstCompletedAt,
    lastPlayedAt,
  };
}
