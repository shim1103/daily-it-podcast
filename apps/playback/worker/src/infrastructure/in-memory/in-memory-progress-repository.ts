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
 * @ensure seed／`upsertProgress`／`upsertProgressIfVersion` で書いた行を episodeId 鍵で保持し、再書込は merge せず置き換える
 * @ensure 行が無い episodeId は `getByEpisodeIds` の Map に載せない
 * @ensure `listUpdatedSince` は `lastPlayedAt > since` の行だけを返し、該当なしは空配列
 * @ensure 書込のたびに行へ新しい版を振る（seed は並び順、旧面の `upsertProgress` も同じ）。版は instance 内で単調増加し、行間で一意
 * @ensure `listChangedAfter` の cursor は 10 進非負整数の文字列で、版がそれより大きい行を版昇順（書込順）で返す
 * @invariant 保持する行は instance ごとに独立し、プロセスを越えて残らない
 */
export class InMemoryProgressRepository implements ProgressRepository {
  private readonly store = new Map<string, VersionedProgress>();
  private lastVersion = 0;

  constructor(seed: readonly ProgressUpsertRow[] = []) {
    for (const row of seed) {
      this.write(row);
    }
  }

  async getByEpisodeIds(
    episodeIds: readonly string[],
  ): Promise<ReadonlyMap<string, EpisodeProgress>> {
    const got = new Map<string, EpisodeProgress>();
    for (const episodeId of episodeIds) {
      const stored = this.store.get(episodeId);
      if (stored !== undefined) {
        got.set(episodeId, stored.progress);
      }
    }
    return got;
  }

  async upsertProgress(row: ProgressUpsertRow): Promise<void> {
    // why: 旧面の書込でも版を進める。経路ごとに版が進まないと、版を期待値にする書込が旧面の書込を見落とす
    this.write(row);
  }

  async listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]> {
    const entries: ProgressUpdatedEntry[] = [];
    for (const [episodeId, { progress }] of this.store) {
      if (progress.lastPlayedAt > since) {
        entries.push({ episodeId, progress });
      }
    }
    return entries;
  }

  async getProgressWithVersion(episodeId: string): Promise<VersionedProgress | null> {
    return this.store.get(episodeId) ?? null;
  }

  async upsertProgressIfVersion(
    row: ProgressUpsertRow,
    expectedVersion: ProgressVersion | null,
  ): Promise<ProgressConditionalWriteResult> {
    const currentVersion = this.store.get(row.episodeId)?.version ?? null;
    if (currentVersion !== expectedVersion) {
      return "conflict";
    }
    this.write(row);
    return "written";
  }

  async listChangedAfter(cursor: ProgressCursor): Promise<ProgressChangeSet> {
    const afterVersion = Number(cursor);
    const changed = [...this.store]
      .filter(([, stored]) => stored.version > afterVersion)
      .sort(([, a], [, b]) => a.version - b.version);
    const last = changed.at(-1);
    return {
      cursor: last === undefined ? cursor : String(last[1].version),
      entries: changed.map(([episodeId, { progress }]) => ({ episodeId, progress })),
    };
  }

  private write(row: ProgressUpsertRow): void {
    this.lastVersion += 1;
    this.store.set(row.episodeId, {
      progress: toProgress(row),
      // why: ProgressVersion は Port が宣言した不透明な brand 型で、採番はこの class だけが行う
      version: this.lastVersion as ProgressVersion,
    });
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
