---
name: WAV→mp3 の Port・runtime 工場を置き encode 本体は残した
date: 2026-09-13T18:17:08
session_id: 99702751-d756-444a-b5c2-b74cc65c6da1
branch: feature/generator-wav-to-mp3-port-runtime
prev: なし
---

## 1. Summary

Application が `os/exec` を持たないよう WAV→mp3 を Port + ffmpeg Adapter stub にし、OS/HTTP 工場を `internal/runtime` へ移して Composition は結線だけにした。ffmpeg 本実装と `ProduceEpisode.Run` 内 encode 呼び出しは Issue に残す。

## 2. Changes

- UseCase が `exec` を直持ちする案を shim 訂正で却下し、Port DI に切り替えた。
- OS 工場を infra や Composition に置く案を却下し、`internal/runtime` と `infra/audio/ffmpeg` を分離した。
- aggregate coverage 緑だけで unit 到達とみなす誤りと、GWT comment 欠落を指摘後に直した。
- `docs/lessons/index.md` は `origin/develop` を正本にし、本 session 分だけ末尾追記（branch 全体の rebase はしない）。
- 検証: commit hook 経由で generator static/unit、playback format/lint/tsc/layers 緑。

### Commits

- `13a477d`
- `a7eb79d`
- `986760b`
