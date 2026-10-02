-- 再生進捗 1 episode = 1 行。列の意味は Decision（docs/decisions の再生進捗正本）を参照。
-- 先勝ち / 後勝ちの merge は Application（worker/src/application/progress/merge-progress.ts）が持ち、
-- D1 adapter は渡された行を保存する（意味は Decision 2026-09-22T19-13-35、置き場は 2026-10-01T18-54-14）。
-- 時刻列は契約境界で UTC 固定幅の ISO 文字列へ正規化して保存する（Decision 2026-10-01T18-54-16）。
-- 本 SQL は列・NULL・PK だけを固定する。
CREATE TABLE episode_progress (
  episode_id TEXT NOT NULL PRIMARY KEY,
  position_sec REAL NOT NULL,
  first_played_at TEXT NOT NULL,
  first_completed_at TEXT,
  last_played_at TEXT NOT NULL
);
