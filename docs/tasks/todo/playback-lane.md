## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`

Access + Vite（TS + Pico.css）+ Worker（list/get）で、contracts に合う fixture または実 Drive から再生できる状態にする。

- [x] web↔worker HTTP 契約（List / Get の TS schema・status 級）
- [x] worker Drive Port + List/Get UseCase + Fake/in-memory Infrastructure（`playback-worker-episodes`。AC は Fake で完了）
- [ ] 実 Google Drive adapter（OAuth・folder ID・Drive API 読取・WAV）— `playback-worker-drive-adapter.md`
- [x] playback 静的検査（Biome + tsc）導入。`pr-c-playback-biome-tsc` で完了
- [ ] web / worker の toolchain（Vite / wrangler 等）を入れる — **未切り出し**（Access 未確定）
- [ ] UI で一覧・再生・原稿表示 — **未切り出し**

HTTP 境界での Domain Error → External `{ code }` 写像、および client 向け表示文は完了済み（`playback-worker-http.md` は完了により削除済み）。

`apps/playback/tsconfig.json` の `lib` は暫定で `["ES2022", "DOM"]` にしている（`worker/src` が `Request`/`Response`/`crypto`/`URL` 等の Web 標準 API 型を要求するため）。wrangler 導入時、`worker` の実行 runtime が Cloudflare Workers に確定したら `@cloudflare/workers-types` への置き換えを再検討する（DOM 固有 API の型が worker 側へ誤って混入する余地を塞ぐため）。

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `playback-worker-episodes.md`（delete 済） | Drive Port + List/Get UseCase + Fake Infrastructure。**済（Fake）** |
| `playback-worker-drive-adapter.md` | Google Drive API 本番読取 Adapter（json + wav） |
| `playback-worker-http.md` | Route / Controller / Domain Error → External `{ code }` 写像、client 向け表示文。**済（完了により削除済み）** |
| `playback-worker-http-refactor.md` | `fetch.ts` 責務分離・複雑性削減（Route抽象化 / error mapping / logging / audio body 作法を分離） |
| `playback-web-api-client.md` | web API Client（Result 型） |

### 依存（実装順）

```text
contracts（済）
  → worker-episodes（済・Fake）
      → worker-http
      → worker-drive-adapter（本番・WAV。http と並行可）
  → web-api-client（Stub で worker-http と並行可）
  → UI（未切り出し。api-client 後）
```

toolchain は worker-http と web-api-client の dev 確認用に後から足してよい。http / api-client の AC は Fake/Stub で完結する。本番 Drive は `playback-worker-drive-adapter.md`。音声ファイルは wav。generator 書込とは共有しない。

### 依存（CI 静的 / 層 / coverage）

```text
PR-A（CI 入口の統一）完了前提
  → PR-E（worker 層検知）: worker-http 実装完了後
  → PR-G（web 層検知）: web-api-client 実装完了後
  → PR-H（unit coverage gate）: worker-http / web-api-client の unit が成立後
```

### Issue 化待ち（今後やるべきこと）

- [ ] PR-E `chore/playback-worker-layer`: worker の **層違反検知**（depcruise 等）を static gate で実行できる状態にする
- [ ] PR-G `chore/playback-web-layer`: web の **層違反検知**（depcruise 等）を static gate で実行できる状態にする
- [ ] PR-H `chore/playback-unit-coverage`: playback の **Unit coverage gate**（Vitest）を導入し、落ちる分岐を最小の unit 追加で埋める
