import type { EpisodeProgress } from "../../entities/models/episode-progress.ts";
import type {
  GetByEpisodeIdsArgs,
  GetByEpisodeIdsResult,
  GetProgressWithVersionArgs,
  GetProgressWithVersionResult,
  InsertProgressIfAbsentArgs,
  ListChangedAfterArgs,
  ListChangedAfterResult,
  ProgressConditionalWriteResult,
  ProgressRepository,
  ProgressRow,
  ProgressUpdatedEntry,
  ProgressVersion,
  ReplaceProgressIfVersionArgs,
  VersionedProgress,
} from "../../application/ports/progress-repository.ts";

/**
 * 進捗行をメモリ上の Map に保持して動く `ProgressRepository`。local development・in-memory mode・
 * UseCase test が、状態を持つ代替として使う。
 *
 * @ensure seed／`upsertProgress`／`insertProgressIfAbsent`／`replaceProgressIfVersion` で書いた行を episodeId 鍵で保持し、再書込は merge せず置き換える
 * @ensure 行が無い episodeId は `getByEpisodeIds` の Map に載せない
 * @ensure `listUpdatedSince` は `lastPlayedAt > since` の行だけを返し、該当なしは空配列
 * @ensure 書込のたびに行へ新しい版を振る（seed は並び順、旧面の `upsertProgress` も同じ）。版は instance 内で単調増加し、行間で一意
 * @ensure `listChangedAfter` は、版が cursor（非負整数）より大きい行を版昇順（書込順）で返す
 * @invariant 保持する行は instance ごとに独立し、プロセスを越えて残らない
 */
export class InMemoryProgressRepository implements ProgressRepository {
  private readonly store = new Map<string, VersionedProgress>();
  private lastVersion = 0;

  constructor(seed: readonly ProgressRow[] = []) {
    for (const row of seed) {
      this.write(row);
    }
  }

  async getByEpisodeIds(args: GetByEpisodeIdsArgs): Promise<GetByEpisodeIdsResult> {
    const progressByEpisodeId = new Map<string, EpisodeProgress>();
    for (const episodeId of args.episodeIds) {
      const stored = this.store.get(episodeId);
      if (stored !== undefined) {
        progressByEpisodeId.set(episodeId, stored.progress);
      }
    }
    return progressByEpisodeId;
  }

  async upsertProgress(row: ProgressRow): Promise<void> {
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

  async getProgressWithVersion(
    args: GetProgressWithVersionArgs,
  ): Promise<GetProgressWithVersionResult> {
    return this.store.get(args.episodeId) ?? null;
  }

  async insertProgressIfAbsent(
    args: InsertProgressIfAbsentArgs,
  ): Promise<ProgressConditionalWriteResult> {
    if (this.store.has(args.row.episodeId)) {
      return "conflict";
    }
    this.write(args.row);
    return "written";
  }

  async replaceProgressIfVersion(
    args: ReplaceProgressIfVersionArgs,
  ): Promise<ProgressConditionalWriteResult> {
    if (this.store.get(args.row.episodeId)?.version !== args.expectedVersion) {
      return "conflict";
    }
    this.write(args.row);
    return "written";
  }

  async listChangedAfter(args: ListChangedAfterArgs): Promise<ListChangedAfterResult> {
    const changed = [...this.store]
      .filter(([, stored]) => stored.version > args.cursor)
      .sort(([, a], [, b]) => a.version - b.version);
    const last = changed.at(-1);
    return {
      cursor: last === undefined ? args.cursor : last[1].version,
      entries: changed.map(([episodeId, { progress }]) => ({ episodeId, progress })),
    };
  }

  private write(row: ProgressRow): void {
    this.lastVersion += 1;
    this.store.set(row.episodeId, {
      progress: toProgress(row),
      // why: ProgressVersion は Port が宣言した不透明な brand 型で、採番はこの class だけが行う
      version: this.lastVersion as ProgressVersion,
    });
  }
}

function toProgress(row: ProgressRow): EpisodeProgress {
  return {
    positionSec: row.positionSec,
    firstPlayedAt: row.firstPlayedAt,
    firstCompletedAt: row.firstCompletedAt,
    lastPlayedAt: row.lastPlayedAt,
  };
}
