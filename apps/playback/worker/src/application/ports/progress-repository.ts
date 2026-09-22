import type { EpisodeProgress } from "../../../../contracts/index.ts";

/**
 * 再生進捗の永続 Port。1 episode = 1 行。鍵は `episodeId` のみ（user 分割なし）。
 * merge（先勝ち / 後勝ち）の意味は Decision を正とし、本 Port は読取面の契約を固定する。
 * 書込 signature は後続 A/C で足す。
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
}

/**
 * A 固定用 stub。常に空 Map を返す（未play相当）。behavior は C が置き換える。
 */
export class StubProgressRepository implements ProgressRepository {
  async getByEpisodeIds(
    _episodeIds: readonly string[],
  ): Promise<ReadonlyMap<string, EpisodeProgress>> {
    return new Map();
  }
}
