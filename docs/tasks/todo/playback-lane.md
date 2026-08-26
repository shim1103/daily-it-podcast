## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`  
deploy・Access の正: `DEPLOY.md`  
React + Hono 方針の正: `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md`

未完了の達成契約は `docs/tasks/todo/playback-*.md` が正。本 lane は進捗 index のみ。decisions は各 task file / 必要時に辿る。GitHub Issue 化しない運用。

Access + Vite（TS + React + Pico.css）+ Worker（Hono、list/get）で、contracts に合う fixture または実 Drive から再生できる状態にする。

- [x] web↔worker HTTP 契約 / worker Port・UseCase・Fake / Route・Error 写像
- [x] 実 Google Drive adapter / 静的検査（Biome + tsc）/ runtime config 境界
- [x] web API Client / UI 一覧・再生・原稿 / 層違反検知（dependency-cruiser）
- [x] deploy / Access 方針の A/B（`wrangler.jsonc`・`worker-entry`・`DEPLOY.md`）
- [x] React + Hono の A（dependency・`routes/app.ts`・RPC client 型契約）
- [x] worker Hono route 定義 / entry cutover / web primitive JSX
- [ ] deploy 前実装・設定（下記 C）
- [ ] 初回手動 deploy 以降（下記 D）
- [ ] React + Hono 残作業（下記 E）— `docs/tasks/todo/playback-*.md`

`apps/playback/tsconfig.json` の `lib` は暫定 `["ES2022", "DOM"]`。wrangler runtime 確定後に `@cloudflare/workers-types` を再検討。

### C: deploy 前

方針・契約は `DEPLOY.md` / decisions / `wrangler.jsonc` 済み。残りは実装・dashboard・投入。

1. wrangler toolchain（package・script・`wrangler types`・Vite build → `web/dist`）
2. 同一 origin 配信の実装完成（assets + `/episodes*`）
3. Workers secret 4 key の投入
4. Access Application / Allow（自分 email・session 30d）の dashboard 設定
5. `wrangler deploy --dry-run` と Access Verification（本番 traffic は載せない）

### D: 後回し

1. 初回手動 `wrangler deploy`（本番 traffic）
2. rollback 手順の文書化
3. logging / observability
4. account の `workers.dev` subdomain 実文字列の確認
5. DAST / penetration test（test URL・Access test identity・攻撃対象が未決）
6. CD / hook による自動 deploy（非 scope）
7. Dependabot / Renovate 等（優先度低）

### E: React + Hono 残作業

worker route / entry / primitive JSX は完了（task file 削除済み）。残り依存順：

1. `docs/tasks/todo/playback-web-view-model-react-hooks.md`（worker route 完了前提。`hc<AppType>()` が `unknown` にならないこと）
2. `docs/tasks/todo/playback-web-feature-component-jsx.md`（1 に依存）
3. `docs/tasks/todo/playback-web-page-jsx-mount.md`（2 に依存）

### 依存（実装順）

```text
contracts / worker / UI / 層検知（済）
  → deploy 前 C
      → 初回手動 deploy（D）
```

音声は wav。generator 書込とは共有しない。

### 依存（CI 静的 / 層 / coverage）

```text
PR-A（CI 入口の統一）完了前提
  → worker / web 層検知（済）
  → PR-H（unit coverage gate、済）
```
