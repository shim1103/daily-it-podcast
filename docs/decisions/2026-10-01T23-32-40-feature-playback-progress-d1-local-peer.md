---
name: pullの差分印は時刻でなく、DBが書込のたびに同じ文で採番する単調増加の`seq`とし、契約はsinceでなく10進非負整数の文字列のcursorにして、文字列から数値への変換はrouteのzodが行う
date: 2026-10-01T23:32:40
branch: feature/playback-progress-d1-local-peer
---

## 1. Decision

1. pull の更新印は、時刻ではなく、DB が書込のたびに同じ文の中で採番する単調増加の整数 `seq` とする。`seq` には索引を張る。同じ `seq` は、行ごとの版（`2026-10-01T18-54-16-feature-playback-progress-d1-local-peer.md` の条件付き書込が使う期待値）も兼ねる。
2. pull の契約は、`since`（時刻）を `cursor` に置き換える。cursor は**10 進非負整数の文字列**で、形式（桁数の上限を含む）の正本は contracts の schema とする。「不透明」とは扱わない。
   1. 初回は cursor を省略し、全件を返す。以後は、応答の `cursor` をそのまま次回に渡す。省略を起点として扱うのは route の zod schema の既定値で、UseCase は cursor を必須で受ける。
   2. 変更が無い応答は、空の `episodes` と、要求と同じ cursor を返す（DB が空の初回は `"0"`）。
   3. cursor は応答の最上位にだけ出す。行の schema と list の embed には `seq` を出さない。
   4. 不正な cursor は 400（`validation_error`）とする。形式は route の zod schema が検証する。
   5. 文字列から数値への変換と省略時の起点の既定は、route の zod schema（`transform`・`default`）が行う。置き場の理由は `2026-08-25T18-42-00-chore-playback-worker-web-layer.md`（値の検証と変換は route の zod が担う）が正。数値から文字列への戻しは、Controller の UseCaseOutput → Response の写しが行う。UseCase・Port・adapter は cursor を非負整数（number）として扱い、変換しない。
3. `clientAt` を使う判断（`2026-09-22T19-13-35-feature-playback-progress.md`）は merge の鍵であり、pull の更新印とは別の問いとして扱う。
4. 現行の route・UseCase の差分鍵（`lastPlayedAt > since`）は、Rejected 2 の案のままである。本 Decision がそれを置き換える。epic の `2026-10-01T18-54-14-feature-playback-progress-application.md` §1.5 は本 Decision に合わせて直してある。移行の実施は別 Issue（#220）で行う（未実施）。
5. 書込量の増加（索引の分だけ、書込 1 回で書く行が増える）は、未確認の push 間隔に依る。数値はここへ載せず、未確認とする。

## 2. Reason

