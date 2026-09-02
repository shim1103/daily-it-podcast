# feat(generator): Lobsters ItemSource Adapter の List を実装する

## 1. Summary

このIssueでは `internal/infrastructure/lobsters` の `ListItemSource.List` を実装し、`hottest.json` → 各 story の `/s/<short_id>.json` → `SourceItem` 変換を通す。完了後、`lobsters` package の Sociable Unit と Narrow Integration が緑になり、`List` が stub の `panic` を返さなくなる。

## 2. Context

- 情報源を GetXAPI から HackerNews・Lobsters・ITmedia の3公式源へ入替える Decision（`docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md`）の実装 task の1本。
- scope-split の A は完了済み。`internal/infrastructure/lobsters/{source_id,error,item_source}.go` が stub で存在し、`item_source.go` の doc コメントに振る舞いと契約値（`MaxStoriesScanned` / `MaxCommentsPerStory` / `CommentDepth` / `apiBaseURL`）が固定されている。
- `item_source_sociable_unit_test.go` に `t.Skip` の RED stub が6本あり、各 Skip 本文に「C が組む double と assert 内容」が書かれている。
- Lobsters は運営元公式で JSON も RSS も提供。招待制で spam・AI 生成投稿がほぼ無い。実測: `/hottest.json` → `/s/<short_id>.json` の `comments[]` に全 comment が `comment_plain`（plain text 済み）込みで入り、story 1 件あたり1追加リクエストで取れる。RSS は本文も comment も空で不採用（`docs/decisions/2026-09-02T14-41-01`）。

## 3. Canonical Sources

- `docs/decisions/2026-09-02T14-41-00-feature-hackernews-api-adapter.md` — 源の入替えと専用 Adapter 方針。
- `docs/decisions/2026-09-02T14-41-01-feature-hackernews-api-adapter.md` — JSON 経路採用、`comment_plain` を main text にする判定軸。
- `docs/decisions/2026-09-02T14-41-02-feature-hackernews-api-adapter.md` — `Context` 行の内容、`links:` に外部 URL、engagement 指標を載せない。
- `docs/decisions/2026-09-02T15-27-00-feature-hackernews-api-adapter.md` — 失敗時挙動。
- `internal/infrastructure/lobsters/item_source.go` — 契約値と振る舞い doc の正本。
- `internal/application/port/item_source.go` — `ItemSource` port 契約。
- `internal/infrastructure/hackernews/item_source.go`（同 branch で並行実装） — 同型の逐次 fetch・`Context` 行組みの参考。
- test 分類・命名は `skills/1:terms/testing-strategy/SKILL.md`、GWT は該当 skill を参照。

## 4. Scope

### In Scope

- `internal/infrastructure/lobsters/item_source.go` の `List` 実装（`panic` を外す）。
- 必要な private helper と API JSON の shape 型。
- `item_source_sociable_unit_test.go` の6本を実装（`t.Skip` を外す）。
- `apps/generator/test/lobsters_narrow_integration_test.go` を新規作成（実 `*http.Client` の外向き HTTP を TLS redirect で観測）。

### Out of Scope

- 契約値の変更。
- comment のスレッド再帰（`CommentDepth` は 1 固定）。
- HackerNews / ITmedia Adapter（別 task file）。
- Broad Integration の3源結線（別 task file）。

## 5. Contract

