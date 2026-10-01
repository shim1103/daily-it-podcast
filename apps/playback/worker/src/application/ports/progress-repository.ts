import type { EpisodeProgress, ProgressPullResponse } from "../../../../contracts/index.ts";

/** pull 応答 1 件。契約の pull episodes 要素と同形。 */
export type ProgressUpdatedEntry = ProgressPullResponse["episodes"][number];

/** upsert 1 行。`episodeId` + 進捗本体。merge せずそのまま永続する。 */
export type ProgressUpsertRow = { readonly episodeId: string } & EpisodeProgress;

/**
 * 再生進捗の永続 Port。1 episode = 1 行。鍵は `episodeId` のみ（user 分割なし）。
 * merge・冪等・完走判定は Application。本 Port は永続能力のみ（Decision 2026-10-01T18:54:14）。
 *
 * @invariant vendor / D1 固有型を露出しない
 */
export interface ProgressRepository {
  /**
   * 指定 episode の進捗行を一括取得する。
   *
   * @ensure 行が無い episodeId は Map に載せない（caller が `progress: null` にする）
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  getByEpisodeIds(episodeIds: readonly string[]): Promise<ReadonlyMap<string, EpisodeProgress>>;

  /**
   * 進捗行をそのまま永続する（create / update の区別なし。merge しない）。
   *
   * @ensure 渡された row を episodeId 鍵で置き換える。応答を返さない
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  upsertProgress(row: ProgressUpsertRow): Promise<void>;

  /**
   * `since` より後に更新された進捗行を返す（HTTP pull の永続側）。
   *
   * @ensure `lastPlayedAt > since` の行だけを返す
   * @ensure 該当なしは空配列（null でない）
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]>;
}

/**
 * Composition 用 zero Stub。読取は空、upsert は no-op、pull は空配列。
 * behavior 検証は Fake（C の test support）が持つ。
 */
export class StubProgressRepository implements ProgressRepository {
  async getByEpisodeIds(
    _episodeIds: readonly string[],
  ): Promise<ReadonlyMap<string, EpisodeProgress>> {
    return new Map();
  }

  async upsertProgress(_row: ProgressUpsertRow): Promise<void> {}

  async listUpdatedSince(_since: string): Promise<readonly ProgressUpdatedEntry[]> {
    return [];
  }
}
