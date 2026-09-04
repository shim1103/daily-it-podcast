---
name: 運用後続の完了は docs・契約固定まで（再 deploy は非 scope）
date: 2026-09-04T02:04:02
branch: feature/playback-deploy-ops-followup
---

## 1. Decision

1. Playback **運用後続**（rollback 文書化・Workers Logs 方針・lane 整理）の完了条件は、Decision の固定と運用 SSOT（`DEPLOY.md`）・必要なら `wrangler.jsonc` の契約更新までとする。
2. 本番への `wrangler deploy` / 再 deploy による反映確認は **非 scope** とする。反映は通常の再 deploy 手順（`DEPLOY.md`）に委ねる。
3. CD / git hook 自動 deploy、DAST、Dependabot / Renovate は本運用後続の完了条件に含めない（採用しない、または別 scope）。

## 2. Reason

1. 方針と手順が無い状態で再 deploy だけしても、次 session が同じ判断を再現できない。logging の目的（context 無し agent が docs だけで同じ答えに至る）に対し、先に Reason / Rejected と latest 手順を残す方が短い。
2. 再 deploy は本番 side effect であり、文書・契約の固定と完了定義を混ぜると「docs は済んだが本番未反映」と「未完了」が二重進捗になる。非 scope と明示すると完了が一意になる。
3. CD・DAST・Dependabot は軸が独立し、Access 付き個人利用や tool 未選定のまま混ぜると checkbox が永久に閉じない。

## 3. Rejected

1. 運用後続の完了に本番再 deploy と永続 log の実機確認を必須にする案 — 文書固定と本番 side effect が同居し、完了定義が膨らむ。
2. DAST / Dependabot / CD を同一 checkbox の完了条件に残す案 — 未決・非採用が進捗 index を腐らせる。
3. Decision 無しで `DEPLOY.md` だけ直す案 — 同じ問い（なぜ rollback が一次か、なぜ常時 observability か、なぜ再 deploy を完了に含めないか）の Reason / Rejected が残らない。
