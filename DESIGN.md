# DESIGN

最終更新: 2026-08-25（playback Feature/Primitive dir 分割と dependency-cruiser 層 gate）

地図・使い方・受け入れ・秘密の名前は `README.md`。Drive に載る表現は `contracts/`。本書は **層・依存・所有・test 配置の規則**だけを書く（パス百科・Drive / HTTP 契約の写しは置かない）。

## 1. システム境界

| 系 | 責務 | 依存してよいもの |
|----|------|------------------|
| `apps/generator` | 取得 → 原稿 → TTS → Drive 書込 | 外部 API / CLI（Infrastructure） |
| `apps/playback/web` | 一覧・再生・原稿表示 | `worker` の HTTP のみ |
| `apps/playback/worker` | Drive 読取 BFF | Drive API（Access 背後） |

禁止: `playback` ↔ `generator` の直接依存。二系統の runtime は互いに import しない。つながるのは Drive 上の file だけ。その形の正本が repo 根 `contracts/`。

例外: `apps/playback/web` の production runtime（`src/` 配下）は `worker` の HTTP のみに依存する。dev-only tooling（`web/vite.config.ts` の dev server middleware）に限り、`worker` の Composition Root を直接 import して dummy backend を local 起動できる。production の HTTP 入口（`worker/src/routes/fetch.ts`）は変更しない。

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
| `playback/web/src/{pages,components/feature,components/primitive,view-models,api,utils,lib}` | frontend skill（role と dir は 1 対 1。Feature/Primitive 分割と層 gate は `docs/decisions/2026-08-25T18-42-00-chore-playback-worker-web-layer.md`） |
| `playback/contracts` | web↔worker HTTP 境界共有型（API Client・Route・Controller・Application・Composition が import。Infrastructure は禁止） |

`playback/web` は Vite + TypeScript（vanilla）+ Pico.css classless。React / Next.js / shadcn は使わない（`docs/decisions/2026-08-18T11-12-00-feature-playback-web.md`）。

依存は内側へ。Composition Root だけが全層を結線する。

`generator` の Entities は generator に閉じる。UI / agent が共有して読む Domains の正は `contracts/`。言語横断の **Domain 型** module（共有 struct / Zod を正本にする）は作らない。

### `contracts/` の読み手

repo 根 `contracts/` は Drive 上の表現（配置・`manuscript.schema.json`）の SSOT。`apps/playback/contracts/`（web↔worker HTTP）とは別物。

| 層 | repo 根 `contracts/` |
|----|----------------------|
| generator **Application** | import する |
| playback worker **Infrastructure**（Drive 読取） | import する |
| generator **Infrastructure**（Drive 保存） | import しない |
| Entities / Composition Root / cmd / playback web | import しない |

禁止: field 手写し・Adapter 隣 snapshot。generator の Drive 書込の層分担は `docs/decisions/2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`。配置は `contracts/drive-layout.md`。

## 3. 外部 I/O

| 役割 | 接続 |
|------|------|
| 情報取得 | TwitterAPI.io（試作）/ GetXAPI（本運用）。Port は `ItemSource`。詳細は `docs/decisions/` の x-api-adoption |
| 原稿 | Cursor CLI（Port は `TextWriter`） |
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
| Unit | 対象ソースの隣（Go: `*_{分類}_test.go`、TS: 隣接 `*.{分類}.test.ts`） |
| Integration / Contract / System・E2E | `apps/generator/test/` または `apps/playback/test/`（web のブラウザ E2E は `web` 配下でも可） |

1. 分類は **file 名**に出す（`sociable_unit` / `narrow_integration` / `contract` / `system_e2e` 等）。dir 名だけに頼らない。runner の収集条件も分類名で絞り、命名忘れを収集漏れとして検出する
2. `integration` 一語で複数分類を兼ねない
3. 実境界に届かないものに `e2e` と付けない
4. Unit を共有 `tests/` に集めない
5. static と Unit は commit と GHA。Integration は push と GHA。Generator race は GHA の Unit 後に実行する。手順の正は `scripts/`。hook と GHA は同じ入口を呼ぶ caller であり、command を YAML や hook へ写さない
6. 片系は `scripts/generator/` と `scripts/playback/` から単独実行できる。root の `check-static.sh` / `test-integration.sh` は片系を呼ぶだけ。root の `test-unit.sh` は composer 契約を実行してから片系 unit を呼ぶ
7. runner は Playback = Vitest（`apps/playback`）、Generator = `go test`
8. generator Unit gate は **statement coverage 90%**（`covermode=atomic`、`-shuffle=on`、`-count=1`）。除外は Composition Root のみ。`error.go` / `names.go` / `constants.go` を名前では除外しない。Integration に Unit 閾値を載せない
9. generator static は `go build ./...` と **depguard** / `errcheck` / `govet` / `gofmt`（`golangci-lint`、`strict` allow）で build・層 import・静的な誤用を block する。Infrastructure が Application から import してよいのは **Port** のみ。playback static は Biome / tsc に加え **dependency-cruiser**（`apps/playback/.dependency-cruiser.mjs`）で層 import を block する
10. generator race gate は `go test -race` で Unit package を実行する。Integration package・Playback・本番 credential を使わない
11. generator module と GitHub Actions runner の Go version は **1.26.6** に固定する

実行手順（hook 導入・コマンド）は `README.md`。

## 6. 文書

| 文書 | 書くこと |
|------|----------|
| README | 地図・使い方・受け入れ・秘密の名前 |
| DESIGN | 本ファイル（規則のみ） |
| contracts/ | Drive 配置・原稿 JSON |

dir 単位の README は置かない。SPEC / PROPOSAL は置かない。
