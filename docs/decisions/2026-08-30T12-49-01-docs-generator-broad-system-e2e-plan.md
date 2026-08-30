---
name: Generator 本番は毎日 07:00 JST、System は週次 07:00 JST と workflow_dispatch
date: 2026-08-30T12:49:01
branch: docs/generator-broad-system-e2e-plan
---

## 1. Decision

1. Generator の **本番** produce workflow は、毎日 **07:00 Asia/Tokyo** に定時実行する。GitHub Actions の `schedule.cron` は UTC で書く（07:00 JST = 22:00 UTC 前日）。入口は専用 script 経由の `cmd/generator`。`workflow_dispatch` も付ける。
2. Generator の **System** workflow は、週 1 回 **月曜 07:00 Asia/Tokyo**（cron は UTC で対応）と `workflow_dispatch` で実行する。必須 Integration / Unit gate には載せない。
3. 暦日・Fetch 基準は runner の `time.Now()` と Application の JST 規則に従う。定時を JST 朝に置くことで、日本の日付境界と運用感覚を揃える。
4. System を branch protection の required check にするかは本 Decision の対象外のままとする。
5. 本 Decision は、先行 Decision（`docs/decisions/2026-08-30T11-56-01-docs-generator-broad-system-e2e-plan.md`）のうち「schedule 頻度は対象外」だった範囲を、上記頻度に限って埋める。

## 2. Reason

1. 本番は毎日の配信が目的なので日次が必要。System は契約回帰が目的なので、毎日本番と二重に vendor / Actions 分を消費しないよう週次を既定にする。
2. cron は UTC 固定だが、運用の「朝 7 時」は JST で決める。日付を UTC 日界に合わせると日本の「今日の episode」とずれる議論が増える。
3. `workflow_dispatch` は失敗再実行と手動確認の入口であり、定時と直交する。

## 3. Rejected

1. System を毎日本番と同じ頻度にする案 — personal free の Actions 分と vendor quota を本番と二重消費する。契約回帰に日次は必須ではない。
2. 定時を UTC 正午など JST 非対応のまま「何時でもよい」で固定する案 — 運用者が毎回換算する。
3. System を pre-push / 必須 Integration に載せる案 — 先行 Decision の gate 分離を壊す。
4. 本番 schedule を Run 実装まで YAML に書かない案 — 注入と入口の契約を先に固定できない。Run 未完時は job が赤になることを前提に、workflow 自体は置く。
