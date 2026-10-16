import type { EpisodeId } from "../../../../domain/index.ts";
import { EpisodeContentError } from "../../entities/errors/episode-content-error.ts";
import type { EpisodeAudio } from "../../entities/models/episode-audio.ts";
import { verifyAudio } from "../audio/verify-audio.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";

export type GetAudioUseCaseInput = { readonly episodeId: EpisodeId };

export type GetAudioUseCaseOutput = EpisodeAudio;

/**
 * 対象 episodeId の mp3 byte を返す。
 *
 * Port は取得したままの byte か undefined を返す。不在と内容の判定はこの use-case が行う。
 *
 * @ensure mp3 が無い時、および内容が `verifyAudio` に不適合な時は {@link EpisodeContentError} を throw する。
 */
export async function getAudio(
  repository: EpisodeRepository,
  input: GetAudioUseCaseInput,
): Promise<GetAudioUseCaseOutput> {
  const audio = await repository.getAudio({ episodeId: input.episodeId });
  if (audio === undefined) {
    throw new EpisodeContentError(`音声が無い: ${input.episodeId}`);
  }
  return verifyAudio(audio, input.episodeId);
}
