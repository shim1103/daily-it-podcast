-- 更新印 seq を足す。seq は pull の差分印（cursor）と、行ごとの版（条件付き書込の期待値）を兼ねる。
-- seq の採番は、書込と同じ文の中で DB が行う（正は docs/decisions/2026-10-01T23-32-40-feature-playback-progress-d1-local-peer.md）。
-- DEFAULT 0 は、SQLite の ADD COLUMN ... NOT NULL が既定値を要求するため、かつ seq を知らない現行 adapter の INSERT を通すため。
-- backfill は、既存行が 0 のままだと cursor "0" の seq > 0 から落ちるため、rowid で一意に埋める。
-- 索引は pull（seq > ?）と採番の最大値取得を全件走査にしないため。UNIQUE にしないのは、現行 adapter が seq=0 で複数行を挿入する間、UNIQUE 違反になるため。
ALTER TABLE episode_progress ADD COLUMN seq INTEGER NOT NULL DEFAULT 0;
UPDATE episode_progress SET seq = rowid;
CREATE INDEX episode_progress_seq_idx ON episode_progress (seq);
