## Playback 実装レーン

参照: docs/daily/2026-08-15T16-23-06-develop.md  
HTTP 契約の正: `apps/playback/contracts/`  
Drive 読みの正: `contracts/drive-layout.md`

Access + Vite（TS + Pico.css）+ Worker（list/get）で、contracts に合う fixture または実 Drive から再生できる状態にする。

- [x] web↔worker HTTP 契約（List / Get の TS schema・status 級）
- [x] worker Drive Port + List/Get UseCase + Fake/in-memory Infrastructure（`playback-worker-episodes`。AC は Fake で完了）
- [ ] 実 Google Drive adapter（OAuth・folder ID・Drive API 読取）— **未達成 / 未切り出し**
- [ ] web / worker の toolchain（Vite / wrangler 等）を入れる — **未切り出し**（Access 未確定）
- [ ] UI で一覧・再生・原稿表示 — **未切り出し**

HTTP 境界での Domain Error → External `{ code }` 写像、および client 向け表示文（例: エピソードが見つからない）は `playback-worker-http.md` のまま。

### Issue 化待ち（詳細は各 file）

| file | 内容 |
|---|---|
| `playback-worker-episodes.md`（worktree から delete 済） | Drive Port + List/Get UseCase + Fake Infrastructure。**済（Fake）** |
| （未切り出し） | 実 Google Drive adapter（OAuth・folder ID・Drive API 読取） |
| `playback-worker-http.md` | Route / Controller / Domain Error → External `{ code }` 写像、client 向け表示文 |
| `playback-web-api-client.md` | web API Client（Result 型） |

### 依存（実装順）

```text
contracts（済）
  → worker-episodes（済・Fake）
      → worker-http
      → 実 Drive adapter（未切り出し）
  → web-api-client（Stub で worker-http と並行可）
  → UI（未切り出し。api-client 後）
```

toolchain は worker-http と web-api-client の dev 確認用に後から足してよい。残 Issue の AC は Fake/Stub で完結する。実 Drive は別切り出し。
