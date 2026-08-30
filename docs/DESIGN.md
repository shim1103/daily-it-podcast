# DESIGN.md

最終更新日時: 2026-05-26
現在 phase: 1

> 技術詳細・実装方針(HOW)を定める。外部挙動(HOW-TO-USE)は `SPEC.md` の守備範囲。

---

## 0. 本文書について

- 読み手: `agent`（実装者）。
- `SPEC.md` の外部挙動を実現するための技術詳細を定義する。
- 権威序列: `WORKFLOW.md` ＞ `PROPOSAL.md` ＞ `SPEC.md` ＞ `DESIGN.md`。

---

## 1. 技術スタック

| 領域 | 採用技術 |
|------|---------|
| 言語 | TypeScript（全システム統一） |
| パッケージ管理 | pnpm workspace（monorepo） |
| 再生 UI | Next.js 14 (App Router) / React / TypeScript |
| UI インフラ | Vercel（将来 deploy） |
| podcast 出力システム | Node.js / TypeScript |
| テスト | Vitest（unit + integration） |
| Lint/Format | ESLint + Prettier |
| 型検査 | tsc --noEmit（CI 必須） |
| CI | GitHub Actions |

---

## 2. モノリポ構成

```
daily-it-podcast/
├── packages/
│   ├── core/                  # 共通型定義・interface・エラークラス
│   ├── orchestrator/          # オーケストレーション本体
│   ├── info-fetcher/          # 情報取得システム
│   ├── script-generator/      # テーマ原稿生成システム
│   ├── manuscript-builder/    # 原稿生成（統合）システム
│   ├── tts/                   # TTS システム
│   └── drive/                 # Drive 保存システム
├── apps/
│   └── web/                   # 再生 UI (Next.js)
├── config/
│   └── podcast.config.ts      # 共通設定ファイル（オーケストレーション参照）
├── docs/
│   ├── WORKFLOW.md
│   ├── PROPOSAL.md
│   ├── SPEC.md
│   └── DESIGN.md
├── pnpm-workspace.yaml
├── package.json
└── turbo.json                 # Turborepo（ビルド・テスト並列化）
```

---

## 3. Interface 定義（packages/core）

### 3.1 テーマ情報テキスト

```typescript
// 情報取得システムの出力形式
interface ThemeInfo {
  source: string;        // 情報源識別子
  title: string;         // テーマタイトル
  rawText: string;       // 情報源から取得した生テキスト
  fetchedAt: string;     // ISO8601
}
```

### 3.2 テーマ原稿

```typescript
// テーマ原稿生成システムの出力形式
interface ThemeScript {
  title: string;
  script: string;        // 読み上げ用テキスト
  durationEstimateSec: number;  // 推定尺（秒）
}
```

### 3.3 原稿.json（Manuscript）

```typescript
interface Manuscript {
  timestamp: string;     // ISO8601
  body: {
    opening: string;
    topics: ThemeScript[];
    closing: string;
  };
}
```

### 3.4 各システムの interface

```typescript
// 情報取得システム
interface InfoFetcher {
  fetch(): Promise<ThemeInfo[]>;
}

// テーマ原稿生成システム
interface ScriptGenerator {
  generate(info: ThemeInfo): Promise<ThemeScript>;
}

// TTS システム
interface TtsService {
  synthesize(manuscript: Manuscript): Promise<Buffer>;
}

// Drive 保存システム
interface DriveService {
  save(audioBuffer: Buffer, manuscript: Manuscript): Promise<string>; // returns episode ID
  listEpisodes(): Promise<EpisodeMetadata[]>;
  getEpisode(id: string): Promise<Episode>;
}
```

---

## 4. 設定・依存性注入

### 4.1 共通設定ファイル（config/podcast.config.ts）

```typescript
export const config: PodcastConfig = {
  duration: { min: 5, max: 30, target: 5 },  // 分
  speakerMode: 'single',
  infoSources: {
    manualText: { enabled: true },
    twitter: { enabled: false },
    mastodon: { enabled: false },
    newsFeed: { enabled: false },
  },
  templateKey: 'default',
  apiProvider: {
    scriptGenerator: 'mock',
    tts: 'mock',
  },
  drive: {
    folderId: process.env.DRIVE_FOLDER_ID ?? '',
  },
  cron: { enabled: false },
};
```

### 4.2 DI パターン

- オーケストレーターはコンストラクタ引数で各 interface を受け取る。
- `config.apiProvider` の値をもとに、起動時に具体クラスを選択してインジェクトする。
- テスト時はモッククラスをそのままインジェクト可能。

