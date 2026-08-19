import { episodeAudioPath, type GetEpisodeResponse } from "../../../../contracts/index.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";

export async function getEpisode(
  repository: EpisodeRepository,
  episodeId: string,
): Promise<GetEpisodeResponse> {
  const manuscript = await repository.getEpisode(episodeId);
  return {
    ...manuscript,
    audioRef: episodeAudioPath(episodeId),
  };
}
