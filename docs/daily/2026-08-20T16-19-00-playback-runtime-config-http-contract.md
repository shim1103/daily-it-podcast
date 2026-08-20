---
name: Playback runtime config HTTP contractの実装とPR準備
date: 2026-08-20T16:19:00
session_id: none
branch: playback-runtime-config-http-contract
prev: 2026-08-20T14-20-00-feature-playback-runtime-config-boundary.md
---

## 1. Summary

runtime config不備を`configuration_error`としてHTTP contractへ追加し、WorkerとWeb Clientのerror mappingを更新した。decision、後続Client task、lessonsを実装結果へ整合させ、PR作成前の検証を完了した。

## 2. Changes

1. `PlaybackRuntimeConfigError`を`ConfigurationError`へ変換し、`500 / configuration_error`として返すmappingを追加した。
2. `UnavailableError`は外部service一時不能の`503 / unavailable`専用として維持した。
3. response bodyのcode制限、cause chain log、credential別secret非漏洩をtestした。
4. 完了済みHTTP contract taskを削除し、後続Web Client taskへ`configuration_error`契約を反映した。
5. 旧decisionへ後続taskによるsuperseding decisionを追記した。
6. Unit `118 passed`、Integration `1 passed`、static check、format、lint、typecheckを検証した。
7. `feat(playback): runtime configのHTTP error contractを分離する` と `docs(playback): HTTP contractの正準記録を更新する` を作成した。

### Commits

1. `de20003`
2. `c73ec80`
