# AGENTS.md

## 権威序列

`WORKFLOW.md` > `PROPOSAL.md` > `SPEC.md` > `DESIGN.md`

矛盾は上位文書を正とする。`PROPOSAL.md` との矛盾は **phase1 へ差し戻し**（自律解消禁止）。

---

## プロジェクト概要

個人利用の日次 IT ニュース podcast 自動生成システム。現在 **本実装フェーズ**。
モック end-to-end 連結は完成済み。gemini-flash TTS・Google Drive API の本実装を進めている。

---

## モノリポ構成

```
packages/core          # 共通型・interface・エラークラス（変更時は全 package に影響）
packages/orchestrator  # DI で各システムを起動・連結
packages/info-fetcher  # ThemeInfo[] を返す
packages/script-generator  # ThemeScript を返す
packages/manuscript-builder  # Manuscript を組み立てる
packages/tts           # Buffer を返す
packages/drive         # Drive 保存・取得
apps/web               # Next.js 再生 UI（App Router）
config/podcast.config.ts  # 唯一の共通設定ファイル
```

---

## 実装制約

- モックは `packages/*/src/mock.ts`、本実装クラスは同 package の `packages/*/src/<name>.ts` に配置する
- `config.apiProvider` の値で起動時に具体クラスを選択して inject する
- secret は環境変数のみ（`.env.local` は `.gitignore`）
- `master` / `production` branch への push・deploy 禁止
- 独自 DB・独自インフラを持たない（生成物は Google Drive に保存）
- TTS は **gemini-flash** を使う
- API provider（原稿生成・TTS）は差し替え可能な前提で設計する
- 定常運用ループに Claude Code は入れない。無人実行は API で行う

## 本実装に必要な環境変数

| 変数名 | 用途 | 取得元 |
|--------|------|--------|
| `GEMINI_API_KEY` | gemini-flash TTS | Google AI Studio |
| `GOOGLE_CLIENT_ID` | Drive OAuth | Google Cloud Console |
| `GOOGLE_CLIENT_SECRET` | Drive OAuth | Google Cloud Console |
| `GOOGLE_REFRESH_TOKEN` | Drive OAuth | OAuth フロー実行後 |
| `DRIVE_FOLDER_ID` | 保存先フォルダ | Drive のフォルダ URL から取得 |

Drive は OAuth 2.0（個人アカウント向け）で実装する。Service Account ではなくリフレッシュトークン方式。

---

## 再生 UI 操作仕様（apps/web）

| 操作 | 動作 |
|------|------|
| エピソード一覧でタップ | 再生画面へ遷移 |
| 音声プレーヤー | 再生・一時停止・シーク |
| 原稿内の単語をタップ | ポップアップに意味の検索結果を表示（MVP はモック） |

- 認証: 自分の Google アカウント login 前提。public 公開しない
- Drive 上のフォルダを走査してエピソード一覧を取得する構造（MVP はモック）

---

## ブランチ戦略

`feature/*` → PR → `develop` → shim が `master` へ merge

agentが作るPRのbase branchは`develop`。

---

## TDD 規約

1. `test: [機能名] — failing tests` を先に commit（red）
2. `feat: [機能名] — implementation` で green にしてから commit
3. テストと実装を同一 commit に混在させない

---

## コマンド

```bash
pnpm install          # 依存解決
pnpm typecheck        # tsc --noEmit（全 package）
pnpm lint             # ESLint
pnpm test             # Vitest（全 package）
pnpm build            # next build（apps/web）
pnpm run orchestrate  # MVP end-to-end 実行
pnpm run dev          # UI 開発サーバー
```

---

## 文書守備範囲（重複ゼロ原則）

| 文書 | 書くこと |
|------|---------|
| `PROPOSAL.md` | WHY / 受け入れ条件（shimのみ編集） |
| `SPEC.md` | shim向け外部挙動・使い方 |
| `DESIGN.md` | 技術詳細・実装方針 |
| `AGENTS.md` | agentが実装時に必要な最小コンテキスト |

コードを見れば分かることは各文書に書かない。
