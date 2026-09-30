---
name: e2e の storageState は人が組み立てて登録し、repo は JSON の雛形と期限確認だけを持つ
date: 2026-09-30T18:07:49
branch: feature/playback-e2e-session-recovery
---

## 1. Decision

1. Secret `PLAYWRIGHT_STORAGE_STATE_JSON` の JSON は、人が雛形から組み立て、`gh secret set` か GitHub の GUI で登録する。登録用の script と、JWT から JSON を組み立てる code は持たない。
2. repo は、JSON の構造の雛形 `apps/playback/cli/access-session/storage-state.example.json` を持つ。cookie の値と domain は placeholder だけで、実値を置かない。
3. 期限確認（`check`）は残す。e2e の最初の step が、Secret の期限を読み、失効なら失敗させる。

## 2. Reason

1. 登録は、session の失効ごとにしか起きない人手の作業である。組み立ての code は、JWT の `exp` から `expires` を作り、host から `domain` を作るなど、Access が cookie を受理する条件を code が肩代わりする。受理されない時、原因が cookie そのものか、code の変換かを、切り分ける対象が増える（この session では、Access が期限内の cookie を受理しなかった。原因は未確認）。
2. 人が JSON を直接見て組み立てれば、登録するものと、Playwright が読むものが同じになる。構造は雛形が示すので、手順のたびに Playwright の仕様を調べ直さなくて済む。
3. 期限の確認は、Secret の中身を読むだけで決まる（副作用も、Access への通信も無い）。CI の最初の step に置けば、期限切れと、Access の不受理を、E2E の失敗の前に分けられる。

## 3. Rejected

1. **登録 script を残し、不安定な部分だけ直す案** — 組み立ての変換を code が持ち続ける限り、原因の切り分け対象が減らない。登録の頻度に対して、script の保守と test の量が見合わない。
2. **雛形を置かず、`DEPLOY.md` の文章だけで構造を伝える案** — JSON の構造は文章より実物のほうが誤りにくい。雛形を test で `check` に通せば、構造が壊れていないことも保てる。
3. **`check` も消し、失効の判定を E2E の失敗画面だけに任せる案** — 期限切れと Access の不受理が、どちらも login 画面に見える。最初の step が期限を先に言えば、切り分けの手数が減る。
