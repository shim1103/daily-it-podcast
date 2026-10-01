import type { ProgressWriteRequest } from "../../../../contracts/index.ts";

/**
 * 進捗 Write（create/update/complete）の Application 入力。
 * HTTP body に episodeId を足した形。UseCase 語彙（Port は持たない）。
 */
export type ProgressWriteCommand = ProgressWriteRequest & {
  readonly episodeId: string;
};
