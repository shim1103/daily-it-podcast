/**
 * 進捗 Write（create / update / complete）の UseCase 入力。
 * 3 つの UseCase が共有する。Controller が Request に `episodeId` を足して作る。
 */
export type ProgressWriteUseCaseInput = {
  readonly episodeId: string;
  readonly positionSec: number;
  readonly clientAt: string;
};

/** 進捗 Write の UseCase 出力。勝ち側の `first*` だけを持ち、`positionSec` は持たない。 */
export type ProgressWriteUseCaseOutput = {
  readonly firstPlayedAt: string;
  readonly firstCompletedAt: string | null;
};
