## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`  
deploy・Access の正: `DEPLOY.md`  
React + Hono 方針の正: `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md`

未完了の達成契約は `docs/tasks/todo/playback-*.md` が正。本 lane は進捗 index のみ。GitHub Issue 化しない運用。

- [x] web↔worker HTTP / Drive adapter / UI / 層検知
- [x] deploy A/B（`wrangler.jsonc`・`worker-entry`・`DEPLOY.md`）
- [x] React + Hono（E）
- [x] deploy 前 C（toolchain・runtime config・Access 設定）
- [ ] deploy Phase 1–4（下記）

### Deploy 残

| Phase | 内容 | 状態 |
|-------|------|------|
| 1 | deploy 前ゲート | `docs/tasks/todo/playback-deploy-pre-gate.md` |
| 2 | 初回 `wrangler deploy`（本番 URL 出現） | 未 |
| 3 | `DEPLOY.md` §5 検証（OTP・一覧・再生） | 未（Phase 2 後） |
| 4 | 運用後続（下記 D） | 未 |

**依存:** Phase 1 → 2 → 3。Phase 4 は 3 後でも可。

#### Phase 2（初回 deploy）

1. `cd apps/playback && npm run build && npx wrangler deploy`
2. 出力 FQDN を記録（`daily-it-podcast.<subdomain>.workers.dev`）
3. Phase 1 で workers.dev を Disabled にした場合はここで Enable

#### Phase 3（本番検証）

`DEPLOY.md` §5 すべて。シークレット窓で許可 email / 拒否 email を分けて確認。

#### Phase 4（D: 運用後続）

1. rollback 手順の文書化
2. logging / observability
3. DAST / penetration test（test URL・identity・攻撃対象が未決 — 保留）
4. CD / hook 自動 deploy（非 scope）
5. Dependabot / Renovate（優先度低）

音声は wav。generator 書込とは共有しない。
