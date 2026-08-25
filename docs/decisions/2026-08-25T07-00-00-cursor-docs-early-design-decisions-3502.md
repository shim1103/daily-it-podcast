---
name: 個人・非 Workspace 前提のため GAS 一本化 runtime は採らない
date: 2026-08-25T07:00:00
branch: cursor/docs-early-design-decisions-3502
---

## 1. Decision

1. Playback / Generator を Google Apps Script（HTML Service + clasp）へ一本化しない。
2. 現行の正は Generator（Go + GHA）と Playback（Vite + Workers）、共有は個人 Drive 上の `contracts/` のみ、とする（`2026-08-15T16-23-06`）。
3. Google Workspace のドメイン ACL を「従業員だけ」公開の SSOT にする前提は、この製品では置かない（運用者は個人であり Workspace を使わない）。

## 2. Reason

1. GAS 案の強みは Workspace 内 identity でのアクセス制御と Spreadsheet/Drive 密結合である。Workspace が無い個人利用では、その強みは「この Google アカウントだけ」に縮み、Access「この email だけ」と同型の別実装になる。
2. Generator は Cursor CLI・Go・長時間 TTS・GHA cron を必要とし、GAS 実行モデルに載らない。UI だけ GAS に移しても台は減らず、入口が GitHub + GAS に増える。
3. HTML Service は現行の Vite + `<audio>` Playback の置き換えとして弱い。wav 配信・Clean Arch / test gate も GAS 側へ移せない。
4. Drive 成果物の SSOT は既に `contracts/` にある。GAS 化は成果物 SSOT を増やさない一方、runtime の癒着で 2 系統の直交性を壊しやすい。

## 3. Rejected

1. Playback を GAS Web App「自分のみ」へ移し、Workers / Access を捨てる案
2. Generator まで Apps Script に寄せる案
3. 「社内向け GAS + clasp + AI 編集」を本 repo の標準形にする案（別 problem の解）
