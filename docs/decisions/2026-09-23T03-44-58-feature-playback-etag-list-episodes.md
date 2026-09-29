---
name: 一覧GET（listEpisodes）にのみ hono/middleware/etag を導入する。音声GETには導入しない
date: 2026-09-23T03:44:58
branch: refactor/playback-worker-hono-design
---

## 1. Decision

`hono/middleware/etag`（`etag()`）を `listEpisodesPath` route にのみ導入する。
`episodeAudioRoutePath`（音声GET）には導入しない。

1. **`app.ts` の `.get(listEpisodesPath, ...)` chainへ `.use(etag())` を挟む。** cache期限
   （`episodeListCacheHeaders` の `max-age=60`）切れ後の revalidation で、内容が変わっていなければ
   `304 Not Modified`（bodyなし）を返せるようにする。
2. **音声GET（`episodeAudioRoutePath`）には導入しない。** 理由は §2 のコスト非対称性による。

## 2. Reason

`node_modules/hono/dist/middleware/etag/index.js` を読むと、この middleware は response body
全体を SHA-1 で digest 計算し ETag を生成する。この計算コストと、導入によって得られる効果は、
一覧JSON と音声mp3で非対称になる。

1. **CPUコストの非対称性。** 一覧JSON（数KB〜数十KB程度）の digest 計算は軽量だが、音声mp3
   （数MB規模）の digest 計算は Cloudflare Workers の CPU時間を有意に消費する。
2. **revalidation が起きる頻度の非対称性。** 一覧JSONは `max-age=60` と短命で、60秒ごとに
   頻繁に revalidation が起きる。音声mp3は `max-age=86400`（1日）と長命で、revalidation
   自体が発生する頻度が低い。導入コスト（毎回の digest 計算）に対し、304化による削減効果を
   得る機会が音声側では少ない。
3. **content変化頻度の非対称性。** 一覧JSONは新しい episode 追加のたびに内容が変わりうる
   （ETag が「実際に変わったか」を判定する意味を持つ）。音声mp3は一度アップロードされた
   episode の音声が後から書き換わることは想定しにくく、`Cache-Control` の長い max-age だけで
   十分に cache され、ETag の追加価値が薄い。

以上3点はいずれも「一覧側は導入コスト対効果が高く、音声側は低い」という同じ向きに揃うため、
route単位で適用を分けることが妥当。

## 3. Rejected

1. **両方の route に一律導入する案** — 音声mp3側の digest 計算コストが、得られる304化効果に
   見合わない。CPU時間はCloudflare Workersの課金・実行時間制限に直結するため、効果の薄い
   route にまでコストを払う理由がない。
2. **どちらにも導入しない案** — 一覧JSON側は revalidation 頻度が高く（60秒ごと）、304化による
   通信量削減効果が積み重なりやすい。導入コストが軽量（数KB〜数十KB程度のdigest計算）である
   ことを踏まえると、見送る理由がない。
