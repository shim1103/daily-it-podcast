---
name: Playback runtime config boundaryの実装とIssue整理
date: 2026-08-20T14:20:00
session_id: none
branch: feature/playback-runtime-config-boundary
prev: 2026-08-19T18-52-00-tmp-branch.md
---

## 1. Summary

Playback Workerのruntime config検証、repository選択、HTTP Error境界を整理し、完了済みtaskを削除した。次のHTTP contract変更は別todoへ分離した。

## 2. Changes

1. runtime configの内部ErrorとExternal Errorを分離し、HTTP boundaryで503 responseへ変換した。
2. InMemory選択を4 key全てundefinedの明示local / unit test modeに限定した。
3. Controllerの不要な`ready` wrapperを削除し、config Error、log、cause chainをtestした。
4. runtime config validationのDecisionを更新した。
5. 完了済みruntime config boundary taskを削除し、`configuration_error` contract変更を新todoへ記録した。
6. Unit `106 passed`、Integration `1 passed`、format、lint、typecheckを検証した。
