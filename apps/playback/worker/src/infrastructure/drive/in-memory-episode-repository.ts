import type {
  EpisodeListItem,
  EpisodeManuscript,
  EpisodeRepository,
} from "../../application/ports/episode-repository.ts";
import { EpisodeNotFoundError } from "../../entities/errors/episode-not-found-error.ts";
import { ManuscriptSchema } from "./manuscript-schema.ts";

type StoredEpisode = {
  json: unknown;
  audio?: Uint8Array;
};

export class InMemoryEpisodeRepository implements EpisodeRepository {
  private readonly episodes = new Map<string, StoredEpisode>();

  put(episodeId: string, json: unknown, audio?: Uint8Array): void {
    this.episodes.set(episodeId, { json, audio });
  }

  async listEpisodes(): Promise<EpisodeListItem[]> {
    const items: EpisodeListItem[] = [];

    for (const [stem, entry] of this.episodes) {
      const parsed = ManuscriptSchema.safeParse(entry.json);
      if (!parsed.success || parsed.data.episodeId !== stem) {
        continue;
      }
      const { body: _body, ...listItem } = parsed.data;
      items.push(listItem);
    }

    return items;
  }

  async getEpisode(episodeId: string): Promise<EpisodeManuscript> {
    const entry = this.episodes.get(episodeId);
    if (entry === undefined) {
      throw new EpisodeNotFoundError(`JSON エントリが無い: ${episodeId}`);
    }

    const parsed = ManuscriptSchema.safeParse(entry.json);
    if (!parsed.success) {
      throw new EpisodeNotFoundError(`原稿 JSON が schema に不適合: ${episodeId}`);
    }
    if (parsed.data.episodeId !== episodeId) {
      throw new EpisodeNotFoundError(`stem と JSON の episodeId が不一致: ${episodeId}`);
    }
    if (entry.audio === undefined) {
      throw new EpisodeNotFoundError(`wav が無い: ${episodeId}`);
    }

    return parsed.data;
  }

  async getEpisodeAudio(episodeId: string): Promise<Uint8Array> {
    const entry = this.episodes.get(episodeId);
    if (entry === undefined) {
      throw new EpisodeNotFoundError(`音声エントリが無い: ${episodeId}`);
    }
    if (entry.audio === undefined) {
      throw new EpisodeNotFoundError(`音声バイトが無い: ${episodeId}`);
    }
    return entry.audio;
  }
}
