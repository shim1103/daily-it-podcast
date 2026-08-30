# DEPLOY

最終更新: 2026-08-30

**運用 SSOT**（Playback deploy・Access、Generator GHA・secret 登録）。地図は `README.md`、層規則は `DESIGN.md`。Reason / Rejected は `docs/decisions/`。進捗は `docs/tasks/todo/*-lane.md`。

Worker 境界契約（`name` / `main` / assets / `/episodes*`）の正本は `apps/playback/wrangler.jsonc` と `apps/playback/worker/src/worker-entry.ts`。本書は写さない。

## 1. Playback 公開形

1. 同一 origin（静的 UI + `/episodes*` API）
2. hostname は `*.workers.dev` のみ（custom domain なし）
3. Web production の `baseUrl` は `""`
4. 利用・共有してよい URL は Access 対象の **本番 hostname のみ**

## 2. Access（入場）

| 項目 | 値 |
|------|-----|
| IdP | メール OTP（One-time PIN） |
| 許可 identity | 自分の email **1 件**（値は repo に書かない） |
| 対象 | 本番 hostname **全体** |
| session | **30 日** |
| app 内 JWT 再検証 | しない |

**公開 URL を出す前に** Access Application と Allow を用意する。

## 3. Playback Worker runtime config

型・検証の正は `PlaybackEnv`（code）。Workers への注入区分:

| key | 注入 |
|-----|------|
| `GOOGLE_OAUTH_CLIENT_ID` | Variable |
| `DRIVE_FOLDER_ID` | Variable |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Secret |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Secret |

production では 4 key 必須。Drive config は Worker のみ。`refresh_token` 失効時は再認可 → `wrangler secret put GOOGLE_OAUTH_REFRESH_TOKEN`。

## 4. Generator process env と GHA 登録

process env の正本は `apps/generator/internal/config/names.go`。`GENERATOR_` 接頭は付けない。

| process env | 区分 |
|------|------|
| `GOOGLE_OAUTH_CLIENT_ID` | Variable |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Secret |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Secret |
| `DRIVE_FOLDER_ID` | Variable |
| `CURSOR_API_KEY` | Secret |
| `GEMINI_API_KEY` | Secret |
| `GETX_API_KEY` | Secret |

GitHub Actions（Settings → Secrets and variables → Actions）:

| 用途 | 登録名 |
|------|------|
| 本番 | process env と同名 |
| test（System） | `TEST_` + 同名（例: `TEST_GETX_API_KEY`、`TEST_DRIVE_FOLDER_ID`） |

workflow が test 登録名を process env 名へ写す。Generator は `TEST_` を知らない。判断: `docs/decisions/2026-08-30T12-49-00`。

credential 付き実 operation は GHA runner のみ。通常 local / Integration gate は実 service を呼ばず local secret を持たない。判断: `docs/decisions/2026-08-27T12-17-00`。

## 5. 定時 / gate 外 workflow

| workflow | 入口 script | いつ | 使う登録 |
|------|------|------|------|
| `generator-produce-episode.yml` | `scripts/generator/produce-episode.sh` | 毎日 07:00 JST（cron UTC `0 22 * * *`）+ `workflow_dispatch` | 本番 Secret / Variable |
| `generator-system.yml` | `scripts/generator/test-system.sh` | 月曜 07:00 JST（cron UTC `0 22 * * 0`）+ `workflow_dispatch` | `TEST_*` |
| `playback-e2e.yml` | `scripts/playback/test-e2e.sh` | 月曜 07:00 JST（cron UTC `0 22 * * 0`）+ `workflow_dispatch` | 下表 `PLAYWRIGHT_*` |

必須 Unit / Integration gate には載せない。判断: `docs/decisions/2026-08-30T12-49-01` / `2026-08-30T16-20-00` / `2026-08-30T16-20-03`。

暦日は JST 運用に合わせる（定時を JST 朝に置く）。`ProduceEpisode.Run` 未完の間、本番 produce 定時は失敗しうる。

### Playback E2E 登録（`PLAYWRIGHT_*`）

方針: `docs/decisions/2026-08-30T16-20-03`。値は repo に書かない。

| GHA 登録名 | 区分 | 意味 |
|------|------|------|
| `PLAYWRIGHT_BASE_URL` | Secret | Access 付き **本番** hostname の origin。`https://daily-it-podcast.<subdomain>.workers.dev` 形式（custom domain なし。preview / version URL は使わない） |
| `PLAYWRIGHT_STORAGE_STATE_JSON` | Secret | Playwright `storageState` の **JSON 本文**（許可 email で OTP 入場したあとの cookie 等） |

**`storageState` の取得（人手・初回または session 失効時）**

1. 許可 email で本番 URL に OTP 入場する（§7）。
2. 同じ browser context で Playwright から `storageState` を書き出す（例: 一時 script で `context.storageState()`、または公式の auth setup）。
3. 出力 JSON を Secret `PLAYWRIGHT_STORAGE_STATE_JSON` に登録する。

**GHA → process.env**

GitHub Secrets は **自動では** `process.env` に入らない。workflow の `env:` で明示写像する。`PLAYWRIGHT_STORAGE_STATE_JSON` は step で file へ書き出し、Playwright が読む path を `PLAYWRIGHT_STORAGE_STATE` に渡す（`playwright.config.ts` は path を読む）。placeholder のみの間は Secret 未登録でも入口は緑でよい。

## 6. Playback で採用しないもの

custom domain / Pages+別 Worker / app 内 Access JWT / Service Token・WARP / preview URL 共有 / CD・git hook 自動 deploy（別 scope）

## 7. Playback 完了定義（Verification）

1. `npm run deploy:dry-run` が通る
2. 許可 email で OTP 入場できる
3. 許可以外の email では入場できない
4. 同一 origin で一覧・再生・原稿が表示できる

## 8. Playback 手順の所在

| Phase | 内容 | 正本 |
|-------|------|------|
| 1 | deploy 前ゲート | `docs/tasks/todo/playback-deploy-pre-gate.md` |
| 2 | 初回 `wrangler deploy` + §7 の browser Verification（OTP・一覧・再生） | `docs/tasks/todo/playback-lane.md` |
| 3 | 運用後続（rollback 文書化等） | 同上 |

初回 deploy: `cd apps/playback && npm run build && npx wrangler deploy`
