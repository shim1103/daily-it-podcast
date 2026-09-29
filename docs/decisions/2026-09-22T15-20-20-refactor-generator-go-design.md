---
name: ProduceEpisode 下位は独立 package に置き、Driven Port を満たせるものは port で受け、取得窓 UseCase は UseCase DI にする
date: 2026-09-22T15:20:20
branch: refactor/generator-go-design
---

## 1. Decision

1. `FetchSourceItems` と `WriteEpisode` は `application` package root に平置きせず、`application/fetch` / `application/writeepisode` へ置き、`manuscript` / `speech` と同型の下位 UseCase package とする。
2. `WriteEpisode`（完成検査 Gate）は `port.EpisodeWriter` を満たす。Composition は raw Adapter を Gate で包み、`ProduceEpisode` は書込面を **`port.EpisodeWriter` だけ**で受ける（manuscript / speech と同型）。
3. `FetchSourceItems`（取得窓の適用）は Driven Port に上げず、`ProduceEpisode` から **UseCase DI**（呼び出し面の interface または具象）で受ける。外側の差し替え面は引き続き `port.ItemSource`。
4. `ProduceEpisode` の公開契約 documentation は観測可能な postcondition / invariant に限る。手順の How（関数呼び出し順の列挙）は契約に書かない。

## 2. Reason

1. 平置きだと「入口 UseCase」と「下位 UseCase」が同じ package に見え、manuscript / speech だけ dir がある非対称が読み手の層理解を壊す。配置の対称は見た目のためではなく、差し替え境界の所在を誤認させないためである。
2. Go の interface は明示 `implements` を要求しない。Gate が `EpisodeWriter` を満たせば、Application 合成と Infrastructure Adapter が同じ型穴に挿さる（manuscript の `TextWriter` と同型）。`ProduceEpisode` が Gate 具象だけを握ると、port 経由の差し替え物語と型が食い違う。
3. 取得窓は外部能力ではなく Application の方針（`backend/application` の UseCase 同士 DI の例そのもの）である。これを Driven Port に上げると「外部 I/O」と「窓の計算」が同じ語彙に混ざり、Port を手段語彙で汚染する。内側の `ItemSource` が infra 差し替え面で足りる。
4. 契約に How を書くと、実装変更のたびに契約が腐る。公開境界の documentation は義務（観測結果）だけを持ち、How は code が SSoT（`coding-style` の公開契約規則）。これは Go 固有というより DbC だが、`(T, error)` と直線的 orchestration では手順書化しやすいので、明示的に禁じた。

## 3. Rejected

1. **fetch / write を root のまま「同層具象だから pointer」と説明する案** — 内側は既に port 経由で差し替え可能であり、非対称の説明にならない。読み手を誤誘導する。
2. **Fetch も `port.ItemSource` だけにして窓を `ProduceEpisode` へ内蔵する案** — 窓の単独検証と変更理由の分離が消える。既存 UseCase 切り出しの根拠を捨てる。
3. **Fetch 用に別 Driven Port（例: now 基準 List）を新設する案** — 外部能力ではなく方針を Port に載せる。Adapter が窓を知る必要が無いのに IF が増える。
4. **`ProduceEpisode` が `*writeepisode.WriteEpisode` 具象のまま、配置だけ dir 化する案** — 配置対称は直るが、text/speech が port・write だけ具象という型の非対称が残る。
5. **契約にパイプライン全文を残す案** — 手順書が契約の顔をし、DRY（実装と契約の二重保守）と Misleading Comment のリスクが残る。
