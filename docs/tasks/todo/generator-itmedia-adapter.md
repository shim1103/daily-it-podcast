# feat(generator): ITmedia NEWS ItemSource Adapter の List を実装する

## 1. Summary

このIssueでは `internal/infrastructure/itmedia` の `ListItemSource.List` を実装し、ITmedia NEWS 速報 RSS 2.0 feed の取得 → `encoding/xml` parse → `SourceItem` 変換を通す。完了後、`itmedia` package の Sociable Unit と Narrow Integration が緑になり、`List` が stub の `panic` を返さなくなる。

## 2. Context

- 情報源を GetXAPI から HackerNews・Lobsters・ITmedia の3公式源へ入替える Decision（`docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）の実装 task の1本。ITmedia は「日本語の報道枠」を担う。
- scope-split の A は完了済み。`internal/infrastructure/itmedia/{source_id,error,item_source}.go` が stub で存在し、`item_source.go` に `feedURL` const（`https://rss.itmedia.co.jp/rss/2.0/news_bursts.xml`）と振る舞い doc が固定されている。「RSS 汎用 Adapter は作らず ITmedia 専用」（`docs/decisions/2026-09-02T14-41-01`）。
- `item_source_sociable_unit_test.go` に `t.Skip` の RED stub が6本ある。
- 実測: この feed は素直な RSS 2.0。`item` に `title` / `link` / `pubDate`（RFC1123Z）があり、`description` は記事による（無い item もある。ある場合は2〜4行の記事要約）。読者 comment は無い。

## 3. Canonical Sources

- `docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md` — 源の入替えと専用 Adapter 方針。
- `docs/decisions/2026-09-02T14-41-01-feature-hackernews-api-adapter.md` — ITmedia 専用 Adapter、RSS 2.0 を標準 `encoding/xml` で読む、汎用 RSS Adapter を作らない根拠。
- `docs/decisions/2026-09-02T14-41-02-feature-hackernews-api-adapter.md` — `Context` 行の内容、`links:` に記事 URL。
- `docs/decisions/2026-09-02T15-27-00-feature-hackernews-api-adapter.md` — 失敗時挙動。
- `internal/infrastructure/itmedia/item_source.go` — `feedURL` と振る舞い doc の正本。
- `internal/application/port/item_source.go` — `ItemSource` port 契約。
- test 分類・命名は `skills/1:terms/testing-strategy/SKILL.md`、GWT は該当 skill を参照。

## 4. Scope

### In Scope

- `internal/infrastructure/itmedia/item_source.go` の `List` 実装（`panic` を外す）。
- RSS 2.0 の `channel` / `item` を読む unexported struct と `encoding/xml` decode。
- `description` の軽い HTML 正規化 helper（token 効率目的）。
- `item_source_sociable_unit_test.go` の6本を実装（`t.Skip` を外す。XML fixture で double を組む）。
- `apps/generator/test/itmedia_narrow_integration_test.go` を新規作成（実 `*http.Client` の外向き HTTP を TLS redirect で観測）。
- `item_source.go` の `List` godoc から「振る舞い（B/C が実装する）:」の箇条書きブロックを削除する（`@require`/`@ensure`/`@invariant` の契約行は残す）。実装コードと `item_source_sociable_unit_test.go` が振る舞いの SSOT。hackernews Adapter（同 branch）が同じ整理を済ませている。

### Out of Scope

- `feedURL` の変更・複数 feed 化（1 outlet 1 adapter。`docs/decisions/2026-09-02T14-41-01`）。
- 別媒体（Publickey / InfoQ 等）の Adapter（lane D 項目。各々専用 Adapter を新設）。
- 記事本文全文の取得（RSS の `description` 要約で足りる。`docs/decisions/2026-09-02T14-41-01` §Rejected 3）。
- HackerNews / Lobsters Adapter（別 task file）。
- Broad Integration の3源結線（別 task file）。

## 5. Contract

