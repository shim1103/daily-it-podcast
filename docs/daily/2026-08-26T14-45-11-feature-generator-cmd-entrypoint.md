---
name: cmd Driving Adapter 完了と OS 成否観測の固定
date: 2026-08-26T14:45:11
session_id: none
branch: feature/generator-cmd-entrypoint
prev: なし
---

## 1. Summary

issue-manager で `cmd/generator` を薄い Driving Adapter として完了し、exit 写像の Sociable Unit・lane 更新・Issue 削除まで行った。続けて OS process の exit/stderr を成否観測の正とする Decision を残し、pr-completion で commit / log / PR まで進めた。dotfiles の logging へ「議論軌跡」責務を追記したが、dotfiles commit は non-scope。

## 2. Changes

1. `main` に公開契約を置き、`run` へ exit 写像を抽出。Sociable Unit で nil→0 / error→非0+stderr を検証した。
2. review 指摘どおり `run` の契約 tag を除去し、lane の未完了 index から完了 Issue を外して todo を削除した。
3. OS & runtime の成否観測 Decision と DESIGN / lane の最新表記を揃えた。
4. static / unit gate pass（unit coverage 91.2%）。`ProduceEpisode.Run` 本体は D のまま。
5. PR #61 を `develop` 向けに作成した。base との merge conflict なし。

### Commits

- `d1b8b41`
- `b3343b7`
- `5bc8e2b`
