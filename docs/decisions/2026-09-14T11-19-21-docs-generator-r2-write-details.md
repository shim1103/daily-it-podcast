---
name: R2 Adapter の失敗は transient 有限 retry・4xx fail-fast・Application 不変・HTTP code 継承・resource id 非露出
date: 2026-09-14T11:19:21
branch: docs/generator-r2-write-details
---

## 1. Decision

1. 先行 Decision `2026-09-14T11-04-30` の §1-6（失敗方針の一行）を、本 file が **R2 Adapter の error / retry 方針**として具体化する。topology・credential・書込意味論は先行を正とし、再掲しない。
2. **generator Infrastructure（R2 Writer）**: network 失敗・HTTP 5xx・429 は **有限回 retry**（情報源 Adapter と同型の transient。回数の定数正本は Adapter / C）。その他 4xx は **fail-fast**（再試行しない）。
3. **generator Application**: Write 失敗の扱い・層写像は **変更しない**（Infra Error を Domain に捏造しない。exit 非 0 の既存契約を維持）。
4. **quota secondary は作らない**: Gemini TextWriter のような「枠枯渇 → 別取得元」は R2 に置かない（storage に代替正本が無い）。先行 `11-04-30` Rejected の storage secondary と同旨。
5. **playback Infrastructure（R2 Repository）**: storage I/O 失敗（認証・network・非 2xx・応答形式不正）は Infrastructure Error を throw する。音声欠落は **`undefined`（throw しない）** — Port 契約を維持。
6. **playback HTTP**: 既存写像を維持する。`configuration_error`（500）/ `unavailable`（503）/ 欠落は既存 NotFound。契約 code の正本は `apps/playback/contracts/`。R2 固有の外部 code は増やさない。
7. **観測・露出**: Error message / log に bucket 名、object key、Account ID、Access Key、secret 実値を載せない（Drive 時代の file id / folder id 非露出と同型）。

## 2. Reason

1. 日次 produce の Put は数回で、Gemini 無料枠枯渇のような product quota secondary が要る負荷ではない。必要なのは一過性の network / 5xx / 429 への粘りだけ。情報源 Adapter が既に transient 有限 retry を持つので、storage も同型に揃えると層の読みが一致する。
2. 4xx（認証誤り・不正 key・client 過り）を retry しても直らない。fail-fast が切り分けを早める。
3. Application を触ると「storage vendor 変更」が Domain / UseCase の変更理由に混ざる。Infra の失敗表現を変えるだけで足りる。
4. playback の欠落（不完全ペア）を Infra Error にすると、公開型書込の妥協（先行 `2026-08-30T23-32-00`）が HTTP 5xx になり、一覧が赤くなる。Port の `undefined` を維持するのが正しい。
5. HTTP 契約 code を増やすと web / E2E が storage 内部事情を知る。既存 `unavailable` で外部一時不能を表せば足りる。
6. resource id を message に出すと Access 配下でも log 流出面が増える。Drive Adapter が既に拒否している不変条件を R2 でも続ける。

## 3. Rejected

1. Gemini 型の枠枯渇 secondary / Drive 退避 — 正本が再び割れる。代替 storage が無い。
2. 全失敗を無制限 retry — 4xx と設定ミスを隠す。
3. Application に R2 専用 Domain Error を足す案 — vendor 変更が Domain 語彙を増やす。
4. 音声欠落を `unavailable` にする案 — 不完全ペアが運用故障に見える。
5. R2 専用の外部 HTTP code を増やす案 — 契約面が storage 事情を漏洩する。
6. Error message に bucket / key を載せて調査を速くする案 — secret・配置の露出面が増える。log の op 種別固定語で足りる。
