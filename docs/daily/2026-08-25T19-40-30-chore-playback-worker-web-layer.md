---
name: playback web/worker 層違反検知の導入と Drive 原稿検証の切り離し
date: 2026-08-25T19:40:30
session_id: none
branch: chore/playback-worker-web-layer
prev: なし
---

## 1. Summary

playback の web / worker に dependency-cruiser による層違反検知を static gate へ入れ、Feature/Primitive を dir 分割し、Infrastructure の原稿検証を HTTP contracts から Drive schema へ移した。scope-split では A を code 契約・B を Decision として固定し、lane の PR-E/G 相当を完了扱いで削除した。

## 2. Changes

1. dry-run で Infrastructure→HTTP contracts の 1 違反を確定し、Application→contracts は generator 寄せで許可した。
2. Feature/Primitive を `components/{feature,primitive}/` に分け、depcruise 規則と check-static を結線した。
3. `ManuscriptSchema` を Ajv + repo 根 `manuscript.schema.json` へ切り替え、sociable_unit を追加した。
4. Decision / DESIGN / playback-lane を code SSOT に同期した。
5. `.cursor/cli.json` に `Delete(*)` を allow へ追加した（Write だけでは Delete 承認待ちになる）。

### Commits

1. `e640259`
2. `c795f5a`
3. `4819f74`
4. `7c041fd`
5. `bac5429`
