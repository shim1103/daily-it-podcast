---
name: R2 の client 口と test peer は app×Scope で写像する（generator=S3、playback=binding、NI 既定は httptest）
date: 2026-09-14T19:29:08
branch: feature/generator-r2-write-adapter
---

## 1. Decision

1. 本番の client 口は先行 Decision `2026-09-14T11-04-30` を正とする。generator Infra は **S3 互換 HTTPS + SigV4**。playback は **Workers R2 binding**。本 file は契約値を写さない。
2. **Sociable Unit**: 同一 process 内。外部依存は `RoundTripper` 等の double。wrangler 起動や実 network peer は置かない。
3. **Narrow Integration** の既定 peer は、現行どおり controllable な `httptest`（実 `*http.Client` / TLS）。**wrangler experimental local S3** は NI の任意強化 peer（署名受理を peer 側で見る）に限り、gate 必須にしない。
4. **Broad Integration** は peer を wrangler local S3 へ上げない。BI が所有するのは合成関係であり、NI が所有する境界 I/O 詳細を再 assert しない（testing-strategy）。
5. **Workers local binding** は playback Adapter とその test の話に閉じる。generator の Writer / `CompletedEpisodeLookup` には binding を持たせない。
6. generator の R2 `CompletedEpisodeLookup` も Writer と同じ **S3 API + 同 credential 形**とする。本番の Compose 結線は列 6 で Writer と **同着**する。手順・Verification の百科は Issue / 列 6 が持ち、本 file には書かない。

## 2. Reason

1. generator は GHA 上の Go CLI、playback は Workers なので、認証注入が非対称になるのは `11-04-30` と同じ必然である。test もその口に揃えないと、本番経路と違うものを緑にして安心する。
2. SU に wrangler を入れると Fault Isolation が壊れ、失敗原因が Adapter か local runtime か判別しにくくなる。同一 process の分岐表は SU、境界機械の I/O は NI、という testing-strategy の分離を保つ。
3. NI の主張は「標準 client が実境界 provider を通ること」である。`httptest` はその provider を controllable peer で満たす既定手段として既に Writer Narrow で使っている。wrangler experimental local S3 は SigV4 受理まで peer が検証できる点で近いが、起動 cost・CI 適性・path-style 対応が別変数なので、既定を置換せず任意強化に留める。
4. BI の HTTP 線は NI と似て見えるが、所有する断言が違う。peer を wrangler に替えても合成関係の検出力は増えず、NI の詳細再 assert になりやすい（minimization）。
5. binding は Workers runtime の注入面である。generator が binding を持つと S3 credential 経路と二重正本になり、`11-04-30` の非対称設計が崩れる。
6. Lookup と Writer を別 credential / 別切替タイミングにすると、不完全ペア判定と書込先がすれ違い、列 6 の同着検証が無意味になる。

## 3. Rejected

1. **BI の storage peer を wrangler local S3 必須にする案** — BI の所有外の I/O 詳細を増やし、NI と二重になる。
2. **SU で wrangler / 実 network を起動する案** — 原因特定が難しく、FIRST の Fast / Repeatable を損なう。
3. **generator に Workers R2 binding を持たせる案** — CLI が Workers runtime に依存し、playback との口が逆転する。
4. **NI 既定 peer を wrangler に置換し `httptest` を捨てる案** — 未実測の experimental 起動を gate の単一障害点にする。強化は任意のまま残す。
5. **Lookup だけ先に本番結線し Writer は Drive のままにする案** — 照会先と書込先が割れ、日次スキップ判定が誤る。
