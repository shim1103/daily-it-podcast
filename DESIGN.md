# DESIGN

最終更新: 2026-08-25（playback Feature/Primitive dir 分割と dependency-cruiser 層 gate／deploy・Access 運用を `DEPLOY.md` へ分離）

地図・使い方・受け入れ・秘密の名前は `README.md`。deploy・Access・公開境界は `DEPLOY.md`。Drive に載る表現は `contracts/`。本書は **層・依存・所有・test 配置の規則**だけを書く（パス百科・Drive / HTTP 契約・運用方針の写しは置かない）。

## 1. システム境界

| 系 | 責務 | 依存してよいもの |
|----|------|------------------|
| `apps/generator` | 取得 → 原稿 → TTS → Drive 書込 | 外部 API / CLI（Infrastructure） |
| `apps/playback/web` | 一覧・再生・原稿表示 | `worker` の HTTP のみ |
| `apps/playback/worker` | Drive 読取 BFF | Drive API（入場境界は `DEPLOY.md`） |

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
| `generator/cmd/generator` | CLI Driving Adapter（薄い入口。成否は OS exit / stderr） |
| `playback/worker/src/entities` 等 | 上に同じ（BFF） |
| `playback/web/src/{pages,components/feature,components/primitive,view-models,api,utils,lib}` | frontend skill（role と dir は 1 対 1。Feature/Primitive 分割と層 gate は `docs/decisions/2026-08-25T18-42-00-chore-playback-worker-web-layer.md`） |
| `playback/contracts` | web↔worker HTTP 境界共有型（API Client・Route・Controller・Application・Composition が import。Infrastructure は禁止） |

`playback/web` は Vite + TypeScript + React + Pico.css classless。`playback/worker` の HTTP 入口は Hono、web↔worker の型同期は Hono RPC。Next.js / shadcn / TanStack は使わない（`docs/decisions/2026-08-26T00-00-00-architecture-reconsider-react-hono.md`）。

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

秘密の HTTP / command 出口は vendor Adapter が持たない。出口契約は `secrettransport` / `commandlaunch`、置き場 runtime と結線は Composition。2軸と配置の正は `docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md`。HTTP × AgentSecrets の正本吸収は `docs/decisions/2026-08-25T19-36-11-feature-generator-agentsecrets-http-transport.md`。Composition 内の表（bindings）と Client/Launcher 組み立て（runtime）の file 分割は `docs/decisions/2026-08-26T14-58-45-feature-generator-agentsecrets-cursor-command-launcher.md`（本文は写さない）。

ブラウザに Drive の長期秘密を置かない。フォルダ ID・OAuth の値は実行設定（`contracts/` 外）。

## 4. 認証の層所有

- UI 入場・公開 hostname・Access ポリシーは `DEPLOY.md`（本書へ再掲しない）
- アプリ内マルチテナント OAuth は作らない
- Drive credential は worker / generator の Infrastructure が持つ（Web は持たない）

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
8. generator Unit gate は **statement coverage 90%**（`covermode=atomic`、`-shuffle=on`、`-count=1`）。除外は Composition Root（`internal/composition/**`）と CLI Driving Adapter（`cmd/**`）のみ。`error.go` / `names.go` / `constants.go` を名前では除外しない。Integration に Unit 閾値を載せない
9. playback Unit gate は **branch coverage**（`@vitest/coverage-v8`）。全体は 100%、外部境界・状態分岐を持つ層（`worker/src/routes/**` 等）は個別 glob で 90% に緩める。global threshold は個別 glob 該当 file も合算した全体値で判定されるため、両者は独立に閾値未達へならない設計にする。型安全のためだけに残る到達不能分岐は `v8 ignore` と理由 comment で除外し、test を書いて無理に通さない。設定は `apps/playback/vitest.config.mjs` の root top-level `test.coverage`（`projects` 配下の個別 project には coverage 設定を持てない Vitest の制約）。Integration に Unit 閾値を載せない
10. generator static は `go build ./...` と **depguard** / `errcheck` / `govet` / `gofmt`（`golangci-lint`、`strict` allow）で build・層 import・静的な誤用を block する。Infrastructure が Application から import してよいのは **Port** のみ。playback static は Biome / tsc に加え **dependency-cruiser**（`apps/playback/.dependency-cruiser.mjs`）で層 import を block する
11. generator race gate は `go test -race` で Unit package を実行する。Integration package・Playback・本番 credential を使わない
12. Go version の正本は `apps/generator/go.mod` の `go` directive。Node version の正本は `apps/playback/.nvmrc`。GitHub Actions は両 file を `go-version-file` / `node-version-file` で参照し、YAML に version 文字列を直書きしない。local の Node version 不一致は `apps/playback/package.json` の `engines` + `.npmrc` の `engine-strict=true` が `npm ci` 時点で検知する

実行手順（hook 導入・コマンド）は `README.md`。

## 6. 文書

| 文書 | 書くこと |
|------|----------|
| README | 地図・使い方・受け入れ・秘密の名前 |
| DESIGN | 本ファイル（層・依存・所有・test 規則のみ） |
| DEPLOY | deploy・Access・公開境界（運用方針の SSOT） |
| contracts/ | Drive 配置・原稿 JSON |

dir 単位の README は置かない。SPEC / PROPOSAL は置かない。運用方針を README / DESIGN へ写さない。
