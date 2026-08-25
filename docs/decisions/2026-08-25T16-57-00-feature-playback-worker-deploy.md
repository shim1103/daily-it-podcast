---
name: deploy・Access 運用の最新方針は DEPLOY.md を SSOT とする
date: 2026-08-25T16:57:00
branch: feature/playback-worker-deploy
---

## 1. Decision

1. Playback の deploy・Access・公開境界の **最新方針** は repo 根 `DEPLOY.md` を唯一の SSOT とする。
2. 文書分業を次に更新する（`docs/decisions/2026-08-15T16-23-08-develop-docs-split.md` の README / DESIGN / contracts 分業を拡張する）。
   1. `README.md` — 地図・使い方・受け入れ・secret 名 inventory
   2. `DESIGN.md` — 層・依存・所有・test 規則
   3. `DEPLOY.md` — deploy・Access・公開境界
   4. `contracts/` — Drive 上の表現
3. wrangler 境界契約（`name` / `main` / assets / `/episodes*` 先回り）の正は A artifact（`apps/playback/wrangler.jsonc` と `apps/playback/worker/src/worker-entry.ts`）とする。`DEPLOY.md` と本 decision はそれを参照し、契約値を写さない。
4. `docs/daily`・`docs/decisions`・`docs/tasks` は session / 判断 / 未完了の **記録**であり、deploy 運用の latest 入口にしない。
5. `apps/playback/README` や `docs/runbooks/` など、運用 latest を増やす置き場は作らない。

## 2. Reason

1. README に Access session・hostname・secret 注入区分まで書くと、地図の変更理由と運用方針の変更理由が同じ file に混ざる。DESIGN に Access 手順相当を書くと層規則の SSOT が腐る。
2. `docs/decisions` は日時付きの凍結記録であり「今の運用を開く入口」にならない。`docs/daily` は session 束ねであり、同じ問いに同じ答えを返す文書ではない。
3. 既存の文書3分業（README / DESIGN / contracts）は Drive・層・地図には足りるが、Cloudflare 公開境界が無所属のまま README / DESIGN へ漏れやすかった。root の大文字 `DEPLOY.md` は DESIGN と同じ寿命（latest・参照透過の内側）で、記録 dir と混ざらない。
4. A の契約値を decision / DEPLOY 本文へ再掲すると、`wrangler.jsonc` を直したときに文書が遅れ、scope-split の「BはAを参照する」に反する。

## 3. Rejected

1. `docs/runbooks/` や `docs/` 配下に latest 運用 SSOT を置く案 — 記録用 dir と latest 運用が混ざる。
2. README / DESIGN に Access・hostname・注入区分を併記し続ける案 — DRY が壊れ、どちらが正か不明になる。
3. `apps/playback/README` を増やす案 — dir 単位 README 禁止（既存 docs-split）と衝突する。
4. decision 本文に Worker `name`・assets path・entry path を正本として書く案 — A 契約の複製になり、更新が二重化する。
5. SPEC / PROPOSAL を常設して運用を書く案 — 既存 Rejected のまま不要。
