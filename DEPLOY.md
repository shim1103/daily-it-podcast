# DEPLOY

最終更新: 2026-09-04

**運用 SSOT**（Playback・Generator の継続運用）。地図は `README.md`、層規則は `DESIGN.md`。Reason / Rejected は `docs/decisions/`。進捗は `docs/tasks/todo/*-lane.md`。

Worker 境界契約（`name` / `main` / assets / `/episodes*` / `observability`）の正本は `apps/playback/wrangler.jsonc` と `apps/playback/worker/src/worker-entry.ts`。本書は写さない。

## 1. Playback 公開形

1. 同一 origin（静的 UI + `/episodes*` API）
2. hostname は `*.workers.dev` のみ（custom domain なし）
3. Web production の `baseUrl` は `""`
4. 利用・共有してよい URL は Access 対象の **本番 hostname のみ**（preview / version URL は共有しない）

## 2. Access（入場）

| 項目 | 値 |
|------|-----|
| IdP | メール OTP（One-time PIN） |
| 許可 identity | 自分の email **1 件**（値は repo に書かない） |
| 対象 | 本番 hostname **全体** |
| session | **30 日** |
| app 内 JWT 再検証 | しない |

許可 / 拒否の証明は人手。Service Token や WARP で Access を迂回しない。

## 3. Playback Worker runtime config

型・検証の正は `PlaybackEnv`（code）。Workers への注入区分:

| key | 注入 |
|-----|------|
| `GOOGLE_OAUTH_CLIENT_ID` | Variable（`*.apps.googleusercontent.com`。`GOCSPX-` は Client Secret 側） |
| `DRIVE_FOLDER_ID` | Variable |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Secret |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Secret |

production では 4 key 必須。Drive config は Worker のみ。`wrangler.jsonc` は `keep_vars: true`（deploy で未指定 Variable を消さない）。

`refresh_token` 失効時: 再認可 → `cd apps/playback && npx wrangler secret put GOOGLE_OAUTH_REFRESH_TOKEN`。

HTTP 切り分け（応答 body は契約 code のみ。詳細は Worker log）:

| 応答 | 意味 |
|------|------|
| `500` / `configuration_error` | runtime config 不足・不正 |
| `503` / `unavailable` | Drive / OAuth token など外部一時不能（CLIENT_ID 取り違え含む） |

### Workers Logs

方針: `docs/decisions/2026-09-04T02-04-01`。契約は `wrangler.jsonc` の `observability`。

1. 永続 log: 上記契約が有効な Version が本番に載っていること（通常の再 deploy で反映）
2. 即時: `cd apps/playback && npx wrangler tail`
3. log に Secret・token・許可 email を出さない

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

GitHub Actions（Settings → Secrets and variables → Actions）:

| 用途 | 登録名 |
|------|------|
| 本番 | process env と同名 |
| test（System） | `TEST_` + 同名 |

workflow が test 登録名を process env 名へ写す。Generator は `TEST_` を知らない。判断: `docs/decisions/2026-08-30T12-49-00`。

credential 付き実 operation は GHA runner のみ。通常 local / Integration gate は実 service を呼ばず local secret を持たない。判断: `docs/decisions/2026-08-27T12-17-00`。

## 5. 定時 / gate 外 workflow

| workflow | 入口 script | いつ | 使う登録 |
|------|------|------|------|
| `generator-produce-episode.yml` | `scripts/generator/produce-episode.sh` | 毎日 07:00 JST（cron UTC `0 22 * * *`）+ `workflow_dispatch` | 本番 Secret / Variable |
| `generator-system.yml` | `scripts/generator/test-system.sh` | 月曜 07:00 JST（cron UTC `0 22 * * 0`）+ `workflow_dispatch` | `TEST_*` |
| `generator-tts-rate.yml` | `scripts/generator/test-tts-rate.sh` | `workflow_dispatch` のみ（cron なし） | `TEST_GEMINI_API_KEY` |
| `playback-e2e.yml` | `scripts/playback/test-e2e.sh` | 月曜 07:00 JST（cron UTC `0 22 * * 0`）+ `workflow_dispatch` | 下表 `PLAYWRIGHT_*` |

必須 Unit / Integration gate には載せない。判断: `docs/decisions/2026-08-30T12-49-01` / `2026-08-30T16-20-00` / `2026-08-30T16-20-03`。

暦日は JST 運用に合わせる。

workflow file を Actions で `workflow_dispatch` するには、**default branch（`develop`）にその yml があること**が必要。

### Generator System（`generator-system.yml`）

