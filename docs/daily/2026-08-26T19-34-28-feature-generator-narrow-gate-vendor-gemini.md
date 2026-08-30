---
name: Gemini synthesizer 境界I/O観測のNarrow分離
date: 2026-08-26T19:34:28
session_id: none
branch: feature/generator-narrow-gate-vendor-gemini
prev: なし
---

## 1. Summary

`docs/tasks/todo/generator-narrow-gate-vendor-gemini.md` の契約に従い、Gemini SpeechSynthesizer の外向き境界I/O観測を sociable unit test から Narrow Integration test へ分離した。issue-managerフローでexecutor実装・reviewer査読を経て、Contract/Acceptance Criteria全項目をmanager自身が現物確認し、issue fileを削除した。

## 2. Changes

1. `apps/generator/test/gemini_narrow_integration_test.go` を新規作成。processenv dummy + httptest TLS + DialTLSContext で実際にPOSTし、POST到達・`x-goog-api-key`ヘッダ非空・成功時非空WAV・error messageにdummy値が出ないことをself-validateする。
2. `synthesizer_sociable_unit_test.go` から境界I/O観測2件（`TestSynthesize_returnsNonEmptyWAV_whenProxyReturnsPCM`, `TestSynthesize_injectsGeminiAPIKeyRealValue_whenCallingUpstream`）を削除し、Narrowとの二重検証をやめた。
3. 残るUnit test（envelope組み立て、空text、retry分岐、error種別）は `fakeSecretTransportClient`（`secrettransport.Client`のSpy、境界I/Oなし）へ書き換え、Unit=Adapter内分岐のみという責務を実際に成立させた。`synthesizer_edge_sociable_unit_test.go` はissue Scope外として不変更。
4. reviewerがcode-review + Contract/AC照合を実施しブロッキング指摘なし。managerが `error message に dummy 値なし` の項目についてassert反転によるmutation検証（実装・testを一時改変して赤くなることを確認後、復元）を行い、self-validateが実効的であることを裏取りした。
5. `bash scripts/generator/test-integration.sh` と `go test ./internal/infrastructure/speech/gemini/ -count=1` の両方がexit 0であることを確認し、issue fileを削除した。

### Commits

- `b600e66`
- `d974187`
