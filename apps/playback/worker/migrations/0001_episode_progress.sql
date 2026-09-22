-- 再生進捗 1 episode = 1 行。列の意味は Decision（docs/decisions の再生進捗正本）を参照。
-- 先勝ち / 後勝ちの merge は Application。本 SQL は列・NULL・PK だけを固定する。
CREATE TABLE episode_progress (
  episode_id TEXT NOT NULL PRIMARY KEY,
  position_sec REAL NOT NULL,
  first_played_at TEXT NOT NULL,
  first_completed_at TEXT,
  last_played_at TEXT NOT NULL
);
