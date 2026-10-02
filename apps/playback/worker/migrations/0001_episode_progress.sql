-- 再生進捗 1 episode = 1 行。列の意味は Decision（docs/decisions の再生進捗正本）を参照。
-- 先勝ち / 後勝ちの merge は Application（worker/src/application/progress/merge-progress.ts）が持ち、
-- D1 adapter は渡された行を保存する（意味は Decision 2026-09-22T19-13-35、置き場は 2026-10-01T18-54-14）。
-- 時刻列は契約境界で UTC 固定幅の ISO 文字列へ正規化して保存する（Decision 2026-10-01T18-54-16）。
-- seq は pull の差分印（cursor）と、行ごとの版（条件付き書込の期待値）を兼ねる。
-- seq の採番は、書込と同じ文の中で DB が行う（正は docs/decisions/2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md）。
-- seq の DEFAULT 0 は、seq を知らない現行 adapter の 5 列 INSERT を、置き換えるまで通すため。
-- seq の索引は pull（seq > ?）と採番の最大値取得を全件走査にしないため。UNIQUE にしないのは、現行 adapter が seq=0 で複数行を挿入する間、UNIQUE 違反になるため。
-- 本 SQL は列・NULL・PK・索引だけを固定する。
CREATE TABLE episode_progress (
  episode_id TEXT NOT NULL PRIMARY KEY,
  position_sec REAL NOT NULL,
  first_played_at TEXT NOT NULL,
  first_completed_at TEXT,
  last_played_at TEXT NOT NULL,
  seq INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX episode_progress_seq_idx ON episode_progress (seq);
