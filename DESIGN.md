# DESIGN

最終更新: 2026-08-30

地図・使い方・受け入れは `README.md`。deploy・Access・GHA 運用・secret 登録は `DEPLOY.md`。Drive 表現は `contracts/`。本書は **層・依存・所有・test 配置の規則**だけを書く。

## 1. システム境界

| 系 | 責務 | 依存してよいもの |
|----|------|------------------|
| `apps/generator` | 取得 → 原稿 → TTS → Drive 書込 | 外部 API / CLI（Infrastructure） |
| `apps/playback/web` | 一覧・再生・原稿表示 | `worker` の HTTP のみ |
| `apps/playback/worker` | Drive 読取 BFF | Drive API（入場境界は `DEPLOY.md`） |

禁止: `playback` ↔ `generator` の直接依存。二系統の runtime は互いに import しない。つながるのは Drive 上の file だけ（形の正本は repo 根 `contracts/`）。

例外: `apps/playback/web` の production runtime（`src/`）は `worker` の HTTP のみ。dev-only（`web/vite.config.ts` middleware）に限り `worker` Composition Root を import して dummy backend を local 起動できる。production HTTP 入口（`worker/src/routes/app.ts`、`worker/src/worker-entry.ts`）は変更しない。

## 2. 層と依存（Clean Arch / DIP）

用語・import 規則の正は skill。ここでは **このリポでの対応**だけ示す。

