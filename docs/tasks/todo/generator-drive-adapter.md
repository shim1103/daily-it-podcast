## Generator: Drive 書込 Adapter（Google Drive API 本番）

Issue draft。`create-issue` で正式化する前の一時置き場。

## 1. Summary

このIssueでは、generator の Infrastructure に Google Drive API への書込 Driven Adapter を実装し、所定フォルダへ `{episodeId}.json` と `{episodeId}.mp3` を置ける状態にする。完了後、Application 所有の書込 Port を本番 Adapter が満たし、HTTP Stub の Unit が Pass する。

## 2. Context

- generator に Drive 向け Port・Adapter・秘密キー名は未実装（`PostSource` 系のみ）
- Fake Drive は本 Issue の対象外。playback の in-memory Fake とも共有しない
- 仮定: Application 書込 Port が未存在のため、本 Issue で最小の書込 Port を Application に定義する。原稿生成・TTS UseCase は別
- 仮定: 同一 `{episodeId}` の再書込は上書き。削除 API は扱わない
- 仮定: 認証は `DESIGN.md` の OAuth refresh。個別キー名が README に無い場合は名前行だけ足してよい（値は書かない）

## 3. Canonical Sources

- `contracts/drive-layout.md` — Drive 配置・命名・ペア（書込側が従う形）
- `contracts/manuscript.schema.json` — 原稿 JSON
- `docs/decisions/2026-08-15T16-23-07-develop-drive-identity.md` — `episodeId` と `date` の分離
- `DESIGN.md` §1・§2・§3 — 2系統分離、層、Drive = Google Drive + OAuth refresh
- `README.md` — 秘密の名前（`Google OAuth（Drive）` / `DRIVE_FOLDER_ID`）
- `docs/decisions/2026-08-16T00-06-30-docs-agentsecrets-secret-export.md` — local はキー名参照。GHA native secrets 寄せは対象外
- `architecture/backend/infrastructure` — Driven Adapter
- `architecture/ports-adapters` — Port 所有は Application。結線は Composition Root
- test 方針 — `testing-strategy` skill

## 4. Scope

### In Scope

- Application 所有の書込 Port（名前は実装時。json + mp3 を所定フォルダへ書く操作）
- Google Drive API を使う Driven Adapter（vendor JSON・file id を Port の外へ出さない）
- 書込前に原稿が `manuscript.schema.json` に適合すること。不適合は成功にしない
- stem と JSON 内 `episodeId` の一致。不一致は成功にしない
- local: 秘密は AgentSecrets のキー名参照（値を code / env に載せない）
- Composition Root が本番 Adapter を結線できること
- Adapter 隣の Sociable Unit（Drive HTTP を Stub）

### Out of Scope

- Cursor CLI / Gemini TTS
- 取得（`PostSource`）の変更
- `cmd/generator` 入口と GHA workflow
- Drive 読取・一覧（playback の責務）
- playback との共有 library / 共有 Port
- 不完全ペアの補償削除
- 実 Google Drive を叩く Integration（本番 credential 禁止）
- Access / playback HTTP

## 5. Contract

**Port（Application 所有・概念）**

| 操作 | 入力 | 成功 | 失敗 |
|---|---|---|---|
| 書込 | 非空 `episodeId`、原稿、mp3 byte | 所定フォルダ直下に `{episodeId}.json` と `{episodeId}.mp3` | JSON 不適合 / stem 不一致 → Domain。Drive I/O → Infrastructure |

**Drive 書込（`drive-layout.md`）**

- フォルダ直下。サブフォルダを作らない
- MIME や Drive file id は Adapter 内部。Port も UseCase も file id を扱わない

**HTTP への非漏出**

- playback HTTP schema・status・`code` を知らない

## 6. Constraints

- `playback` を import しない。共有は `contracts/` のファイル契約のみ
- 秘密値を code に書かない。README の変数名のみ
- Infrastructure が Application から import してよいのは Port のみ（`DESIGN.md` §5 depguard）
- 2ファイル書込の途中失敗を「成功」にしない。補償削除は作らない

## 7. Acceptance Criteria

- [ ] AC-1: Stub が Drive 成功を返すとき、json と mp3 が同じ stem でフォルダ直下へ書かれる
- [ ] AC-2: 原稿が `manuscript.schema.json` 不適合なら Drive を呼ばず失敗する
- [ ] AC-3: JSON 内 `episodeId` と stem が不一致なら Drive を呼ばず失敗する
- [ ] AC-4: Drive HTTP 失敗は Infrastructure Error になり、vendor 型が Port 外へ出ない
- [ ] AC-5: 秘密値ではなくキー名だけが Adapter / README から観測できる
- [ ] AC-6: Composition Root が本番 Adapter を Port 実装として結線できる

## 8. Verification

```bash
cd apps/generator && go test ./internal/infrastructure/... ./internal/composition/...
```

- Drive HTTP は `httptest` 等の Stub。実 Drive は実行しない
- `./scripts/test-unit.sh` が Pass（既存 coverage / depguard を壊さない）
- 本番 credential を読まない（`scripts/test-unit.sh` / `test-integration.sh` の invariant）

## 9. Dependencies

- 先行: `contracts/drive-layout.md` / `manuscript.schema.json`（済）
- 並行可: Cursor CLI / Gemini Infrastructure
- 後続: 原稿→TTS→書込を通す UseCase と `cmd` / GHA（本 Adapter を呼ぶ）
- 共有しない: `playback-worker-drive-adapter.md`（読取。言語・runtime・方向が違う）

## 10. Risks

- json 成功・mp3 失敗で不完全ペアが残る → 全体を失敗とし、補償削除は後続に残す
- OAuth キー名が README に無く実装者が値を file に置く → 名前行だけ README に足し、値は AgentSecrets / GHA secrets

## 11. Notes

- playback の Fake / 本番読取 Adapter を generator から再利用しない
- 次アクション: `create-issue` で Issue 化（`shim gh` は shim 明示まで実行しない）
