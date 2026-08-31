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

2026-08-15 以降の **Playback + Generator** が本流。それ以前の mock MVP 文書・実装は削除済み（rewrite 起点: `2631c16`）。

## 技術選定

| 役割 | 選定 |
|------|------|
| 再生 UI | Vite + TypeScript + React + Pico.css（classless） |
| UI の裏側 | Cloudflare Workers（Hono、Drive 代理） |
| UI 入場 | Cloudflare Access（`DEPLOY.md`） |
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

| 知りたいこと | 正本 |
|------|------|
| 層・依存・test 配置・error 3 層 | `DESIGN.md` |
| deploy・Access・GHA 運用・secret 登録 | `DEPLOY.md` |
| Drive のファイル契約 | `contracts/` |
| Playback HTTP 契約 | `apps/playback/contracts/` |
| 未完了 index | `docs/tasks/todo/*-lane.md` |
| 再発する判断 | `docs/decisions/` |
| 全体図 | `docs/diagrams/architecture.mmd` |

## Branch

| branch | 役割 |
|--------|------|
| **`develop`** | **SSOT**（実装・doc・workflow の正本。PR の base） |
| `master` | release（shim が `develop` から merge） |

`feature/*` → PR（base: **`develop`**）→ `master` は shim が release。

GitHub の default branch を `develop` にすること（`workflow_dispatch` と定時 workflow の正本）。

## 使い方

1. **再生:** Access 入場 → 一覧 → 再生・原稿表示。手順は `DEPLOY.md`
2. **生成:** GHA 定時 / 手動。UI からは起動しない。成果物は `contracts/` に従う。運用は `DEPLOY.md`
3. **Playback local:** `cd apps/playback && npm ci && npm run dev`（Node は `.nvmrc`）
4. **generator:** Go は `apps/generator/go.mod`。`golangci-lint` を PATH へ
5. **hook:** `./scripts/install-hooks.sh`
6. **検証入口:** `./scripts/check-static.sh` / `./scripts/test-unit.sh` / `./scripts/test-integration.sh`（詳細・閾値は `DESIGN.md`、credential 付き・E2E 定時は `DEPLOY.md`）

## 受け入れ

- [x] GHA 本番 Secret / Variable 登録済（Generator produce・Playback E2E。`DEPLOY.md`）
- [ ] Access 入場は `DEPLOY.md` の Verification を満たす
- [ ] Generator 成功後、Playback で一覧・再生・原稿表示できる
- [ ] Drive 上の形は `contracts/` に従う

## 制約

- 非公開・独自 DB なし・マルチユーザーなし
- `master` への無断 push 禁止。本番 deploy 方針は `DEPLOY.md`