---

## 5. モック実装方針（MVP）

各システムのモッククラスは `packages/*/src/mock.ts` に配置する。

| システム | モック挙動 |
|---------|----------|
| InfoFetcher | 固定の `ThemeInfo[]` を返す |
| ScriptGenerator | 固定の `ThemeScript` を返す |
| ManuscriptBuilder | 上2者のモック出力から固定 `Manuscript` を組み立てる |
| TtsService | 空の `Buffer` を返す |
| DriveService | save は `"mock-episode-id"` を返す。listEpisodes はモックエピソード配列を返す |
| 再生 UI | `DriveService.listEpisodes()` のモックデータを静的インポートで代替 |

---

## 6. テスト設計（Vitest）

### 6.1 テスト観点（Given-When-Then）

各システム境界に対して以下を検証する。

**情報取得システム**
- Given: 有効な情報源設定 / When: fetch() 実行 / Then: ThemeInfo[] が返る

**テーマ原稿生成システム**
- Given: 有効な ThemeInfo / When: generate() 実行 / Then: ThemeScript が返る

**原稿生成システム**
- Given: InfoFetcher モック + ScriptGenerator モック / When: build() 実行 / Then: Manuscript が返り、timestamp・opening・topics・closing が揃っている

**TTS システム**
- Given: 有効な Manuscript / When: synthesize() 実行 / Then: Buffer が返る

**Drive 保存システム**
- Given: Buffer + Manuscript / When: save() 実行 / Then: episode ID が返る
- Given: 保存済みエピソード / When: listEpisodes() 実行 / Then: EpisodeMetadata[] が返る

**オーケストレーション（end-to-end）**
- Given: 全システムがモック・設定が有効 / When: orchestrate() 実行 / Then: Drive に保存され episode ID が返る

### 6.2 テストファイル配置

```
packages/<name>/src/__tests__/<name>.test.ts
```

---

## 7. CI（GitHub Actions）

```yaml
# .github/workflows/ci.yml
jobs:
  ci:
    steps:
      - pnpm install
      - pnpm typecheck   # tsc --noEmit（全 package）
      - pnpm lint        # ESLint
      - pnpm test        # Vitest（全 package）
      - pnpm build       # next build（apps/web）
```

---

## 8. エラー設計

```typescript
// packages/core/src/errors.ts
export class PodcastError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly cause?: unknown,
  ) {
    super(message);
    this.name = 'PodcastError';
  }
}

export class InfoFetchError extends PodcastError {}
export class ScriptGenerationError extends PodcastError {}
export class TtsError extends PodcastError {}
export class DriveError extends PodcastError {}
```

- 各システムは固有の error class を throw し、上位で握り潰さない。
- オーケストレーターはエラーをログに記録してから再 throw する。

---

## 9. セキュリティ

- secret（`DRIVE_FOLDER_ID` 等）は環境変数に外出しし、`.env.local` は `.gitignore` に含める。
- `DRIVE_FOLDER_ID` は MVP では必須ではないが、設定口は用意する。
- 外部入力（将来の情報取得システム）は各 InfoFetcher 実装内で validation する。
- 依存パッケージの脆弱性スキャンは CI に `pnpm audit` を追加する。

---

## 10. 再生 UI 設計（apps/web）

### 10.1 ページ構成

```
/                    # エピソード一覧（DriveService.listEpisodes() 利用）
/episode/[id]        # エピソード再生（音声プレーヤー + 原稿表示）
```

### 10.2 コンポーネント構成

```
apps/web/src/
├── app/
│   ├── page.tsx              # エピソード一覧画面
│   └── episode/[id]/
│       └── page.tsx          # エピソード再生画面
├── components/
│   ├── EpisodeList.tsx        # エピソード一覧
│   ├── AudioPlayer.tsx        # 音声プレーヤー
│   ├── ManuscriptViewer.tsx   # 原稿表示（タップで意味検索）
│   └── MeaningPopup.tsx       # 意味検索ポップアップ
└── lib/
    └── drive.ts              # DriveService の web 向けラッパー（MVP ではモックを静的インポート）
```

### 10.3 意味検索（MVP）

- テキストの単語をタップ → 選択テキストを取得 → `MeaningPopup` にポップアップ表示。
- MVP ではポップアップ内に「モック: {選択テキスト} の意味」を表示する（実 API なし）。
- 将来実装: 辞書 API または LLM による意味説明に差し替え。
