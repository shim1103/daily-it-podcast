---
name: processenv は親 environ を継承し Composition 結線形は変えない
date: 2026-08-30T17:40:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `processenv.Launcher` の child environment は、親 process の environ を継承したうえで、inject secret を同名上書きする。
2. `InheritedEnvNameAllow`（PATH / HOME / TMPDIR）のみを渡す方式は採らない。PR80 Cursor CLI probe の成功条件は親 env 継承であり、allowlist のみ（env -i 相当）では Cursor API に到達できない（実測: `Failed to reach the Cursor API`）。
3. `NewSecretEnvLauncherFactory(secretValue, lookupEnv)` と Composition（`newCursorTextWriter`）の結線形は変えない。他 vendor secret の deny list を Composition に足さない。
4. 本 Decision は、processenv の「allowlist + secret だけ」を前提にした Narrow の観測契約を、上記継承＋inject に置き換える。

## 2. Reason

1. PR80 probe は親 env 継承で Cursor API に届いた。allowlist のみは実測で届かない。
2. Composition 以下の結線は既に完成しており、Cursor 結線へ他 vendor の env 名を持ち込むと SRP を壊す。

## 3. Rejected

1. allowlist（PATH/HOME/TMPDIR）のみを維持する案 — GHA で Cursor API に届かない。
2. Composition に他 vendor secret の deny list を渡す案 — 結線形を壊し、Cursor composition が無関係な secret 名を知る。
