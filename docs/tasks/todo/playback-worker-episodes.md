## Playback worker: Drive 読取（List / Get JSON / Get 音声）

Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

worker の Application + Drive Infrastructure で、Drive 上の `{episodeId}.json` / `{episodeId}.mp3` を読み、List 用目録・Get 用原稿・Get 用音声 byte を返す UseCase と Port を実装する。HTTP status や Route は知らない。

## 2. Context

- HTTP 契約（path・Response schema・External `code`）は `apps/playback/contracts/` に固定済み
- List は JSON 目録のみ。mp3 の有無は Get で初めて見る（decision 参照）
- 不完全ペア専用 `code` は作らない。Get で渡せない件は内側で「不在」として扱い、HTTP 層が `episode_not_found` に写す
- wrangler 起動・Route 結線は別 task（`playback-worker-http.md`）

## 3. Canonical Sources

- `apps/playback/contracts/` — web↔worker HTTP 境界（Response shape・path 定数。field 写し禁止）
- `contracts/drive-layout.md` — Drive 読取（一覧 `*.json`、1 件は JSON + 音声）
- `contracts/manuscript.schema.json` — 原稿 JSON（Infrastructure が検証）
- `docs/decisions/2026-08-17T17-40-00-feature-playback-web.md` — List/Get 分割・error 方針
- `DESIGN.md` §2 — worker 層 dir（entities / application / infrastructure / composition）
- test 方針 — `testing-strategy` skill

## 4. Scope

### In Scope

- Application が所有する Episode 読取 Port（名前は実装時に決定。Application 所有を守る）
- `ListEpisodes` UseCase: `*.json` 列挙 → `ListEpisodesResponseSchema` を満たす item だけ返す。schema 不適合 file は行に出さない
- `GetEpisode` UseCase: `{episodeId}.json` + 対応 mp3。JSON 不適合・音声無しは Domain 不在。成功時 `GetEpisodeResponseSchema` 相当 + 音声 byte
- `GetEpisodeAudio`（または同等）: mp3 byte を返す。成功 body は `audio/mpeg`
- Fake / in-memory Drive Adapter で Application Unit test 可能にする
- Drive file id を Response に載せない

### Out of Scope

- HTTP Route / Controller / status 変換（→ `playback-worker-http.md`）
- `apps/playback/web/` 一切
- UI / ViewModel / `<audio>` 配線
- Access / JWT 検証
- List の sort 順（契約未固定。実装で勝手に決めない）
- Range / cache / 署名 URL
- `apps/playback/contracts/` の schema 変更
- Google Drive API 本番 Adapter / OAuth refresh（→ `playback-worker-drive-adapter.md`）
- wrangler.toml・本番 deploy

## 5. Contract

**Port（Application 所有・概念）**

| 操作 | 入力 | 成功 | 失敗（内側） |
|---|---|---|---|
| List | なし | 目録 item 配列（List schema 相当） | Drive I/O → Infrastructure Error |
| Get JSON | `episodeId`（非空） | Get Response 相当（`audioRef` は HTTP 層が付与してよい。UseCase は path 文字列 `episodeAudioPath(id)` を返してよい） | 不在 / JSON 不適合 / 音声無し → Domain 不在。Drive I/O → Infrastructure |
| Get 音声 | `episodeId` | `audio/mpeg` byte | 不在 / mp3 無し → Domain 不在。Drive I/O → Infrastructure |

**Drive 読取（`drive-layout.md`）**

- 一覧: `*.json` の stem = `episodeId`。音声は見ない
- 1 件: stem 一致の json + mp3。不適合・欠落は返さない

**HTTP への非漏出**

- UseCase / Infrastructure は status・External `code` を知らない

## 6. Constraints

- `playback/contracts` を Application / Infrastructure から import してよいのは Response 組み立てに必要な型・schema のみ。Route 層の責務を Application に混ぜない
- generator と直接依存しない。共有は Drive ファイルのみ
- 秘密の値を code に書かない（README の変数名のみ）

## 7. Acceptance Criteria

- [ ] AC-1: Fake Drive で List が schema 不適合 JSON を行に含めない
- [ ] AC-2: Fake Drive で Get が json のみ（mp3 無し）を成功にしない（Domain 不在）
- [ ] AC-3: Fake Drive で Get 成功時、返却原稿が `manuscript.schema.json` に適合する
- [ ] AC-4: Get 音声成功時、byte が取得できる（形式は Infrastructure の責務。HTTP 層は `audio/mpeg` を付与）
- [ ] AC-5: Application Unit test が Port Fake のみで Pass する（Route 不要）

## 8. Verification

- `cd apps/playback && npm run test:unit` — worker Application / Infrastructure 隣の `*.test.ts` が Pass
- Integration（実 Drive）は本 Issue 必須にしない。Fake で AC を満たす

## 9. Dependencies

- 先行: `apps/playback/contracts/`（済）
- 後続: `playback-worker-http.md`（本 task の UseCase を HTTP に載せる）
- 後続: `playback-worker-drive-adapter.md`（同じ Port の Google Drive API 実装）

## 10. Risks

- List item を raw JSON から組む際、top-level だけ読んで schema 不適合を通す risk → Infrastructure で schema 検証し、落ちた file は List から除外

## 11. Notes

- `audioRef` の意味: Get JSON Response の不透明 URI。正準相対 path は `episodeAudioPath(episodeId)`。web は分解しない。byte は別 GET
- 次アクション: `create-issue` で Issue 化（`shim gh` は shim 明示まで実行しない）
