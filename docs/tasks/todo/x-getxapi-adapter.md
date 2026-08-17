## Issue3 draft — GetXAPI PostSource Adapter

GitHub Issue 未作成。作成時はこの本文を使う。

**Title（案）:** `feat(generator): GetXAPI で PostSource を実装する`  
**type / priority（案）:** `feat` / （`workflow/constants.toml` 確定後に合わせる）

```markdown
## 1. Summary

このIssueでは、本運用 vendor の GetXAPI 向け Adapter で既存の `PostSource` 契約を満たし、Composition から TwitterAPI.io 実装を残したまま結線できるようにする。
完了後、AgentSecrets proxy 経由で監視 user の時間窓内オリジナル投稿を `models.Post` として取得できる。

## 2. Context

- Port・Domain 型・監視定数・後回し範囲は stub / decision 済み。
- TwitterAPI.io Adapter は `infrastructure/x/twitterapiio/` に実測済み。HTTP は `agentsecrets.Client.Do`、秘密値は保持しない。
- UseCase `application.FetchWatchedPosts` は rebase 後の `origin/develop`（PR #14）で既存。Port だけを呼び、本 Adapter の path と独立。
- GetXAPI User Tweets（docs）は `GET https://api.getxapi.com/twitter/user/tweets`。query は `userId`（推奨）と `cursor`。auth は `Authorization: Bearer`。page は `has_more` / `next_cursor`。
- TwitterAPI.io は `X-API-Key` のため `Inject.Headers` だった。GetXAPI は Bearer が公式と一致するので `Inject.Bearer: secretnames.GetXAPIKeyName` を使う（proxy test が同キー名で先例）。
- 秘密キー名の正は README / `secretnames.GetXAPIKeyName`（`GETX_API_KEY`）。新規名を足さない。
- 仮定: User Tweets の並びは新しい順。`createdAt` が `since` を下回ったら以降 page を呼ばない（twitterapiio と同じ打ち切り）。docs は順序を明記していない。

## 3. Canonical Sources

