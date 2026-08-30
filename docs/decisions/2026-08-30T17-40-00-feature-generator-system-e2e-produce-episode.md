---
name: processenv は親 environ を継承し他 vendor secret だけ落とす
date: 2026-08-30T17:40:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `processenv.Launcher` の child environment は、親 process の environ を継承したうえで、Composition が渡す **denied parent env 名**（他 vendor の secret）だけを除去し、inject secret を上書きする。
2. `InheritedEnvNameAllow`（PATH / HOME / TMPDIR）のみを渡す方式は採らない。PR80 Cursor CLI probe の成功条件は親 env 継承であり、allowlist のみ（env -i 相当）では Cursor API に到達できない（実測: `Failed to reach the Cursor API`）。
3. denied 名の正は Composition が `config` の secret key 名から供給する（GetX / Gemini / Google OAuth secret）。`CURSOR_API_KEY` は inject で上書きする。
4. 本 Decision は、processenv の「allowlist + secret だけ」を前提にした Narrow の観測契約を、上記継承＋denied に置き換える。

## 2. Reason

1. PR80 probe は `env` で HOME/PATH/TMPDIR を上書きしつつ **親 env を継承**していた。`env -i` で最小集合だけにした試験ではなかったのに「CURSOR_API_KEY のみで足りる」と読める結論が残った。
2. System 実測で allowlist のみの child は Cursor API 到達に失敗した。probe 成功条件へ戻すのが Least Astonishment。
3. 親を全継承すると GetX / Gemini / OAuth secret が Cursor child へ漏れる。denied 名で切れば Least Privilege を保つ。

## 3. Rejected

1. allowlist（PATH/HOME/TMPDIR）のみを維持する案 — GHA で Cursor API に届かない。
2. 親 environ を無フィルタ継承する案 — 他 vendor secret が Cursor へ漏れる。
3. System だけ親継承に分岐する案 — production produce も同じ processenv を通る。分岐は二重契約になる。
