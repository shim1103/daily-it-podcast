import type { GetEpisodeResponse, ListEpisodesResponse } from "../../../../contracts/index.ts";

export type EpisodeListItem = ListEpisodesResponse["episodes"][number];
export type EpisodeManuscript = Omit<GetEpisodeResponse, "audioRef">;

export interface EpisodeRepository {
  listEpisodes(): Promise<EpisodeListItem[]>;
  getEpisode(episodeId: string): Promise<EpisodeManuscript>;
  getEpisodeAudio(episodeId: string): Promise<Uint8Array>;
}
