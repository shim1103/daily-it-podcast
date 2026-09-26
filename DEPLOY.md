# DEPLOY

最終更新: 2026-09-16

## Scope

本書が持つのは **Playback・Generator の継続運用 SSOT**——deploy 手順、Access・secret・Variable 登録、GHA workflow の一覧と発火条件、rollback。「いつ」「何の値で」「どう動かすか」を扱う。

Non-scope（書かない・写さない）:

- 地図・使い方 → `README.md`
- 層・依存・技術選定・test 配置の規則 → `DESIGN.md`
- Reason / Rejected・再発する判断 → `docs/decisions/`
- 未完了 index → [wiki Issue #192](https://github.com/shim1103/daily-it-podcast/issues/192)
- Worker 境界契約（`name` / `main` / assets / route / `observability` の値） → `apps/playback/wrangler.jsonc` / `worker-entry.ts`（本書は参照のみ、値は写さない）

音声の配置契約は mp3（`contracts/episode-layout.md`）。**現行 runtime の正本はR2**（書込・読取とも）。残る未完了実測は[wiki Issue #192](https://github.com/shim1103/daily-it-podcast/issues/192)を正とする。Google OAuth / Drive のcodebaseは削除済み（`2026-09-16`）。GHA側の`GOOGLE_OAUTH_*` / `TEST_GOOGLE_OAUTH_*` / `DRIVE_FOLDER_ID` / `TEST_DRIVE_FOLDER_ID`登録は未使用のまま残っている（削除は別途）。登録手順の百科はここに書かない。

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

型・検証の正は `PlaybackEnv`（code）。R2 読取は Worker binding（S3 key を Worker に持たない）。binding 契約の正本は `apps/playback/wrangler.jsonc` の `r2_buckets`（方針は Decision `2026-09-14T11-04-30`）。`wrangler.jsonc` は `keep_vars: true`（deploy で未指定 Variable を消さない）。

HTTP 切り分け（応答 body は契約 code のみ。詳細は Worker log）:

| 応答 | 意味 |
|------|------|
| `500` / `configuration_error` | runtime config 不足・不正（`EPISODES` binding 未設定等） |
| `503` / `unavailable` | R2 など外部一時不能 |

### Workers Logs

契約は `wrangler.jsonc` の `observability`（常時 ON）。

1. 永続 log: 上記契約が有効な Version が本番に載っていること（通常の再 deploy で反映）
2. 即時: `cd apps/playback && npx wrangler tail`
3. log に Secret・token・許可 email を出さない

## 4. Generator process env と GHA 登録

process env の正本は `apps/generator/internal/config/names.go`。`GENERATOR_` 接頭は付けない。

| process env | 区分 |
|------|------|
| `CURSOR_API_KEY` | Secret |
| `GEMINI_API_KEY` | Secret |
| `SPARE_GEMINI_API_KEY` | Secret |
| `R2_ACCOUNT_ID` | Variable |
| `R2_BUCKET` | Variable |
| `R2_ACCESS_KEY_ID` | Secret |
| `R2_SECRET_ACCESS_KEY` | Secret |

GitHub Actions（Settings → Secrets and variables → Actions）の `TEST_` 接頭辞は、**値が本番と異なる** credential/variable にのみ付ける（Decision `2026-09-16T00-39-21`）。値が本番と同一の `CURSOR_API_KEY` / `SPARE_GEMINI_API_KEY` / `R2_ACCOUNT_ID` は test workflow も本番登録名をそのまま使う。値が本番と異なるのは `TEST_GEMINI_API_KEY` / `TEST_R2_BUCKET` / `TEST_R2_ACCESS_KEY_ID` / `TEST_R2_SECRET_ACCESS_KEY` のみ。

workflow ごとの credential 一覧・fallback 順序は Decision `2026-09-16T00-39-21` §1-1 の表を正とする。

R2 key は上表（値は書かない）。本番 write/lookup の Composition 結線は R2 固定（Decision `2026-09-14T12-49-26`）。

`GEMINI_API_KEY` は TextWriter/TTS 共通の primary（free）、`CURSOR_API_KEY` は TextWriter の 2 段目 fallback、`SPARE_GEMINI_API_KEY` は TextWriter/TTS 共通の final fallback（paid）。fallback 順序の理由は `GeminiConfig`（`internal/config/config.go`）の invariant と Decision `2026-09-16T00-39-21` を正とする。

credential 付き実 operation は GHA runner のみ。通常 local / Integration gate は実 service を呼ばず local secret を持たない。

## 5. 定時 / gate 外 workflow

| workflow | 入口 script | いつ | 使う登録 |
|------|------|------|------|
| `generator-produce-episode.yml` | `scripts/generator/produce-episode.sh` | 毎日 05:00 JST（cron UTC `0 20 * * *`）+ `workflow_dispatch` | 本番 Secret / Variable |
| `generator-system.yml` | `scripts/generator/test-system.sh` | master への `push` + `workflow_dispatch` | `TEST_*` |
| `generator-tts-rate.yml` | `scripts/generator/test-tts-rate.sh` | `workflow_dispatch` のみ（cron なし） | `TEST_GEMINI_API_KEY` |
| `generator-draft-rate.yml` | `scripts/generator/test-draft-rate.sh` | `workflow_dispatch` のみ（cron なし） | `CURSOR_API_KEY` |
| `playback-smoke.yml` | `npm run test:smoke`（webServer: `npm run dev:smoke`） | master 向け `pull_request` + `workflow_dispatch` | `CLOUDFLARE_API_TOKEN` `TEST_R2_ACCOUNT_ID` |
| `playback-deploy.yml` | `scripts/playback/deploy.sh` | `generator-system.yml` 成功後（`workflow_run`）+ `workflow_dispatch` | `CLOUDFLARE_API_TOKEN` |
| `playback-e2e.yml` | `scripts/playback/test-e2e.sh` | `playback-deploy.yml` 成功後（`workflow_run`）+ `workflow_dispatch` | 下表 `PLAYWRIGHT_*` |

必須 Unit / Integration gate には載せない。`playback-smoke.yml` は master 向け PR で走る。CD 連鎖は `master push → generator-system.yml → playback-deploy.yml → playback-e2e.yml` の順（各段は前段の成功時のみ発火。失敗時は後続を走らせない）。判断: `docs/decisions/2026-09-15T03-51-41` / `2026-09-15T05-36-32` / `2026-09-15T05-04-17` / `2026-09-17T14-30-00`。

`workflow_run` は GitHub 仕様上、**default branch（`master`）に存在する workflow 定義**を見て発火判定する。`generator-system.yml` → `playback-deploy.yml` → `playback-e2e.yml` の連鎖は、3 つの yml すべてが master へ merge されるまで有効化されない（feature branch 上の差分だけでは動かない）。

暦日は JST 運用に合わせる。

`workflow_dispatch` は yml **file 名**が `master` にあれば通る。`gh workflow run <yml> --ref <feature-branch>` で feature branch 側の tree（yml の env・step・script 変更込み）が実行され、master への merge / checkout は要らない。`HTTP 404 workflow not found on the default branch` は file 名自体が master に無い完全新規 workflow のときだけで、その場合は yml を先に master へ載せる。

### Generator System（`generator-system.yml`）

`-tags=system` の system test を **1 回ずつ通すだけ**。「壊れていないか」だけを測り、PASS 率は定常で測らない。1 回でも FAIL なら run が赤。master push 契機の CD 連鎖（§5 冒頭）の起点であり、成功しないと `playback-deploy.yml` が発火しない。判断: `docs/decisions/2026-09-03T14-45-00` / `16-30-00` / `2026-09-17T14-30-00`。

- 実体は `TestProduceEpisodeSystem`（`//go:build system`）1 本。`composition.NewProduceEpisodeFromEnvWithTopicCount`（`SYSTEM_TEST_TOPIC_COUNT` で topic 数を任意指定できる。未指定時は疎通確認用の 1 topic）→ `Run` を 1 度通し、実 5 情報源 → 原稿（Gemini→Cursor→Gemini fallback）→ Gemini TTS（fallback）→ R2 書込 の疎通と通し経路の R2 実到達を見る。Fetch 窓に SourceItem 0 件だった日は `no_source_items` Domain Error で PASS 扱い（fetch は疎通しており system は壊れていない）。他の error は system 故障として赤。
- 必要 credential は config 契約の全 key（`CURSOR_API_KEY` / `TEST_GEMINI_API_KEY` / `SPARE_GEMINI_API_KEY` / `TEST_R2_*`）。1 つでも欠けたら Skip。
- cron の 1 回通しが **2 週連続で落ちたら** bug 扱いで Issue 化する。1 週だけの赤は provider 起因として再 `workflow_dispatch` する。
- 赤になったら故障区間に応じて `generator-tts-rate.yml`（TTS 側）/ `generator-draft-rate.yml`（Cursor 原稿側）を手動 dispatch して切り分ける。CD 連鎖が止まる（`playback-deploy.yml` が発火しない）ので、切り分けを優先する。
- 定時緑化を運用目標にするのは課金枠移行後。無料枠のうちは「dispatch で回せたとき緑」で可。

一次 evidence は GHA run URL（`go test -v` の `t.Logf` と `$GITHUB_STEP_SUMMARY`）。

### Gemini TTS rate 計測（`generator-tts-rate.yml` / `TestGeminiTTSRate`）

`system && ratemeasure` の dispatch 専用。**e2e 1 回通しが TTS 区間で落ちた／不安定なときの事後調査**に使う。cron は持たない（RPM 圧迫回避）。

- 本番 topic 束（`TTS_DOUBLE` = `max` / `tgt` / `min` で尺帯を選ぶ。既定 `max`）を `runs` 回 `SynthesizeAll` し、Adapter が `err == nil` で返る率が `pass_threshold`（既定 0.8）以上なら緑。
- 待機系パラメータの既定は `callGap` 20s / `retryBackoffBase` 60s / `retryBackoffMax` 3m。`generator-tts-rate.yml` の `inputs.default` が SSOT。429 が続くときは dispatch input でこれらを上げて所要の変化を観測する。
- dispatch 例: `gh workflow run generator-tts-rate.yml -f runs=10 -f double=max [-f call_gap= -f retry_backoff_base= -f retry_backoff_max= -f pass_threshold=]`。

env は `TEST_GEMINI_API_KEY` 直読み（本番 `GEMINI_API_KEY` を計測へ流さない）。判断: `docs/decisions/2026-09-03T14-45-00` / `14-46-00`。

### Cursor draft rate 計測（`generator-draft-rate.yml` / `TestCursorAPIDraftRate`）

`system && ratemeasure` の dispatch 専用。**e2e 1 回通しで Cursor 原稿が尺下限割れ／件数外れを再発したときの事後調査**と prompt の A/B に使う。cron は持たない（Cursor draft は 1 回数分）。Cursor は Cloud Agents HTTP API 移行済みで `agent` binary install は不要（判断: `docs/decisions/2026-09-03T17-03-33`）。

- 固定擬似ソース → `ComposeBriefWithTemplate(items, variant)` → `Write` → `ManuscriptDraftFromWriterOutput` を `runs` 回直列に通し、valid Draft が返る率が `pass_threshold`（既定 0.8）以上なら緑。
- `prompt_variant` = `default`（現行 `constants.TextWriterBriefPrompt`）/ `a`（`apps/generator/test/system/testdata/brief_prompt_variant_a.txt`）。下限マージンの薄さが続くなら variant で detail 目安を上げた prompt を検証してから `const` へ反映する。
- `*cursorapi.Error` の `Op=="do"`（API へ到達すらできない環境要因）はその回を分母から除外する。全回が除外なら Skip。
- dispatch 例: `gh workflow run generator-draft-rate.yml -f runs=5 [-f prompt_variant=a -f pass_threshold=]`。

env は `CURSOR_API_KEY` 直読み（値が本番と同一のため`TEST_` 接頭辞を持たない。Decision `2026-09-16T00-39-21`）。判断: `docs/decisions/2026-09-03T14-45-00` / `14-47-00` / `2026-09-16T00-39-21`。

### Playback smoke（`playback-smoke.yml`）

deploy 前（master 向け PR）に「Access 以外の本当の e2e」を1本で確かめる。判断: `docs/decisions/2026-09-15T05-36-32`。

- 実体は `npm run test:smoke`（Playwright）。webServer は `npm run dev:smoke`（`web/vite.smoke.config.ts`）で、origin は `npm run dev` と同じ localhost:3000。
- `web/vite.smoke.config.ts` は `createApp()`（override 無し。本番と同じ composition 経路）を、`env.EPISODES` に実 TEST R2 binding（`test/support/create-remote-test-r2-binding.ts`、`getPlatformProxy` の `remoteBindings: true`、bucket は `daily-it-podcast-dev`）を注入して呼ぶ。`wrangler dev` は使わない（Hono app の `fetch(req, env)` を Vite dev server 上で直接呼ぶだけで足りる）。実 Cloudflare 認証（`CLOUDFLARE_API_TOKEN` / `TEST_R2_ACCOUNT_ID`）が必要。
- TEST bucket は episode 件数を固定 assert しない。「一覧応答が runtime config error にならないこと」を常時確認し、「episode が 1 件以上あるときだけ選択・再生が例外にならないこと」を追加確認する。credential 未登録なら一覧が `configuration_error`（500）になり smoke は赤くなる。

### Playback E2E（`PLAYWRIGHT_*`）

OTP 手動・push 時 storageState。値は repo に書かない。

`playback-smoke.yml` が一覧・原稿・再生・seek の UI 機能を deploy 前に確認するため、E2E は「Access session を経由して本番 R2 へ実際に到達できるか」の 1 test（一覧表示）だけを持つ。原稿・再生・seek の観測は smoke 側にある。

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

安定 fixture（`apps/playback/test/e2e/fixtures/stable-episode/`）は本番 R2 bucket **直下**に置く。配置契約は mp3（`contracts/episode-layout.md`）。日次 produce が増えても fixture pair は残す。

## 6. 再 deploy

```bash
cd apps/playback
npm run build
npx wrangler deploy
```

Variable / Secret の値変更は Dashboard または `wrangler secret put`。code だけの更新は上記で足りる。`observability` 契約の本番反映もこの手順。

## 7. rollback

`wrangler rollback` で全量戻す。command 名・flag は現行 `npx wrangler --help` / 公式で確認する。

1. `cd apps/playback`
2. `npx wrangler versions list` で戻したい Version ID を特定する
3. `npx wrangler rollback <version-id>` で全量戻す（preview / version URL は共有しない）
4. Access 付き本番 hostname で一覧・原稿・再生を人手 smoke する

二次: 既知 good commit を checkout して §6 の再 deploy。

対象外: Variable / Secret / Access（Dashboard または `wrangler secret put` 等の別手順）。

## 8. Playback で採用しないもの

custom domain / Pages+別 Worker / app 内 Access JWT / Service Token・WARP / preview URL 共有 / DAST / Dependabot・Renovate

Reason / Rejected: `docs/decisions/2026-08-25T17-10-00`（公開境界）、`docs/decisions/2026-09-04T02-04-02`（運用後続の完了境界）。

CD（master push 契機の自動 `wrangler deploy`）は `docs/decisions/2026-09-15T05-04-17` で採用へ転換した。自動 deploy は `playback-deploy.yml`（§5 参照。`generator-system.yml` 成功後にのみ発火。Decision `2026-09-17T14-30-00`）。手動手順は §6。
