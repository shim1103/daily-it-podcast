## Issue3 draft — GetXAPI PostSource Adapter

GitHub Issue 未作成。作成時はこの本文を使う。

**Title（案）:** `feat(generator): GetXAPI で PostSource を実装する`  
**type / priority（案）:** `feat` / （`workflow/constants.toml` 確定後に合わせる）

```markdown
## 1. Summary

このIssueでは、本運用 vendor の GetXAPI 向け Adapter で既存の `PostSource` 契約を満たし、Composition から TwitterAPI.io 実装を残したまま切り替え可能にする。
完了後、Infrastructure 経由で監視 user の時間窓内オリジナル投稿を `models.Post` として取得できる。

## 2. Context

- Port・Domain 型・監視定数・後回し範囲は stub / decision 済み。
- TwitterAPI.io Adapter（Issue1）で Port 契約を実測済み。同じ契約を GetXAPI で満たす。
- UseCase（`FetchWatchedPosts`）は Port 依存のみで、本 Adapter と path が独立。
- 当面はオリジナル投稿のみ（Reply / Repost / 引用は後回し）。
- 秘密キー名の正は README / `secretnames.GetXAPIKeyName`（`GETX_API_KEY`）。

## 3. Canonical Sources

- Port 契約 — `apps/generator/internal/application/port/post_source.go`
- 戻り型 — `apps/generator/internal/entities/models/post.go`
- 窓・監視 id — `apps/generator/internal/entities/constants/`
- 採用API — `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md`
- 後回し範囲 — `docs/decisions/2026-08-15T17-43-09-feature-x-api-adoption.md`
- 試作 Adapter（参照実装）— `apps/generator/internal/infrastructure/x/twitterapiio/`
- Composition 結線先例 — `apps/generator/internal/composition/twitterapiio.go`
- 秘密名 — `apps/generator/internal/infrastructure/secretnames/names.go`、`README.md`
- 層・依存 — `DESIGN.md`
- 設計哲学（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/philosophy/SKILL.md` および同 dir の `design-philosophy.md`
- 書き方・公開契約 documentation（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/coding-style/SKILL.md`（`naming.md` / `comments.md` / `function-design.md`）
- test方針（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md`（levels / contracts / naming-and-layout / credential 等は同 dir）

## 4. Scope

### In Scope

- `infrastructure/x/getxapi/` に HTTP client と `PostSource` 実装
- cursor ページングと `since` による打ち切り
- raw → `models.Post` 変換
- Infrastructure Error への変換
- Composition でのこの Adapter の選択・結線切り替え（`GETX_API_KEY`）
- Port 契約を検証する test

### Out of Scope

- `application/port/`・`entities/` の変更
- `infrastructure/x/twitterapiio/` の削除・無関係リファクタ
- UseCase の変更
- trends / Reply / Repost / 引用 / profile cache
- upsert・DB・media ローカル保存本体
- GHA cron 全体・本番 `WatchUserIDs` の差し替え

## 5. Contract

- 既存 `PostSource.ListByUser` の `@require` / `@ensure` / `@invariant` を変えない・満たす。
- 戻りは `[]models.Post` のみ。vendor 固有型を Application へ出さない。
- API key は秘密名 `GETX_API_KEY`（`secretnames.GetXAPIKeyName`）。code・commit に値を書かない。

## 6. Constraints

- 書いてよい path: `apps/generator/internal/infrastructure/x/getxapi/**`、および Composition の **Adapter 選択・結線切り替えのみ**。
- `port/`・`entities/`・`twitterapiio/` の無関係変更は禁止。
- 後回し decision の対象を実装しない。
- philosophy / coding-style / testing-strategy は上記絶対 path を正とし、Issue 本文や実装コメントへ同内容を再定義しない。
- 公開境界の契約は Port 宣言箇所のみを SSoT とする（coding-style）。

## 7. Acceptance Criteria

- [ ] AC-1: `PostSource` を実装する型が `getxapi` 配下にあり、Composition から注入できる。
- [ ] AC-2: `ListByUser` がオリジナル投稿のみ返し、Reply / Repost / 引用を含まない。
- [ ] AC-3: 戻り要素の `CreatedAt` がすべて `since` 以上である。
- [ ] AC-4: 戻りが `models.Post` の field のみで、vendor 型が Application から参照されない。
- [ ] AC-5: 秘密は `GETX_API_KEY` 経由のみで、リポジトリに key 値がない。
- [ ] AC-6: TwitterAPI.io Adapter の実装が残っており、削除されていない。
- [ ] AC-7: Port 契約を検証する test が、testing-strategy に沿って追加され pass する。

## 8. Verification

- test の Scope×Sociability・配置・credential 扱いは `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md` に従う。
- `apps/generator` 配下で、追加した Port 契約向け test を実行し pass する。
- go 未導入環境なら、実行手順と結果を報告に残す（代替: CI またはローカルで同 test）。

## 9. Dependencies

- blocked by: なし（Port 契約と Issue1 実測は既存。path は独立）。
- related: Issue1 TwitterAPI.io Adapter（`apps/generator/internal/infrastructure/x/twitterapiio/`）、Issue2 UseCase（`FetchWatchedPosts`）。

## 10. Risks

- GetXAPI の response が公式 X API v2 互換想定と実測でずれる → Infrastructure 内で変換し、Port 契約は維持する。
- reverse-engineered API の一時障害 → Infrastructure Error に閉じ、Port 契約は維持する。
- Composition 切り替えで試作 Adapter を誤って消す → AC-6 と「twitterapiio 削除禁止」で防ぐ。

## 11. Notes

- UseCase・後回し取得・GHA cron は follow-up。このIssueの完了条件に入れない。
- GitHub Issue 作成は `shim gh create-issue`（本 draft の Markdown 本文を stdin へ）。この todo 作成だけでは作成しない。
```
