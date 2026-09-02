# feat(generator): HackerNews ItemSource Adapter の List を実装する

## 1. Summary

このIssueでは `internal/infrastructure/hackernews` の `ListItemSource.List` を実装し、`topstories.json` → 個別 item / comment 取得 → `SourceItem` 変換を通す。完了後、`hackernews` package の Sociable Unit と Narrow Integration が緑になり、`List` が stub の `panic` を返さなくなる。

## 2. Context

- 情報源を GetXAPI から HackerNews・Lobsters・ITmedia の3公式源へ入替える Decision（`docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）の実装 task の1本。
- scope-split の A は完了済み。`internal/infrastructure/hackernews/{source_id,error,item_source}.go` が stub で存在し、`item_source.go` の doc コメントに振る舞いと契約値（`MaxStoriesScanned` / `MaxCommentsPerStory` / `CommentDepth` / `apiBaseURL`）が固定されている。
- `item_source_sociable_unit_test.go` に `t.Skip` の RED stub が7本あり、各 Skip 本文に「C が組む double と assert 内容」が書かれている。
- HackerNews Firebase API は認証不要・rate limit 無し・無料（実測確認済み）。story item は外部リンク型だと `text` が空になり、`kids` を辿った comment 本文が main text になる。

## 3. Canonical Sources

- `docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md` — 源の入替えと専用 Adapter 方針。
- `docs/decisions/2026-09-02T14-41-01-feature-hackernews-api-adapter.md` — JSON 経路採用、comment を main text にする判定軸。
- `docs/decisions/2026-09-02T14-41-02-feature-hackernews-api-adapter.md` — `Context` 行の内容、`links:` に外部 URL、engagement 指標を載せない。
- `docs/decisions/2026-09-02T15-27-00-feature-hackernews-api-adapter.md` — 失敗時挙動（Error return のみ / retry は transient HTTP のみ / 個別 item 失敗は要素を落として続行）。
- `internal/infrastructure/hackernews/item_source.go` — 契約値と振る舞い doc の正本。
- `internal/application/port/item_source.go` — `ItemSource` port 契約（`@require`/`@ensure`/`@invariant`）。
- `internal/infrastructure/x/getxapi/post_source.go`（git history: `git show HEAD~N` で辿れる。削除済み） — 逐次 fetch・pagination・`Context` 行組みの実装参考。
- `apps/generator/test/getxapi_narrow_integration_test.go`（削除済み。git history） — Narrow Integration の TLS redirect double の書き方。
- test 分類・命名は `skills/1:terms/testing-strategy/SKILL.md`、GWT は該当 skill を参照。

## 4. Scope

### In Scope

- `internal/infrastructure/hackernews/item_source.go` の `List` 実装（`panic` を外す）。
- 必要な private helper（page 取得、item decode、`Context` 行組み、時刻変換）。
- `item_source.go` 内の unexported 型（API JSON の shape）。
- `item_source_sociable_unit_test.go` の7本を実装（`t.Skip` を外し、`stubRoundTripper` で double を組む）。
- `apps/generator/test/hackernews_narrow_integration_test.go` を新規作成（実 `*http.Client` の外向き HTTP を TLS redirect で観測。認証 header 無しでよいことを assert）。

### Out of Scope

- 契約値（`MaxStoriesScanned` 等）の変更。変えるなら Decision が先。
- comment のスレッド再帰（`CommentDepth` は 1 固定。lane D 項目）。
- composite の source またぎ sort（lane D 項目）。
- Lobsters / ITmedia Adapter（別 task file）。
- Broad Integration の3源結線（別 task file）。

## 5. Contract

- `func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error)` — signature は変更しない。
- `@ensure` は port 契約に従う: 各 `SourceItem.SourceID` は `hackernews.SourceID`、`OccurredAt` は UTC かつ `since` 以上、該当なしは非 nil の空 slice。
- `OccurredAt = time.Unix(item.time, 0).UTC()`。
- `Context` の行（`docs/decisions/2026-09-02T14-41-02` 準拠、取れた行だけ書く）: `item_id` / `actor_id` / `actor_name`（`by`） / `title` / `text`（story `text` と comment 本文を改行連結。空なら行ごと省略） / `permalink`（`https://news.ycombinator.com/item?id=<id>`） / `links`（story の `url` があれば）。score / `descendants` は載せない。
- 取得範囲: `topstories.json` の先頭 `MaxStoriesScanned` 件の id を対象。`type=="story"` かつ `!deleted && !dead` かつ `time>=since.Unix()` を残す。残した story ごとに `kids` の先頭 `MaxCommentsPerStory` 件を `item/<id>.json` で取得（`CommentDepth=1`）。
- HTML 正規化: entity unescape + `<p>`→改行 + タグ除去の軽い正規化のみ（token 効率目的。full readability はしない）。
- 失敗時（`docs/decisions/2026-09-02T15-27-00` 準拠）: `*hackernews.Error{Op, Err}` を return。`client.Do` error / 5xx は 1 回だけ即再試行。4xx / parse 失敗は即 return。個別 item / comment の取得失敗はその要素を落として `List` を続行。`topstories.json` の取得失敗は `List` ごと失敗。
- `ctx` は全 HTTP リクエストへ伝播（`http.NewRequestWithContext`）。Adapter 独自の timeout は持たない。

