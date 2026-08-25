# DEPLOY

最終更新: 2026-08-25

Playback の **deploy・Access・公開境界**の latest SSOT。地図は `README.md`、層規則は `DESIGN.md`。判断の Reason / Rejected は `docs/decisions/`（記録）。本書に無い運用方針を README / DESIGN へ書かない。

境界契約（Worker `name` / `main` / assets / `/episodes*` 先回り）の正本は次とする。本書は写さない。

1. `apps/playback/wrangler.jsonc`
2. `apps/playback/worker/src/worker-entry.ts`

関連 decision:

1. 文書分業・本書を SSOT にする — `docs/decisions/2026-08-25T16-57-00-feature-playback-worker-deploy.md`
2. 同一 origin・Access・secret — `docs/decisions/2026-08-25T17-10-00-feature-playback-worker-deploy.md`

## 1. 公開形

1. 同一 origin（静的 UI + `/episodes*` API）
2. hostname は `*.workers.dev`（account subdomain は Cloudflare 側の既存値）
3. custom domain は使わない
4. Web production の `baseUrl` は `""`（相対 path）
5. 利用・共有してよい URL は Access 対象の **本番 hostname のみ**。preview / version URL は使わない・公開しない

## 2. Access（入場）

| 項目 | 値 |
|------|-----|
| IdP | Cloudflare Access **メール OTP** |
| 許可 identity | **自分の email 1 つだけ**（値は Access ポリシー側。repo に書かない） |
| それ以外 | Deny（Access default-deny） |
| 対象 | 本番 hostname **全体** |
| session | **30 日** |
| app 内 JWT 再検証 | しない |
| Service Token / WARP / 追加 IdP | しない |

公開 URL を出す前に Access Application と Allow ポリシーを用意する。

## 3. Secret（Playback Worker）

実行時 key の正は `PlaybackEnv`（`apps/playback/worker/src/composition/runtime-config.ts`）。inventory 名一覧は `README.md`。

| key | 注入 |
|-----|------|
| `GOOGLE_OAUTH_CLIENT_ID` | Workers **secret** |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Workers **secret** |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Workers **secret** |
| `DRIVE_FOLDER_ID` | Workers **secret** |

production で in-memory repository mode は使わない（4 key 必須）。Drive secret は Worker のみ。Web は持たない。

## 4. 採用しないもの

1. custom domain
2. Pages UI + 別 Worker API（別 origin）
3. app 内 Access JWT 検証
4. Service Token / WARP / 追加 IdP
5. preview / version URL の利用・共有
6. CD / git hook による自動 deploy（別 scope）
7. README / DESIGN への本方針の再掲
8. A 契約値（`wrangler.jsonc` の字段）の本書への再掲

## 5. Verification（本番 traffic 前の確認観点）

1. `wrangler deploy --dry-run` が通る
2. 許可 email で OTP 入場できる
3. 許可以外の email では入場できない
4. 同一 origin で一覧・再生・原稿が表示できる（実装完了後）

本番へ traffic を載せる `wrangler deploy` 実行そのもの、rollback 手順の詳細、logging / observability の詳細は後続 scope（初回手動 deploy 以降）とする。