- Port 契約 — `apps/generator/internal/application/port/post_source.go`
- 戻り型 — `apps/generator/internal/entities/models/post.go`
- 窓・監視 id — `apps/generator/internal/entities/constants/`
- 採用API — `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md`
- 後回し範囲 — `docs/decisions/2026-08-15T17-43-09-feature-x-api-adoption.md`
- 試作 Adapter（参照実装）— `apps/generator/internal/infrastructure/x/twitterapiio/`
- 試作 Adapter の sociable unit — `apps/generator/internal/infrastructure/x/twitterapiio/post_source_sociable_unit_test.go`
- Composition 結線先例 — `apps/generator/internal/composition/twitterapiio.go`
- AgentSecrets proxy — `apps/generator/internal/infrastructure/agentsecrets/proxy.go`
- 秘密名 — `apps/generator/internal/infrastructure/secretnames/names.go`、`README.md`
- UseCase（触らない）— `apps/generator/internal/application/fetch_watched_posts.go`
- GetXAPI User Tweets — https://docs.getxapi.com/docs/users/user-tweets
- 層・依存 — `DESIGN.md`
- 設計哲学（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/philosophy/SKILL.md` および同 dir の `design-philosophy.md`
- 書き方・公開契約 documentation（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/coding-style/SKILL.md`（`naming.md` / `comments.md` / `function-design.md`）
- test方針（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md`（levels / contracts / naming-and-layout / credential 等は同 dir）

## 4. Scope

### In Scope

- `infrastructure/x/getxapi/` に `PostSource` 実装（`agentsecrets.Client` 経由。vendor へ直接 `net/http` しない）
- User Tweets: `userId` + 空 cursor 省略、`has_more` / `next_cursor` で page、`createdAt` が `since` 未満で打ち切り
- raw → `models.Post`（`entities.urls[].expanded_url` → `URLs`。docs sample の `media[]` がある場合は `models.Media` へ写す）
- オリジナル判定: `isReply` と `quoted_tweet` 非 null を除外。`retweeted_tweet` は docs sample に無い。実測で同 field があれば除外し、無ければ Posts tab の応答に従う
- Infrastructure Error（twitterapiio の `Error{Op, Err}` と同型の package 内型）
- `composition.NewGetXAPIPostSource` の追加。twitterapiio factory は残す
- Port 契約を検証する sociable unit（twitterapiio の test 形を踏襲）

### Out of Scope

- `application/port/`・`entities/`・`application/fetch_watched_posts.go` の変更
- `infrastructure/x/twitterapiio/` の削除・無関係リファクタ
- `agentsecrets` / `secretnames` の変更（既存 `GetXAPIKeyName` と `Inject.Bearer` で足りる）
- cmd 入口がどちらを呼ぶかの切り替え（cmd 入口 Issue）
- `/twitter/user/tweets_and_replies` および `/twitter/user/tweets/complete`
- trends / Reply / Repost / 引用の取得・profile cache
- upsert・DB・media ローカル保存本体
- GHA cron 全体・本番 `WatchUserIDs` の差し替え

## 5. Contract

- 既存 `PostSource.ListByUser` の `@require` / `@ensure` / `@invariant` を変えない・満たす。
- 戻りは `[]models.Post` のみ。vendor 固有型を Application へ出さない。
- 外向き HTTP は `agentsecrets.Client.Do`。`TargetURL` は https。`Inject.Bearer` にキー名 `GETX_API_KEY` だけを載せる。`Inject.Headers` は使わない。
- 秘密値は Adapter が保持しない。code・commit に値を書かない。
- 該当なしは空 slice（nil ではない）。twitterapiio と同じ。

## 6. Constraints

- 書いてよい path: `apps/generator/internal/infrastructure/x/getxapi/**`、および `composition` の GetXAPI factory 追加のみ。
- `port/`・`entities/`・`twitterapiio/`・UseCase・`agentsecrets`・`secretnames` は変更禁止。
- 後回し decision の対象を実装しない。
- philosophy / coding-style / testing-strategy は上記絶対 path を正とし、Issue 本文や実装コメントへ同内容を再定義しない。
- 公開境界の契約は Port 宣言箇所のみを SSoT とする（coding-style）。

## 7. Acceptance Criteria

- [ ] AC-1: `getxapi` 配下に `PostSource` 実装があり、`composition.NewGetXAPIPostSource` から注入できる。
- [ ] AC-2: `ListByUser` がオリジナル投稿のみ返し、Reply / 引用を含まない。
- [ ] AC-3: 戻り要素の `CreatedAt` がすべて `since` 以上である。境界値は含む。下限を下回った page の次は呼ばない。
- [ ] AC-4: 戻りが `models.Post` の field のみで、vendor 型が Application から参照されない。
- [ ] AC-5: proxy へ `X-AS-Inject-Bearer: GETX_API_KEY` が渡り、`X-AS-Inject-Header-X-API-Key` は空。リポジトリに key 値がない。
- [ ] AC-6: TargetURL は `https://api.getxapi.com/twitter/user/tweets` で、`userId` を付け、初回 cursor を query に付けない。
- [ ] AC-7: TwitterAPI.io の実装と `NewTwitterAPIIOPostSource` が残っている。
- [ ] AC-8: Port 契約を検証する sociable unit が twitterapiio と同配置・GWT で追加され pass する。

## 8. Verification

- test の Scope×Sociability・配置・credential 扱いは `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md` に従う。
- twitterapiio と同じく、対象ソース隣の sociable unit（external test package、`httptest` の proxy double）にする。
- `apps/generator` 配下で追加 test を実行し pass する。
- go 未導入環境なら、実行手順と結果を報告に残す（代替: CI またはローカルで同 test）。

## 9. Dependencies

- blocked by: なし（Port・proxy・秘密名・UseCase・twitterapiio 実測は `origin/develop` に既存）。
- related: TwitterAPI.io Adapter、`FetchWatchedPosts`、cmd 入口 Issue（未作成）。

## 10. Risks

- GetXAPI の tweet 形は User Tweets docs が正。decision の「公式 X API v2 schema 互換」と docs sample が食い違う → Adapter は docs / 実測の field を変換し、Port 契約は維持する。
- `retweeted_tweet` が docs sample に無い → 無いなら除外しない。実測で出たら twitterapiio と同じく除外する。
- page 順が新しい順でない → `CreatedAt < since` で打ち切ると取りこぼす。実装時に 1 page 内の時刻で判定し、test で固定する。
- reverse-engineered API の一時障害 → Infrastructure Error に閉じ、Port 契約は維持する。

## 11. Notes

- cmd が GetXAPI を呼ぶ切替、UseCase 変更、後回し取得、GHA cron は follow-up。このIssueの完了条件に入れない。
- GitHub Issue 作成は `shim gh create-issue`（本 draft の Markdown 本文を stdin へ）。この todo 作成だけでは作成しない。
```
