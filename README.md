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
Playback (Vite + React + Cloudflare)
  Cloudflare Access（メール OTP・許可 email = 自分だけ）
  → UI → Workers（Drive 読取の代理）
```

旧実装（Next.js monorepo / 意味検索など）は `archive/2026-08-15-pre-rewrite` tag に凍結済みです。本流はここから Playback + Generator 構成へ作り直します。

## 技術選定

| 役割 | 選定 |
|------|------|
| 再生 UI | Vite + React + TypeScript + Tailwind |
| UI の裏側 | Cloudflare Workers（Drive 代理） |
| UI 入場 | Cloudflare Access（メール OTP、自分の email のみ） |
| 生成 | Go CLI + GitHub Actions cron |
| 保存 | 個人 Google Drive |
| 原稿生成 | Cursor CLI（非対話） |
| 音声合成 | Google Gemini TTS |

やめるもの: 意味検索、工程ごとの多 package 分割、Next.js 前提、マルチユーザー／他人の Drive。

## リポジトリ構成（予定）

```text
apps/playback/     # 再生 UI
apps/generator/    # 日次生成 CLI（Go）
contracts/         # Drive 上の原稿など境界契約
```

## Branch

| branch | 役割 |
|--------|------|
| `develop` | 作業の base（本流） |
| `master` | release |

`feature/*` → PR（base: `develop`）→ 必要なら `master` へ（shim が release）。

## 必要な秘密情報（名前のみ）

| 変数 | 用途 | 置き場所の目安 |
|------|------|----------------|
| Cloudflare Access 許可 email | UI 入場 | Access ポリシー |
| Google OAuth（Drive 用 refresh 等） | Drive 読取・書込 | Workers / GHA secrets |
| `DRIVE_FOLDER_ID` | 保存先フォルダ | 同上 |
| `CURSOR_API_KEY` | 原稿生成（Cursor CLI） | GHA secrets |
| `GEMINI_API_KEY` | TTS | GHA secrets |

値の取得手順や画面操作の詳細は仕様書側に書きます。

## 制約

- public に誰でも聴けるようにはしない
- 独自 DB・独自インフラを持たない（生成物は Google Drive）
- マルチユーザー化しない
- `master` への無断 push / 本番 deploy はしない（release は shim）

## 文書

仕様・設計が固まり次第、次を置きます。

| 文書 | 内容 |
|------|------|
| PROPOSAL | WHY / 受け入れ条件 |
| SPEC | 使い方・外部挙動 |
| DESIGN | 実装方針・依存・契約 |
| AGENTS | agent 向け最小コンテキスト |

README は地図と入口だけです。UI の詳細・原稿ルール・podcast 設定の細部はここには書きません。
