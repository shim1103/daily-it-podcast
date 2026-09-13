## 1. Summary

このIssueでは、Publickey / TechCrunch / CloudWatch の stub `List` を本番取得・写像へ置き換え、HackerNews / Lobsters の写像を behavior test で固定する。完了後、5 源すべてが `SourceItem`（`SourceBody` + `Meta`）契約どおりに材料を返し、A の stub 足場 test は残らない。

## 2. Context

事実:
1. composition は既に 5 Adapter を結線している。新 3 源の `List` は空 slice を返す stub。
2. HackerNews / Lobsters は `toSourceItem` 実装済み。写像方針は Decision `2026-09-13T17-14-10`。
3. 字段形・URL の Purpose 側は Decision `2026-09-13T17-14-00`。型の正本は `models/source_item.go`。

仮定（作業を止めない）:
1. 各 feed URL 定数は現行 package 定数で到達可能。
2. 情報源 retry は Decision `2026-09-02T15-27-00` を 5 源に適用する（追加判断不要）。

## 3. Canonical Sources

1. `apps/generator/internal/entities/models/source_item.go` — `SourceItem` / `SourceBody` 型の正本（A）
2. `apps/generator/internal/infrastructure/{publickey,techcrunch,cloudwatch}/` — `SourceID`・feed URL・stub `List`（A）
3. `apps/generator/internal/infrastructure/{hackernews,lobsters}/item_source.go` — 実装済み写像の正本（A）
4. Decision `2026-09-13T15-08-55` — 採用源セット（B）
5. Decision `2026-09-13T17-14-00` — Purpose 側 URL / 空の意味 / Meta 空可（B）
6. Decision `2026-09-13T17-14-10` — 写像を固定する方針。表は Decision に置かない（B）
7. Decision `2026-09-02T15-27-00` — 情報源 Adapter の retry / fallback なし（B）
8. Decision `2026-09-02T14-41-02` — Adapter は外部記事 HTML を fetch しない（B）
9. test 方針 — `skills/1:terms/testing-strategy`（再掲しない）

## 4. Scope

### In Scope

1. Publickey / TechCrunch / CloudWatch の `List` 実装（feed 取得・FetchWindow 相当の since フィルタ・写像）
2. 新 3 源の A stub 足場 test を behavior test へ置換
3. HackerNews / Lobsters 写像の behavior assert を §5 表と一致させる（不足分のみ）

### Out of Scope

1. TextWriter へどう prompt を渡すか・P1/P2 文言（D・本分類の non-scope）
2. Application の `Links` 件数に応じた先 fetch
3. Adapter 内の外部記事 HTML scrape
4. composition 結線のやり直し（既に 5 源）
5. `SourceItem` 型の再設計

## 5. Contract

外から見える契約（既存 Port / 型は変えない）:
1. 各 Adapter は `port.ItemSource.List(ctx, since)` を満たす
2. 戻り `[]models.SourceItem` の字段は `source_item.go` どおり
3. 該当なしは空 slice（nil ではない）
4. 一覧/feed 本体の失敗は `List` ごと Infrastructure Error。個別 item 失敗は要素落とし（`15-27-00`）

### 実装前の raw→field 表（本 Issue が正本。実装後は各 `toSourceItem` が正本）

#### hackernews（Verification。実装正本は `hackernews.toSourceItem`）

| raw | 行き先 |
|---|---|
| `title` | `Summary` |
| `text`（正規化後） | `Detail.Text` |
| `url` | `Detail.Links` |
| top-level comment 本文（≤`MaxCommentsPerStory`） | `Discourse.Text`（結合） |
| HN item permalink | `Discourse.Links` |
| `id` / `by` | `Meta` |
| `score` / `descendants` / `kids` / `type` | 捨てる |

#### lobsters（Verification。実装正本は `lobsters.toSourceItem`）