`-tags=system` の system test を **1 回ずつ通すだけ**。「壊れていないか」だけを測り、PASS 率は定常で測らない。1 回でも FAIL なら run が赤。判断: `docs/decisions/2026-09-03T14-45-00` / `16-30-00`。

- cron の 1 回通しが **2 週連続で同じ test を落としたら** bug 扱いで Issue 化する。1 週だけの赤は provider 起因として再 `workflow_dispatch` する。
- 赤になったら `generator-tts-rate.yml` を手動 dispatch して原因（provider の一時劣化か・retry / callGap の詰めか）を切り分ける。
- 定時緑化を運用目標にするのは課金枠移行後。無料枠のうちは「dispatch で回せたとき緑」で可。

一次 evidence は GHA run URL（`go test -v` の `t.Logf` と `$GITHUB_STEP_SUMMARY`）。

### Gemini TTS rate 計測（`generator-tts-rate.yml` / `TestGeminiTTSRate`）

`system && ratemeasure` の dispatch 専用。**System の 1 回通しが落ちた／不安定なときの事後調査**に使う。cron は持たない（RPM 圧迫回避）。

- 本番 topic 束（`TTS_DOUBLE` = `max` / `tgt` / `min` で尺帯を選ぶ。既定 `max`）を `runs` 回 `SynthesizeAll` し、Adapter が `err == nil` で返る率が `pass_threshold`（既定 0.8）以上なら緑。
- 待機系パラメータの既定は `callGap` 20s / `retryBackoffBase` 60s / `retryBackoffMax` 3m。`generator-tts-rate.yml` の `inputs.default` が SSOT。429 が続くときは dispatch input でこれらを上げて所要の変化を観測する。
- dispatch 例: `gh workflow run generator-tts-rate.yml -f runs=10 -f double=max [-f call_gap= -f retry_backoff_base= -f retry_backoff_max= -f pass_threshold=]`。

env は `TEST_GEMINI_API_KEY` 直読み（本番 `GEMINI_API_KEY` を計測へ流さない）。判断: `docs/decisions/2026-09-03T14-45-00` / `14-46-00`。

### Playback E2E（`PLAYWRIGHT_*`）

方針: `docs/decisions/2026-08-30T16-20-03`。値は repo に書かない。

| GHA 登録名 | 区分 | 意味 |
|------|------|------|
| `PLAYWRIGHT_BASE_URL` | Secret | Access 付き本番 hostname の origin |
| `PLAYWRIGHT_STORAGE_STATE_JSON` | Secret | Playwright `storageState` の JSON 本文 |

local 実行時の path env 名は `PLAYWRIGHT_STORAGE_STATE`（GHA には登録しない）。workflow が JSON Secret → file → その path に写す。

**`storageState` 更新（初回・session 失効時）**

1. 許可 email で本番 URL に OTP 入場する
2. Playwright で `storageState` を書き出す（例: headed browser で `CF_Authorization` 付与後に保存）
3. JSON 本文を Secret `PLAYWRIGHT_STORAGE_STATE_JSON` に登録する

手動確認: `gh workflow run playback-e2e.yml --ref <branch>`（Secret 付き）。

安定 fixture（`apps/playback/test/e2e/fixtures/stable-episode/`）は本番 `DRIVE_FOLDER_ID` **直下**に json+wav を置く。日次 produce が増えても残す。

## 6. 再 deploy

```bash
cd apps/playback
npm run build
npx wrangler deploy
```

Variable / Secret の値変更は Dashboard または `wrangler secret put`。code だけの更新は上記で足りる。`observability` 契約の本番反映もこの手順。

## 7. rollback

方針: `docs/decisions/2026-09-04T02-04-00`。command 名・flag は現行 `npx wrangler --help` / 公式で確認する。

1. `cd apps/playback`
2. `npx wrangler versions list` で戻したい Version ID を特定する
3. `npx wrangler rollback <version-id>` で全量戻す（preview / version URL は共有しない）
4. Access 付き本番 hostname で一覧・原稿・再生を人手 smoke する

二次: 既知 good commit を checkout して §6 の再 deploy。

対象外: Variable / Secret / Access（Dashboard または `wrangler secret put` 等の別手順）。

## 8. Playback で採用しないもの

custom domain / Pages+別 Worker / app 内 Access JWT / Service Token・WARP / preview URL 共有 / CD・git hook 自動 deploy / DAST / Dependabot・Renovate

Reason / Rejected: `docs/decisions/2026-08-25T17-10-00`（公開境界）、`docs/decisions/2026-09-04T02-04-02`（運用後続の完了境界）。
