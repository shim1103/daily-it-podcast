---
name: playback worker fetch.ts は ArrayBuffer 正規化を境界で確定する
date: 2026-08-18T17:38:00
branch: feature/playback-worker-http
---

## 1. Decision

1. 音声 byte（`Uint8Array`）を `fetch.ts` の **Route boundary** で `ArrayBuffer` に正規化して `Response` に渡す。
2. `Uint8Array.buffer` は TS 上 `ArrayBuffer | SharedArrayBuffer` を含み得るため、`Response` の body 契約を満たす型を実行経路で確定する。
3. 正規化処理は Route の外向き責務（wiring/HTTP返却）を崩さない範囲で helper に退避し、`fetch.ts` の肥大を避ける。

## 2. Reason

1. 型不一致は境界（`Response` body）で顕在化し、実行前に検出できるべき。
2. `SharedArrayBuffer` 由来でも `ArrayBuffer` を返すことで、型と観測可能な body の両方を契約に揃える。
3. audio body composition の詳細は routing / error mapping / logging へ混ざるべきでないため、Route boundary 内に閉じる。

## 3. Rejected

1. `Uint8Array` のまま `Response` に渡して runtime / 型の許容差に賭ける。
2. 正規化を Controller / UseCase 側へ移し、HTTP境界を曖昧化する。
3. `ArrayBuffer` 化と `Response` 作成を Route へ散在させ、DRY と責務境界を崩す。

