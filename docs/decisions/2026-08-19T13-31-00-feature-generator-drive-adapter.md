---
name: Drive Adapter は contracts の manuscript.schema.json を直接 import する
date: 2026-08-19T13:31:00
branch: feature/generator-drive-adapter
---

## Status

**Superseded** by `2026-08-19T15-00-00-feature-generator-drive-adapter-layer-split.md`。generator の Drive **保存** Adapter は schema を import しない。`go:embed` + snapshot 禁止は Application 側の schema 読み取りで継続する。

## 1. Decision

generator の Drive Adapter は repo 根 `contracts/manuscript.schema.json` を直接 import して検証する。Adapter 隣の schema snapshot と byte 一致 Unit は置かない。Go は `contracts` package が JSON を `go:embed` し、byte 列だけを公開する。jsonschema engine は Adapter に残す。

本判断は `2026-08-19T13-00-00` の「読み手は Drive Adapter」を維持し、「Go package にしない / snapshot を置く」を上書きする。読み手の表は `DESIGN.md` を正とする。

## 2. Reason

snapshot は正本の複製であり DRY を弱める。byte 一致 Unit は値の写し確認で検出力が無い。`go:embed` は module 内の同 package dir しか読めないため、JSON と同じ dir に byte 公開だけの package を置く。Infrastructure だけがそれを import し、Entities / Application は import しない。

## 3. Rejected

1. Adapter 隣へ schema を copy し byte 一致 Unit で drift を見る案
2. `go generate` で snapshot を同期する案（まだ複製が残る）
3. jsonschema compiler を `contracts` package へ移す案（Drive I/O と無関係な engine が正本 dir に入る）
4. generator の Go module を repo 根へ上げて embed する案（module 境界の変更が本 Adapter より大きい）
