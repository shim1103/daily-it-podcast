# DEPLOY

最終更新: 2026-08-29

Playback の **deploy・Access・公開境界**の運用 SSOT。地図は `README.md`、層規則は `DESIGN.md`、進捗は `docs/tasks/todo/playback-lane.md`。判断の Reason / Rejected は `docs/decisions/`。本書に無い運用方針を README / DESIGN へ書かない。

Worker 境界契約（`name` / `main` / assets / `/episodes*` 先回り）の正本は `apps/playback/wrangler.jsonc` と `apps/playback/worker/src/worker-entry.ts`。本書は写さない。

## 1. 公開形

1. 同一 origin（静的 UI + `/episodes*` API）
2. hostname は `*.workers.dev` のみ（custom domain なし）
3. Web production の `baseUrl` は `""`
4. 利用・共有してよい URL は Access 対象の **本番 hostname のみ**（preview / version URL は使わない）

## 2. Access（入場）

| 項目 | 値 |
|------|-----|
| IdP | メール OTP（One-time PIN） |
| 許可 identity | 自分の email **1 件**（値は repo に書かない） |
| 対象 | 本番 hostname **全体** |
| session | **30 日** |
| app 内 JWT 再検証 | しない |

**公開 URL を出す前に** Access Application と Allow を用意する（deploy 前に保存可。OTP 検証は hostname live 後）。

## 3. Runtime config（Playback Worker）

key 名・repo 全体の inventory は `README.md`。型・検証の正は `PlaybackEnv`（code）。Workers への注入区分のみ:

| key | 注入 |
|-----|------|
| `GOOGLE_OAUTH_CLIENT_ID` | Variable |
| `DRIVE_FOLDER_ID` | Variable |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Secret |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Secret |

production では 4 key 必須（in-memory に落とさない）。Drive config は Worker のみが持つ。

**refresh_token** は OAuth 認可フローで事前取得する長命 credential。自動 rotation はしない。失効時は再認可 → `wrangler secret put GOOGLE_OAUTH_REFRESH_TOKEN`。

## 4. 採用しないもの

custom domain / Pages+別 Worker / app 内 Access JWT / Service Token・WARP / preview URL 共有 / CD・git hook 自動 deploy（別 scope）

## 5. 完了定義（Verification）

本番として「deploy ready」と言う条件:

1. `npm run deploy:dry-run` が通る
2. 許可 email で OTP 入場できる
3. 許可以外の email では入場できない
4. 同一 origin で一覧・再生・原稿が表示できる

## 6. 手順の所在（進捗は lane）

| Phase | 内容 | 正本 |
|-------|------|------|
| 1 | deploy 前ゲート（whoami・dry-run・OAuth consent） | `docs/tasks/todo/playback-deploy-pre-gate.md` |
| 2 | 初回 `wrangler deploy` | `docs/tasks/todo/playback-lane.md` |
| 3 | §5 の browser 検証 | 同上 |
| 4 | rollback 文書化・logging 等 | 同上 |

初回 deploy コマンド: `cd apps/playback && npm run build && npx wrangler deploy`
