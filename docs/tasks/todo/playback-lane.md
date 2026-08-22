## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`

Access + Vite（TS + Pico.css）+ Worker（list/get）で、contracts に合う fixture または実 Drive から再生できる状態にする。

- [x] web↔worker HTTP 契約（List / Get の TS schema・status 級）
- [x] worker Drive Port + List/Get UseCase + Fake/in-memory Infrastructure（`playback-worker-episodes`。AC は Fake で完了）
- [x] worker Route / Controller / Domain Error → External `{ code }` 写像（`playback-worker-http`）
- [x] `fetch.ts` 責務分離・複雑性削減（`playback-worker-http-refactor`）
- [x] 実 Google Drive adapter（`GoogleDriveEpisodeRepository`。OAuth・folder ID・Drive API 読取・WAV）
- [x] playback 静的検査（Biome + tsc）導入。`pr-c-playback-biome-tsc` で完了
- [x] worker runtime config 境界と `configuration_error` の HTTP contract（PR #36 / PR #38）
- [ ] web API Client の応答処理 — `playback-web-api-client.md`
- [ ] web / worker の toolchain（Vite / wrangler 等）を入れる — **未切り出し**（Access 未確定）
- [ ] UI で一覧・再生・原稿表示 — **未切り出し**

`apps/playback/tsconfig.json` の `lib` は暫定で `["ES2022", "DOM"]` にしている（`worker/src` が `Request`/`Response`/`crypto`/`URL` 等の Web 標準 API 型を要求するため）。wrangler 導入時、`worker` の実行 runtime が Cloudflare Workers に確定したら `@cloudflare/workers-types` への置き換えを再検討する（DOM 固有 API の型が worker 側へ誤って混入する余地を塞ぐため）。

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `playback-web-api-client.md` | API Client の応答処理（status 分類 → parse → schema 検証 → `ApiResult`） |

### 未確定仕様

- [ ] DAST / penetration test は、test deployment URL、Cloudflare Access を通る test identity、許可された攻撃対象が未決定。これらを決定するまで Issue 化しない

### 依存（実装順）

```text
contracts / worker（済）
  → web-api-client ┐
  → toolchain      ┴→ UI
                        → wrangler / deploy（Access 確定後。未切り出し）
```

web-api-client と toolchain は互いに依存せず並行できる。web-api-client の AC は Stub `fetch`、UI の AC は Stub API Client で完結する。wrangler と deploy だけが Access 確定待ち。音声ファイルは wav。generator 書込とは共有しない。

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
