## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`  
deploy・Access の正: `DEPLOY.md`（decision: `docs/decisions/2026-08-25T16-57-00-feature-playback-worker-deploy.md` / `2026-08-25T17-10-00-feature-playback-worker-deploy.md`）

Access + Vite（TS + Pico.css）+ Worker（list/get）で、contracts に合う fixture または実 Drive から再生できる状態にする。

- [x] web↔worker HTTP 契約（List / Get の TS schema・status 級）
- [x] worker Drive Port + List/Get UseCase + Fake/in-memory Infrastructure（`playback-worker-episodes`。AC は Fake で完了）
- [x] worker Route / Controller / Domain Error → External `{ code }` 写像（`playback-worker-http`）
- [x] `fetch.ts` 責務分離・複雑性削減（`playback-worker-http-refactor`）
- [x] 実 Google Drive adapter（`GoogleDriveEpisodeRepository`。OAuth・folder ID・Drive API 読取・WAV）
- [x] playback 静的検査（Biome + tsc）導入。`pr-c-playback-biome-tsc` で完了
- [x] worker runtime config 境界と `configuration_error` の HTTP contract（PR #36 / PR #38）
- [x] UI で一覧・再生・原稿表示（一覧 page 1 つに統合。component 構成・audio 取得方式・URL 同期は `docs/decisions/2026-08-25T05-10-48-feature-playback-ui-structure.md`）
- [x] deploy / Access 方針の A/B（`wrangler.jsonc`・`worker-entry`・`DEPLOY.md`・decision 2本）
- [ ] web API Client の応答処理 — `playback-web-api-client.md`（file 欠落なら再切り出し）
- [ ] deploy 前実装・設定（下記 C）— **Issue 未作成**
- [ ] 初回手動 deploy 以降（下記 D）— **Issue 化しない／後で決める**

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

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `playback-web-api-client.md` | API Client の応答処理（status 分類 → parse → schema 検証 → `ApiResult`） |

### 依存（実装順）

```text
contracts / worker / UI（済）
  → web-api-client
  → deploy 前 C（toolchain・secret・Access dashboard・dry-run）
      → 初回手動 deploy（D）
```

web-api-client の AC は Stub `fetch` で完結できる。音声は wav。generator 書込とは共有しない。

web の role ↔ dir 対応と Pico.css の導入方式は `docs/decisions/2026-08-20T19-29-21-playback-web-layer-layout.md`。

### 依存（CI 静的 / 層 / coverage）

```text
PR-A（CI 入口の統一）完了前提
  → PR-E（worker 層検知）: 前提充足済み
  → PR-G（web 層検知）: web-api-client 実装完了後
  → PR-H（unit coverage gate）: web-api-client の unit が成立後
```

### Issue 化待ち（今後やるべきこと）

- [ ] PR-E `chore/playback-worker-layer`: worker の **層違反検知**（depcruise 等）を static gate で実行できる状態にする
- [ ] PR-G `chore/playback-web-layer`: web の **層違反検知**（depcruise 等）を static gate で実行できる状態にする
- [ ] PR-H `chore/playback-unit-coverage`: playback の **Unit coverage gate**（Vitest）を導入し、落ちる分岐を最小の unit 追加で埋める
