---
name: 原稿 TextWriter を Cursor CLI から Cloud Agents REST へ移し CLI / commandlaunch を廃止する
date: 2026-09-03T17:03:33
branch: feature/generator-cursor-cli-to-http-api
---

## 1. Decision

1. 日次原稿の `port.TextWriter` 実装を Cursor CLI（`agent` command + `commandlaunch`）から **Cloud Agents REST**（`https://api.cursor.com/v1/agents`）へ移す。Port の signature は変えない。契約値（dir・constructor・定数・不在 package）の正本は A（`internal/infrastructure/manuscript/cursorapi` および削除後の不在）とする。
2. 呼び出し形は **no-repo**（`repos` / `env` omit）、**毎回 create**（agent 再利用・前回 conversation は持たない）、完了観測は **SSE**（`GET .../runs/{runId}/stream`）で終端 `result.text` を text 断片とする。
3. secret 結線は既存 HTTP Adapter（`gemini`）と同型とする。Composition が `Reveal()` 済み API key と、全体 `Timeout` を置かない `*http.Client` を Adapter へ渡す。CLI 時代の child env inject（N2）・`processenv` factory 経路は廃止する。
4. model は Adapter 定数で `composer-2.5` に固定する。runtime の `GET /v1/models` による存在確認はしない。
5. HTTP retry は情報源 Adapter の最小方針を基にし、Cursor の 429 だけ粘る。`client.Do` error と **idempotent な GET** の 5xx は 1 回即再試行。429 は有限回 + backoff（`Retry-After` があれば尊重）。**POST create の 5xx / 曖昧 timeout は再試行しない**（非 idempotent・二重 agent 回避）。401 / 403 / 400、run 終端 error、空 text は再試行しない。SSE 途中断からの再 create もしない。
6. 長時間待ち・streaming 向け Client は全体 `Timeout` を置かず、各 request は `ctx` 伝播のみとする。短時間用の共有 30s HTTP Client には乗せない。run 全体の上限は GHA job / process cancel に委ねる。
7. 指示対象を失った CLI 経路を残さない。`commandlaunch` / `processenv` / `cursorcli`、`probe-cursor-cli` workflow・script、System workflow の Cursor CLI install 手順を削除する。
8. 本 Decision は次を置き換える（旧 file 本文は書き換えない。読み手は本 file を正とする）。
   1. `docs/decisions/2026-08-18T16-30-00-feature-cursor-text-writer.md` の Rejected「Cloud API / SDK で複雑性を増やさない」および CLI argv / stdout envelope を正とする結論のうち、**transport が CLI である部分**。
   2. Cursor CLI 専用の秘密供給・`commandlaunch` 結線に関する後続 Decision（例: `2026-08-22T11-55-22`、`2026-08-25T13-53-55`、`2026-08-29T10-50-44`、`2026-08-29T13-48-57`、`2026-08-29T13-48-58`、`2026-08-29T14-51-24`）のうち、**command 出口・N2 inject・processenv factory を前提とする部分**。HTTP 境界一般（Reveal 済み primitive を Adapter へ渡す）は維持する。

non-scope（本 Decision で決めない）: Cursor が返した text が Domain の Draft として invalid な場合の扱い、DomainError の変更。

## 2. Reason

### なぜ REST Cloud Agents か

goal は Cursor CLI と `commandlaunch` を廃止し HTTP 経由で Cursor を使うこと。Go に first-party SDK は無く、公式の HTTP 面は Cloud Agents API（public beta）である。TS/Python SDK や sidecar を足すと runtime が増え、既存 generator（単一 Go process + 標準 `net/http`）の対称性を壊す。Local SDK runtime は結局 local executor に戻り、command 廃止と矛盾しやすい。

Cloud Agents は chat-completions ではなく agent workflow である（公式 Overview）。現行 CLI は `--mode ask` の断片生成だったが、HTTP 面に同等の ask-only エンドポイントは無い。no-repo + prompt で text を取り `result.text` を断片とするのが、Port を変えずに載せる最短経路である。品質・token は未実測であり lane（D）へ残す。未実測を理由に transport 移行自体を止めない（goal が明示済み）。

