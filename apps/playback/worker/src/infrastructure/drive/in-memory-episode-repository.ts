import type {
  EpisodeRepository,
  RawManuscriptDetail,
  RawManuscriptEntry,
} from "../../application/ports/episode-repository.ts";

type StoredEpisode = {
  json: unknown;
  audio?: Uint8Array;
};

/**
 * local development / unit test 用に、原稿 json / wav をメモリ上に保持する `EpisodeRepository`。
 *
 * 真の外部境界（ここでは Map への格納・取り出し）だけを担い、schema 適合・stem 一致・不正 JSON・
 * wav 欠落の判定はしない。判定は use-case（`application/use-cases/*`）が行う。
 */
export class InMemoryEpisodeRepository implements EpisodeRepository {
  private readonly episodes = new Map<string, StoredEpisode>();

  put(episodeId: string, json: unknown, audio?: Uint8Array): void {
    this.episodes.set(episodeId, { json, audio });
  }

  async listManuscripts(): Promise<RawManuscriptEntry[]> {
    const entries: RawManuscriptEntry[] = [];
    for (const [stem, entry] of this.episodes) {
      entries.push({ stem, json: entry.json });
    }
    return entries;
  }

  async getManuscript(episodeId: string): Promise<RawManuscriptDetail | undefined> {
    const entry = this.episodes.get(episodeId);
    if (entry === undefined) {
      return undefined;
    }
    return { json: entry.json, hasAudio: entry.audio !== undefined };
  }

  async getEpisodeAudio(episodeId: string): Promise<Uint8Array | undefined> {
    return this.episodes.get(episodeId)?.audio;
  }
}
