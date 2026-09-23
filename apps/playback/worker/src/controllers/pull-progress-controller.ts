import type { ProgressPullResponse } from "../../../contracts/index.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type PullProgressUseCase = (since: string) => Promise<ProgressPullResponse>;

export type PullProgressController = (since: string) => Promise<ProgressPullResponse>;

/**
 * 進捗 pull を返す Controller を組み立てる。
 *
 * @require since は呼び出し側（Hono route の zValidator）が契約 schema で検証済み
 * @ensure 戻り関数は契約の ProgressPullResponse を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createPullProgressController(useCase: PullProgressUseCase): PullProgressController {
  return async function pullProgressController(since: string): Promise<ProgressPullResponse> {
    try {
      return await useCase(since);
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
