import { index, integer, primaryKey, real, sqliteTable, text } from "drizzle-orm/sqlite-core";

/**
 * 進捗 D1 表の DDL 正本（Drizzle schema）。migration は `drizzle-kit generate` がここから生成する。
 * 表名・列名の定数は `progress-d1-constants.ts` がここから導く。Port には出さない。
 */
export const episodeProgressTable = sqliteTable(
  "episode_progress",
  {
    episodeId: text("episode_id").notNull(),
    positionSec: real("position_sec").notNull(),
    firstPlayedAt: text("first_played_at").notNull(),
    firstCompletedAt: text("first_completed_at"),
    lastPlayedAt: text("last_played_at").notNull(),
    // why: pull の差分印（cursor）と、行ごとの版（条件付き書込の期待値）を兼ねる
    // todo: 現行 adapter が seq を知らない間だけ、5 列 INSERT を通すために既定値 0 を置く。adapter が seq を採番して書くようになったら外す
    seq: integer("seq").notNull().default(0),
  },
  (table) => [
    // why: 列側の `.primaryKey()` は drizzle-kit(rc) が NOT NULL を出さない。SQLite は非 INTEGER の PK に NULL を許すため、表側の宣言で NOT NULL PRIMARY KEY を出す
    primaryKey({ columns: [table.episodeId] }),
    // why: pull（seq > ?）と採番の最大値取得を全件走査にしないための索引
    // todo: 現行 adapter が seq=0 で複数行を挿入する間は UNIQUE にできない。adapter が seq を採番して書くようになったら UNIQUE を検討する
    index("episode_progress_seq_idx").on(table.seq),
  ],
);
