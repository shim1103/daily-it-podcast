---
name: 進捗の mode が d1 の時は進捗 D1 binding を必須にし、未結線を永続しない実装へ落とさず設定不足として弾く
date: 2026-10-01T19:22:21
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 進捗の mode が `d1` の時（本番経路は episode `r2` ＋ 進捗 `d1` を明示する）は、進捗の永続 binding（`EPISODE_PROGRESS`、D1）を必須にする。未結線は config module の検証が `PlaybackRuntimeConfigError` で弾く。永続しない実装（廃止した no-op の `StubProgressRepository` など）へ無言で落とさない。500 `configuration_error` への写像は `2026-08-20T12-57-00-playback-runtime-config-validation.md` に従い、ここへ写さない。
2. 進捗の mode が `in-memory` の時は env の中身を見ず、Map で動く in-memory の進捗 repository を選ぶ（no-op の Stub を廃止した理由は `2026-10-02T13-06-35-feature-playback-progress-d1-local-peer.md`）。binding の有無で選択を変えない。
3. 進捗 repository の選択も、episode repository と同じく config module の検証を通して決め、binding の有無を選択側で再判定しない。検証は episode と進捗で独立する（`2026-10-02T16-13-40-feature-playback-progress-d1-local-peer.md`）。
4. 進捗の mode を episode の mode から独立した明示設定にする判断は `2026-10-02T16-13-40-feature-playback-progress-d1-local-peer.md` に従い、ここへ写さない。

## 2. Reason

1. 従来の進捗 Stub（廃止済み）は何も永続しない。書込は何も保存せず、送られた `clientAt` をそのまま勝ち側の `first*` として HTTP 200 で返し、pull は空を返す。呼び出し側からは正常応答に見え、保存できたように見えて消える。設定の欠落が観測できず、利用者のデータだけが静かに失われる。判定結果を無言で代替実装へ逃がすと設定漏れが観測不能になる、という防御設計の原則（`.claude/skills/1:terms/error-handling/defensive-design.md` §2.4）と、未指定を安全側の既定値へ黙って読み替えない原則（同 §7.3）に当たる。
2. 進捗に D1 を選んだのに binding が欠けた場合、episode の `EPISODES` 欠落（episode が `r2` の時）は既に throw している。進捗だけ落とすと、同じ「選んだ資源に必要な binding の欠落」への扱いが非対称になり、設定退行や誤 deploy で R2 は止まるのに進捗だけ 200 のまま消える。
3. Controller 一式は request ごとに組み立てるため、設定不足は最初の request から 500 として現れる。設定の欠落を、永続しない 200 の長い継続ではなく、直後に見える失敗へ変える。
4. 進捗の in-memory を env 非依存にするのは、local と unit が D1 を持たずに動く前提を保つため。in-memory は専用の明示 option がある時だけ許す既存の方針（`2026-08-20T12-57-00-playback-runtime-config-validation.md`）と、本番設定を暗黙に Fake へ落とさない方針（`2026-08-19T17-38-00-playback-repository-selection.md`）に揃い、binding の有無という暗黙の条件で選択を変えない。
5. episode と進捗でそれぞれ、必須 binding の一覧と欠落時の Error が config module の検証の一箇所に閉じ、判定が二重化しない。

## 3. Rejected

1. **binding が無ければ mode に関わらず Stub へ落とす案**（従来の挙動） — 進捗に D1 を明示した本番経路でも、設定退行や誤 deploy が永続しない 200 になる。review でも Must fix と指摘された。
2. **binding 欠落を警告ログだけ出して続行する案** — Stub の sentinel 200 は呼び出し側から正常に見える。ログは request の応答を変えず、利用者の保存喪失は止まらない。
3. **進捗に D1 を選んでも任意 binding のまま、選択側で有無を分岐し続ける案** — `EPISODES` との非対称が残り、必須 binding の判定が config module と選択側に二重化する。
4. **進捗の D1 必須を episode の `r2` mode に従属させる案**（以前の答え。commit `c6be266`） — episode が `r2` なら進捗の mode に関わらず D1 binding を要求するので、進捗に D1 を使わない場面でも D1 を要求し、組合せが固定される。進捗の mode を独立した明示設定に分けたため採らない。