## 6. Constraints

- Infrastructure → Application は `port` のみ import（`.golangci.yml` depguard。`internal/infrastructure` prefix は許可されるが設計規律で守る）。
- 逐次取得。並行化しない（`docs/decisions/2026-09-02T15-27-00` §4）。
- secret を持たない（HackerNews は認証不要）。
- vendor 固有型（API JSON の struct）を package 外へ露出しない。

## 7. Acceptance Criteria

- [ ] `List` が `panic` を含まず、port 契約（`SourceID` / `OccurredAt` UTC≥since / 非 nil 空 slice）を満たす。
- [ ] `type!="story"`（job / poll）・`deleted` / `dead` の item が結果に現れない。
- [ ] `time == since` は含み、`time == since-1s` は除外する（境界）。
- [ ] `kids` を `MaxCommentsPerStory` 超で与えたとき、comment 取得数が上限で頭打ちになる。
- [ ] story の `text` が空でも comment 本文が `Context` の `text:` 行に入り、`title:` のみにならない。
- [ ] story の `url` があれば `Context` に `links:` 行が出る。無ければ出ない。
- [ ] `client` nil / 非200 / 壊れた JSON の各経路で `*hackernews.Error` が返り、`Error()` が `"hackernews:"` prefix、`Unwrap()` が非 nil。
- [ ] 個別 comment 取得が1件失敗しても、その story の他 comment と他 story は結果に残る。
- [ ] `topstories.json` の取得が失敗したら `List` 全体が `*hackernews.Error` を返す。
- [ ] `item_source_sociable_unit_test.go` の7本が実装され緑（`t.Skip` を外す）。
- [ ] `apps/generator/test/hackernews_narrow_integration_test.go` が実 `*http.Client` の外向き HTTP を TLS redirect で観測し、GET が届き認証 header 無しで成功することを assert して緑。
- [ ] Narrow が secret なし（`apps/generator/test/` の build tag なし）で、Integration gate と coverage 分母に入る。

## 8. Verification

```bash
cd apps/generator
go build ./...
go vet ./...
go test ./internal/infrastructure/hackernews/... -run TestList -v
go test ./test/ -run TestHackerNews -v
golangci-lint run ./...
gofmt -l .
```

- 3源 Adapter 全部が揃うまで `go test ./...` の `test` package は broad integration の3 SKIP が残る（別 task で解消）。この task の完了判定は `hackernews` package と `TestHackerNews*` Narrow が緑になること。
- coverage: `./scripts/test-unit.sh`（generator Unit gate: statement 90%）。閾値割れなら Narrow の分母算入を確認。

## 9. Dependencies

- Blocked by: なし（scope-split A 完了済み）。
- Related: `generator-lobsters-adapter.md` / `generator-itmedia-adapter.md`（同型・並行可）、`generator-source-adapters-wiring.md`（この3本の後）。

## 10. Risks

- N+1 の逐次取得が遅い（上位30 story + comment で数百リクエスト）。→ 契約値の足切りで有界。GHA job timeout に収まるか Narrow / 手元実測で確認。収まらなければ並行化を lane から起票（`docs/decisions/2026-09-02T15-27-00` §Rejected 6）。
- HackerNews の `time` が topstories のランキング順で並んでおらず時刻順でないため、`since` フィルタで全 `MaxStoriesScanned` 件を走査する必要がある。「古いのが来たら break」はできない。

## 11. Notes

- `getxapi` Adapter（削除済み）の `toSourceItem` / pagination loop が最も近い実装参考。git history から取れる。
- comment を Context に入れる件数・連結順（ランク順 or 時系列）は Decision で「先頭 `MaxCommentsPerStory` 件」までしか固定していない。連結順は `kids` の順（HackerNews のランク順）でよい。
- delivery の kind 分類 type switch（`internal/delivery`）は現ブランチに無い（未 merge の `feature/generator-system-e2e-produce-episode` にある）。その branch が merge される時に `hackernews`/`lobsters`/`itmedia` を type switch に足し `getxapi` を除く必要がある — これは merge 時の作業で、この task の scope 外（lane D に記録済み）。
