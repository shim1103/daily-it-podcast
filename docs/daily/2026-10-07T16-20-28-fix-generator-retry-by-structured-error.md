---
name: generator の原稿生成と TTS の失敗を、再試行・fallback・即 error へ分ける方針に直し、重複を共通 helper へ集めた
date: 2026-10-07T16:20:28
session_id: c027c4ad-bbd5-4a02-88da-e4b1fc198eef
branch: fix/generator-retry-by-structured-error
prev: 2026-10-06T17-02-33-develop.md
---

## 1. Summary

epic #225 の AC 6（retry する error の範囲を決め、範囲外は待たずに fallback へ切り替える）を、Gemini 原稿・Cursor 原稿・Gemini TTS の 3 Adapter へ実装し、AC 6 を達成にした。きっかけは run 37396892485 で、quota 超過の 429 を 3 回待ってから fallback していた事実である。方針は Decision `generator-genai-api-failure-retry-or-fallback` に記録した。重複は `writerretry` へ集め、TTS の dir を原稿側と同じ構成へ統合した。

## 2. Changes

1. run 37396892485 の log を読むと、fallback は起きていた（01:22:39〜01:22:46）。待っても戻らない 429 を、自前の backoff（1s・2s・4s）で 3 回待ってから fallback しただけだった
2. 原稿側の 429 に `Retry-After` が付くかは未実測。付かなければ、原稿の 429 は実質すぐ fallback になる
3. Gemini TTS の `gemini-3.8-flash-tts`・`gemini-3.8-flash-lite-tts` の存在と、3.1 からの破壊的変更（入力は逐語の transcript、出力の既定が RIFF header 付き WAV、`response_format` の形）は、公式 docs で確かめた。API key が無く実測していない。移行は #225 の AC 7 に足した
4. 実装は package ごとに sub-agent へ並行で委譲した（Gemini 原稿・Cursor 原稿・TTS・application 層）。報告は diff と `go test -count=1 ./...`・`golangci-lint` で独立に確かめた。agent の指摘で、geminiapi と Cursor の stream 取得の 401/403 が Decision と食い違うことが分かり、test から直した
5. 検証は、`go build`・`gofmt`・`go vet`・`go test -count=1 ./...`・`-race`（manuscript）・`golangci-lint`（0 issues）が全て通った。TTS の file 統合では、test 関数 55 本が統合前後で名前の集合まで同一であることを確かめた
6. 旧 Decision 3 件の file 名を slug へ改め、docs・workflow・script・DEPLOY の参照を張り替えた。daily は更新していない（shim の指示）
7. `agent-standards` の skills を更新した（未 commit）。`documentation/contract`・`comment`・`decision`、`coding-style/function-design`、`2:platform/go/documentation-layout`・`SKILL`。lesson は #211 へ 12 行を追記した。wiki #192 へ Q9・B18 を足し、#225 の AC 6 を達成にした
8. 予期しない変更が 2 つあった。`docs/daily` の変更の revert と、TTS の `constants_sociable_unit_test.go` の削除は、shim が行った
9. sandbox 内では `go test` の build cache と `golangci-lint` の cache が書けないため、`GOCACHE`・`GOLANGCI_LINT_CACHE` を `$TMPDIR` へ向けた。`gh` は TLS の検証に届かないので sandbox 外で実行した
10. `8664252` に、`git mv` で stage 済みだった旧 Decision 3 件の rename が混ざった。履歴は書き換えていない

### Commits

- `17f145b`
- `66243bd`
- `8152ee6`
- `ac7e87d`
- `c971898`
- `774d5f8`
- `57a0027`
- `70f739e`
- `7cc4d26`
- `328e3af`
- `ba97ada`
- `b70b038`
- `96cbc75`
- `7a33c34`
- `0d45681`
- `cc5940d`
- `0ec86aa`
- `c3c5d88`
- `8664252`
- `70b0556`
- `d0140c7`
- `25d0799`
- `254101a`
- `8aae63c`
- `c8c965f`
