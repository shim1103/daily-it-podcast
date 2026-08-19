# refactor(generator): X Adapter を ItemSource.List に載せる

GitHub Issue 化はしない。契約の正は `docs/decisions/2026-08-19T13-25-20-refactor-generator-source-port.md`。Port / `SourceItem` / `FetchSourceItems` は実装済み。再設計しない。

## 1. Summary

このIssueでは、GetXAPI と TwitterAPI.io の Driven Adapter を `port.ItemSource.List` に載せ、監視 user と vendor JSON→`SourceItem` を Infrastructure に閉じる。完了後、両 Adapter と Composition がコンパイルし、Application は X user id を知らないまま `List` 1回で配列を受け取る。

## 2. Context

事実:

1. Application は `ItemSource.List(ctx, since)` と `models.SourceItem{SourceID, OccurredAt, Context}` に既に移行済み。`Post` / `PostSource.ListByUser` / `FetchWatchedPosts` は削除済み
2. 両 X Adapter の `List` は空 slice を返す Stub。`ctx` と `since` は Dummy 引数。vendor 取得の参照実装は各 `post_source.go` 末尾の `todo:` block comment（旧 `ListByUser`）に残す
3. `WatchUserIDs` は `entities/constants` に残っている。Application は参照しない
4. Composition factory は `NewGetXAPIItemSource` / `NewTwitterAPIIOItemSource` で `port.ItemSource` を返す

user が明示した決定:

1. 必須 schema は `SourceID` と `OccurredAt` だけ。それ以外は `Context`（非 schema text）。optional な構造化 field は Port に置かない
2. `OccurredAt` は各情報源データに付いている発生時刻。取得日時でも CLI `now` でもない。`now` は UseCase が `since` を作るためだけ
3. `item_id` / `actor_id` は TextWriter が同一人物を推論するために `Context` へ書く。必須 schema ではない
4. GetXAPI と TwitterAPI.io は必要な vendor field だけ取得・decode する。相手に無いものを揃えない
5. 第2情報源 Adapter と並列 composite は作らない

仮定:

1. GetXAPI の `author` に表示名が常にあるかは未確認。decode して取れたら `Context` に書く。無ければ行を書かない。dummy 名は置かない
2. TwitterAPI.io の fixture には `author.userName` がある。decode するなら `Context` へ。今の `rawAuthor` は `ID` だけでもよい
3. vendor HTTP は今どおり user 単位 page。Port の `List` 1回は vendor 1 call を意味しない

## 3. Canonical Sources

1. `docs/decisions/2026-08-19T13-25-20-refactor-generator-source-port.md` — Port 契約（必須2 field、`Context`、`OccurredAt` の意味、`SourceID "x"`）
2. `docs/decisions/2026-08-17T17-15-01-feature-x-getxapi-adapter.md` — `x/` は facade ではない。vendor 切替は Composition
3. `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md` — オリジナルのみ、24h、試作/本運用。戻り形は本 decision が上書き
4. `apps/generator/internal/application/port/item_source.go` — `ItemSource.List` の SSoT
5. `apps/generator/internal/entities/models/source_item.go` — `SourceItem`
6. `apps/generator/internal/entities/constants/fetch_window.go` — `since` 幅。`OccurredAt` ではない
7. `apps/generator/internal/entities/constants/watch_users.go` — 移す対象の dummy id
8. `apps/generator/internal/application/fetch_source_items.go` — UseCase。触らない
9. `architecture/ports-adapters` — Port 所有は Application。Adapter 追加は mechanism のみ
10. test 方針は `testing-strategy` skill を参照する

## 4. Scope

### In Scope

1. `WatchUserIDs` を `infrastructure/x/` 配下へ移す（両 vendor が参照する情報源設定。facade ではない）
2. GetXAPI が `List(ctx, since)` を実装し、watch users を内部で全部集め 1 配列で返す
3. TwitterAPI.io も同契約。`media` は vendor に無いので `Context` に書かない
4. vendor JSON → `SourceItem`（`toPost` / `models.Post` / `models.Media` は使わない。旧 `toPost` は各 `post_source.go` の `todo:` block comment を `toSourceItem` へ書き換えて移植する）
5. filter 専用 field（`isReply` / quoted / retweeted）は `Context` に出さない
6. 隣の sociable unit を `List` 契約へ付け替える
7. Composition が両 Adapter を `port.ItemSource` として結線できること（factory 名は実装済み）

### Out of Scope

1. `ItemSource` / `SourceItem` / `FetchSourceItems` の再設計
2. 第2情報源 Adapter（ニュース・RSS 等）
3. 複数情報源の並列 composite
4. Cursor TextWriter / TTS / Drive / cmd / GHA
5. GitHub Issue 化
6. vendor 共通 pager / 正規化 package（既存 decision）

## 5. Contract

**Port（実装済み。従うだけ）**

`List(ctx, since) ([]models.SourceItem, error)`

1. `since` は `OccurredAt` の inclusive 下限
2. `SourceID` 非空。`OccurredAt` は UTC かつ `since` 以上
3. 該当なしは空 slice（nil ではない）
4. 必須 field が空の要素を成功として返さない
5. vendor 型・watch user 一覧を Port の外へ出さない

**X 共通**

1. `SourceID` は `"x"`。vendor 名（`getxapi` / `twitterapiio`）は使わない
2. `OccurredAt` は vendor `createdAt` を parse した時刻。HTTP した時刻ではない
3. `List` は `WatchUserIDs` を内部で回す。引数に `userID` は無い
4. 先頭 user の取得が error なら即 return。成功分の部分結果は返さない。以降の user は呼ばない
5. Reply / Repost / 引用は今どおり含めない
6. `createdAt < since` で page を打ち切る仮定は現状維持

