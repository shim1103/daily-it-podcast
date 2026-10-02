import type {
  ProgressChangeSet,
  ProgressConditionalWriteResult,
  ProgressCursor,
  ProgressRepository,
  ProgressUpdatedEntry,
  ProgressUpsertRow,
  ProgressVersion,
  VersionedProgress,
} from "../../application/ports/progress-repository.ts";

type EpisodeProgress = ProgressUpdatedEntry["progress"];

/**
 * 進捗行をメモリ上の Map に保持して動く `ProgressRepository`。local development・in-memory mode・
 * UseCase test が、状態を持つ代替として使う。
 *
 * @ensure seed／`upsertProgress` した行を episodeId 鍵で保持し、再 upsert は merge せず置き換える
 * @ensure 行が無い episodeId は `getByEpisodeIds` の Map に載せない
 * @ensure `listUpdatedSince` は `lastPlayedAt > since` の行だけを返し、該当なしは空配列
 * @invariant 保持する行は instance ごとに独立し、プロセスを越えて残らない
 */
export class InMemoryProgressRepository implements ProgressRepository {
  private readonly store = new Map<string, EpisodeProgress>();

  constructor(seed: readonly ProgressUpsertRow[] = []) {
    for (const row of seed) {
      this.store.set(row.episodeId, toProgress(row));
    }
  }

  async getByEpisodeIds(
    episodeIds: readonly string[],
  ): Promise<ReadonlyMap<string, EpisodeProgress>> {
    const got = new Map<string, EpisodeProgress>();
    for (const episodeId of episodeIds) {
      const progress = this.store.get(episodeId);
      if (progress !== undefined) {
        got.set(episodeId, progress);
      }
    }
    return got;
  }

  async upsertProgress(row: ProgressUpsertRow): Promise<void> {
    this.store.set(row.episodeId, toProgress(row));
  }

  async listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]> {
    const entries: ProgressUpdatedEntry[] = [];
    for (const [episodeId, progress] of this.store) {
      if (progress.lastPlayedAt > since) {
        entries.push({ episodeId, progress });
      }
    }
    return entries;
  }

  // todo: C が実装（in-memory は単調増加する版の採番）へ置換したら、この zero return と todo を消す
  async getProgressWithVersion(_episodeId: string): Promise<VersionedProgress | null> {
    return null;
  }

  // todo: C が実装（in-memory は単調増加する版の採番）へ置換したら、この zero return と todo を消す
  async upsertProgressIfVersion(
    _row: ProgressUpsertRow,
    _expectedVersion: ProgressVersion | null,
  ): Promise<ProgressConditionalWriteResult> {
    return "conflict";
  }

  // todo: C が実装（in-memory は単調増加する版の採番）へ置換したら、この zero return と todo を消す
  async listChangedAfter(cursor: ProgressCursor): Promise<ProgressChangeSet> {
    return { cursor, entries: [] };
  }
}

function toProgress(row: ProgressUpsertRow): EpisodeProgress {
  return {
    positionSec: row.positionSec,
    firstPlayedAt: row.firstPlayedAt,
    firstCompletedAt: row.firstCompletedAt,
    lastPlayedAt: row.lastPlayedAt,
  };
}
