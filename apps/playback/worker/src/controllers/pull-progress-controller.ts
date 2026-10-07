import type { ProgressPullQuery, ProgressPullResponse } from "../../../contracts/index.ts";
import type {
  PullProgressUseCaseInput,
  PullProgressUseCaseOutput,
} from "../application/use-cases/pull-progress.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type PullProgressUseCase = (
  input: PullProgressUseCaseInput,
) => Promise<PullProgressUseCaseOutput>;

export type PullProgressController = (query: ProgressPullQuery) => Promise<ProgressPullResponse>;

function toPullProgressUseCaseInput(query: ProgressPullQuery): PullProgressUseCaseInput {
  return { since: query.since };
}

function toProgressPullResponse(output: PullProgressUseCaseOutput): ProgressPullResponse {
  return {
    episodes: output.episodes.map((entry) => ({
      episodeId: entry.episodeId,
      progress: {
        positionSec: entry.progress.positionSec,
        firstPlayedAt: entry.progress.firstPlayedAt,
        firstCompletedAt: entry.progress.firstCompletedAt,
        lastPlayedAt: entry.progress.lastPlayedAt,
      },
    })),
  };
}

/**
 * 進捗 pull を返す Controller を組み立てる。
 *
 * @require query は呼び出し側（Hono route の zValidator）が契約 schema で検証済み
 * @ensure 戻り関数は契約の ProgressPullResponse を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createPullProgressController(useCase: PullProgressUseCase): PullProgressController {
  return async function pullProgressController(
    query: ProgressPullQuery,
  ): Promise<ProgressPullResponse> {
    try {
      return toProgressPullResponse(await useCase(toPullProgressUseCaseInput(query)));
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
