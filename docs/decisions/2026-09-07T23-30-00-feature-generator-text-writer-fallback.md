---
name: Cursor create の枯渇シグナルに HTTP 400 + usage_limit_exceeded を加える
date: 2026-09-07T23:30:00
branch: feature/generator-text-writer-fallback
---

## 1. Decision

1. `cursorapi.TextWriter` の create（POST /v1/agents）が **HTTP 400 かつ応答 body に `usage_limit_exceeded` を含む**とき、その infra error を vendor 非依存の番兵 `port.ErrSourceExhausted` で wrap する。先行 Decision `2026-09-07T19-06-00` §4 が定めた 401/403 の分岐に、この 1 条件を加える形。判定は `bytes.Contains(raw, []byte("usage_limit_exceeded"))` で、Cursor の error JSON の構造（`error.code`）を型で parse しない。
2. それ以外の 400（body に `usage_limit_exceeded` を含まない）は従来どおり番兵で wrap せず、`cursorapi.Error`（op `create_status`）のまま `ProduceEpisode` へ伝播させて run を赤にする。429 / 5xx / SSE 途中断 / 到達不可（`Op=="do"`）も従来どおり。
3. UseCase（`manuscript.TextWriter`）と Composition は変更しない。切り替え条件が「番兵 `port.ErrSourceExhausted`」であることは不変で、その番兵を返す条件が cursorapi 内で 1 つ増えるだけ。切り替えは高々 1 回、secondary は現行 Flash-Lite（`geminiapi.ModelID`）のまま。

先行 Decision `2026-09-07T19-06-00` を supersede しない。§4 の「Cursor が確実に使えないときだけ退避する」という趣旨は維持し、その「確実に使えない」の観測点を 401/403 に加えて「400 + `usage_limit_exceeded`」へ拡張する。先行 Decision の file 本文は書き換えない。

non-scope: `usage_limit_exceeded` 以外の 400 error code の分類（`billing_*` 等が将来現れたら都度この Decision を継ぐ）。Cursor の usage-based pricing を有効化する運用判断。Gemini 原稿の品質・token・尺の実測（先行 Decision の non-scope のまま）。

## 2. Reason

### なぜ 400 + `usage_limit_exceeded` を枯渇シグナルに加えるか

先行 Decision `2026-09-07T19-06-00` §4 は「subscription / Pro / key の失効はすべて create の 401/403 として現れる（`2026-09-04T15-05-00` の daily で 401 を実証）」を根拠に 401/403 限定とした。その後 System e2e（`generator-system.yml` run 34132953055）で、**Background Agent の利用枠が尽きたケースは 400** で返ることが判明した。body は

```
{"error":{"code":"usage_limit_exceeded","message":"Usage-based pricing required. Background Agent requires at least $2 remaining until your hard limit. ..."}}
```

で、「原稿を Cursor から取得できない」という点で 401/403 と同じ状態である。ここで退避しないと、利用枠が尽きている間は毎日 System / 本番 produce が `cursorapi: create_status: 400` で赤くなり、fallback を用意した意味（Cursor が使えなくても原稿を出し続ける）が失われる。よって 401/403 と同じ扱いにする。

### なぜ 400 全部ではなく `usage_limit_exceeded` を含むものだけか

先行 Decision `2026-09-07T19-06-00` §Rejected 3 と同じ懸念。400 は request が malformed（model 名の版落ち・スキーマ変更・no-repo の形の崩れ）でも返る。400 を丸ごと番兵にすると、Cursor 側の request が壊れているだけの日も毎回 Gemini 原稿へ落ち、primary の品質を捨てる頻度が上がる。`usage_limit_exceeded` という Cursor 固有の code 文字列を 1 つ見るだけなら、malformed request の 400 は従来どおり赤で止まり、枠喪失の 400 だけが退避へ回る。

### なぜ error JSON を型で parse しないか

先行 Decision `2026-09-07T19-06-00` §Reason「なぜ番兵 error か」と同型。cursorapi の error 応答構造（`error.code` のネスト）に依存すると、Cursor が応答形を変えるたびに parse を直す羽目になる。`bytes.Contains` で code 文字列の存在だけを見れば、`error.code` が `error.reason` へ改名されても `usage_limit_exceeded` という値が body のどこかにある限り拾える。この判定は「枯渇を知っている cursorapi 自身」に閉じ、UseCase にも Composition にも漏れない。

### なぜ Decision を分けたか（`2026-09-07T19-06-00` へ追記しない）

先行 Decision は「fallback の設計（層・番兵・model・retry）」を丸ごと固定した 1 判断単位。今回は「その番兵を返す条件を 1 つ足す」という後続の小さな判断で、実測（run 34132953055）を根拠に持つ。先行 Decision 本文を書き換えると「いつ何を根拠に条件が増えたか」が消えるため、`decisions.md` の Decision Record の粒度に従い別 file として継ぐ。

## 3. Rejected

1. **400 を丸ごと番兵で wrap する案** — malformed request（model 版落ち・スキーマ変更）の 400 まで退避対象になり、Cursor の request が壊れているだけの日も Gemini 原稿へ落ちる。`usage_limit_exceeded` の文字列 1 つを見れば枠喪失だけを拾える。
2. **Cursor error JSON を struct で parse し `error.code == "usage_limit_exceeded"` を厳密判定する案** — cursorapi が応答形を変えるたびに parse を直す。`bytes.Contains` なら code の値さえ body に残っていれば拾え、構造変更に強い。
3. **先行 Decision `2026-09-07T19-06-00` §4 を「400/401/403」へ書き換える案** — いつ何を根拠に 400 が加わったかが消える。実測（run 34132953055）を持つ後続判断として別 file で継ぐ。
4. **System test 側で 400 `usage_limit_exceeded` を PASS 扱いにする案（`no_source_items` と同型）** — System は「壊れていないか」を測るもので、fallback が設計どおり Gemini へ切り替わって Drive へ到達するなら緑であるべき。400 を test 側で握り潰すと fallback 経路が System で一度も検証されない。
5. **Cursor の usage-based pricing を有効化して 400 自体を消す案** — 「完全無料枠で原稿を出し続ける」要件（先行 Decision §Reason）に反する。枠が尽きたら Gemini へ退避するのが本 fallback の目的。
