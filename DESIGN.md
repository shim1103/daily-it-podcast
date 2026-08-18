# DESIGN

最終更新: 2026-08-18（coverage 除外を Composition Root のみへ）

地図・使い方・受け入れ・秘密の名前は `README.md`。Drive に載る表現は `contracts/`。本書は **層・依存・所有・test 配置の規則**だけを書く（パス百科・Drive / HTTP 契約の写しは置かない）。

## 1. システム境界

| 系 | 責務 | 依存してよいもの |
|----|------|------------------|
| `apps/generator` | 取得 → 原稿 → TTS → Drive 書込 | 外部 API / CLI（Infrastructure） |
| `apps/playback/web` | 一覧・再生・原稿表示 | `worker` の HTTP のみ |
| `apps/playback/worker` | Drive 読取 BFF | Drive API（Access 背後） |

禁止: `playback` ↔ `generator` の直接依存。共有は Drive 上のファイルのみ（`contracts/`）。

## 2. 層と依存（Clean Arch / DIP）

用語・import 規則の正は skill。ここでは **このリポでの対応**だけ示す。

- Backend 層: [architecture/backend](file:///Users/shim0729/.claude/skills/architecture/backend/SKILL.md)（Entities / Application / Infrastructure / Composition Root / Route・Controller）
- Port（Repository IF）は **Application が所有**し、Infrastructure が実装する → [ports-adapters](file:///Users/shim0729/.claude/skills/architecture/ports-adapters.md) / [application](file:///Users/shim0729/.claude/skills/architecture/backend/application.md)
- Ring・依存方向: [ring-model](file:///Users/shim0729/.claude/skills/architecture/ring-model.md)
- Frontend 役割: [architecture/frontend](file:///Users/shim0729/.claude/skills/architecture/frontend/SKILL.md)（page / feature / view-model / api-client / utils / lib）

| 置き場 | skill 上の層 |
|--------|----------------|
| `generator/internal/entities` | Entities |
| `generator/internal/application` | Application（UseCase + Port IF） |
| `generator/internal/infrastructure` | Infrastructure |
| `generator/internal/composition` | Composition Root |
| `generator/cmd/generator` | 起動入口 |
| `playback/worker/src/entities` 等 | 上に同じ（BFF） |
| `playback/web/src/{pages,components,api,utils}` | frontend skill |
| `playback/contracts` | web↔worker HTTP 境界共有型（API Client と Route / Controller のみ import） |

`playback/web` は Vite + TypeScript（vanilla）+ Pico.css classless。React / Next.js / shadcn は使わない（`docs/decisions/2026-08-18T11-12-00-feature-playback-web.md`）。

依存は内側へ。Composition Root だけが全層を結線する。

`generator` の Entities は generator に閉じる。UI / agent が共有して読む Domains の正は `contracts/`（言語横断の型モジュールは作らない）。

## 3. 外部 I/O

| 役割 | 接続 |
|------|------|
| 情報取得 | TwitterAPI.io（試作）/ GetXAPI（本運用）。Port は `PostSource`。詳細は `docs/decisions/` の x-api-adoption |
| 原稿 | Cursor CLI |
| TTS | Gemini |
| Drive | Google Drive + OAuth refresh |

ブラウザに Drive の長期秘密を置かない。フォルダ ID・OAuth の値は実行設定（`contracts/` 外）。

## 4. 認証

- UI: Cloudflare Access（メール OTP）。許可 identity は自分のみ（手順の詳細はここに書かない）
- アプリ内マルチテナント OAuth は作らない
- Drive credential は worker / generator の Infrastructure が持つ

## 5. Test 配置

分類・FIRST・二重最小化の正: [testing-strategy](file:///Users/shim0729/.claude/skills/testing-strategy/SKILL.md)  
配置・命名の正: [naming-and-layout](file:///Users/shim0729/.claude/skills/testing-strategy/naming-and-layout.md)  
Scope × Sociability: [levels](file:///Users/shim0729/.claude/skills/testing-strategy/levels.md)

このリポでの適用だけ:

| Scope | 置き場 |
|-------|--------|
| Unit | 対象ソースの隣（Go: `*_test.go`、TS: 隣接 `*.test.ts`） |
| Integration / Contract / System・E2E | `apps/generator/test/` または `apps/playback/test/`（web のブラウザ E2E は `web` 配下でも可） |

1. 分類は **file 名**に出す（`narrow_integration` / `contract` / `system_e2e` 等）。dir 名だけに頼らない
2. `integration` 一語で複数分類を兼ねない
3. 実境界に届かないものに `e2e` と付けない
4. Unit を共有 `tests/` に集めない
5. Unit は commit と GHA。Integration は push と GHA。手順の正は `scripts/`。hook と GHA は同じ入口を呼ぶ caller であり、command を YAML や hook へ写さない
6. 片系は `scripts/generator/` と `scripts/playback/` から単独実行できる。root の `check-static.sh` / `test-integration.sh` は片系を呼ぶだけ。root の `test-unit.sh` は composer 契約を実行してから片系 unit を呼ぶ
7. runner は Playback = Vitest（`apps/playback`）、Generator = `go test`
8. generator Unit gate は **statement coverage 90%**（`covermode=atomic`）。除外は Composition Root のみ。`error.go` / `names.go` / `constants.go` を名前では除外しない。Integration に Unit 閾値を載せない
9. generator static は **depguard**（`golangci-lint`、`strict` allow。他 linter は enable しない）で §2 の層 import を block する。Infrastructure が Application から import してよいのは **Port** のみ。playback 側の同等 gate は置かない

実行手順（hook 導入・コマンド）は `README.md`。

## 6. 文書

| 文書 | 書くこと |
|------|----------|
| README | 地図・使い方・受け入れ・秘密の名前 |
| DESIGN | 本ファイル（規則のみ） |
| contracts/ | Drive 配置・原稿 JSON |

dir 単位の README は置かない。SPEC / PROPOSAL は置かない。
