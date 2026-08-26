## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`  
deploy・Access の正: `DEPLOY.md`（decision: `docs/decisions/2026-08-25T16-57-00-feature-playback-worker-deploy.md` / `2026-08-25T17-10-00-feature-playback-worker-deploy.md`）  
React + Hono + Hono RPC 導入の正: `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md`。toolchain version 固定は `docs/decisions/2026-08-26T01-00-00-architecture-reconsider-react-hono.md`

Access + Vite（TS + React + Pico.css）+ Worker（Hono、list/get）で、contracts に合う fixture または実 Drive から再生できる状態にする。

- [x] web↔worker HTTP 契約（List / Get の TS schema・status 級）
- [x] worker Drive Port + List/Get UseCase + Fake/in-memory Infrastructure（`playback-worker-episodes`。AC は Fake で完了）
- [x] worker Route / Controller / Domain Error → External `{ code }` 写像（`playback-worker-http`）
- [x] `fetch.ts` 責務分離・複雑性削減（`playback-worker-http-refactor`）
- [x] 実 Google Drive adapter（`GoogleDriveEpisodeRepository`。OAuth・folder ID・Drive API 読取・WAV）
- [x] playback 静的検査（Biome + tsc）導入。`pr-c-playback-biome-tsc` で完了
- [x] worker runtime config 境界と `configuration_error` の HTTP contract（PR #36 / PR #38）
- [x] web API Client の応答処理（`docs/decisions/2026-08-20T13-44-08-playback-web-api-client.md`）
- [x] UI で一覧・再生・原稿表示（一覧 page 1 つに統合。component 構成・audio 取得方式・URL 同期は `docs/decisions/2026-08-25T05-10-48-feature-playback-ui-structure.md`）
- [x] web / worker の層違反検知（`dependency-cruiser`）を static gate で実行。Feature/Primitive dir 分割と Drive 原稿検証の HTTP 切断を含む（`docs/decisions/2026-08-25T18-42-00-chore-playback-worker-web-layer.md`）
- [x] deploy / Access 方針の A/B（`wrangler.jsonc`・`worker-entry`・`DEPLOY.md`・decision 2本）
- [ ] deploy 前実装・設定（下記 C）— **Issue 未作成**
- [ ] 初回手動 deploy 以降（下記 D）— **Issue 化しない／後で決める**
- [ ] React + Hono + Hono RPC 導入（下記 E）— Issue 6本作成済み

`apps/playback/tsconfig.json` の `lib` は暫定で `["ES2022", "DOM"]`。wrangler runtime 確定後に `@cloudflare/workers-types` への置き換えを再検討する。

### C: やっていないこと（deploy 前・Issue 未作成）

方針・契約は `DEPLOY.md` / decisions / `wrangler.jsonc` 済み。残りは実装・dashboard・投入。

1. wrangler toolchain（package・script・`wrangler types`・Vite build → `web/dist`）
2. 同一 origin 配信の実装完成（assets + `/episodes*`）
3. Workers secret 4 key の投入
4. Access Application / Allow（自分 email・session 30d）の dashboard 設定
5. `wrangler deploy --dry-run` と Access Verification（本番 traffic は載せない）

### D: これから決める／後回し

1. 初回手動 `wrangler deploy`（本番 traffic）
2. rollback 手順の文書化
3. logging / observability
4. account の `workers.dev` subdomain 実文字列の確認
5. DAST / penetration test（test URL・Access test identity・攻撃対象が未決）
6. CD / hook による自動 deploy（非 scope）
7. Dependabot / Renovate 等の dependency 自動更新 PR 機構（後で決めるかもしれない。個人開発の現規模では優先度低）

### E: React + Hono + Hono RPC 導入（Issue 6本作成済み）

方針は `docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md`。A 区分（dependency 追加・`routes/app.ts`・`api/playback-rpc-client.ts` の型契約固定）は完了済み。残りは以下6 Issue（C 区分）。

worker 系（直列）：

1. `docs/tasks/todo/playback-worker-hono-route-definition.md`
2. `docs/tasks/todo/playback-worker-hono-entry-cutover.md`（1 に依存）

web 系（依存順）：

3. `docs/tasks/todo/playback-web-view-model-react-hooks.md`
4. `docs/tasks/todo/playback-web-primitive-component-jsx.md`（3 と並行可）
5. `docs/tasks/todo/playback-web-feature-component-jsx.md`（3・4 に依存）
6. `docs/tasks/todo/playback-web-page-jsx-mount.md`（5 に依存）

worker 系と web 系は互いに独立して進行できる（`PlaybackApiClient` interface が不変のため）。

### 依存（実装順）

```text
contracts / worker / UI / 層検知（済）
  → deploy 前 C（toolchain・secret・Access dashboard・dry-run）
      → 初回手動 deploy（D）
```

音声ファイルは wav。generator 書込とは共有しない。

web の role ↔ dir 対応は `docs/decisions/2026-08-25T18-42-00-chore-playback-worker-web-layer.md`（Feature/Primitive 分割）。Pico.css 導入は `docs/decisions/2026-08-20T19-29-21-playback-web-layer-layout.md`。

### 依存（CI 静的 / 層 / coverage）

```text
PR-A（CI 入口の統一）完了前提
  → worker / web 層検知（済。`apps/playback/.dependency-cruiser.mjs` + `scripts/playback/check-static.sh`）
  → PR-H（unit coverage gate、済）
```

- [x] PR-H `chore/playback-unit-coverage`: playback の **Unit coverage gate**（Vitest branch coverage、`@vitest/coverage-v8`）を導入し、落ちる分岐を最小の unit 追加で埋めた。gate 定義は `DESIGN.md` §5（Test 配置）9 項目目を正とする
