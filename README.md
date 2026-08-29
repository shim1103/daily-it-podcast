# daily-it-podcast

個人利用の日次 IT ニュース podcast を自動生成し、自分だけが聴けるようにする仕組みです。商用・公開サービスではありません。

## 全体の形

生成と再生は別系統です。つながるのは個人 Google Drive 上のファイルだけです。UI 経由では生成しません。

```text
Generator (Go + GitHub Actions cron)
  取得 → 原稿 (Cursor CLI) → 音声 (Gemini TTS) → 保存
        ↓
  個人 Google Drive（音声 + 原稿）
        ↑
Playback (Vite + TypeScript + React + Cloudflare)
  Access → UI → Workers（Hono、Drive 読取の代理）
```

旧実装は `archive/2026-08-15-pre-rewrite` に凍結。本流は Playback + Generator へ作り直し。

## 技術選定

| 役割 | 選定 |
|------|------|
| 再生 UI | Vite + TypeScript + React + Pico.css（classless） |
| UI の裏側 | Cloudflare Workers（Hono、Drive 代理） |
| UI 入場 | Cloudflare Access（詳細は `DEPLOY.md`） |
| 生成 | Go CLI + GitHub Actions cron |
| 保存 | 個人 Google Drive |
| 原稿 | Cursor CLI（非対話） |
| 音声 | Google Gemini TTS |

## リポジトリ

```text
apps/playback/contracts/ # web↔worker HTTP
apps/playback/web/       # Vite UI
apps/playback/worker/    # BFF
apps/generator/          # Go CLI
contracts/               # Drive 上の表現（SSOT）
.github/workflows/
```

層・依存・所有・test 配置の規則 → `DESIGN.md`  
deploy・Access・公開境界 → `DEPLOY.md`  
Drive のファイル契約 → `contracts/`  
Playback HTTP 契約 → `apps/playback/contracts/`

## Branch

| branch | 役割 |
|--------|------|
| `develop` | base |
| `master` | release |

`feature/*` → PR（base: `develop`）→ `master` は shim が release。

## 使い方

1. **再生:** Access 入場（`DEPLOY.md`）→ 一覧 → 再生・原稿表示（意味検索なし）。deploy 手順・進捗は `DEPLOY.md` と `docs/tasks/todo/playback-lane.md`
2. **生成:** GHA cron / 手動。UI からは起動しない。成果物は `contracts/` に従う
3. **Playback 依存:** `cd apps/playback && npm ci`。Node version は `apps/playback/.nvmrc` を正本にする（`nvm use` 等で合わせる）
4. **Playback 起動（local）:** `cd apps/playback && npm run dev` で `localhost:3000`
5. **Playback deploy 前確認:** `cd apps/playback && npm run deploy:dry-run`（詳細は `docs/tasks/todo/playback-deploy-pre-gate.md`）
6. **generator Go:** version は `apps/generator/go.mod` を正本にする（`go version` で確認）
7. **generator lint:** `golangci-lint` を PATH に入れる（例: `brew install golangci-lint`）
8. **hook 導入:** `./scripts/install-hooks.sh`
9. **static（commit / GHA）:** `./scripts/check-static.sh`（片系: `./scripts/generator/check-static.sh`）
10. **Unit（commit / GHA）:** `./scripts/test-unit.sh`（composer 契約のあと片系: `./scripts/generator/test-unit.sh`、`./scripts/playback/test-unit.sh`）
11. **generator condition coverage（local のみ）:** `./scripts/generator/report-condition-coverage.sh`。`gobco v1.3.4` で generator Unit package（`./cmd/...`、`./internal/...`）の Boolean condition を report する。threshold はなく、hard gate ではない。既存の statement coverage gate はこの report と別に維持する
12. **generator race（GHA）:** `./scripts/generator/test-race.sh`
13. **Integration（push / GHA）:** `./scripts/test-integration.sh`（片系: `./scripts/generator/test-integration.sh`、`./scripts/playback/test-integration.sh`）。generator gate は secret なし Narrow のみ。HTTP vendor の SU/NI latest 化と Composition HTTP 移行の残作業は `docs/tasks/todo/generator-lane.md`
condition coverage report は、構文として認識できる Boolean condition を対象にする。未使用 function は検出できず、`select` も対象外である。したがって完全な branch coverage ではない。

test 配置・gate の規則は `DESIGN.md`。

## 受け入れ

- [ ] Access 入場は `DEPLOY.md` の Verification を満たす
- [ ] Generator 成功後、Playback で一覧・再生・原稿表示できる
- [ ] Drive 上の形は `contracts/` に従う

## Runtime config inventory

| 変数 | 用途 | 区分 | 注入元 |
|------|------|------|--------|
| `GOOGLE_OAUTH_CLIENT_ID` | Drive OAuth | Variable | GitHub Actions Variables / Workers Variables |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Drive OAuth | Secret | GitHub Actions Secrets / Workers Secrets |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Drive OAuth | Secret | GitHub Actions Secrets / Workers Secrets |
| `DRIVE_FOLDER_ID` | 保存先 | Variable | GitHub Actions Variables / Workers Variables |
| `CURSOR_API_KEY` | 原稿 | Secret | GitHub Actions Secrets |
| `GEMINI_API_KEY` | TTS | Secret | GitHub Actions Secrets |
| `GETX_API_KEY` | X 投稿取得（GetXAPI） | Secret | GitHub Actions Secrets |

この表は運用上のinventoryであり、実行時契約のSSOTではない。必要なkey・型・検証はruntimeごとのconfiguration boundaryが所有する。Playback Workerの注入区分は`DEPLOY.md`。

Generatorのproduction情報源はGetXAPIのみとする。

credential付き実operationはGitHub Actions runnerだけで実行する。通常のlocal開発と自動testは実serviceを呼ばず、local secretを持たない。

## 制約

- 非公開・独自 DB なし・マルチユーザーなし
- `master` への無断 push 禁止。本番 deploy 方針は `DEPLOY.md`

## 文書

| 文書 | 内容 |
|------|------|
| README.md | 地図・使い方・受け入れ・runtime config inventory（本ファイル） |
| DESIGN.md | 層・依存・所有・test 方針 |
| DEPLOY.md | deploy・Access・公開境界 |
| contracts/ | Drive に載る wav/json |
| apps/playback/contracts/ | web↔worker HTTP |
| docs/decisions/ | 再発する判断の正（Playback list の concept / 視覚言語は `2026-08-28T19-20-00` / `2026-08-28T19-20-01`。本文は写さない） |

dir ごとの README は置かない。層の詳細は skill を正とする（`DESIGN.md` 参照）。
