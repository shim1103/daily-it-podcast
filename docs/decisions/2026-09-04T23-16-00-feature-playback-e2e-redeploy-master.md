---
name: 音声配信は HTTP Range request に応える。任意位置 seek はこれ無しでは効かない
date: 2026-09-04T23:16:00
branch: feature/playback-e2e-redeploy-master
---

## 1. Decision

`apps/playback/worker/src/routes/audio-response.ts`（`createAudioResponse`）は、request の `Range` header を解釈し、`bytes=N-M` / `bytes=N-` を `206 Partial Content` + `Content-Range` + `Content-Length` で返す。開始位置が総 byte 数以上なら `416 Range Not Satisfiable`。header が無い、または解釈できない形式なら Range を無視し従来どおり `200` で全体を返す。応答は常に `Accept-Ranges: bytes` を持つ。

## 2. Reason

1. `<audio>` 要素は、配信元が `Accept-Ranges: bytes` を返さない（＝ Range request に応えない）と判断すると、`currentTime` への代入を「seek 不可能」として無視する。実機の Chrome devtools で `seeked` event を観測したところ、`currentTime` 代入後に `seeking`→`seeked` は発火するが、`seeked` 時点の `currentTime` が代入値ではなく `0` のままだった。任意位置 seek は JS 側の実装（`readyState` を待つ・`seeked` を待つ等）をどれだけ正しく書いても、配信側が Range に応えない限り効かない。
2. 本番の音声は 24MB 前後の非圧縮 wav（PCM）で、全体配信のダウンロードには数秒かかる。Range 対応が無いと、topic の sec bar を押すたびに実質「先頭から聴き直す」挙動になり、seek 機能そのものが成立しない。
3. `416` を明示するのは RFC 7233 準拠のためで、`Content-Range: bytes */total` を返すことでクライアントに実際のリソースサイズを伝える。逆順 range（`bytes=5-2` 等）は仕様上曖昧なため、開始位置に丸めて `206` で応える（クライアントの奇妙な入力で配信を止めない）。

## 3. Rejected

1. 音声を圧縮フォーマット（opus/aac 等）へ変換しファイルサイズを下げる — ファイルサイズを減らしても Range 非対応である限り「全体を毎回ダウンロードしてから seek」という構造は変わらず、根本原因を残す。将来ファイルがさらに大きくなれば同じ症状が再発する。
2. frontend 側で「seek 位置を含む range だけ fetch して blob URL を再構成する」独自実装をする — HTTP の標準機構（Range request）で解決できる問題を独自実装で肩代わりする理由がない。Rule of Least Power に反する。
3. E2E の timeout を伸ばして症状を隠す — CI の帯域が細いから遅いのではなく、Range 非対応で「常に全体を先頭から取得し直す」設計自体が遅い。本番ユーザーの体験（seek 反応の遅さ）も同じ原因で悪化しており、timeout を伸ばしても直らない。
