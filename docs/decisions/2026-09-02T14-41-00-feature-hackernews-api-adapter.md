---
name: 情報取得源を GetXAPI 単独から HackerNews・Lobsters・ITmedia NEWS の3公式源へ入替え、GetXAPI を撤去する
date: 2026-09-02T14:41:00
branch: feature/hackernews-api-adapter
---

## 1. Decision

Generator の情報取得源を次のとおり入替える。

1. **源を3本にする。** 国際・議論の HackerNews、国際・議論の Lobsters、日本・報道の ITmedia NEWS。3本とも認証不要・無料・運営元公式の API / RSS だけを叩く。役割分担は HackerNews / Lobsters が議論、ITmedia が日本語の報道。
2. **GetXAPI（および X vendor 経路一式）を撤去する。** `infrastructure/x/`、`config` の源向け credential（`SourceConfig` / `GETX_API_KEY`）、`DEPLOY.md` と GHA workflow の当該 Secret 行を削除する。復活は git 履歴から行い、無効化した code を温存しない。
3. **源ごとに専用 Adapter を置き、facade を作らない。** 置き場・`SourceID` 値・取得上限・base URL は `infrastructure/hackernews/`・`infrastructure/lobsters/`・`infrastructure/itmedia/` の stub を正本とする。「RSS 汎用 Adapter」は作らず、ITmedia も 1 源 1 Adapter とする。切替・合成は Composition Root の composite `ItemSource` のみが行い、Application は源個数・種類を知らない。

派生する2つの選択は別 Decision に置く。取得経路（JSON か RSS か）の判定と ITmedia 専用 Adapter の根拠は「先行 Decision（`2026-09-02T14-41-01-feature-hackernews-api-adapter.md`）」、`Context` への書き込み内容と外部 URL の扱いは「先行 Decision（`2026-09-02T14-41-02-feature-hackernews-api-adapter.md`）」を正とする。

`DESIGN.md` §3「外部 I/O」表の「情報取得」行と `README.md` の取得記述を本 Decision に合わせて更新する。契約値（源の `SourceID`・取得上限・base URL）は Adapter stub を正本とし、地図文書には写さない。

## 2. Reason

GetXAPI は X の公開データを session token 経由で取る非公式 proxy であり、X の Developer Agreement / Policy（"Scraping the Services without the prior consent of Twitter is expressly prohibited"）に反する。提供元自身が「unofficial API を使うと account 停止の risk がある」と明記している。先行 Decision（`2026-08-15T16-39-20-feature-x-api-adoption.md`）は「取得用 X account を本 account と分離（BAN 対策）」とし、停止される前提で運用する割り切りを記録していた。

philosophy §4-2 Rule of Least Power は「vendor 固有 API・独自経路より標準規格・公式 API を優先する」「不安の解消のために強力な実装へ乗り換えるのは目的への適合ではない」と定める。GetXAPI は reverse-engineered proxy で X frontend 変更に同時停止し、規約 risk と secret 管理コスト（AgentSecrets proxy 注入、`DEPLOY.md` Secret 登録）を伴う。認証不要・無料・公式の3源へ移せば、規約 risk と secret 管理が同時に消え、X frontend への相関 failure も無くなる。先行 Decision（`2026-08-15T16-39-20-feature-x-api-adoption.md`）の Rejected 末尾が「真の冗長化には別種 route（RSS bridge 等）が必要」と述べていた方向そのものである。

3源の選定は「議論・comment があるか」「日本の議論か国際の議論か（どちらも要る）」「AI slop ではなく技術のトレードオフを真面目に扱っているか」で比較した。

- **HackerNews（国際・議論）**: 運営元が Firebase で公式提供。認証不要・rate limit 無し・無料。story は外部リンクが主だが `kids` を辿れば comment 本文が API で取れ、議論の密度が主価値。票と moderation の文化で AI slop は上位に残りにくい。
- **Lobsters（国際・議論）**: 運営元公式で JSON も RSS も提供。招待制のため spam・AI 生成投稿がほぼ無く、HackerNews より小規模で技術密度が高い。
- **RSS(ITmedia NEWS)（日本・報道）**: `rss.itmedia.co.jp` の公式 RSS。読者 comment 文化は無く議論は担わないが、`description` に記事要約が付き、「日本で話題になった IT の出来事」を事実として供給する。議論軸を HackerNews / Lobsters が、日本語の報道軸を ITmedia が埋め、2軸を3本で覆う。

源ごと専用 Adapter・facade なしは、先行 Decision（`2026-08-17T17-15-01-feature-x-getxapi-adapter.md`）の「vendor ごとに Adapter を置き facade にしない」を対象を X vendor から3源へ移して引き継いだもの。composite `ItemSource` を残すのは先行 Decision（`2026-08-30T11-20-00-feature-generator-composition-produce-episode-wiring.md`）が Rejected で「composite も廃して Adapter を直接渡す案」を退け「複数源が入ったとき Application 側の変更を不要にする境界を先に置く」としていた通りで、今回3源で活きる。`ItemSource` port の契約（`SourceID` + `OccurredAt` 必須、残りは opaque `Context`。先行 Decision `2026-08-19T13-25-20-refactor-generator-source-port.md`）は撤去の影響を受けず、そのまま3源を収容する。

GetXAPI 撤去に伴い先行 X 関連 Decision（`2026-08-15T16-39-20-feature-x-api-adoption.md`、`2026-08-15T17-43-09-feature-x-api-adoption.md`、`2026-08-17T13-12-00-feature-x-post-source-adapter.md`、`2026-08-17T17-15-01-feature-x-getxapi-adapter.md`）の X vendor 固有の記述は指示対象を失う。旧 file の本文は書き換えない（`decisions.md` §8）。読み手は X vendor 経路が撤去済みであることを本 Decision から知る。

## 3. Rejected

1. **GetXAPI を無効化して code を温存する案** — 「一時廃止」を理由に `infrastructure/x/` と config を残すと、生きていない結線が読み手に「これは使うのか」を問わせ続ける。復活は git 履歴から実装し直せる。破棄と延期を同じ操作にしない（`scope-split` §D）。
2. **X vendor 経路を1源として3源に加える案** — 規約 risk と secret 管理コストを抱えたまま。philosophy §4-2 が退ける「不安の解消のための強力な経路の温存」。撤去して公式3源に寄せる。
3. **Zenn 記事一覧 API（`zenn.dev/api/articles`）を源にする案** — 運営元非公式の隠し API で、GetXAPI と同じ reverse-engineered 依存になる。Zenn を使うなら運営元公式の RSS 経由に限る。
4. **Qiita API v2 / Zenn RSS を今回の源に加える案** — どちらも公式だが、母集団に入門記事・buzz 記事・AI 生成記事が多く、「技術のトレードオフを真面目に扱う」条件にはフィルタ（stocks 数・topic 指定）前提でしか合わない。今回の3源で議論軸・報道軸を覆えており、YAGNI。日本語の議論源が要ると確定した時点で lane から再検討する。
