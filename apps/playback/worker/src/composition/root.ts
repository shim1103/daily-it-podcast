import { getEpisode } from "../application/use-cases/get-episode.ts";
import { getEpisodeAudio } from "../application/use-cases/get-episode-audio.ts";
import { listEpisodes } from "../application/use-cases/list-episodes.ts";
import { createGetEpisodeAudioController } from "../controllers/get-episode-audio-controller.ts";
import { createGetEpisodeController } from "../controllers/get-episode-controller.ts";
import { createListEpisodesController } from "../controllers/list-episodes-controller.ts";
import { InMemoryEpisodeRepository } from "../infrastructure/drive/in-memory-episode-repository.ts";

const repository = new InMemoryEpisodeRepository();

export const listEpisodesController = createListEpisodesController(() =>
  listEpisodes(repository),
);

export const getEpisodeController = createGetEpisodeController((episodeId) =>
  getEpisode(repository, episodeId),
);

export const getEpisodeAudioController = createGetEpisodeAudioController(
  (episodeId) => getEpisodeAudio(repository, episodeId),
);
