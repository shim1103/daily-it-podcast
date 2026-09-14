---
name: ffmpeg Encoder 本実装と ProduceEpisode encode 結線を完了した
date: 2026-09-14T12:11:08
session_id: none
branch: feature/generator-audio-mp3-encode-write
prev: なし
---

## 1. Summary

WAV→MP3 の ffmpeg Adapter を本実装し、`ProduceEpisode` が ConcatWAV 直後に Port 経由で encode してから WriteEpisode へ mp3 を渡す状態にした。失敗時 progress の Start/Done 再検証は検出力が薄いため削った。

## 2. Changes

- unit/broad は fake LookPath/Run。実 ffmpeg bitstream の Narrow/System 観測は未実施。
- 検証: `scripts/generator/check-static.sh` と `scripts/generator/test-unit.sh` 緑（coverage 91.9%）。commit hook 経由でも generator/playback gate 緑。
- shim 指摘: 失敗時「Start あり Done なし」は early return の構造帰結であり、共通 helper も無いので専用 case は不要。存在確認ではなく検出力で判断する。

### Commits

- `41dea8e`
- `cc4cd30`
- `405e7dd`
- `db5974b`
- `f93e756`
- `9120905`
