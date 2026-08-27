---
name: runtime configとsecret管理境界の再構成
date: 2026-08-27T17:09:47
session_id: none
branch: docs/env-secret-management-reconsider
prev: なし
---

## 1. Summary

Generatorのcredential付き実operationをGitHub Actionsへ限定し、runtime configの分類、configuration boundary、HTTP Adapterの依存、production sourceをA/Bとして固定した。後続実装をAgentSecrets削除、TwitterAPI.io削除、Cursor CLIのGitHub Actions実測、runtime config loader実装の4契約へ分割した。

## 2. Changes

1. GeneratorのConfig、Secret、Loader、validation Error、production UseCase factoryの境界契約をcode artifactとして追加した。最初の単一`contract.go`構成は責務過多だったため、名前、Config、Secret、Loader、Errorの変更理由ごとに分離した。
2. PlaybackのCloudflare VariablesとSecretsをbinding型として分離し、production VariablesをCloudflare側で維持する設定へ揃えた。
3. local AgentSecrets廃止、Variables／Secrets分類、Generator configuration boundary、`secrettransport`廃止、TwitterAPI.io削除のDecision Recordを作成し、README、DESIGN、DEPLOYのlatest policyを更新した。
4. C01〜C04の達成契約を`docs/tasks/todo/`へ作成し、Generator laneへ依存順と並列可能範囲を登録した。Dに属する未実測値と後続実装は変更していない。
5. commit hookでGenerator static／Unit、Playback format／lint／typecheck／依存層／Unitを確認した。push hookでGenerator／Playback Integrationがpassした。

### Commits

1. `843cf30`
2. `942233e`
