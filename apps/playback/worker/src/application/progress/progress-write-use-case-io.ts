import type { UtcInstant } from "../../../../domain/index.ts";

/**
 * 進捗 Write（create / update / complete）の UseCase 入力。
 * 3 つの UseCase が共有する。Controller が Request に `episodeId` を足して作る。
 */
export type ProgressWriteUseCaseInput = {
  readonly episodeId: string;
  readonly positionSec: number;
  readonly clientAt: UtcInstant;
};

/** 進捗 Write の UseCase 出力。勝ち側の `first*` だけを持ち、`positionSec` は持たない。 */
export type ProgressWriteUseCaseOutput = {
  readonly firstPlayedAt: UtcInstant;
  readonly firstCompletedAt: UtcInstant | null;
};
