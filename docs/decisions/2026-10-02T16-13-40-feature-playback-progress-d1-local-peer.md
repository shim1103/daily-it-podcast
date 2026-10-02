---
name: 進捗の永続先（D1 | in-memory）は episode の永続先と独立した明示設定で選び、既定値も binding の有無も使わない
date: 2026-10-02T16:13:40
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. 進捗の永続先の選択（`in-memory` | `d1`）は、episode の永続先の選択（`in-memory` | `r2`）とは別の明示設定として runtime config に持つ。どちらも呼び出し側が常に明示し、検証も互いに独立させる。不正な値は未指定を含めて `PlaybackRuntimeConfigError` で弾く。
2. binding の必須条件も独立させる。R2 binding（`EPISODES`）は episode の mode が `r2` の時だけ、D1 binding（`EPISODE_PROGRESS`）は進捗の mode が `d1` の時だけ必須にする。
3. 2 つの mode の組合せは制限しない。本番経路は episode `r2` ＋ 進捗 `d1` を明示して選ぶ。
4. 進捗の mode を既定値で補わず、env の binding の有無から暗黙に選ばない。

## 2. Reason

1. episode の原稿（R2）と進捗（D1）は別の資源で、変更理由・欠落した時の失敗・使う場面が異なる。選択を 1 つの mode に束ねると、2 つの資源の組合せが固定され、片方の都合でもう片方の設定が決まる（`design-philosophy.md` §2-1 Orthogonality）。
2. 進捗を episode の mode に従属させると、進捗に D1 を使わない場面でも D1 binding を要求してしまう。たとえば R2 の実物は読みたいが進捗は書きたくない smoke、逆に episode は in-memory のまま進捗だけ D1 で試す local が、設定として表せない。独立にすれば、使う資源の分だけ binding を要求できる。
3. 設定の欠落を無言で別の実装へ落とさず Error にする既存の原則を、進捗にも episode と同じ形で適用できる。判定結果を無言で代替実装へ逃がすと設定漏れが観測不能になる防御設計の原則（`.claude/skills/1:terms/error-handling/defensive-design.md` §2.4）と、未指定を既定値へ黙って読み替えない原則（同 §7.3）に当たる。本番設定を暗黙に Fake へ落とさない `2026-08-19T17-38-00-playback-repository-selection.md`、in-memory は専用の明示 option がある時だけ許す `2026-08-20T12-57-00-playback-runtime-config-validation.md`、進捗 D1 の未結線を設定不足として弾く `2026-10-01T19-22-21-feature-playback-progress-d1-local-peer.md` が、進捗にも同じ形で効く。
4. 既定値で補うと、呼び出し側が mode を書き忘れた時に、意図しない側の実装が無言で選ばれる。binding の有無で選ぶと、誤結線（binding の付け忘れ・付け間違い）が「in-memory で動く」ように見え、観測できない。どちらも、同種の判断をいつ・どこで行っても同じ結論にする一貫性（`design-philosophy.md` §4-5）を壊す。明示させれば、設定の不足はそのまま Error として現れる。
5. 2 つの mode の検証が同じ config module に閉じるので、必須 binding の一覧と欠落時の Error は一箇所に残り、選択側で再判定しない。

## 3. Rejected

1. **進捗を episode の `r2` mode に従属させる案**（現行。commit `c6be266` で、r2 mode の時に D1 binding を必須にした形） — 資源の組合せが固定され、進捗に D1 を使わない場面でも D1 binding を要求してしまう。未結線を弾く意図そのものは、進捗の mode が `d1` の時の必須として残す（`2026-10-01T19-22-21-feature-playback-progress-d1-local-peer.md`）。
2. **進捗の mode 未指定を既定値（in-memory や d1）で補う案** — 無言 fallback になる。in-memory に補うと保存が消える設定が、d1 に補うと D1 を持たない場面が、書き忘れのまま通る。
3. **env の binding の有無から自動で選ぶ案** — 暗黙で、誤結線が観測できない。binding を付け忘れた本番が in-memory で動いてしまう。
