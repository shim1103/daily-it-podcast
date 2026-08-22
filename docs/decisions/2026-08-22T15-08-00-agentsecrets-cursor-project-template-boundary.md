---
name: Cursor 専用 AgentSecrets project と repo の env template を分離する
date: 2026-08-22T15:08:00
branch: develop
---

## 1. Decision

1. Cursor CLI の local secret 注入は、`~/.agentsecrets-projects/cursor/.agentsecrets/project.json` が指す Cursor 専用 AgentSecrets project だけを正とする。この project は `CURSOR_API_KEY` だけを持つ。
2. repo root の `.env.example` は AgentSecrets project の選択や secret 注入の allowlist として扱わない。`agentsecrets env --` は、実行時 current directory 直下の `.agentsecrets/project.json` が指す active project の全 secret を注入する。
3. したがって repo root の `.env.example` に `CURSOR_API_KEY` を載せない。Cursor 専用 project が repo 外にあることと、Cursor の注入境界は `docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md` を正とする。

## 2. Reason

1. 実測で `~/.agentsecrets-projects/cursor` からの `agentsecrets env --` は `CURSOR_API_KEY` だけを注入し、Generator の Cursor Adapter も同 dir を working directory に指定する。一方、repo root の `.agentsecrets/project.json` は `daily-it-podcast` project を指す。root `.env.example` の記載は、この project 選択を変えない。
2. `.env.example` と project config は責務が対照的である。前者は人間が必要な env 名を把握するための template、後者は AgentSecrets が runtime で解決・注入する secret 集合を決める設定である。同じ `CURSOR_API_KEY` を root template と repo 外 project の両方で表すと、どちらが注入範囲の正か不明になる。これは DRY と Least Astonishment に反する（`philosophy` §2-2、§4-5）。
3. Cursor CLI へ必要最小の secret だけを渡すには、専用 project の active project 化が必要である。root template の記載で境界が制御されるように見せると、実際の権限境界と documentation が乖離する。これは Least Privilege に反する（`philosophy` §4-3）。

## 3. Rejected

1. repo root の `.env.example` を Cursor 専用 project の secret inventory として残す案。AgentSecrets の active project を決めないため、注入範囲の SSOT にならない。
2. Cursor 専用 project の `.agentsecrets/project.json` を repo 内へ移す案。clone 状態と secret 境界が結び付き、project dir が無い環境で wrapper の境界が静かに外れる。
3. `.env.example` へ全 project の secret 名を再び列挙する案。runtime ごとの secret 契約を repo 共通 schema へ戻し、`2026-08-19T17-37-00-playback-runtime-secret-boundary.md` の Decision 1 と矛盾する。