- `func (s *ListItemSource) List(ctx context.Context, since time.Time) ([]models.SourceItem, error)` — signature 変更なし。
- `@ensure` は port 契約に従う: `SourceID` は `lobsters.SourceID`、`OccurredAt` は UTC かつ `since` 以上、該当なしは非 nil の空 slice。
- `OccurredAt` は story の `created_at`（offset 付き RFC3339 相当）を UTC 化した値。
- `Context` の行（`docs/decisions/2026-09-02T14-41-02` 準拠、取れた行だけ）: `item_id`（`short_id`） / `actor_id` / `actor_name`（`submitter_user`） / `title` / `text`（`comment_plain` 上位 N 件を改行連結。`description_plain` があれば先頭に足す。空なら行省略） / `permalink`（`short_id_url` か `comments_url`） / `links`（story の `url` があれば）。`score` / `comment_count` は載せない。
- 取得範囲: `hottest.json` の `created_at >= since` を残し、先頭 `MaxStoriesScanned` 件を対象。各 story の `/s/<short_id>.json` を取得し、`comments[]` から `is_deleted` / `is_moderated` でない comment を先頭 `MaxCommentsPerStory` 件まで、`comment_plain` を本文に使う（`CommentDepth=1`）。
- 失敗時（`docs/decisions/2026-09-02T15-27-00` 準拠）: `*lobsters.Error{Op, Err}` を return。`client.Do` error / 5xx は 1 回だけ即再試行。4xx / parse 失敗は即 return。個別 story 詳細の取得失敗はその story を落として `List` を続行。`hottest.json` の取得失敗は `List` ごと失敗。
- `ctx` を全 HTTP リクエストへ伝播。Adapter 独自 timeout なし。

## 6. Constraints

- Infrastructure → Application は `port` のみ import。
- 逐次取得。並行化しない。
- secret を持たない（Lobsters JSON は認証不要）。
- API JSON の struct を package 外へ露出しない。
- `comment_plain` があるため HTML 除去処理は原則不要。`comment`（HTML）を使わない。

## 7. Acceptance Criteria

- [ ] `List` が `panic` を含まず、port 契約を満たす。
- [ ] `created_at == since` は含み、`since` より前は除外する（境界）。
- [ ] `is_deleted` / `is_moderated` の comment が結果に現れない。
- [ ] comment 本文が `comment_plain` から取られている（`comment` の HTML ではない）。
- [ ] comment 取得数が `MaxCommentsPerStory` で頭打ちになる。
- [ ] 該当なしで非 nil の空 slice を返す。
- [ ] `client` nil / 非200 / 壊れた JSON の各経路で `*lobsters.Error` が返り、`Error()` が `"lobsters:"` prefix、`Unwrap()` 非 nil。
- [ ] 個別 story 詳細の取得が1件失敗しても、他 story は結果に残る。
- [ ] `hottest.json` の取得が失敗したら `List` 全体が `*lobsters.Error` を返す。
- [ ] `item_source_sociable_unit_test.go` の6本が実装され緑。
- [ ] `apps/generator/test/lobsters_narrow_integration_test.go` が実 HTTP を TLS redirect で観測して緑。
- [ ] Narrow が secret なしで Integration gate と coverage 分母に入る。

## 8. Verification

```bash
cd apps/generator
go build ./...
go vet ./...
go test ./internal/infrastructure/lobsters/... -run TestList -v
go test ./test/ -run TestLobsters -v
golangci-lint run ./...
gofmt -l .
```

- 完了判定は `lobsters` package と `TestLobsters*` Narrow が緑になること（broad integration の SKIP は別 task で解消）。
- coverage: `./scripts/test-unit.sh`。

## 9. Dependencies

- Blocked by: なし（scope-split A 完了済み）。
- Related: `generator-hackernews-adapter.md` / `generator-itmedia-adapter.md`（同型・並行可）、`generator-source-adapters-wiring.md`（この3本の後）。

## 10. Risks

- Lobsters の `hottest.json` は数十件しか返さない（実測）。`MaxStoriesScanned`（stub の値）がほぼ全件に相当する日がある。`since` フィルタで少数・0 件になる日もある。0 件は非 nil 空 slice で正常。
- `created_at` の time zone offset（`-05:00` 等）の parse 漏れで `OccurredAt` がずれる。UTC 化を必ず通す。

## 11. Notes

- HackerNews Adapter と実装形が近い（一覧取得 → 個別詳細 → `Context` 組み）。同 branch で並行実装するなら `Context` 行組み helper の重複を確認し、必要なら共通化を検討（ただし2 Adapter 時点では YAGNI 寄り。3つ目の ITmedia は RSS で形が違う）。
- delivery の kind 分類 type switch は現ブランチに無い（merge 時作業。lane D 記録済み）。
