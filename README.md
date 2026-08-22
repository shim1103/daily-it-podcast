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
Playback (Vite + TypeScript + Cloudflare)
  Cloudflare Access（メール OTP・許可 email = 自分だけ）
  → UI → Workers（Drive 読取の代理）
```

旧実装は `archive/2026-08-15-pre-rewrite` に凍結。本流は Playback + Generator へ作り直し。

## 技術選定

| 役割 | 選定 |
|------|------|
| 再生 UI | Vite + TypeScript + Pico.css（classless） |
| UI の裏側 | Cloudflare Workers（Drive 代理） |
| UI 入場 | Cloudflare Access（メール OTP、自分のみ） |
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

層・依存・test 配置の規則 → `DESIGN.md`  
Drive のファイル契約 → `contracts/`  
Playback HTTP 契約 → `apps/playback/contracts/`

## Branch

| branch | 役割 |
|--------|------|
| `develop` | base |
| `master` | release |

`feature/*` → PR（base: `develop`）→ `master` は shim が release。

## 使い方

1. **再生:** Access（メール OTP）→ 一覧 → 再生・原稿表示（意味検索なし）
2. **生成:** GHA cron / 手動。UI からは起動しない。成果物は `contracts/` に従う
3. **Playback 依存:** `cd apps/playback && npm ci`
4. **Playback 起動:** `cd apps/playback && npm run dev` で `localhost:3000` の再生 UI が立つ
5. **generator Go:** Go `1.26.6`
6. **generator lint:** `golangci-lint` を PATH に入れる（例: `brew install golangci-lint`）
7. **hook 導入:** `./scripts/install-hooks.sh`
8. **static（commit / GHA）:** `./scripts/check-static.sh`（片系: `./scripts/generator/check-static.sh`）
9. **Unit（commit / GHA）:** `./scripts/test-unit.sh`（composer 契約のあと片系: `./scripts/generator/test-unit.sh`、`./scripts/playback/test-unit.sh`）
9. **generator race（GHA）:** `./scripts/generator/test-race.sh`
10. **Integration（push / GHA）:** `./scripts/test-integration.sh`（片系: `./scripts/generator/test-integration.sh`、`./scripts/playback/test-integration.sh`）

test 配置・gate の規則は `DESIGN.md`。

## 受け入れ

- [ ] 自分以外は Access で入れない
- [ ] Generator 成功後、Playback で一覧・再生・原稿表示できる
- [ ] Drive 上の形は `contracts/` に従う

## 秘密（名前のみ）

| 変数 | 用途 | 置き場所 |
|------|------|----------|
| Access 許可 email | UI 入場 | Access ポリシー |
| `GOOGLE_OAUTH_CLIENT_ID` | Drive OAuth | Workers / GHA secrets |
| `GOOGLE_OAUTH_CLIENT_SECRET` | Drive OAuth | 同上 |
| `GOOGLE_OAUTH_REFRESH_TOKEN` | Drive OAuth | 同上 |
| `DRIVE_FOLDER_ID` | 保存先 | 同上 |
| `CURSOR_API_KEY` | 原稿 | GHA secrets |
| `GEMINI_API_KEY` | TTS | GHA secrets |
| `TWITTER_IO_API_KEY` | X 投稿取得（試作 TwitterAPI.io） | AgentSecrets（local）/ GHA secrets |
| `GETX_API_KEY` | X 投稿取得（GetXAPI） | AgentSecrets（local）/ GHA secrets | 

secret名の一覧は運用上のinventoryであり、実行時のSSOTではない。必要なsecretと注入経路の契約は各runtimeが所有する（`docs/decisions/2026-08-19T17-37-00-playback-runtime-secret-boundary.md`）。

local 開発時の値は AgentSecrets（OS keychain + zero-knowledge cloud sync）が保持する。agent は `.env` / `secrets/**` / `~/.ssh/**` を読めない（`.claude/settings.json` 等の deny）。値の登録・確認は `agentsecrets` CLI を shim 自身が実行する。

## 制約

- 非公開・独自 DB なし・マルチユーザーなし
- `master` への無断 push / 本番 deploy 禁止

## 文書

| 文書 | 内容 |
|------|------|
| README.md | 地図・使い方・受け入れ（本ファイル） |
| DESIGN.md | 層・依存・所有・test 方針 |
| contracts/ | Drive に載る wav/json |
| apps/playback/contracts/ | web↔worker HTTP |

dir ごとの README は置かない。層の詳細は skill を正とする（`DESIGN.md` 参照）。
