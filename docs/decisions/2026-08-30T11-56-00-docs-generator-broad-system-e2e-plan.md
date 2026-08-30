---
name: Generator Integration gate は secret なし Narrow と Broad を載せる
date: 2026-08-30T11:56:00
branch: docs/generator-broad-system-e2e-plan
---

## 1. Decision

1. Generator の Integration gate（pre-push と GitHub Actions の Integration workflow、およびそれらが呼ぶ `scripts/generator/test-integration.sh`）が実行してよいのは、**credential / secret 値を使わない Narrow Integration と Broad Integration** とする。
2. secret なし Broad Integration は pre-push からも外さない。
3. 本 Decision は、先行 Decision（`docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`）のうち「gate は Narrow のみ」の範囲を置き換える。同 file の「実 AgentSecrets / keychain を要する suite は gate に載せない」「収集境界の正本は code」は維持する。
4. tag 名・path・収集 command の正本は code 契約を参照する。本 Decision に写さない。

## 2. Reason

1. 先行 Decision が gate を Narrow に限った主因は、実 AgentSecrets / keychain 経路と無人 CI を混ぜないことだった。その後の Decision（`docs/decisions/2026-08-27T12-17-00-docs-env-secret-management-reconsider.md`）で local secret 供給と local 実 operation を廃し、secret なし auto test だけが local / CI の共通面になった。secret なし Broad は Narrow と同じ面に乗る。
2. Broad は Narrow では検出できない配線・状態伝播・error 伝播を見る。secret なし（httptest / fake child 等）なら Repeatable であり、gate から外すと配線回帰が PR 前に落ちない。
3. 「重いから pre-push から外す」は case 数と runtime の問題であり、Scope 名で除外する理由にならない。重さは Broad の case を削って解く。pre-push と GHA で Broad の有無を分けると、どちらが正の gate か判別できなくなる。

## 3. Rejected

1. gate を Narrow のみのままにし Broad は GHA だけにする案 — secret なし Broad を Narrow と別扱いする根拠が、AgentSecrets 分離前提の失効後に残らない。
2. Broad が重い想定で pre-push から外す案 — 未実測の重さで収集境界を割り、pre-push 緑と GHA 緑の意味がずれる。
3. Broad を必須 Unit workflow に載せる案 — Scope が Integration なのに Unit gate へ隠し、分類と runner の対応が壊れる。
