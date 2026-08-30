---
name: Generator の System / E2E は CI gate に載せない
date: 2026-08-26T17:45:00
branch: docs/infra-test-discussion
---

## 1. Decision

1. Generator の System / E2E は **CI gate（pre-push・GitHub Actions の必須 Integration / Unit workflow）に載せない**。
2. System / E2E を実行する場合の入口は local 専用とし、gate 用 Integration 入口と混ぜない。収集境界の正本は code 契約を参照する。
3. 過去 project の「PR CI に e2e を載せる」運用を、本 repo の正にしない。

## 2. Reason

1. 既存 Decision は commit = Unit、push/GHA = Integration、E2E は gate に入れない（YAGNI）。System を今 CI 必須にすると、その Decision と矛盾し、gate 条件が二重化する。
2. local 実物の主経路は AgentSecrets + OS keychain である。GitHub-hosted runner にその keychain は無く、System を CI 必須にすると skip か別 runtime（processenv）へすり替わり、「System を CI で証明した」意味が変わる（Least Astonishment）。
3. System の目的は Unit / Integration では不可能な最終結果に限る。最終 UseCase（ProduceEpisode 本体）が未完のまま System を CI に固定しても、検証対象が定まらない。

## 3. Rejected

1. 過去 Next.js project の ci-checks（PR に e2e）を本 repo へ移植する案 — 過去基準であり、本 repo の gate Decision と keychain 前提に合わない。
2. System を Integration workflow に相乗りさせる案 — Scope 名が Integration のまま System を隠し、分類と runner の対応が壊れる。
3. processenv + CI secret で「CI 上の System」を今必須化する案 — 入口→出口の最終 postcondition が未固定のまま供給路だけ先に増える。
