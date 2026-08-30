---
name: Generator の GHA 保存名は process env 名と同一、test 用だけ TEST_ 接頭
date: 2026-08-30T12:49:00
branch: docs/generator-broad-system-e2e-plan
---

## 1. Decision

1. Generator が読む process environment の key は `apps/generator/internal/config/names.go` を正とする。`GENERATOR_` 接頭は付けない。
2. GitHub Actions の **本番** Secrets / Variables の名前は、上記 process env 名と同一とする。
3. GitHub Actions の **test** Secrets / Variables の名前は、同一語幹に `TEST_` 接頭を付ける（例: process `GETX_API_KEY` ← Secret `TEST_GETX_API_KEY`）。
4. workflow が test 用保存名を process env 名へ写す。application / config は `TEST_` を知らない。
5. Secret と Variable の区分は先行 Decision（`docs/decisions/2026-08-27T12-17-01-docs-env-secret-management-reconsider.md`）を維持する。

## 2. Reason

1. configuration boundary は既に process env 名で契約している。GHA 保存名を別語彙（`GENERATOR_` 等）にすると、inventory・workflow・config の3箇所で名前が分岐する。
2. test と本番を同じ GitHub 名で共有すると流用禁止を破る。接頭 `TEST_` だけを GHA 保存層に置けば、process 契約は変えずに隔離できる。
3. application に `TEST_` を読ませると、本番 path と test path で Load 契約が二重になる。写像は workflow（caller）の責務で足りる。

## 3. Rejected

1. 全 key に `GENERATOR_` を付ける案 — config 契約とズレ、既存 inventory を壊す。
2. test も本番と同じ GHA 名を使う案 — 本番値の test 流用が起きやすい。
3. config に `TEST_*` を直接読ませる案 — 保存区分と runtime 契約が混ざる。
4. test 用だけ別語彙（`E2E_*` 等）にする案 — process 名との対応が読み手の再判断になる。`TEST_` + 同一語幹で足りる。
