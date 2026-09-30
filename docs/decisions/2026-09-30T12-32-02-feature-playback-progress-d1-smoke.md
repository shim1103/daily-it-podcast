---
name: 進捗D1の一過性疎通は共有 config の binding 面で往復し、独立な remote 経路で読み戻して local の模擬でないことを確かめる
date: 2026-09-30T12:32:02
branch: feature/playback-progress-d1-smoke
---

## 1. Decision

1. 疎通は、共有 config（`test/support/wrangler.smoke.jsonc`）の D1 binding（`EPISODE_PROGRESS`）を `getPlatformProxy` の remote binding で取り、adapter が呼ぶ最小面（`prepare` / `bind` / `first` / `run`）で、書く・読む・消す・消えたことを読む、を工程ごとに独立した case で通す。`wrangler d1 execute` の CLI 実行だけで済ませない。
2. 同じ疎通で、binding で書いた行が、binding とは独立な remote 経路（wrangler CLI の `--remote`）でも見えることを確かめる。binding が実 Cloudflare でなく local の模擬に落ちていれば、ここで失敗する。専用 config は使わない。
3. 同じ判定を R2 の `EPISODES` binding にも、同じ dispatch で適用する。
4. 共有 config の D1・R2 の entry には `remote: true` を明記する。疎通が local の模擬への落ち込みを Red として観測したため、足した。常設 `playback-smoke` も同じ config を使うので、R2 は実 TEST bucket へ届く形になる。
5. test 固定・PASS 後削除・merge 意味を見ない点は、R2 疎通の Decision（`2026-09-16T10-45-14-feature-r2-smoke-migrate-cutover` / `2026-09-16T13-46-41-feature-r2-smoke-remove-and-migrate`）と同じ答えとし、本 file へ写さない。表の準備は別 Decision（`2026-09-30T13-30-00-feature-playback-progress-d1-smoke`）が持つ。

## 2. Reason

1. CLI は SQL を直接流すだけで、adapter が実際に呼ぶ面（先行 Decision `2026-09-23T08-06-58-feature-playback-progress` §1-2 が A で固定した `prepare` / `bind` / `first` / `run`）を通らない。CLI が緑でも、binding 経由の到達や面の不一致は見逃す。
2. wrangler は、local で simulate できる binding が remote resource へ届くには、binding 定義に `remote: true` が要ると警告する（wrangler 4.127.0）。共有 config の D1・R2 entry には `remote` の指定が無かった。local の模擬では、D1 は空 DB で表が無く write が落ちるが、R2 は空 bucket でも put → get が成功して緑になる。binding 経由の往復だけでは「実物へ届いた」と言えない。独立な remote 読み取りだけが、届いたことを保証する。
3. 実 dispatch（GitHub Actions run 36673616434）で実測した。TEST D1 へ migration を remote で適用した直後でも、binding 経由の書き込みは `no such table: episode_progress` で落ち、binding は local の模擬を見ていた。R2 は、binding で put した object が `--remote` の読み取りで存在せず、put は local の模擬に入っていた。読解が実測で裏付けられたため、`remote: true` を足した。
4. 専用 config で `remote: true` を明記して疎通すると、検査したい共有 config の欠陥を隠す。疎通は、常設 smoke が実際に使う config を通して初めて、その config の正しさを確かめられる。
5. 書く・読む・消す・消えたことを読む、を1 case に束ねると、失敗した工程が case 名から分からない。工程ごとに独立した case にすれば、名前で失敗箇所を特定できる。前提は各 case の Given で整え、case 同士の実行順に依存させない。

## 3. Rejected

1. **CLI（`wrangler d1 execute --remote`）だけで疎通を見る案** — adapter が呼ぶ面を通らず、結線ミスが緑で通る。
2. **専用 config で `remote: true` を明記して疎通する案** — 共有 config の `remote` 欠落を隠し、常設 smoke が実物へ届いているかを確かめられない。
3. **binding 経由の往復だけで remote 到達を確認したことにする案** — R2 は local の模擬でも put → get が通り、緑のまま実物へ届いていない状態を見逃す。
4. **共有 config へ先に `remote: true` を足してから疎通する案** — 足す前に Red を観測しないと、足す必要があったのか分からず、常設 smoke の挙動を根拠なく変える。Red が出た時だけ足す。
