import type { ListEpisodesResponse } from "../../../../contracts/index.ts";
import type { EpisodeRepository } from "../ports/episode-repository.ts";

export async function listEpisodes(repository: EpisodeRepository): Promise<ListEpisodesResponse> {
  const episodes = await repository.listEpisodes();
  return { episodes };
}