**`Context`（非 schema。取れた行だけ。欠けた行は書かない）**

行の語形（Adapter 内の出力規約。Port は `string` としか知らない）:

```text
item_id: {tweet id}
actor_id: {author.id}
actor_name: {表示名。取れたときだけ}
text: {本文}
permalink: {投稿 URL}
links: {expanded_url を空白区切り。1つ以上あるときだけ}
media: {media URL を空白区切り。1つ以上あるときだけ}
```

GetXAPI: 上記のうち取れたもの。`media` は `media[].url` が空でないとき。

TwitterAPI.io: `media` 行は書かない（field が無い）。`links` は expanded_url があるときだけ。

`item_id` と `actor_id` は TextWriter が「この投稿とこの投稿は同一人物」と推論するための材料である。fetch 用の必須 schema ではない。

**変換表（GetXAPI）**

| vendor | 行き先 |
|---|---|
| 固定 `"x"` | `SourceID` |
| `createdAt` | `OccurredAt` |
| `id` | Context `item_id` |
| `author.id` | Context `actor_id` |
| `author` 表示名（取れたら） | Context `actor_name` |
| `text` | Context `text` |
| `url` | Context `permalink` |
| `entities.urls[].expanded_url` | Context `links` |
| `media[].url` | Context `media` |
| `isReply`, `quoted_tweet` | 捨てる（filter） |
| like / lang / profile | 取得しない |

**変換表（TwitterAPI.io）**

| vendor | 行き先 |
|---|---|
| 固定 `"x"` | `SourceID` |
| `createdAt` | `OccurredAt` |
| `id` | Context `item_id` |
| `author.id` | Context `actor_id` |
| `author.userName`（取るなら） | Context `actor_name` |
| `text` | Context `text` |
| `url` | Context `permalink` |
| `entities.urls[].expanded_url` | Context `links` |
| `media` | 無い。書かない。空 media も作らない |
| `isReply`, `quoted_tweet`, `retweeted_tweet` | 捨てる（filter） |

## 6. Constraints

1. Application / Entities に X user id・tweet 形・permalink 型・media 型を戻さない
2. `WatchUserIDs` を Application から読まない
3. 両 vendor の mechanism を共通 package にまとめない
4. 秘密値を code に書かない
5. Infrastructure が Application から import してよいのは Port のみ
6. `Context` を JSON にして Port 側で parse する形にしない

## 7. Acceptance Criteria

- [ ] AC-1: 両 Adapter が `port.ItemSource` を満たし、`ListByUser` / `models.Post` / `models.Media` が generator から消えている
- [ ] AC-2: `List` 1回で `WatchUserIDs` 全 user の original が 1 配列に載る。引数に user id は無い
- [ ] AC-3: 両 vendor の各要素 `SourceID` は `"x"`。vendor 名ではない
- [ ] AC-4: `OccurredAt` は vendor `createdAt`。`OccurredAt < since` の要素は含まれない。Reply / Repost / 引用は含まれない
- [ ] AC-5: GetXAPI の成功経路 `Context` に `item_id` / `actor_id` / `text` が含まれる。media URL があるとき `media` 行がある
- [ ] AC-6: TwitterAPI.io の成功経路 `Context` に `media` 行が無い
- [ ] AC-7: 必須 field（`SourceID` / `OccurredAt`）が空の要素を成功にしない
- [ ] AC-8: 先頭 watch user の取得失敗で error を返し、部分結果を返さない
- [ ] AC-9: `WatchUserIDs` が Application / Entities から消え、X infra にある
- [ ] AC-10: Composition の factory が `port.ItemSource` として結線できる
- [ ] AC-11: `cd apps/generator && go test ./internal/infrastructure/x/... ./internal/composition/...` が pass
- [ ] AC-12: `scripts/test-unit.sh` が pass（coverage / depguard を壊さない）

## 8. Verification

```bash
cd apps/generator && go test ./internal/infrastructure/x/... ./internal/composition/... ./internal/application/...
```

```bash
./scripts/test-unit.sh
```

1. Drive / 実 X HTTP は叩かない。既存どおり proxy Stub
2. 本番 credential を読まない

## 9. Dependencies

blocked by: なし。Port / UseCase は先行済み。

blocks: Cursor TextWriter が `Context` を読む前提。TextWriter 実装自体はこの Issue の対象外。

## 10. Risks

1. GetXAPI に表示名が無い → dummy を置かず `actor_name` 行を省略する
2. `List` が全 watch user を内部で回すと sociable unit の Given が「引数の user id」から「定数の watch 一覧」に変わる。定数を test で差し替えられる置き方にする
3. 途中で旧 `ListByUser` を残して両対応しない。Stub の空 `List` を本番取得に置き換え、参照 block comment を削除する

## 11. Notes

採らない: Port に optional の `Permalink` / `Media` / `Actor` を足す。必須以外は `Context`。

follow-up（この Issue ではやらない）: 第2情報源、並列 composite。Port 対応の確認用捏造例（実装しない）:

1. ニュース: `SourceID="news"`、`OccurredAt=PublishedAt`、`Context` に `org_id` / `org_name` / 居るときだけ `actor_id` / `actor_name`
2. RSS: `SourceID="rss"`、`OccurredAt=PubDate`、author が空なら actor 行を書かない

既存 `toPost` を `SourceItem` の薄い wrapper にして `Post` 相当 field を復活させない。
