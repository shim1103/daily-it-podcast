import type { ProgressWriteRequest, ProgressWriteResponse } from "../../../contracts/index.ts";
import type {
  ProgressWriteUseCaseInput,
  ProgressWriteUseCaseOutput,
} from "../application/progress/progress-write-use-case-io.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type ProgressWriteUseCase = (
  input: ProgressWriteUseCaseInput,
) => Promise<ProgressWriteUseCaseOutput>;

export type ProgressWriteController = (
  episodeId: string,
  body: ProgressWriteRequest,
) => Promise<ProgressWriteResponse>;

function toProgressWriteUseCaseInput(
  episodeId: string,
  body: ProgressWriteRequest,
): ProgressWriteUseCaseInput {
  return { episodeId, positionSec: body.positionSec, clientAt: body.clientAt };
}

function toProgressWriteResponse(output: ProgressWriteUseCaseOutput): ProgressWriteResponse {
  return { firstPlayedAt: output.firstPlayedAt, firstCompletedAt: output.firstCompletedAt };
}

/**
 * 進捗 Write（create / update / complete）用 Controller を組み立てる。
 * 操作の違いは渡す use case が担う。本 factory は写像だけを共有する。
 *
 * @require episodeId / body は呼び出し側（Hono route の zValidator）が契約 schema で検証済み
 * @ensure 戻り関数は検証済み入力で ProgressWriteResponse を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createProgressWriteController(
  useCase: ProgressWriteUseCase,
): ProgressWriteController {
  return async function progressWriteController(
    episodeId: string,
    body: ProgressWriteRequest,
  ): Promise<ProgressWriteResponse> {
    try {
      return toProgressWriteResponse(await useCase(toProgressWriteUseCaseInput(episodeId, body)));
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
