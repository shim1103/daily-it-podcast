import type { EpisodeId } from "../../../domain/index.ts";
import type {
  GetAudioUseCaseInput,
  GetAudioUseCaseOutput,
} from "../application/use-cases/get-audio.ts";
import type { EpisodeAudio } from "../entities/models/episode-audio.ts";
import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type GetAudioUseCase = (input: GetAudioUseCaseInput) => Promise<GetAudioUseCaseOutput>;

export type GetAudioController = (episodeId: EpisodeId) => Promise<EpisodeAudio>;

/**
 * 音声 byte を返す Controller を組み立てる。
 *
 * @require episodeId は呼び出し側（Hono route の zValidator）が契約 schema で検証済み
 * @ensure 戻り関数は検証済み episodeId で音声 byte を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createGetAudioController(useCase: GetAudioUseCase): GetAudioController {
  return async function getAudioController(episodeId: EpisodeId): Promise<EpisodeAudio> {
    try {
      return await useCase({ episodeId });
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
