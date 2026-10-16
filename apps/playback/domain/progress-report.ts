import type { ClientAt } from "./client-at.ts";
import type { PositionSec } from "./position-sec.ts";

/**
 * 端末からの**進捗報告**。「`clientAt` の時点で、再生位置が `positionSec` だった」という事実。
 * どの episode の報告かは、報告を受ける層が別に持つ（HTTP では path の `episodeId`）。
 */
export type ProgressReport = {
  readonly positionSec: PositionSec;
  readonly clientAt: ClientAt;
};