- Backend: [architecture/backend](file:///Users/shim0729/.claude/skills/architecture/backend/SKILL.md)
- Port は Application 所有・Infrastructure 実装 → [ports-adapters](file:///Users/shim0729/.claude/skills/architecture/ports-adapters.md)
- Ring: [ring-model](file:///Users/shim0729/.claude/skills/architecture/ring-model.md)
- Frontend: [architecture/frontend](file:///Users/shim0729/.claude/skills/architecture/frontend/SKILL.md)

| 置き場 | skill 上の層 |
|--------|----------------|
| `generator/internal/entities` | Entities |
| `generator/internal/application` | Application（UseCase + Port IF） |
| `generator/internal/application/build` | Builder helper（Gate ではない） |
| `generator/internal/config` | Configuration Boundary |
| `generator/internal/infrastructure` | Infrastructure |
| `generator/internal/composition` | Composition Root |
| `generator/cmd/generator` | CLI Driving Adapter |
| `playback/worker/src/entities` 等 | 上に同じ（BFF） |
| `playback/web/src/{pages,components/feature,components/primitive,view-models,api,utils,lib}` | frontend（role と dir は 1 対 1） |
| `playback/contracts` | web↔worker HTTP 境界共有型（Infrastructure は import 禁止） |

`playback/web` は Vite + TypeScript + React + Pico.css classless。`playback/worker` 入口は Hono、型同期は Hono RPC。Next.js / shadcn / TanStack は使わない。Playback UI の concept / 視覚言語は `docs/decisions/` を正とし、本書へ写さない。

依存は内側へ。Composition Root だけが全層を結線する。未完了 index は `docs/tasks/todo/*-lane.md`。

`generator` Entities は generator に閉じる。言語横断の共有 Domain 型 module は作らない。UI / agent が共有して読む形は `contracts/`。

### `contracts/` の読み手

repo 根 `contracts/` は Drive 表現の SSOT。`apps/playback/contracts/`（HTTP）とは別物。

| 層 | repo 根 `contracts/` |
|----|----------------------|
| generator **Application** | import する |
| playback worker **Infrastructure**（Drive 読取） | import する |
| generator **Infrastructure**（Drive 保存） | import しない |
| Entities / Composition Root / cmd / playback web | import しない |

禁止: field 手写し・Adapter 隣 snapshot。配置は `contracts/drive-layout.md`。

## 3. 外部 I/O

| 役割 | 接続 |
|------|------|
| 情報取得 | 公式 API / RSS の複数源（HackerNews・Lobsters・ITmedia NEWS）。Port は `ItemSource`。源ごとに専用 Adapter、facade なし（RSS 汎用 Adapter も作らない）。複数源 merge は Composition の composite。Application は源個数を知らない。源の選定理由は `docs/decisions/` |
| 原稿 | Cursor CLI（Port `TextWriter`） |
| TTS | Gemini |
| Drive | Google Drive + OAuth refresh |

`generator/internal/config` が startup で process environment を一度だけ読み、検証済み capability Config を Composition へ渡す。HTTP Adapter は `*http.Client` と必要な capability config / credential だけを受け取る。保存元・environment key は知らない。Cursor CLI 経路の secret 生値は Composition → `processenv` closure に閉じ、`cursorcli` は inject 名・argv だけを持つ。

ブラウザに Drive credential を置かない。注入の運用は `DEPLOY.md`。

## 4. 認証の層所有

- UI 入場・hostname・Access は `DEPLOY.md`
- アプリ内マルチテナント OAuth は作らない
- Drive credential は worker / generator の Infrastructure が持つ（Web は持たない）

## 5. Test 配置

分類・FIRST・最小化の正: [testing-strategy](file:///Users/shim0729/.claude/skills/testing-strategy/SKILL.md)（naming / levels 含む）。

このリポでの適用だけ:

| Scope | 置き場 |
|-------|--------|
| Unit | 対象ソースの隣（Go: `*_{分類}_test.go`、TS: `*.{分類}.test.ts`） |
| Integration / Contract / System | Generator: `apps/generator/test/`。Playback Integration: `apps/playback/test/integration/`。Playback browser E2E: `apps/playback/test/e2e/`（`test/contract/` は置かない） |

1. 分類は **file 名**に出す（`sociable_unit` / `narrow_integration` / `broad_integration` / `contract` / `system` / `e2e`）。`integration` 一語で兼ねない。Generator System 分類語は `system`。Playback browser E2E 分類語は `e2e`（Generator の `system` と混線させない）
2. 実境界に届かないものに `e2e` と付けない
3. Unit を共有 `tests/` に集めない
4. 手順の正は `scripts/`。hook と GHA は同じ入口を呼ぶ caller。toolchain command を YAML / hook へ写さない
5. root の `check-static.sh` / `test-unit.sh` / `test-integration.sh` は片系を呼ぶだけ。Unit composer 契約のあと片系 unit
6. runner: Playback = Vitest（Unit / Integration）、Playwright（browser E2E）。Generator = `go test`
7. generator Unit gate: statement coverage 90%（`covermode=atomic`、`-shuffle=on`、`-count=1`）。除外は Composition Root（`internal/composition/**`）・CLI Driving Adapter（`cmd/**`）・build tag 付き suite・Broad Integration 以上。`error.go` / `names.go` / `constants.go` を名前では除外しない。secret なし Narrow（`apps/generator/test/` の tag なし）は production code のカバー分母に含める。local の condition report は `scripts/generator/report-condition-coverage.sh`（`gobco v1.3.4`）。threshold なし・hard gate ではない。未使用 function / `select` は対象外で完全な branch coverage ではない
8. playback Unit gate: branch coverage（全体 100%、外部境界層は glob で 90%）。secret なし Narrow Integration を分母に含める（SU + NI 合算）。Broad 以上・E2E は分母に入れない。設定は `apps/playback/vitest.config.mjs`
9. generator static: `go build` + golangci（depguard / errcheck / govet / gofmt）。Infrastructure→Application は Port のみ。playback static: Biome / tsc / dependency-cruiser
10. generator race: `go test -race` は Unit package のみ
11. Go / Node version の正本は `go.mod` / `.nvmrc`。GHA は `*-version-file` で参照
12. Integration **gate**（pre-push / GHA Integration）: secret なし Narrow + Broad。System / Playback E2E / 本番 produce は gate 外（収集・入口の正は code と `DEPLOY.md`）。Playback Vitest Integration project は `system_e2e` を収集しない

実行手順の入口一覧は `README.md`。credential 付き定時・secret 名は `DEPLOY.md`。
