---
name: 進捗WriteのHTTP失敗は既存codeのまま。行なしupdateは404、重複createは冪等200
date: 2026-09-22T18:58:38
branch: docs/playback-audio-history
---

## 1. Decision

1. 進捗 Write / pull の失敗を表す `PlaybackHttpErrorCode` は増やさない。既存の External Error 4 種と status 対応（`ValidationError`→400 `validation_error`、`NotFoundError`→404 `episode_not_found`、`ConfigurationError`→500、`UnavailableError`→503）へ畳む。
2. update（PATCH）で進捗行が無いときは **404**（Domain の行不在 → External `NotFoundError`）。契約の Error 型・写像の正本は worker の Domain Error と `mapInternalErrorToExternal` とし、ここへ写さない。
3. create（POST）で既に進捗行があるときは Error にせず **冪等に成功（200）** とし、勝ち側の `first*` を返す。先勝ち／後勝ちの merge と両立させる。

## 2. Reason

1. 進捗同期失敗は user 向け通知を出さない（先行 Decision `2026-09-19T19-12-30-docs-playback-audio-history.md`）。新しい code を足しても UI 分岐が増えず、web の code 写像と status 表だけが膨らむ。形の不正も意味ルール違反（skew 等）も、client が直せる／再送しても無駄、という区別は現状の補助同期では status 400 系に畳んで足りる。
2. 行なし update は「その進捗資源がまだ無い」状態である。400（操作順不正）より 404（資源不在）の方が HTTP の読みと一致し、無限 retry しても成功しないことが status から分かる。
3. 重複 create を 409 等で落とすと、再送や複数端末の初回 play が失敗扱いになり、先勝ち merge の前提（後から来た正当な時刻も受理しうる）と衝突する。既行への create を成功扱いにすれば再送が安全になる。

## 3. Rejected

1. **422 や進捗専用 HTTP code を新設する案** — user 通知も retry 方針の分岐も今は増えない。既存 4 code と web 写像を壊すだけのコストになる。
2. **行なし update を 400 `validation_error` にする案** — 資源不在と入力形不正が同じ code に混ざり、観測と将来の扱い分けがしにくい。
3. **重複 create を 409 / 専用 code で拒否する案** — 冪等再送と複数端末初回が壊れ、merge 方針と合わない。
