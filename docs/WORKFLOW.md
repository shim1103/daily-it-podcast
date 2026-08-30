# WORKFLOW.md

最終更新日時: 2026-05-26

> **最上位・不変文書（IMMUTABLE / HIGHEST AUTHORITY）**
> このファイルは project 非依存の汎用 workflow 規約である。
> podcast / domain / api / language / tools に依存しない。
> 任意の委任タスク（script・application・prompt・動画 等）で再利用する。
> agent はこのファイルを編集してはならない。

---

## 0. 読み手と目的

- 読み手: `shim`（人間）と `agent`（委任先 AI）の双方。
- 目的: 委任タスクを 2 phase で進行させ、文書を常に最新・無矛盾に保つための運用規約を定義する。

---

## 1. 文書セット（固定ファイル名）

| ファイル名 | 役割 | 抽象度 | 主編集者 |
|---|---|---|---|
| `WORKFLOW.md` | 最上位・不変の workflow 規約 | ― | （不変） |
| `PROPOSAL.md` | 企画書。目的(WHY)と受け入れ条件(WHAT) | 最も抽象 | `shim` |
| `SPEC.md` | 仕様書。shim 向けの使い方・外部挙動(HOW-TO-USE) | 中間 | `agent` |
| `DESIGN.md` | 設計実装計画書。技術詳細・実装方針(HOW) | 最も具体 | `agent` |
| `QUESTIONS.md` | agent から shim への質問点 | ― | `agent` 記入 / `shim` 回答反映後に削除 |

- 参照は常に固定ファイル名で名指しする。
- WORKFLOW.md 自身は固有名詞・タスク固有語を含まない。

---

## 2. 権威序列（AUTHORITY ORDER）

序列: `WORKFLOW.md` ＞ `PROPOSAL.md` ＞ `SPEC.md` ＞ `DESIGN.md`

- 矛盾が発生した場合、上位文書を正とし、下位文書を更新して矛盾を解消する。
- `SPEC.md` ↔ `DESIGN.md` の矛盾: agent が自律的に解消する。
- `PROPOSAL.md` との矛盾: agent は自律解消してはならない。phase1 へ差し戻す（§5）。

---

## 3. 文書の守備範囲（重複ゼロ原則）

- 各文書は抽象度が異なり、内容を一切重複させない。
  - `PROPOSAL.md`: なぜ作るか + 受け入れ条件（WHY / WHAT）。
  - `SPEC.md`: shim が読めば使い方が分かる外部挙動。技術詳細を書かない。
  - `DESIGN.md`: 技術詳細・実装方針。これを見て実装する。
- 重複を検知したら、守備範囲に基づき正しい文書へ寄せ、他方から除去する。

---

## 4. 文書の鮮度（FRESHNESS）

- すべての変更を直ちに該当文書へ反映する。
- 反映時に矛盾が生じたら §2 の序列で即解消し、全文書を常に最新・無矛盾に保つ。
- shim の回答は chat で行い、agent が該当文書へ反映する。`QUESTIONS.md` は回答が反映された部分のみ削除する。
- `docs/diagrams/architecture.mmd` はアーキテクチャに影響する変更のたびに随時更新する。`SPEC.md`・`DESIGN.md` との整合を常に保つ。

---

## 5. Phase 定義

### Phase 1 — Plan

- 進入条件: shim が phase1 を明示宣言する。
- agent の動作: `WORKFLOW.md` と `PROPOSAL.md` のみを入力とし、shim への途中質問なしで `SPEC.md` `DESIGN.md` `QUESTIONS.md` を生成する。
- agent は `PROPOSAL.md` を編集しない（shim が明示した時のみ編集する）。
- shim の動作: 生成物を review し、必要な準備（api・権限・mcp 等の用意と secret の配置）を行い、`QUESTIONS.md` に chat で回答する。その後 承認 または 修正・再調査依頼 を返す。
- 終了条件: shim が承認するまで反復する。`QUESTIONS.md` に未回答が残る間は承認しない。

### Phase 2 — Implement

- 進入条件: shim が phase2 を明示宣言する。
- agent の動作: `SPEC.md` と `DESIGN.md` から実装を開始し、受け入れ条件を満たすまで shim への途中質問なしで loop する。
  - 成果物（コード）・`SPEC.md`・`DESIGN.md`・`QUESTIONS.md` を更新してよい。
  - `PROPOSAL.md` は不変。
  - 不明点が生じても停止せず、仮定(assumption)を置いて実装を続行し、その仮定を `QUESTIONS.md` に記録する。
  - `PROPOSAL.md` との矛盾が判明した場合のみ、phase2 を中断し phase1 へ差し戻す。
- shim の動作: 成果物を review し、`QUESTIONS.md` に回答し、承認 または 再実装依頼 を返す。
- 終了条件: shim が承認するまで反復する。

---

## 6. Phase 制御

- 現在の phase は shim の明示宣言でのみ決定する。
- 遷移 trigger は shim の承認のみ。
- phase2 中の `PROPOSAL.md` 矛盾は phase1 へ差し戻す唯一の経路である。

---

## 7. 実装方針（project 非依存・基本方針）

### 7.1 バージョン管理

- 本番反映用ブランチへの push を禁止する。
- deploy を禁止する。
- 統合用ブランチを持ち、一定の粒度で commit する。

### 7.2 TDD

- 実装前に plan し、その後 TDD で進める。
- テストは Given-When-Then 記法で観点を明示する。
- カバレッジ数値目標は設けず、テスト観点の網羅を優先する。

### 7.3 設計品質

- カプセル化：内部詳細を隠蔽し、必要な interface のみ公開する。
- 責務分離：static / dynamic な関数を分け、再利用単位を component として切り出す。
- error は専用の error class で表現し、握り潰さない。
- 設計と実装の重複・矛盾を極力なくす。

### 7.4 非機能要件

- 保守性・拡張性・信頼性・移植性を常時意識する。

### 7.5 CI/CD

- 静的検査・整形・型検査・テスト（unit / integration 等）を CI で実行する。
- 型のある言語では typecheck を必須とする。

### 7.6 セキュリティ

- セキュリティに十分配慮する。最低限:
  - secret は環境変数に外出しし、バージョン管理から除外する。
  - 依存パッケージの脆弱性を scan する。
  - 外部入力を validation する。
  - 最小権限の原則を適用する。