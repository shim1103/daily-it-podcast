---
name: web API Clientはstatusを単一情報源にし、契約codeをweb語彙へ1対1で変換する
date: 2026-08-20T13:44:08
branch: playback-web-api-client
---

## 1. Decision

1. HTTP境界の失敗は `HTTP-error(status, code)` → `API-error(code)` → `UI-error(表示文)` の3段で扱う。web API Clientが持つのは中段だけとする
2. API-errorは応答の `status` だけを情報源にする。失敗bodyの `code` はparseしない
3. 契約 `PlaybackHttpErrorCode` をweb語彙として素通しせず、同名でも別layerの型として宣言し、写像表で1対1に変換する
4. `fetchAudio` の成功dataは `Blob` とする。web側は音声形式を知らない
5. baseUrlはfactory `createPlaybackApiClient` のdepsで1度だけ受け取り、各methodの引数にしない
6. TS test fileは `*.{Scope}_{Sociability}.test.ts` 形式で分類を名前に持たせる
7. 契約に無いstatus分類が届いた時、`toApiErrorCode` は既存codeへ倒さずthrowする

## 2. Reason

1. 表示文への変換はDelivery Mechanismの責務であり、ViewModelは境界共有型をimportできない。API-errorをwebに閉じた語彙として持たないと、契約型がUIまで漏れる
2. workerは `status` と `code` を1対1で返すため、bodyを読んでも新しい情報が無い。二重の情報源はどちらが正かの規則を要求する
3. 素通しは契約enumの改名・意味変更をweb全体へ直接伝播させる。写像表を1箇所に置くと、契約が増えた時にcompileが変換の追加を強制する
4. `Response.blob()` はContent-Typeを保持するため、web側が音声形式のliteralを書かずに済む。`ArrayBuffer` だと下流がMIMEを与える必要が出る
5. baseUrlは隠すべき具体値であり、各methodが受け取ると同じ知識が複数箇所へ散る。factoryに閉じるとURL正規化も構造的に1箇所へ集まる
6. Scope × Sociabilityがfile名から判別できることは分類規約の必須条件で、Go側は既に満たしていた。runnerの収集条件も分類名で絞り、命名忘れを検出可能にする
7. callerはcodeごとにretry・再入力・報告を選ぶ。未知の分類を既存codeへ倒すと、起きていない失敗を伝えて誤った操作を選ばせる。想定外の前提違反はResultではなくthrowで大域脱出させ、公開methodがcatchして`invalid_response`へ落とす

## 3. Rejected

1. 失敗bodyの `code` をparseして採用する案。statusと二重の情報源になり、不一致時の優先規則が新たに必要になる
2. 契約 `PlaybackHttpErrorCode` をweb語彙として直接再exportする案。層の境界が消え、契約変更の影響範囲がweb全体へ広がる
3. `fetchAudio` が `ArrayBuffer | Uint8Array` を返す案。呼び出し側に分岐を強制し、同じ目的の値が複数の型で表現される
4. Content-Typeを検証して不一致を失敗にする案。API Clientの責務はstatus確認・parse・schema検証であり、同一repo内workerの契約遵守は前提にできる
5. baseUrlを各methodの引数に置く案。callerが同じ値を持ち回り、URL正規化がmethodの数だけ重複する
6. 未知statusの分類logicをClient側に再実装する案。分類は契約の責務であり、Client側でassertすると障害時に原因unitを特定できない
7. 未知の分類を`unavailable`へ倒す案。戻り型は守れるが、worker正常時でも障害と誤報し、callerのretry判断を誤らせる