1. client 時刻を基準にすると、offline や retry で遅れて届く書込（`clientAt` が、他端末が持つ cursor より古い）が、次の pull の条件に当たらない。他端末は、その書込を取りこぼす。後勝ちの `lastPlayedAt` が動かず `first_completed_at` だけが動く変更も、client 時刻の列に差分として現れず、完走の事実を取りこぼす。
2. 時刻ベースの更新印は、どの時計（client・Worker・DB）で打っても、議論が残る。別 isolate の少し遅れた時刻による取りこぼし、同一ミリ秒の衝突、時計ずれである。`seq` は時計を使わず、DB の単一スレッドの上で厳密に単調になる。時刻の議論そのものが要らなくなる（`design-philosophy.md` §4-4 は非決定要素を境界の外へ押し出すことを求める）。
3. `seq` は、行ごとの版を兼ねる。時刻の印だと、版と差分印で別の列が要る。D1 のローカル実測（実 SQLite）で、同じ期待値の条件付き UPDATE を 10 本同時に走らせた結果、勝者はちょうど 1 本で、`seq` は全行で一意だった。本番 D1 での同実測は未確認である。
4. `seq` の採番は最大値を引くので、索引が無ければ全件走査になる。実測では、1000 行で索引なしが `rows_read` 1000、索引ありが 1 だった。pull（`seq > ?`）は定期的に走るので、索引は採番と pull の両方に効く。
5. cursor を 10 進文字列として契約に出し、内部は整数として扱うのは、実態が既に不透明ではないからである。契約の schema は形式を 10 進の桁数上限つきで固定していて、client だけが中身を解釈しないに過ぎない。それなのに Port と adapter まで不透明な文字列で通すと、形式の知識が contracts・D1 adapter・In-Memory の 3 箇所に散り、各 adapter が `String`／`Number` を持つ。検証と変換を境界に置けば（入力は route の zod、出力は Controller の写し）、変換は境界の 1 組だけで、adapter は `seq` をそのまま比較して返す薄い永続のままになる。桁数の上限は、数値へ変換しても精度が落ちない範囲に収まる。
6. 更新印の方式を後から変えられる自由は、10 進文字列という形式の範囲内でだけ残る。形式ごと変えるなら契約の変更（A）として扱う。この縮小は受け入れる。不透明を保つには、adapter が cursor を符号化・検証し、不正値を 400 へ写す経路を新設する必要があり、HTTP の公開 error code を増やさない方針（`2026-09-22T18-58-38-feature-playback-progress.md`）と釣り合わない。
7. UseCase が cursor を必須で受けるのは、省略可能という性質が HTTP の query の都合で、永続の能力ではないからである。起点の既定を route の zod schema に置けば、UseCase の入力に「省略」と「指定」の 2 状態が無く、入力から読める。
8. 初回を全件にするのは、list の契約（別機能）を変えずに済み、重複転送が session あたり 1 回で済むからである。`seq` を行や embed に出さないのも、同じ理由で list の契約を動かさないためである。
9. `clientAt` は「操作が起きた順」を表し、merge に必要な時刻である。pull の更新印が答えるのは「この cursor 以降に、server が何を受けたか」で、別の問いである。記録上、pull の更新印を client 時刻にした理由は無く、更新印の列が表に無かった現状の帰結だった。
10. 書込量は増える。索引は書込 1 回で書く行を 1 つ増やし、実測では索引ありの更新が `rows_written` 2（表と索引）だった。D1 の課金は rows written が rows read より桁違いに高い。ただし push の間隔は未決で、書込量の総量は未確認である。

## 3. Rejected

1. **全件 pull のみで、差分を持たない案** — 応答の payload と D1 の読取が、行数 × ポーリング回数に比例して増える。shim が却下した。初回だけ全件にするのは採用案に含まれ、これとは違う。ポーリング間隔・D1 の読取上限の数値は未確認で、ここへは載せない。
2. **client 時刻（`last_played_at`）を差分印にする案**（現行） — 遅れて届く書込と、`first_completed_at` だけが動く変更を、他端末が取りこぼす。
3. **Worker の時計の `updated_at` を差分印にする案**（旧答え） — Worker は複数の isolate で動き、時計が揃わない。別 isolate の少し遅れた時刻が、既に取った cursor より小さくなり、取りこぼす。
4. **DB の時計の `updated_at` を差分印にする案**（旧答え） — 同一ミリ秒に複数の書込が当たると、印が衝突する。行ごとの版（条件付き書込の期待値）にも向かない。
5. **cursor を契約上の JSON 数値にする案** — 契約の型を変える範囲が、A で並べた schema と web に及ぶ。内部で整数として扱うことと別の問いで、契約は 10 進文字列のままにする。以前は「契約が `seq` という内部の方式に縛られる」を却下理由にしていたが、契約の schema が既に 10 進の形式を固定しているので、その理由は根拠を失った。
6. **list の応答に cursor を足し、list から差分を始める案** — 別機能の契約を変える。必要になれば、後から移れる。
7. **cursor を Port と adapter まで不透明な文字列で通し、adapter が `String`／`Number` で変換する案**（旧答え） — 契約の schema が形式を固定していて、不透明になっていない。変換と形式の知識が contracts・D1 adapter・In-Memory に散る。不透明を保つ場合は、不正値を 400 へ写す経路の新設が要り、公開 error code を増やさない方針と釣り合わない。
8. **UseCase が cursor の省略を起点として扱う案**（旧答え） — 省略可能は HTTP の query の都合で、永続の能力ではない。UseCase の入力に「省略」と「指定」の 2 状態が残る。
9. **文字列から数値への変換を、Controller の Request → UseCaseInput で行う案**（旧答え） — 規約の標準形だが、route の zod が既に検証した値を Controller が再び解釈する段が増える。置き場の比較は `2026-08-25T18-42-00-chore-playback-worker-web-layer.md` が持つ。
