---
name: Generator System Test は Integration gate 外の GHA 専用入口とする
date: 2026-08-30T11:56:01
branch: docs/generator-broad-system-e2e-plan
---

## 1. Decision

1. Generator の System Test は **CI 必須の Integration / Unit gate**（pre-push・必須 Integration / Unit workflow）に載せない。
2. System Test の収集・実行入口は gate 用 Integration 入口と分け、code 契約（build tag 付き suite と専用 script）を正とする。
3. System Test が行う credential 付き実 operation の実行場所は GitHub Actions に限る。通常の local 開発と Integration gate は実 service を呼ばず local secret を必要としない。前提の正は先行 Decision（`docs/decisions/2026-08-27T12-17-00-docs-env-secret-management-reconsider.md`）。
4. System Test の実行可能入口は `cmd/generator`（built binary または同等の `go run`）を **subprocess** とする。test process 内で Composition を直呼びして Driving Adapter を飛ばさない。
5. 本 Decision は、先行 Decision（`docs/decisions/2026-08-26T17-45-00-docs-infra-test-discussion.md`）のうち「System を CI 必須 gate に載せない」「gate 入口と混ぜない」を維持する。同 file が前提にしていた AgentSecrets + local keychain による local 実物 System は、`2026-08-27T12-17-00` 以降の前提では採らない。
6. System を branch protection の required check にするか、main merge 後の schedule 頻度は本 Decision の対象外とする。

## 2. Reason

1. System は Unit / Integration では検証できない最終結果に限定する。それを必須 Integration gate に混ぜると、secret なし Narrow / Broad の Repeatable な失敗と、credential・外部 service 起因の失敗が同じ赤になる。
2. local に secret を置く経路は既に廃されている。System を local 手動や AgentSecrets に戻すと、secret 配布面と運用 cost が再び増える。
3. Composition 直呼びは Application 合成の Broad に近く、CLI Driving Adapter（exit code・stderr・env 読取境界）を通らない。System の入口契約は `cmd/generator` の process 境界である。
4. required check 化や週次 schedule は「いつ merge を止めるか／どれだけ回すか」の運用頻度の問いであり、「どこで・どの入口で・secret をどう扱うか」とは別問いである。未固定のまま本 Decision に書き込むと推測を確定したように見える。

## 3. Rejected

1. System を Integration workflow や pre-push に相乗りさせる案 — Scope 名が Integration のまま System を隠し、gate が credential 依存になる。
2. local で AgentSecrets / keychain / `.env` により System を回す案 — `2026-08-27T12-17-00` が廃した local secret・local 実 operation を復活させる。
3. test 内で `NewProduceEpisodeFromEnv` を直呼びして System とする案 — `cmd` を通らず、Driving Adapter の契約が未検証のまま「System」と呼ぶ。
4. 本番 GHA secret を System に流用する案 — 本番永続・本番権限を test が触る。禁止事項と同型。具体 inventory は別 scope。
