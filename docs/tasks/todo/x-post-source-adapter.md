## Issue1 draft — TwitterAPI.io PostSource Adapter

GitHub Issue 未作成。作成時はこの本文を使う。

**Title（案）:** `feat(generator): TwitterAPI.io で PostSource を実装する`  
**type / priority（案）:** `feat` / （`workflow/constants.toml` 確定後に合わせる）

```markdown
## 1. Summary

このIssueでは、TwitterAPI.io 向け Adapter で既存の `PostSource` 契約を満たし、Composition から結線できる状態にする。
完了後、Infrastructure 経由で監視 user の時間窓内オリジナル投稿を `models.Post` として取得できる。

## 2. Context

- Port・Domain 型・監視定数・後回し範囲は manager 側で stub / decision 済み。
- 試作 vendor は TwitterAPI.io。本運用 GetXAPI は別Issue。
- 当面はオリジナル投稿のみ（Reply / Repost / 引用は後回し）。

## 3. Canonical Sources

- Port 契約 — `apps/generator/internal/application/port/post_source.go`
- 戻り型 — `apps/generator/internal/entities/models/post.go`
- 窓・監視 id — `apps/generator/internal/entities/constants/`
- 採用API — `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md`
- 後回し範囲 — `docs/decisions/2026-08-15T17-43-09-feature-x-api-adoption.md`
- 層・依存 — `DESIGN.md`
- 設計哲学（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/philosophy/SKILL.md` および同 dir の `design-philosophy.md`
- 書き方・公開契約 documentation（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/coding-style/SKILL.md`（`naming.md` / `comments.md` / `function-design.md`）
- test方針（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md`（levels / contracts / naming-and-layout / credential 等は同 dir）

## 4. Scope

### In Scope

- `infrastructure/x/twitterapiio/` に HTTP client と `PostSource` 実装
- cursor ページングと `since` による打ち切り
- raw → `models.Post` 変換
- Infrastructure Error への変換
- Composition でのこの Adapter 結線（`TWITTERAPI_IO_API_KEY`）
- Port 契約を検証する test

### Out of Scope

- `application/port/`・`entities/` の変更
- UseCase（監視 user 一括取得）— 別Issue
- GetXAPI Adapter — 別Issue
- trends / Reply / Repost / 引用 / profile cache
- upsert・DB・media ローカル保存本体
- GHA cron 全体・本番 `WatchUserIDs` の差し替え

## 5. Contract

- 既存 `PostSource.ListByUser` の `@require` / `@ensure` / `@invariant` を変えない・満たす。
- 戻りは `[]models.Post` のみ。vendor 固有型を Application へ出さない。
- API key は env `TWITTERAPI_IO_API_KEY`。code・commit に値を書かない。

## 6. Constraints

- 書いてよい path: `apps/generator/internal/infrastructure/x/twitterapiio/**`、および Composition のこの Adapter 結線に限る変更。
- `port/`・`entities/` は変更禁止。
- 後回し decision の対象を実装しない。
- philosophy / coding-style / testing-strategy は上記絶対 path を正とし、Issue 本文や実装コメントへ同内容を再定義しない。
- 公開境界の契約は Port 宣言箇所のみを SSoT とする（coding-style）。

## 7. Acceptance Criteria

- [ ] AC-1: `PostSource` を実装する型が `twitterapiio` 配下にあり、Composition から注入できる。
- [ ] AC-2: `ListByUser` がオリジナル投稿のみ返し、Reply / Repost / 引用を含まない。
- [ ] AC-3: 戻り要素の `CreatedAt` がすべて `since` 以上である。
- [ ] AC-4: 戻りが `models.Post` の field のみで、vendor 型が Application から参照されない。
- [ ] AC-5: 秘密は `TWITTERAPI_IO_API_KEY` 経由のみで、リポジトリに key 値がない。
- [ ] AC-6: Port 契約を検証する test が、testing-strategy に沿って追加され pass する。

## 8. Verification

- test の Scope×Sociability・配置・credential 扱いは `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md` に従う。
- `apps/generator` 配下で、追加した Port 契約向け test を実行し pass する。
- go 未導入環境なら、実行手順と結果を報告に残す（代替: CI またはローカルで同 test）。

## 9. Dependencies

- blocked by: なし（Port / 型 / 定数は既存）。
- related: UseCase 一括取得 Issue、GetXAPI Adapter Issue（未作成可）。

## 10. Risks

- TwitterAPI.io の media 有無が docs と実レスポンスでずれる → 無ければ `Media` 空で契約どおりとし、有れば写す。
- reverse-engineered API の一時障害 → Infrastructure Error に閉じ、Port 契約は維持する。

## 11. Notes

- 本運用 GetXAPI・UseCase・後回し取得は follow-up。このIssueの完了条件に入れない。
```
