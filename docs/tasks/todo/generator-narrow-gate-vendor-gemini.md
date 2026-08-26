## 1. Summary

`apps/generator/test/gemini_narrow_integration_test.go` を追加し、processenv dummy API key + httptest TLS で Gemini SpeechSynthesizer の外向き境界を Integration gate（secret なし Narrow）で self-validate する。既存 sociable unit から境界 I/O 観測を移し、Unit と Narrow の責務を分離する。

## 2. Context

1. Synthesizer は `secrettransport.Client` で API key を注入して外向き POST する
2. `synthesizer_sociable_unit_test.go` が processenv + DialTLSContext + httptest で header/body/WAV を検証しており、Unit/Narrow が混在している

## 3. Canonical Sources

1. `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`
2. `testing-strategy/levels.md` / `credential.md`
3. `apps/generator/internal/infrastructure/speech/gemini/synthesizer.go`
4. `apps/generator/internal/infrastructure/secrettransport/processenv/client.go`
5. `apps/generator/internal/infrastructure/speech/gemini/synthesizer_sociable_unit_test.go`（移管元）

## 4. Scope

### In Scope

1. Narrow: `apps/generator/test/gemini_narrow_integration_test.go`（build tag なし）
2. Narrow へ移す観測: processenv dummy、DialTLSContext redirect、POST 到達、API key header 存在、成功時非空 WAV
3. Unit に残す観測: 空 text、retry 分岐、envelope 組み立ての論理、error 種別（境界 I/O なし、または Client Stub）
4. Unit から移管 case を削除し二重検証をやめる

### Out of Scope

1. Gemini 実 API / AgentSecrets / Broad・System・E2E

## 5. Contract

1. Integration gate exit 0
2. POST + API key header 非空 + 非空 WAV
3. error message に dummy 値なし
4. Unit = Adapter 内分岐のみ。Narrow = 境界 I/O のみ

## 6. Constraints

1. dummy 値を failure message に出さない
2. `local_real` / AgentSecrets を使わない

## 7. Acceptance Criteria

- [ ] Narrow file が CI gate で収集・実行できる
- [ ] upstream で POST と API key header 存在
- [ ] 成功時 WAV 契約
- [ ] error message に dummy 値なし
- [ ] **責務分離**: sociable unit から processenv+httptest 観測を Narrow へ移し、Unit に残さない
- [ ] **責務分離**: Unit は Adapter 内分岐のみ。Narrow は境界 I/O のみ

## 8. Verification

```bash
bash scripts/generator/test-integration.sh
go test ./internal/infrastructure/speech/gemini/ -count=1
```

## 9. Dependencies

- `docs/decisions/2026-08-26T17-42-00-docs-infra-test-discussion.md`

## 10. Risks

1. body/header 観測ずれ → 「POST到達 + header存在 + WAV非空」に最小化

## 11. Notes

GitHub Issue 化しない。本 file が契約の正。
