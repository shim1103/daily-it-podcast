---
name: Cursor CLI TextWriter Adapterの実装とCLI境界の秘密供給
date: 2026-08-22T12:30:00
session_id: none
branch: feature/generator-cursor-text-writer
prev: 2026-08-18T08-10-53-feature-generator-cursor-manuscript.md
---

## 1. Summary

`TextWriter` Portを満たすCursor CLI Driven Adapterを実装し、秘密供給をAgentSecretsのexec wrapper経由へ移した。CLI境界の秘密供給機構を`agentsecrets`へ新設し、Cursor専用projectの分離で他vendorの鍵がCursor CLIへ渡らない状態にした。完了したtask draftを削除した。

## 2. Changes

1. `TextWriter` PortのCursor CLI Adapterを追加した。exec seamを注入形にし、envelope解釈と失敗の畳み込みをAdapter内へ閉じた。
2. Adapterのreviewでmutationを実測し、production結線の破壊・stderr漏洩・nil-interface trapの3つが既存testで検出できないことを確認して是正した。
3. `agentsecrets`にCLI境界の秘密供給を新設した。exec wrapperのargv構築と、目的別project設定dirの解決を所有させた。
4. 実binaryで`env` subcommandに絞り込みflagが無いことと、`exec` subcommandが秘密値をstdoutへ返すことを確認し、後者を採らない判断にした。
5. 子processのenvを明示構築へ切り替え、親環境の暗黙全継承を断った。`cmd.Env`が未設定のときだけ全継承する挙動をprobeで実証した。
6. Cursor専用AgentSecrets projectを作成し、注入範囲を実測で対比した。repo rootで`Injecting 5 secrets`、Cursor専用project dirで`Injecting 1 secret: CURSOR_API_KEY`。
7. 位置情報の解決を`agentsecrets`側へ寄せ、Composition Rootが環境変数を読む構造を解消した。既存4 factoryが環境を読んでいないことを数えて判断した。
8. CLI境界の秘密供給を定めるdecisionを追加した。先行decisionは書き換えず、§4のみを上書きする形にした。
9. generator static check、generator unit coverage `91.1%`、`go build` / `go vet` / `gofmt`を検証した。

### Commits

1. `ff6f4c3`
2. `7271d2f`
3. `8454596`
4. `41897c0`
