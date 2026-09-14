---
name: mp3 移行の encode / Drive purge 入口を追加した
date: 2026-09-14T17:22:42
session_id: audio-mp3-cutover-migrate-pr-completion
branch: chore/audio-mp3-cutover-migrate
prev: なし
---

## 1. Summary

`.cache` wav→本番 Encoder→mp3 の local 入口と、Drive wav 削除＋完成ペア検証の `workflow_dispatch` を追加した。使い捨て実装は `hack/` に置き、Issue / DEPLOY へ入口 path を寄せた。本番 Drive の人手移行と同着 deploy 自体は未実行。

## 2. Changes

1. 変換は shell argv 複製をやめ、既存 `EncodeWAVToMP3` へ bytes を渡す形にした
2. purge+verify は `target=prod|test` で Drive OAuth 4 key だけを読む。完成ペアは同 stem の json+mp3 以外を拒否する
3. GWT を case 事実付きに揃え、shell test は `test-*-sociable-unit.shell` 命名にした

### Commits

- `dd09434`
- `d4325f7`
- `f0eed3d`
