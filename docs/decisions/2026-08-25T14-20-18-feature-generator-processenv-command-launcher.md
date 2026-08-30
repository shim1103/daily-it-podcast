---
name: local AgentSecrets 復元は出口軸で Issue を分離する
date: 2026-08-25T14:20:18
branch: feature/generator-processenv-command-launcher
---

## 1. Decision

1. local AgentSecrets の Composition 結線復元は、**出口軸で Issue を分ける**。command 出口は `commandlaunch.Launcher` 実装と Cursor 専用 project 結線を1 Issue とする。HTTP 出口は `secrettransport.Client` 実装としての AgentSecrets proxy 結線を別 Issue とする。1 Issue に両出口をまとめない。
2. production（processenv）側も出口ごとに Issue を分けた現状を維持する。HTTP production は process-env HTTP transport Issue、command production は process-env command launcher（完了）とする。local と production を同一 Issue に混ぜない。
3. 依存関係は次を正とする。
   1. process-env command launcher（完了）→ local AgentSecrets command launcher。command 契約と Cursor Adapter の `Launcher` 依存が先にある。
   2. process-env HTTP transport → local AgentSecrets HTTP transport。HTTP Adapter が `secrettransport.Client` に依存してから、local 実装を差し込める。Adapter 切替と processenv 実装は production HTTP Issue が所有する。
   3. local command Issue と local HTTP Issue は**互いに blocked しない**。出口契約が違い、Cursor 専用 project は command 側だけの知識である。
4. 境界・契約・2軸の正準は既存 Decision を SSOT とする（本 Decision は Issue 分割と依存だけを追加する）。`2026-08-22T18-35-00`、`2026-08-22T18-44-00`、`2026-08-22T11-55-22`、`2026-08-22T15-08-00`、`2026-08-25T13-53-55`。

## 2. Reason

1. 出口が違うと契約・失敗モード・検証境界が違う。command は child process / project cwd / allowlist、HTTP は request Inject / proxy。1 Issue にすると AC と Narrow Integration が混線し、close 判定が曖昧になる（`philosophy` §3-1 S）。
2. 2軸モデル（置き場 × 出口）では、同じ置き場（AgentSecrets）でも出口ごとに実装スロットが別である。Issue を置き場だけで束ねると、出口直交性が Issue 境界で崩れる（`2026-08-25T13-53-55`）。
3. production HTTP で Adapter を `secrettransport` へ切替えないまま local HTTP だけ結線すると、Adapter が具象 `agentsecrets.Client` のまま残り、DIP が完成しない。local HTTP は production HTTP の Adapter 切替後に置く。
4. local command は Cursor 専用 project という HTTP に無い制約を持つ。HTTP Issue へ混ぜると Out of Scope が肥大し、専用 project 決定（`2026-08-22T15-08-00`）の検証が薄くなる。

## 3. Rejected

1. local command と local HTTP を1 Issue にまとめる案。出口・契約・AC が同居し、並列実装と review 分割ができない。
2. local AgentSecrets 全体を process-env HTTP / command Issue の追記 scope にする案。production と local の完了条件が混ざり、production gate を local keychain 前提にしやすい。
3. local HTTP を production HTTP より先に単独完了扱いにする案。Adapter が `secrettransport` 未依存のままでは「結線復元」が最終系に届かない。
