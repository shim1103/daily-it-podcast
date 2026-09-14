import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";

/**
 * 対象 episodeId の mp3 byte を返す。
 *
 * Port は mp3 byte か「無し（undefined）」を返すだけなので、不在の Domain Error 化はこの
 * use-case が行う。
 *
 * @ensure mp3 が無い時は {@link EpisodeContentError} を throw する。
 */
export async function getAudio(
  repository: EpisodeRepository,
  episodeId: string,
): Promise<Uint8Array> {
  const audio = await repository.getAudio(episodeId);
  if (audio === undefined) {
    throw new EpisodeContentError(`音声が無い: ${episodeId}`);
  }
  return audio;
}
