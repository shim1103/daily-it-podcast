import type { EpisodeProgress } from "../../../../contracts/index.ts";
import type {
  ProgressRepository,
  ProgressUpdatedEntry,
  ProgressUpsertRow,
} from "./progress-repository.ts";

/**
 * C の UseCase test 向け in-memory Fake。A 契約面へ置かない（Decision 2026-10-01T18:54:14）。
 *
 * @ensure seed / upsert した行を get・list で読める。merge しない
 */
export function createFakeProgressRepository(
  seed: readonly ProgressUpsertRow[] = [],
): ProgressRepository {
  const store = new Map<string, EpisodeProgress>();
  for (const row of seed) {
    store.set(row.episodeId, toProgress(row));
  }

  return {
    async getByEpisodeIds(
      episodeIds: readonly string[],
    ): Promise<ReadonlyMap<string, EpisodeProgress>> {
      const got = new Map<string, EpisodeProgress>();
      for (const episodeId of episodeIds) {
        const progress = store.get(episodeId);
        if (progress !== undefined) {
          got.set(episodeId, progress);
        }
      }
      return got;
    },

    async upsertProgress(row: ProgressUpsertRow): Promise<void> {
      store.set(row.episodeId, toProgress(row));
    },

    async listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]> {
      const entries: ProgressUpdatedEntry[] = [];
      for (const [episodeId, progress] of store) {
        if (progress.lastPlayedAt > since) {
          entries.push({ episodeId, progress });
        }
      }
      return entries;
    },
  };
}

function toProgress(row: ProgressUpsertRow): EpisodeProgress {
  return {
    positionSec: row.positionSec,
    firstPlayedAt: row.firstPlayedAt,
    firstCompletedAt: row.firstCompletedAt,
    lastPlayedAt: row.lastPlayedAt,
  };
}
