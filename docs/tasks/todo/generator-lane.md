## Generator 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md

取得 → Cursor CLI 原稿 → Gemini TTS → Drive 書込を Go CLI + GHA で通す。

- [x] go.mod（module path）と `PostSource` / `Post` / 監視定数の境界 stub
- [ ] 情報取得 Adapter（Issue1 draft: `x-post-source-adapter.md`。Issue 未作成）
- [ ] 監視 user 一括取得 UseCase（Issue2 draft: `x-fetch-watched-posts-usecase.md`。Issue 未作成）
- [ ] GetXAPI Adapter（Issue3・下に詳細。Issue / draft md 未作成）
- [ ] cmd 入口
- [ ] Cursor CLI / Gemini / Drive の Infrastructure
- [ ] GHA workflow で定期または手動実行

### 未完了: Issue3 GetXAPI Adapter（本運用差し替え）

Issue1（TwitterAPI.io）で Port 契約を実測したあと、同じ `PostSource` を GetXAPI で満たす。create-issue template の正式 draft はまだ作らない。終わってない task として境界だけ残す。

- 目的: 本運用 vendor を GetXAPI にし、Composition から切り替え可能にする
- 依存: Issue1 完了後が安全（契約・変換の実測後）。path は独立
- 所有 path: `apps/generator/internal/infrastructure/x/getxapi/`
- 触ってよい: 上記 dir、Composition の **Adapter 選択・結線切り替えのみ**（twitterapiio 実装を消さない／無関係リファクタしない）
- 触禁止: `application/port/`、`entities/`、`infrastructure/x/twitterapiio/` の無関係変更、後回し decision 対象
- 受け入れの正: 既存 `PostSource` 契約（変えない・満たす）。vendor 型非漏出
- 秘密: 作成時に env 名を固定（例案 `GETXAPI_API_KEY`）。値は secrets。code に書かない
- やること: HTTP client、cursor / since 打ち切り、raw → `models.Post`、Infrastructure Error、契約用 test、Composition 切り替え
- やらない: UseCase、trends、Reply/Repost/引用、profile cache、upsert/DB、media ローカル保存本体、GHA cron 全体
- 参照: `docs/decisions/2026-08-15T16-39-20-feature-x-api-adoption.md`、`docs/decisions/2026-08-15T17-43-09-feature-x-api-adoption.md`、Port / Post / constants、Issue1 draft
- 次アクション: Issue1 検証後に create-issue template で正式 draft → gh 作成
