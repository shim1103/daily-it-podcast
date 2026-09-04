---
name: Generator CLI 失敗は Delivery で kind 付き stderr へ写し、processenv は stderr head を cause 診断へ載せる
date: 2026-08-30T18:30:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. Generator CLI に client UI は無い。Internal Error を External へ写す場所は Driving Adapter の Delivery（`internal/delivery`）である。`cmd/generator` は `delivery.Format` を呼んで stderr へ書き、exit 非0 にするだけとする。
2. External stderr は次の行契約とする。`kind` は `domain` / `infrastructure` / `config` / `unknown`。`op` は Domain・Infra の `Op` または単一 `config.Error` の key。無ければ行ごと省略。`message` は `err.Error()` の1行化。`cause[i]` は Unwrap 連鎖。
3. kind 判定は具象型で行う。`*entities/errors.Error` は domain、`*config.Error` / `*config.Errors` は config、cursorcli / processenv / gemini / getxapi / gdrive / oauth の `*Error` は infrastructure、それ以外は unknown。外側の typed error を先に採用する。
4. Application は Infrastructure Error を Domain Error へ変換しない。Infra は return のみ。
5. `processenv.Launcher` は child stderr を Discard しない。失敗時は先頭最大 300 byte を cause 診断へ載せる。secret 値・stdin 本文は Error に載せない。stderr に secret 値が含まれる場合は head 自体を省略する。
6. Decision `2026-08-18T16-30-00` §1-5 の「stderr の内容文字列を上位へ写さず、内部診断として扱う」は、本 Decision が supersede する。成功/失敗の観測面（stdout JSON + exit code）は旧 Decision のままである。

## 2. Reason

1. playback の `mapInternalError` → `createHttpErrorResponse` と同型で、変換は境界に1箇所だけ置く。cmd が infra 型を import すると既存 invariant と depguard 例外が増える。
2. 失敗を `generator: %v` 1行だけにすると、Domain / Infra / config の切り分けが GHA log から不能になる。kind 行は分類を固定し、message / cause は診断本文を残す。
3. Cursor CLI 失敗の実因は child stderr にある。Discard すると Delivery まで届かない。300 byte head と secret 含有時の省略は、診断と非漏洩の衝突を固定する。

## 3. Rejected

1. Application で Infra を Domain へ畳む案 — 層の変換境界を内側へずらす。Application は Infra 型を知らない。
2. cmd が具象 Error を type switch する案 — Driving Adapter が infra を import する。
3. stderr 全文を無制限に Error へ載せる案 — secret 混入面が広がり、観測 payload が膨らむ。
4. 旧 Decision どおり stderr を上位へ写さない案 — System 失敗の切り分けが exit code だけになる。
