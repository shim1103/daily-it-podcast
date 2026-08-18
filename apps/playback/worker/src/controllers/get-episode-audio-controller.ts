import { mapInternalErrorToExternal } from "./map-internal-error.ts";
import { parseGetEpisodeRequest } from "./parse-get-episode-request.ts";

export type GetEpisodeAudioUseCase = (
  episodeId: string,
) => Promise<Uint8Array>;

export type GetEpisodeAudioController = (
  body: unknown,
) => Promise<Uint8Array>;

/**
 * 音声 byte を返す Controller を組み立てる。
 *
 * @require useCase は確定した episodeId で音声 byte を返す
 * @ensure 戻り関数は unknown を schema 検証し、音声 byte を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createGetEpisodeAudioController(
  useCase: GetEpisodeAudioUseCase,
): GetEpisodeAudioController {
  return async function getEpisodeAudioController(
    body: unknown,
  ): Promise<Uint8Array> {
    const request = parseGetEpisodeRequest(body);
    try {
      return await useCase(request.episodeId);
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
