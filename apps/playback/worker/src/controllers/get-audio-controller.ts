import { mapInternalErrorToExternal } from "./map-internal-error.ts";

export type GetAudioUseCase = (episodeId: string) => Promise<Uint8Array>;

export type GetAudioController = (episodeId: string) => Promise<Uint8Array>;

/**
 * 音声 byte を返す Controller を組み立てる。
 *
 * @require episodeId は呼び出し側（Hono route の zValidator）が契約 schema で検証済み
 * @ensure 戻り関数は検証済み episodeId で音声 byte を返す。Internal は External に変換して throw する
 * @invariant HTTP status と Response object を作らない
 */
export function createGetAudioController(useCase: GetAudioUseCase): GetAudioController {
  return async function getAudioController(episodeId: string): Promise<Uint8Array> {
    try {
      return await useCase(episodeId);
    } catch (error) {
      throw mapInternalErrorToExternal(error);
    }
  };
}
