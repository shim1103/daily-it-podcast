---
name: 再生進捗の正本はR2外のD1状態行（1 episode=1行）とし、UI信号とGetのprogress形をそれで表す
date: 2026-09-19T16:25:18
branch: feature/playback-progress
---

## 1. Decision

1. 再生進捗（listening progress）の正本は、episode content（R2の原稿json・mp3・将来のcatalog master）とは別storeに置く。実装storeは Cloudflare **D1** の状態表1本とする。
2. 利用者は Access 配下の単一userであり、端末横断でも同一のsingletonとする。行の鍵は `episodeId` のみとし、`userId` は持たない。
3. 1 episode につき最新状態を1行で持つ。append-onlyのevent履歴表は本 Decision の範囲に含めない（読者が無い間は作らない）。
4. 状態行が表す意味は次とする。契約の型・列定義の正本は後続の schema / migration artifact とし、本 Decision は意味だけを固定する。
   1. `positionSec` — 共有された最後の再生カーソル。完走後も0へ戻さず、末尾の位置のまま残す
   2. `lastPlayedAt` — 最後に進捗が更新された日時
   3. `firstPlayedAt` — 初めて進捗が付いた日時（一度入ったら不変、という運用前提。書込mergeの正は別軸）
   4. `firstCompletedAt` — 初めて完走扱いになった日時（未完走はnull。一度入ったら不変、という運用前提）
5. 全長 `durationSec` は原稿側の既存正本のまま進捗storeへ複製しない。
6. list向けGetが返す `progress` の意味は次とする。字段の正式schemaは契約artifactを正とし、ここへ写さない。
   1. 未play（行なし）→ `progress: null`
   2. 行あり → `positionSec` / `firstPlayedAt` / `firstCompletedAt`（null可） / `lastPlayedAt`
   3. 「完走したことがあるか」はAPIにbooleanを増やさず、UIが `firstCompletedAt != null` で導出する
7. UIへの写し方（進捗dataの見え方）は次とする。見た目の寸法・Icon字形の正本はUI実装側とする。
   1. 行上段の配信日（原稿の `date`）の右に、完走経験があれば印（✔️）を出す
   2. 進捗があるとき、play周り下段の日付は `lastPlayedAt` を `yyyy/mm/dd` に整形して出す
   3. `positionSec < durationSec` なら現在位置UI（barおよび `現在/全長`）。`positionSec >= durationSec` なら末尾停止UI（再再生を示すIcon＋全長）

## 2. Reason

1. 原稿json・mp3はGeneratorが書くほぼ不変のcontentである。userのlistening状態を同じobjectへ混在させると、書込主体・cache寿命・訂正手順が衝突し、contentと進捗の変更理由が1箇所に同居する（SRP / 直交の破綻）。
2. catalog master や list専用の集約jsonへ進捗を埋め込むと、日次のcontent更新と高頻度の進捗更新が同一表現を汚し、一覧cache方針（短いTTLのcatalog）と進捗の新鮮さ要求を同じknobで回せなくなる。
3. Access単一利用者・端末横断同一履歴という前提では、進捗のパーティション鍵にuser識別子は不要である。鍵を増やすとGet/Writeと将来の移行コストだけが増え、今の要件を満たす最短ではない。
4. いまlist UIが要るのは各episodeの最新カーソルと「触った／完走した」時刻であり、event列のfoldではない。1行状態ならlist joinが単純で、件数も日次+1のスケールに対して十分小さい。
5. 「完走したことがあるか」と「最後に再生したのはいつか」は別信号である。日付1つに畳むとどちらも読めなくなる。完走経験は上段配信日横の印、最終再生は下段の `lastPlayedAt` に分けると、content日付とlistening日付も混線しない。
6. 完走後に `positionSec` を0へ戻すと、UIのseek位置が末尾に残る体験と永続状態が食い違う。末尾のまま残し、`positionSec` と `durationSec` の比較で末尾停止と再聴中を導出する方が、永続と表示の意味が一致する。
7. Getに `completedOnce` booleanを足すと `firstCompletedAt` と同義の二重表現になる。null可否で足りる導出をAPI契約へ載せない。
8. `firstPlayedAt` を状態行に含めるのは、APIが今すぐ表示しなくても、初回playの事実はwrite時に確定し、後から列追加migrateするより「知っている事実を行に残す」方が安い。出さない字段をAPIへ載せる義務は生じない（APIはKISS、DB行は必要な事実を持てる）。

## 3. Rejected

1. **原稿json / mp3 / catalog master に進捗を書く案** — contentとuser状態のlifecycleが衝突し、Generator正本をuser操作で書き換えさせることになる。
2. **browserだけ（localStorage等）を進捗の正本にする案** — 端末横断の同一履歴という要件を満たせない。
3. **KVを進捗の正本にする案** — 最新1値の保存自体は可能だが、日時付き状態の一括読取・将来の時刻順参照・状態行の拡張をSQLで扱う方がD1に寄る。KV+D1の二重正本は今の要件に対して過剰である。
4. **event sourcingのみ（履歴appendをfoldして最新を都度導出）をlistの正本にする案** — listが要るのは最新状態であり、全event foldは読取コストと実装複雑さだけが増える。履歴表が要るのは観測・分析の読者が現れてからで足りる。
5. **完走後に `positionSec` を0へresetする案** — UIが末尾にseekを残す方針と矛盾し、「末尾停止」と「頭から再聴」の永続表現が曖昧になる。
6. **Getのprogressに `completedOnce` を併載する案** — `firstCompletedAt` からの導出と重複する。
7. **user鍵（Access email等）で行を分割する案** — 単一usersingleton前提では不要。前提が崩れたときに初めて鍵を足す。