| raw | 行き先 |
|---|---|
| `title` | `Summary` |
| `description_plain` | `Detail.Text` |
| `url` | `Detail.Links` |
| `comment_plain`（deleted/mod 除外・上限） | `Discourse.Text`（結合） |
| `short_id_url` | `Discourse.Links` |
| `comments_url` | 捨てる |
| `short_id` / `submitter_user` | `Meta` |
| `score` / `flags` / `tags` / `comment_count` / HTML `description` | 捨てる |

#### publickey

| raw | 行き先 |
|---|---|
| `title` + Atom `summary` | `Summary` |
| Atom `content`（text 化） | `Detail.Text` |
| `link[rel=alternate]` | `Detail.Links` |
| （議論 URL 無し） | `Discourse` 空 |
| `id` / `author` | `Meta` |

#### techcrunch

| raw | 行き先 |
|---|---|
| `title` | `Summary` |
| `description` | `Detail.Text` |
| `link` | `Detail.Links` |
| （議論 URL 無し） | `Discourse` 空 |
| `guid`（item_id。記事 URL にしない） / `dc:creator` | `Meta` |
| `category` / 空の `content:encoded` | 捨てる |

#### cloudwatch

| raw | 行き先 |
|---|---|
| `title` | `Summary` |
| `description` | `Detail.Text` |
| `link`（`?ref` 無し） | `Detail.Links` |
| `rdf:about`（`?ref=rss` 付き） | 捨てる |
| （議論 URL 無し） | `Discourse` 空 |
| （作者無し） | `Meta` 空でよい |
| `dc:subject` | 捨てる |

## 6. Constraints

1. 1 媒体 1 Adapter。RSS 汎用 facade を新設しない（`15-08-55` / 先行 Adapter 方針）
2. `SourceID`・feed URL の正本は各 package 定数。Issue に値を写して正本にしない
3. engagement（score 等）を `SourceItem` に載せない
4. Decision 本文へ写像表を戻さない
5. A/B で固定されていない Port・型・dir を新設しない

## 7. Acceptance Criteria

- [ ] AC-1: Publickey / TechCrunch / CloudWatch の `List` が stub ではなく、§5 表どおり `SourceItem` を返す
- [ ] AC-2: since 窓外の item は結果に含まれない
- [ ] AC-3: 該当 0 件のとき空 slice かつ error nil
- [ ] AC-4: feed 本体の transient HTTP 失敗は `15-27-00` どおり最小 retry のあと Infrastructure Error
- [ ] AC-5: 新 3 源の stub 足場 test が behavior test に置き換わっている（stub-only のまま残っていない）
- [ ] AC-6: HackerNews / Lobsters の写像 assert が §5 表と矛盾しない
- [ ] AC-7: composition 経由の Fetch が 5 源を含む経路で compile / 関連 SU が通る

## 8. Verification

```bash
cd apps/generator
go test ./internal/infrastructure/publickey/... ./internal/infrastructure/techcrunch/... ./internal/infrastructure/cloudwatch/... ./internal/infrastructure/hackernews/... ./internal/infrastructure/lobsters/... ./internal/composition/... ./internal/application/...
```

coverage gate がある場合は既存 `scripts/generator/test-unit.sh`（または repo 正の unit gate）も pass。

## 9. Dependencies

blocked by: なし（A/B 確定済み）

blocks: TextWriter prompt の P1/P2 反映（別 D / 将来 C。本 Issue の Out of Scope）

## 10. Risks

1. feed schema が cache 実測と違う日がある → Narrow または SU で実 body の代表形を固定し、壊れやすい HTML scrape に逃げない
2. CloudWatch RDF / Publickey Atom の parse 差分 → 媒体ごとに Adapter 内で閉じ、共通 facade を作らない

## 11. Notes

1. 実装後は §5 表の正本を各 Adapter code へ移し、本 Issue の表は Verification 照合用の記録になる
2. follow-up（本 Issue 外）: `Links` 件数に応じた Application 先 fetch、composite の時系列 sort
