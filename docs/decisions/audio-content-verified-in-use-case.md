---
name: "取得した音声 byte の内容（空・mp3 の署名）は、どの層が保証するか"
description: "原稿と同じく、application の純関数 verifyAudio が UseCase で検査し、不適合は EpisodeContentError にする。Adapter は I/O の成功と応答の形状までを保証し、EpisodeAudio は Uint8Array の別名のままにする。"
---

## 1. Decision

1. 音声の内容の検査（空でないこと、mp3 の署名を持つこと）は、`application/audio/verify-audio.ts` の純関数 `verifyAudio` が行う。`getAudio` UseCase が、Port の生 byte に適用する。不適合は `EpisodeContentError`（原稿の不適合と同じ型）で、理由は message で区別する。
2. Port の `GetAudioResult` は、取得したままの生 byte（`Uint8Array | undefined`）とする。R2 と In-Memory の Adapter は、I/O の成功と応答の形状までを保証し、内容は判断しない。
3. 署名は、ID3v2 tag の先頭3 byte か、MPEG audio の frame sync（先頭 byte が `0xFF` で2 byte 目の上位3 bit が立つ）のどちらか。判定は先頭の byte だけで、再生可能性は保証しない。
4. `EpisodeAudio`（worker の `entities/`）は `Uint8Array` の別名のままにする。brand にしない。routes は `Uint8Array` を要求する。

## 2. Reason

1. 空の byte（0 byte）のオブジェクトは、検査するコードが無く、HTTP 200 の空の音声として返っていた。generator が置く配置契約を信頼するだけでは、この穴は塞がらない。
2. 内容の判定は platform に依らない domain 側の判断である。R2 Adapter に置くと In-Memory Adapter にも同じ判定が要り、複製になる。原稿の `verifyManuscript` は、両 Adapter が共有する application の純関数として前例がある。音声も同じ形にすれば、Adapter の方針（I/O だけ）が揃う。
3. 検査を先頭の署名に絞るのは、費用を抑えつつ、空や別形式の file を弾くには足りるからである。全 frame の走査や decode は、費用が大きい。署名の規則は MPEG の frame header の仕様についての記憶に基づき、一次資料では検証していない。実際の fixture（libmp3lame の無音 frame、先頭 `0xFF 0xE3`）が通ることだけを test で確かめた。
4. `EpisodeAudio` を brand にする根拠（`entities.md` §10-1 の「同じ不変が繰り返し破られた」証拠）が無い。検査の入口が `verifyAudio` に1つなので、後で brand に上げる時の差分も小さい。
5. routes は、`Response` の組立てと Range の切り出しに byte の view だけを使う。最小権限（`criteria.md` §2-5）で `Uint8Array` が足りる。routes は entities を import できず、wire の契約にも `EpisodeAudio` の概念が無い。

## 3. Rejected

1. **R2 Adapter の内部で内容を保証する案** — In-Memory Adapter に同じ判定が複製される。Adapter は I/O だけを担うという既存の方針（R2 の read Adapter）と衝突する。
2. **内容を検査しない案**（旧答え） — 空の byte が HTTP 200 で返る穴が残る。
3. **全 frame の走査や decode で再生可能性を保証する案** — 費用が大きく、署名の規則の確度も検証できていない段階では過剰である。再生できない音声が実害になってから足す。
4. **`EpisodeAudio` を brand にする案** — 上の Reason 4。
5. **routes に `EpisodeAudio` を要求させる案** — routes は entities を import できず、byte の view だけを使うので、型を強める利益が無い。