### なぜ no-repo・毎回 create・SSE か

原稿生成は repo clone・PR・follow-up を必要としない。1 日 1 回の produce で前回 conversation は不要（YAGNI）。毎回 create すれば agentId 永続化・resume・MCP 再注入の分岐が消える。完了観測は公式 SSE が round-trip を減らし、poll は rate を削る。実装が SSE で行き詰まった時の fallback は将来の別判断とし、初回の正は SSE とする。

### なぜ secret / Client を gemini 同型にするか

HTTP 境界では既に Composition が検証済み値を `Reveal()` して Adapter へ渡す。CLI だけ「Go が値を持たない」特殊則を残すと、出口が HTTP なのに結線だけ command 時代の形が残り、Orthogonality を破る。N2 / allowlist / `processenv` は child process のための知識であり、HTTP 化で指示対象が消える。消えた前提の間接を残さない（KISS）。

Cursor run は分単位になり得る。共有 30s Client に載せると正常完了前に切れる。全体 Timeout を置かない Client で `ctx` に委ねるのは、情報源 Adapter が独自 timeout を持たない方針（先行 Decision `2026-09-02T15-27-00`）と同型で、最終防壁を GHA / signal に置く。関数名は vendor に依存させず「Timeout を置かない Client」として Composition に置く。

### なぜ model list 確認をしないか

現行 CLI も定数固定だった。日次 1 回で起動時に models を列挙しても、稀な破壊的変更への過剰防衛になる。失敗したら Infra Error で落ち、定数を直す方が分岐が少ない（KISS）。

### なぜ retry が gemini 全コピーではないか

gemini の重装 retry は TTS 無料枠など vendor 固有前提への対処だった。情報源 3 Adapter は transient を 1 回拾うだけで足りると既に決めた（`2026-09-02T15-27-00`）。Cursor は公式が 429 に backoff を勧めるため、429 だけ粘る。一方 `POST /v1/agents` は非 idempotent で、timeout 後の盲目再 create は二重 agent・二重課金になり得る。安全側は即失敗である。

### なぜ CLI / probe / install を消すか

HTTP 化後、fake `agent` binary・Cursor CLI install・probe-cursor-cli は指示対象を失う。死んだ経路を残すと「どちらが正か」が二重になる。不在を契約（A）として固定する。

## 3. Rejected

1. **TypeScript / Python SDK または sidecar を generator から呼ぶ案** — Go 単一 process の境界が増え、既存 HTTP Adapter と非対称。goal の「HTTP」は REST で足りる。
2. **Local SDK / 引き続き CLI executor を裏に残す案** — commandlaunch 廃止と矛盾する。
3. **repo 付き cloud agent（clone・PR）案** — 原稿専用に不要な git/tool 面が増える（YAGNI）。
4. **agent 再利用・follow-up run で context を積む案** — 1 日 1 回・前回不要。永続化と resume 分岐が過剰。
5. **完了観測を poll のみにする案** — 実装は単純だが HTTP 回数と rate 消費が増える。初回の正は SSE。
6. **CLI 時代の processenv / N2 inject を HTTP 後も残す案** — child process が無いのに command 出口契約が残る。指示対象喪失。
7. **runtime で `GET /v1/models` してから create する案** — 日次 1 回への過剰防衛。定数固定失敗→code 修正で足りる。
8. **gemini と同型の重装 retry を POST create にも掛ける案** — 非 idempotent。二重起動リスクが費用対効果を上回る。
9. **情報源と同型で 429 を非 retry にする案** — Cursor docs が 429 backoff を推奨する。情報源の「429 は仕様変更の兆候」前提がここには無い。
10. **probe-cursor-cli や System の CLI install を HTTP 化後も残す案** — 死んだ経路の二重 SSOT。
