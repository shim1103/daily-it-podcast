## Issue2 draft — 監視 user 一括取得 UseCase

GitHub Issue 未作成。作成時はこの本文を使う。

**Title（案）:** `feat(generator): 監視 user の Post 一括取得 UseCase を追加する`  
**type / priority（案）:** `feat` / （`workflow/constants.toml` 確定後に合わせる）

```markdown
## 1. Summary

このIssueでは、`WatchUserIDs` と `FetchWindow` を使い、既存の `PostSource` を呼んで監視対象全員のオリジナル投稿を集める UseCase を追加する。
完了後、Application 層だけで「定数の全 user × 時間窓」の取得手順を Fake Port 付きで実行・検証できる。

## 2. Context

- Port・Domain 型・監視定数は manager 側で stub 済み。
- TwitterAPI.io Adapter（Issue1）とは path が独立しており、Port さえあれば並行実装できる。
- UseCase は Infra・vendor・env を知らない。

## 3. Canonical Sources

- Port 契約 — `apps/generator/internal/application/port/post_source.go`
- 戻り型 — `apps/generator/internal/entities/models/post.go`
- 窓・監視 id — `apps/generator/internal/entities/constants/`
- 採用API・取得方針 — `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md`
- 後回し範囲 — `docs/decisions/2026-08-15T17-43-09-feature-x-api-adoption.md`
- Issue1 Adapter — `apps/generator/internal/infrastructure/x/twitterapiio/`、結線は `apps/generator/internal/composition/`
- 層・依存 — `DESIGN.md`
- 設計哲学（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/philosophy/SKILL.md` および同 dir の `design-philosophy.md`
- 書き方・公開契約 documentation（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/coding-style/SKILL.md`（`naming.md` / `comments.md` / `function-design.md`）
- test方針（絶対参照・再定義禁止） — `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md`（levels / contracts / naming-and-layout / test-doubles 等は同 dir）

## 4. Scope

### In Scope

- `apps/generator/internal/application/` 配下への UseCase 追加（監視 user 全員分の `PostSource.ListByUser` 呼び出し）
- `WatchUserIDs` と `now - FetchWindow`（または同等の since 算出）を引数組み立てに使う
- Port の Fake / Stub による UseCase unit test

### Out of Scope

- `application/port/`・`entities/` の変更
- `infrastructure/**` の実装・変更
- Composition 結線（Issue1 の Adapter 結線に含める）
- TwitterAPI.io / GetXAPI の HTTP・env
- trends / Reply / Repost / 引用 / profile cache
- upsert・DB・media ローカル保存・GHA cron
- 本番 `WatchUserIDs` の差し替え（dummy のままでよい）

## 5. Contract

- 依存は `PostSource` と Entities のみ。Infrastructure を import しない。
- 各 `WatchUserIDs` 要素について `ListByUser(ctx, userID, since)` を呼ぶ。
- `since` は `FetchWindow` に基づく（契約の時間窓と一致）。
- 公開する UseCase 入口がある場合、契約 documentation は宣言箇所のみ（coding-style）。Port 契約は再掲・変更しない。

## 6. Constraints

- 書いてよい path: `apps/generator/internal/application/` 配下の **UseCase とその test のみ**（`port/` は変更禁止）。
- `infrastructure/**`・`composition/**`・`entities/**`・`port/` は変更禁止。
- philosophy / coding-style / testing-strategy は上記絶対 path を正とし、Issue 本文や実装コメントへ同内容を再定義しない。
- 後回し decision の対象を実装しない。

## 7. Acceptance Criteria

- [ ] AC-1: UseCase が `constants.WatchUserIDs` の各 id に対して `PostSource.ListByUser` を呼ぶ。
- [ ] AC-2: `since` が `FetchWindow` に基づいて決まる。
- [ ] AC-3: UseCase の source が `infrastructure` を import していない。
- [ ] AC-4: Fake / Stub の `PostSource` で unit test が追加され pass する。
- [ ] AC-5: `port/`・`entities/` に差分がない。

## 8. Verification

- test の Scope×Sociability・Double・配置は `/Users/shim0729/.claude/skills/testing-strategy/SKILL.md` に従う。
- UseCase 向け unit test を実行し pass する。
- go 未導入環境なら、実行手順と結果を報告に残す。

## 9. Dependencies

- blocked by: なし（Port 契約があれば足りる。Issue1 Adapter 完了は不要）。
- related: Issue1 Adapter（`apps/generator/internal/infrastructure/x/twitterapiio/`）、GetXAPI Adapter Issue（未作成可）。

## 10. Risks

- Issue1 と同時変更で `application/` 配下の file 名が衝突する → UseCase 用の新規 file 名をこのIssueで固定し、`port/` は触らない。
- エラー時に一部 user だけ成功する挙動が曖昧 → 最初の失敗で返すか全 user 試すかを UseCase 契約に一文で固定する（実装前に宣言箇所へ書く）。

## 11. Notes

- Adapter・Composition・vendor は Issue1。このIssueは手順（Application）だけ。
- 一 user 失敗時の方針は Notes の Risks どおり、UseCase 公開境界の契約に含める。
```
