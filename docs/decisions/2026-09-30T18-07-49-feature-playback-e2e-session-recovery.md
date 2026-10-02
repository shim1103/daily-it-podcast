---
name: e2e の storageState は人が組み立てて登録し、repo は JSON の雛形・期限確認・expires の変換だけを持つ
date: 2026-09-30T18:07:49
branch: feature/playback-e2e-session-recovery
---

## 1. Decision

1. Secret `PLAYWRIGHT_STORAGE_STATE_JSON` の JSON は、人が雛形から組み立て、`gh secret set` か GitHub の GUI で登録する。登録用の script と、JWT から JSON を組み立てる code は持たない。
2. repo は、JSON の構造の雛形 `apps/playback/cli/access-session/storage-state.example.json` を持つ。cookie の値と domain は placeholder だけで、実値を置かない。
3. 期限確認（`check`）は残す。e2e の最初の step が、Secret の期限を読み、失効なら失敗させる。
4. 人が変換する項目は無い。`expires` には DevTools の日時をそのまま貼り、Unix 秒への変換は、E2E を実行する step が `normalize` で行う。変換するのは Access cookie の `expires` だけで、`value` と `domain` には触れない。

## 2. Reason

1. 登録は、session の失効ごとにしか起きない人手の作業である。組み立ての code は、JWT の `exp` から `expires` を作り、host から `domain` を作るなど、Access が cookie を受理する条件を code が肩代わりする。受理されない時、原因が cookie そのものか、code の変換かを、切り分ける対象が増える（この session では、Access が期限内の cookie を受理しなかった。原因は未確認）。
2. 人が JSON を直接見て組み立てれば、登録するものと、Playwright が読むものが同じになる。構造は雛形が示すので、手順のたびに Playwright の仕様を調べ直さなくて済む。
3. 期限の確認は、Secret の中身を読むだけで決まる（副作用も、Access への通信も無い）。CI の最初の step に置けば、期限切れと、Access の不受理を、E2E の失敗の前に分けられる。
4. `expires` を人が Unix 秒へ変換する運用は、実 run で破れた（雛形の `0` が置き換わらないまま登録され、期限が 1970 年の失効になった）。変換は 1 項目・純関数で、test で保てる。JWT から複数項目を作る組み立て（Reason 1）と違い、Access が受理する条件には関わらない。人の手から外せば、変換の誤りがなくなる。

## 3. Rejected

1. **登録 script を残し、不安定な部分だけ直す案** — 組み立ての変換を code が持ち続ける限り、原因の切り分け対象が減らない。登録の頻度に対して、script の保守と test の量が見合わない。
2. **雛形を置かず、`DEPLOY.md` の文章だけで構造を伝える案** — JSON の構造は文章より実物のほうが誤りにくい。雛形を test で `check` に通せば、構造が壊れていないことも保てる。
3. **`check` も消し、失効の判定を E2E の失敗画面だけに任せる案** — 期限切れと Access の不受理が、どちらも login 画面に見える。最初の step が期限を先に言えば、切り分けの手数が減る。
4. **`expires` も人が Unix 秒へ変換する案** — 変換の command を調べ、日時の形（小数秒・`Z`）に合わせる手間があり、置き換え忘れも起きた。人が貼るだけで済む形にできる。
