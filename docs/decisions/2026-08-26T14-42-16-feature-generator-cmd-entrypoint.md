---
name: cmd/generator の成否観測は OS process の exit code と stderr を正とする
date: 2026-08-26T14:42:16
branch: feature/generator-cmd-entrypoint
---

## 1. Decision

1. `apps/generator/cmd/generator` の成否を caller（local shell / GHA runner）が観測する正本は、**OS process の exit code** と **stderr への error 文言**である。
2. Go **runtime** は `main` の起動と通常終了（成功時 exit 0）を担う。**`os` package** は OS process 境界（`os.Exit`・`os.Stderr`・signal）への窓口である。成否の契約は runtime 内部状態ではなく OS 観測面に置く。
3. 失敗写像は薄い: UseCase の non-nil error → stderr へ1行 → `os.Exit` 非0。Domain / Infra の分類や exit code の細分化は cmd に持たない（入口責務の正は `docs/decisions/2026-08-25T23-12-31-feature-generator-cmd-usecase-boundary.md`）。
4. **local と GHA で binary・写像は同一**。違うのは parent（shell vs Actions runner）が exit / stderr をどこで読むかだけである。

## 2. Reason

1. CLI / GHA の契約は「process が何で終わったか」である。Go runtime の goroutine 状態や panic stack を正本にすると、caller ごとに解釈が割れ、薄い Driving Adapter の責務（wire + 終了）を超える。
2. `os.Exit` は OS に exit status を渡す。成功時に `os.Exit(0)` を書かず `main` の自然終了に任せるのは、runtime の通常終了経路を使い、失敗時だけ明示的に非0へ写すためである。
3. stderr は「人が読む失敗」、stdout は将来の機械出力余地を汚さない側、という Unix 慣習に合わせる。local の terminal と GHA の job log はどちらも process の stderr を受け取る。
4. signal 付き `context` は OS が process へ届ける Interrupt / SIGTERM を cancel に写す。local の Ctrl+C と GHA の cancel / timeout が同じ入口契約に乗る。

## 3. Rejected

1. 失敗を stdout だけに出し exit 0 のままにする案 — GHA step が成功扱いになる。観測面が壊れる。
2. Domain / Infra ごとに exit code を細分化する案 — cmd に分類責務が入り、薄い入口（B-cmd）を厚くする。必要なら別 Issue。
3. `ProduceEpisode.Run` の panic stub を recover して成功扱いにする案 — 失敗を隠す。panic は error→stderr→非0 経路とは別物とし、Run 本体は D のまま（既存 Rejected と同旨）。
4. local 用と GHA 用で入口の写像を分ける案 — 同一 binary で足りるのに分岐が増え、Orthogonality が壊れる。
