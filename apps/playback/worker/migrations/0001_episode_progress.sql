-- 再生進捗 1 episode = 1 行。列の意味は Decision（docs/decisions の再生進捗正本）を参照。
-- 先勝ち / 後勝ちの merge は worker/src/infrastructure/d1/d1-progress-repository.ts が 1 文の upsert で持つ
-- （意味は Decision 2026-09-22T19-13-35）。本 SQL は列・NULL・PK だけを固定する。
CREATE TABLE episode_progress (
  episode_id TEXT NOT NULL PRIMARY KEY,
  position_sec REAL NOT NULL,
  first_played_at TEXT NOT NULL,
  first_completed_at TEXT,
  last_played_at TEXT NOT NULL
);
