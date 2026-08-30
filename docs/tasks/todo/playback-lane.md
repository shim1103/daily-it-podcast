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
- [ ] deploy Phase 1–3（下記）

### Deploy 残

| Phase | 内容 | 状態 |
|-------|------|------|
| 1 | deploy 前ゲート | `docs/tasks/todo/playback-deploy-pre-gate.md` |
| 2 | 初回 `wrangler deploy` + `DEPLOY.md` §7 Verification（OTP・一覧・再生） | 未 |
| 3 | 運用後続（下記） | 未（Phase 2 後） |

**依存:** Phase 1 → 2 → 3。

#### Phase 2（初回 deploy + Verification）

1. `cd apps/playback && npm run build && npx wrangler deploy`
2. 出力 FQDN を記録（`daily-it-podcast.<subdomain>.workers.dev`）
3. Phase 1 で workers.dev を Disabled にした場合はここで Enable
4. `DEPLOY.md` §7 すべて。シークレット窓で許可 email / 拒否 email を分けて確認。

#### Phase 3（運用後続）

1. rollback 手順の文書化
2. logging / observability
3. DAST / penetration test（test URL・identity・攻撃対象が未決 — 保留）
4. CD / hook 自動 deploy（非 scope）
5. Dependabot / Renovate（優先度低）

### Integration / E2E 方針

```text
gate = secret なし Narrow + Broad（Decision 2026-08-30T16-20-00）
E2E = gate 外・週次月曜 07:00 JST + dispatch（同 Decision）
分類語 = e2e（同 Decision）
coverage 分母 = SU + secret なし NI（Decision 2026-08-30T16-20-01）
Page BI 当面なし / API Stub=SU・真 HTTP=NI（Decision 2026-08-30T16-20-02）
OTP 手動・週次 storageState・Drive=Worker（Decision 2026-08-30T16-20-03）
deploy = 3 Phase（Decision 2026-08-30T16-20-04）
```

### C（達成契約）

1. [x] Drive Narrow Integration（`apps/playback/test/*gdrive*narrow_integration*.test.ts`。達成契約 file は完了削除）
2. [ ] API Client Narrow Integration — `docs/tasks/todo/playback-narrow-integration-api-client.md`
3. [x] Worker Broad Integration 正常系（`apps/playback/test/*broad_integration*.test.ts` の Composition 正常系。達成契約 file は完了削除）
4. [ ] 認証済み browser E2E — `docs/tasks/todo/playback-e2e-browser-authenticated.md`

音声は wav。generator 書込とは共有しない。
