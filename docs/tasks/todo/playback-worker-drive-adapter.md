## Playback worker: Drive 読取 Adapter（Google Drive API 本番）

Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

このIssueでは、worker Infrastructure に Google Drive API への読取 Driven Adapter を実装し、既存の `EpisodeRepository` を本番 Drive で満たす。完了後、List / Get JSON / Get 音声が Drive HTTP Stub の Unit で Pass し、HTTP status や Route は知らない。

## 2. Context

- HTTP 契約は `apps/playback/contracts/` に固定済み
- Port / UseCase / in-memory Fake は済。todo `playback-worker-episodes.md` は delete 済み。本 Issue は本番 Adapter のみ
- 音声ファイルの正は `{episodeId}.wav`（`contracts/drive-layout.md`）。mp3 は使わない
- List は JSON 目録のみ。wav の有無は Get で見る（`drive-layout.md`）
- 仮定: 認証は `DESIGN.md` の OAuth refresh。Workers は native secrets（AgentSecrets 単一化は対象外）
- 仮定: 個別 OAuth キー名が README に無い場合は名前行だけ足してよい（値は書かない）

## 3. Canonical Sources

- `contracts/drive-layout.md` — Drive 読取（一覧 `*.json`、1件は JSON + `{episodeId}.wav`）
- `contracts/manuscript.schema.json` — 原稿 JSON（Infrastructure が検証）
- `apps/playback/worker/src/application/ports/episode-repository.ts` — 満たす Port
- `apps/playback/worker/src/infrastructure/drive/in-memory-episode-repository.ts` — Fake。本番は同じ Port を置換
- `apps/playback/contracts/` — Response shape。field 写し禁止。Route 責務は混ぜない
- `docs/decisions/2026-08-17T17-40-00-feature-playback-web.md` — List/Get 分割・error 方針
- `docs/decisions/2026-08-18T11-17-00-feature-tts-speech-synthesizer-wav.md` — 保存形式は WAV
- `docs/decisions/2026-08-18T12-48-00-feature-playback-worker-episodes.md` — 不在は `EpisodeNotFoundError` に畳む
- `DESIGN.md` §1・§2・§3 — worker は Drive 読取 BFF。OAuth refresh。秘密は Infrastructure
- `README.md` — 秘密の名前（`Google OAuth（Drive）` / `DRIVE_FOLDER_ID`）
- `docs/decisions/2026-08-16T00-06-30-docs-agentsecrets-secret-export.md` — GHA / Workers native secrets 寄せは対象外
- `architecture/backend/infrastructure` — Driven Adapter
- `architecture/ports-adapters` — 結線は Composition Root
- test 方針 — `testing-strategy` skill

## 4. Scope

### In Scope

- Google Drive API を使う `EpisodeRepository` の本番実装
- 一覧: 所定フォルダ直下の `*.json`。音声の有無は見ない。schema 不適合・stem 不一致は行に出さない
- 1件: stem 一致の json + wav。JSON 不適合・音声無しは Domain 不在（`EpisodeNotFoundError`）
- 音声 byte（WAV）。成功時の HTTP `Content-Type` は Route 側（本 Adapter は byte のみ）
- OAuth・`DRIVE_FOLDER_ID` は env/secrets（README の名前のみ）。Drive file id を Response / Port 戻り値に載せない
- Composition Root が Fake の代わりに本番 Adapter を結線できること
- Adapter 隣の Unit（Drive HTTP を Stub）

### Out of Scope

- Port / UseCase / in-memory Fake（済）
- HTTP Route / Controller / status 変換（→ `playback-worker-http.md`）
- `apps/playback/web/` 一切
- generator 書込 Adapter（→ `generator-drive-adapter.md`）
- Access / JWT 検証
- List の sort 順（契約未固定。実装で勝手に決めない）
- Range / cache / 署名 URL
- wrangler 本番 deploy
- 実 Google Drive を叩く Integration（本番 credential 禁止）

## 5. Contract

Port の入出力は `EpisodeRepository` に従う。本 Issue はそれを Google Drive API で満たす。

**Drive 読取（`drive-layout.md`）**

- 一覧: `*.json` の stem = `episodeId`。音声は見ない
- 1件: stem 一致の json + wav。不適合・欠落は返さない

**HTTP への非漏出**

- UseCase / Infrastructure は status・External `code` を知らない
- Drive file id・フォルダ id を外へ出さない

## 6. Constraints

- generator と直接依存しない。共有は Drive ファイルと `contracts/` のみ
- Fake Adapter の本番置換。Port 形状を変えて Fake を壊さない
- 秘密の値を code に書かない（README の変数名のみ）
- 読む拡張子は `.wav`
- Workers 上で Google 公式 Node SDK が動かない前提を固定しない。Least Power で足りる手段（REST fetch 等）を選ぶ

## 7. Acceptance Criteria

- [ ] AC-1: Stub 上の適合 JSON が一覧に出る
- [ ] AC-2: Stub 上の schema 不適合 JSON が一覧に出ない
- [ ] AC-3: json のみ（wav 無し）の Get が Domain 不在になる
- [ ] AC-4: Get 成功時、返却原稿が `manuscript.schema.json` に適合する
- [ ] AC-5: Get 音声成功時、wav byte が取れる
- [ ] AC-6: Drive HTTP 失敗は Infrastructure Error になる
- [ ] AC-7: Composition Root が本番 Adapter を Port 実装として結線できる
- [ ] AC-8: Application Unit（Fake）が本 Issue の変更後も Pass する

## 8. Verification

```bash
cd apps/playback && npm run test:unit
```

- worker Infrastructure 隣の `*.test.ts` が Pass
- Drive HTTP は Stub。実 Drive は実行しない
- 本番 credential を読まない

## 9. Dependencies

- 先行: `EpisodeRepository` + Fake（済）
- 並行可: `playback-worker-http.md`（Fake UseCase で完結）
- 共有しない: `generator-drive-adapter.md`（書込。言語・runtime・方向が違う）

## 10. Risks

- Drive の page token を Application に漏らす → 列挙は Adapter 内で閉じ、Port は完成した目録だけ返す
- file id が `audioRef` や JSON に混入する → `audioRef` は HTTP 層の path。Drive id は Adapter 内部だけ
- Fake と本番で「不在」判定がずれる → 判定は Port / `EpisodeNotFoundError` 契約に合わせ、Adapter は mechanism だけ持つ

## 11. Notes

- 次アクション: `create-issue` で Issue 化（`shim gh` は shim 明示まで実行しない）
- 実 Drive 結合は別 follow-up（credential を gate に載せない）
- HTTP `Content-Type` の正は `episodeAudioContentType`（`audio/wav`）。本 Adapter は byte だけ返す
