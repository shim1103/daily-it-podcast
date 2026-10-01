import type {
  EpisodeProgress,
  ProgressPullResponse,
  ProgressWriteRequest,
  ProgressWriteResponse,
} from "../../../../contracts/index.ts";

/** 進捗 Write（create/update/complete）の永続入力。HTTP body に episodeId を足した形。 */
export type ProgressWriteCommand = ProgressWriteRequest & {
  readonly episodeId: string;
};

/** pull 応答 1 件。契約の pull episodes 要素と同形。 */
export type ProgressUpdatedEntry = ProgressPullResponse["episodes"][number];

/**
 * 再生進捗の永続 Port。1 episode = 1 行。鍵は `episodeId` のみ（user 分割なし）。
 * merge（先勝ち / 後勝ち）の意味は Decision を正とし、本 Port は永続面の入出力契約を固定する。
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
   * 進捗行を create または update する（HTTP create/update の永続側）。
   *
   * @ensure 永続後の勝ち側 `firstPlayedAt` / `firstCompletedAt` を返す
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  writeProgress(command: ProgressWriteCommand): Promise<ProgressWriteResponse>;

  /**
   * 完走を記録する（HTTP complete の永続側）。
   *
   * @ensure 永続後の勝ち側 `firstPlayedAt` / `firstCompletedAt` を返す
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  completeProgress(command: ProgressWriteCommand): Promise<ProgressWriteResponse>;

  /**
   * `since` より後に更新された進捗行を返す（HTTP pull の永続側）。
   *
   * @ensure 該当なしは空配列（null でない）
   * @ensure storage I/O 失敗は Infrastructure Error を throw する
   */
  listUpdatedSince(since: string): Promise<readonly ProgressUpdatedEntry[]>;
}

/**
 * in-memory／unit 用の代替 `ProgressRepository`。何も永続しない。
 *
 * @ensure 読取は空、書込は zero value、pull は空配列を返す
 * @invariant 本番経路（r2 mode）の Composition Root は選ばない（永続されない応答を HTTP 成功で返さない）
 */
export class StubProgressRepository implements ProgressRepository {
  async getByEpisodeIds(
    _episodeIds: readonly string[],
  ): Promise<ReadonlyMap<string, EpisodeProgress>> {
    return new Map();
  }

  async writeProgress(_command: ProgressWriteCommand): Promise<ProgressWriteResponse> {
    // what: 永続しない代替が返す epoch。契約 schema の firstPlayedAt を満たすためだけの sentinel
    return { firstPlayedAt: "1970-01-01T00:00:00.000Z", firstCompletedAt: null };
  }

  async completeProgress(_command: ProgressWriteCommand): Promise<ProgressWriteResponse> {
    return { firstPlayedAt: "1970-01-01T00:00:00.000Z", firstCompletedAt: null };
  }

  async listUpdatedSince(_since: string): Promise<readonly ProgressUpdatedEntry[]> {
    return [];
  }
}
