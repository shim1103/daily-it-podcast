import type { EpisodeRepository } from "../ports/episode-repository.ts";

export async function getEpisodeAudio(
  repository: EpisodeRepository,
  episodeId: string,
): Promise<Uint8Array> {
  return repository.getEpisodeAudio(episodeId);
}
