---
name: Gemini 以外の実境界を実 secret で通す System test を full run とは別に持つ
date: 2026-09-02T16:57:00
branch: feature/generator-system-e2e-produce-episode
---

## 1. Decision

1. `apps/generator/test/system/` に **「Gemini 以外 full」System test** を新規で置く（build tag `//go:build system`）。`ProduceEpisode.Run` を、GetX / Cursor CLI / OAuth / Google Drive は **実 secret の production Adapter**、`SpeechSynthesizer` だけ **無音 WAV を返す test fake** に差し替えて 1 回通す。
2. 検証対象は次の実到達である。
   1. GetX API 実 fetch が Fetch 窓内で SourceItem ≥1 を返す。
   2. Cursor CLI 実 draft が `ManuscriptDraftFromWriterOutput` を通る JSON を返す。
   3. OAuth refresh→access token が実際に発行される。
   4. Drive へ `{episodeId}.json` → `{episodeId}.wav` を実書込し、実読取で JSON の episodeId が stem と一致し、本 run の成果物を実削除できる。
3. 各実境界の所要時間を段階ログ（`t.Logf`）で出す。GetX fetch / Cursor draft / Drive 書込＋読取＋削除それぞれの経過を測る。
4. 同日完成ペア skip（先行 Decision `2026-08-30T23-30-00`）に当たると新規 produce にならないので、test は run 前に test Drive folder の当日ペアを掃除するか、skip 発生時に `t.Skip` ではなく明示 fail で「当日ペアが残っている」と知らせる。
5. Gemini TTS の実到達は本 test の対象外。正は TTS 単体 test（先行 Decision `2026-09-02T13-57-00`）と full run（`system && full`）。

## 2. Reason

1. Gemini 無料枠 RPD=15 が緑化 blocker で、有料枠移行を待つ間も「Gemini 以外は完璧か」を確かめたい。full run（`system && full`）は毎回 Gemini を十数回焼くので、有料枠前に回せない。
2. GetX / Cursor / OAuth / Drive はそれぞれ Narrow Integration を持つが、Narrow は fake upstream（`httptest` + `DialTLSContext` redirect）で、実 secret・実 API の到達も所要時間も測れない。特に Drive 書込（create + upload media）は過去 System run が全て TTS 手前で落ちたため一度も実到達していない。
3. `ProduceEpisode.NewProduceEpisode` は `SpeechSynthesizer` を引数で受けるので、`_test.go` 側で production Adapter を結線しつつ speech だけ fake に差し替えられる。Composition に test 用の分岐を足す必要がない。
4. 無音 WAV fake で TTS を 0 リクエストにすれば、Gemini の quota を一切消費せず全経路の配線・実 I/O・実削除まで 1 回で通せる。TTS 部分の実到達だけは分業（TTS 単体 test）で担保済み。

## 3. Rejected

1. TTS を実 API で 1 本だけ焼く案 — 無料枠 RPD=15 を 1 消費する。shim は「GeminiKey 以外」を確かめたいので Gemini は 0 リクエストにする。
2. Drive 書込だけの独立 System test にする案 — GetX / Cursor の実到達と所要時間が取れない。目的は「Gemini 以外すべて」。
3. full run（`system && full`）の Gemini を毎回 fake にして 1 本化する案 — full run は「入口から出口まで実物」の到達 test という役割なので、実物でない speech を混ぜない。別 test として分ける。
4. Composition に speech override の口を足す案 — production Composition に test 専用の seam を持ち込む。`_test.go` 側の結線で足りる。