- `func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error)` — signature 変更なし。
- `@ensure` は port 契約に従う: `SourceID` は `itmedia.SourceID`、`OccurredAt` は UTC かつ `since` 以上、該当なしは非 nil の空 slice。
- `OccurredAt` は `item.pubDate`（RFC1123Z、例 `Wed, 02 Sep 2026 12:00:00 +0900`）を UTC 化した値。
- `Context` の行（`docs/decisions/2026-09-02T14-41-02` 準拠、取れた行だけ）: `title` / `text`（`description` の HTML を軽く正規化。無ければ行省略） / `permalink`（`item.link`） / `links`（`item.link`）。`description` が無い item は `title` と `permalink` / `links` のみで `SourceItem` 化する。
- 取得範囲: `feedURL` を1回 GET し、`item.pubDate >= since` を残す。件数上限は設けない（feed は数十件）。
- HTML 正規化: entity unescape + タグ除去の軽い正規化のみ（full readability はしない。`docs/decisions/2026-09-02T14-41-02` §Reason）。
- 失敗時（`docs/decisions/2026-09-02T15-27-00` 準拠）: `*itmedia.Error{Op, Err}` を return。`client.Do` error / 5xx は 1 回だけ即再試行。4xx / 壊れた XML は即 return。feed は1本なので「一覧 endpoint の失敗 = `List` ごと失敗」。
- `ctx` を HTTP リクエストへ伝播。Adapter 独自 timeout なし。

## 6. Constraints

- Infrastructure → Application は `port` のみ import。
- RSS parser は標準 `encoding/xml`。外部ライブラリ（`gofeed` 等）を足さない（`docs/decisions/2026-09-02T14-41-01` §Rejected 5）。
- secret を持たない（RSS は認証不要）。
- RSS の struct を package 外へ露出しない。

## 7. Acceptance Criteria

- [ ] `List` が `panic` を含まず、port 契約を満たす。
- [ ] `pubDate == since` は含み、`since` より前は除外する（境界）。
- [ ] `pubDate` の time zone offset（`+0900` 等）を UTC 化している。
- [ ] `description` が空の item で `Context` が `title` + `permalink` + `links` 行のみで組まれる（`text:` 行が出ない）。
- [ ] `description` がある item で `text:` 行に HTML 除去済みの要約が入る。
- [ ] `Context` に `links: <item.link>` 行が出る。
- [ ] 該当なしで非 nil の空 slice を返す。
- [ ] 非200 / 壊れた XML の各経路で `*itmedia.Error` が返り、`Error()` が `"itmedia:"` prefix、`Unwrap()` 非 nil。
- [ ] `item_source_sociable_unit_test.go` の6本が実装され緑。
- [ ] `apps/generator/test/itmedia_narrow_integration_test.go` が実 HTTP を TLS redirect で観測して緑。
- [ ] Narrow が secret なしで Integration gate と coverage 分母に入る。

## 8. Verification

```bash
cd apps/generator
go build ./...
go vet ./...
go test ./internal/infrastructure/itmedia/... -run TestList -v
go test ./test/ -run TestITmedia -v
golangci-lint run ./...
gofmt -l .
```

- 完了判定は `itmedia` package と `TestITmedia*` Narrow が緑になること。
- coverage: `./scripts/test-unit.sh`。

## 9. Dependencies

- Blocked by: なし（scope-split A 完了済み）。
- Related: `generator-hackernews-adapter.md` / `generator-lobsters-adapter.md`（並行可）、`generator-source-adapters-wiring.md`（この3本の後）。

## 10. Risks

- ITmedia RSS は UTF-8 宣言 + BOM や実体参照の混在があり得る。`encoding/xml` の `Decoder` に `CharsetReader` を設定するか、UTF-8 前提でよいか Narrow で実データを見て確認。
- `feedURL` が 302 リダイレクトを返す構成（実測で `news.xml` は 302）。`news_bursts.xml` は 200 を確認済みだが、`*http.Client` の default redirect follow に任せる。
- `description` に相対 URL・`rssad` トラッキング付き URL が混じることがある。`links:` は `item.link`（記事 URL）を使い、`description` 内の URL は加工しない。

## 11. Notes

- HackerNews / Lobsters は JSON、ITmedia は RSS(XML) で parse 経路が違う。`Context` 行組みの共通化は3 Adapter が出揃ってから検討（`docs/decisions/2026-09-02T14-41-01` の「三度現れてから」）。
- 別媒体（Publickey 等）を足す時は `infrastructure/publickey/` を新設。この Adapter を「汎用化」しない。
- delivery の kind 分類 type switch は現ブランチに無い（merge 時作業。lane D 記録済み）。
