---
name: Generator episode validationの実装とComposition Root結線
date: 2026-08-20T17:28:00
session_id: none
branch: generator-episode-validation
prev: 2026-08-20T14-20-00-feature-playback-runtime-config-boundary.md
---

## 1. Summary

GeneratorのDrive保存前に原稿schema・stem・空値をApplicationで検証し、Composition Rootからvalidationを迂回できないUseCase結線へ整理した。完了済みtodoを削除した。

## 2. Changes

1. `WriteEpisode` UseCaseとDomain Errorを追加し、検証成功時だけ`EpisodeWriter`を呼ぶ構造にした。
2. Drive保存AdapterからApplication validationを除去し、raw Adapterとして責務を限定した。
3. Composition Rootのfactoryを`NewGoogleDriveWriteEpisode`へ変更し、UseCaseを返す形にした。
4. schema validatorの依存とApplication用depguard allowを追加した。
5. Application・Composition・Drive Adapterのtestを整備した。
6. generator static check、generator unit coverage `90.3%`、repository unit、format、lint、typecheckを検証した。

### Commits

1. `769e2aa`
2. `bfd2241`
