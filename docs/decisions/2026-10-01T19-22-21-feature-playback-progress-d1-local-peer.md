---
name: 本番経路の進捗 D1 binding は必須にし、未結線を永続しない Stub へ落とさず設定不足として弾く
date: 2026-10-01T19:22:21
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 本番経路（r2 mode 固定）では、進捗の永続 binding（`EPISODE_PROGRESS`、D1）を `EPISODES` と同じく必須にする。未結線は `validatePlaybackEnv` の r2 分岐が `PlaybackRuntimeConfigError` で弾く。永続しない `StubProgressRepository` へ無言で落とさない。500 `configuration_error` への写像は `2026-08-20T12-57-00-playback-runtime-config-validation.md` に従い、ここへ写さない。
2. in-memory mode は env の中身を見ず、永続しない `StubProgressRepository` を選ぶ。binding の有無で選択を変えない。
3. 進捗 repository の選択も、episode repository と同じく `validatePlaybackEnv` を通して決める。binding の有無を選択側で再判定しない。

## 2. Reason

1. 進捗の Stub は何も永続しない。書込は何も保存せず、送られた `clientAt` をそのまま勝ち側の `first*` として HTTP 200 で返し、pull は空を返す。呼び出し側からは正常応答に見え、保存できたように見えて消える。設定の欠落が観測できず、利用者のデータだけが静かに失われる。判定結果を無言で代替実装へ逃がすと設定漏れが観測不能になる、という防御設計の原則（`.claude/skills/1:terms/error-handling/defensive-design.md` §2.4）と、未指定を安全側の既定値へ黙って読み替えない原則（同 §7.3）に当たる。
2. 同じ composition の `EPISODES` 欠落は既に throw している。進捗だけ落とすと、同じ「本番に必要な binding の欠落」への扱いが非対称になり、設定退行や誤 deploy で R2 は止まるのに進捗だけ 200 のまま消える。
3. Controller 一式は request ごとに組み立てるため、設定不足は最初の request から 500 として現れる。設定の欠落を、永続しない 200 の長い継続ではなく、直後に見える失敗へ変える。
4. in-memory を env 非依存にするのは、local と unit が D1 を持たずに動く前提を保つため。in-memory は専用の明示 option がある時だけ許す既存の方針（`2026-08-20T12-57-00-playback-runtime-config-validation.md`）と、本番設定を暗黙に Fake へ落とさない方針（`2026-08-19T17-38-00-playback-repository-selection.md`）に揃い、binding の有無という暗黙の条件で選択を変えない。
5. 選択を `validatePlaybackEnv` へ通すと、必須 binding の一覧と欠落時の Error が config module の一箇所に閉じ、episode と progress で判定が二重化しない。

## 3. Rejected

1. **binding が無ければ mode に関わらず Stub へ落とす案**（従来の挙動） — r2 mode でも設定退行や誤 deploy が永続しない 200 になる。review でも Must fix と指摘された。
2. **binding 欠落を警告ログだけ出して続行する案** — Stub の sentinel 200 は呼び出し側から正常に見える。ログは request の応答を変えず、利用者の保存喪失は止まらない。
3. **r2 mode でも D1 を任意 binding のまま、選択側で有無を分岐し続ける案** — `EPISODES` との非対称が残り、必須 binding の判定が config module と選択側に二重化する。
