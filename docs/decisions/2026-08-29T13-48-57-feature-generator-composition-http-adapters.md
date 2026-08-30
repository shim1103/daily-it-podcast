---
name: Cursor CLI 経路の child env 名（読取key / 注入名）と継承allowlist の所有を、変更理由ごとに config / cursorcli / commandlaunch へ分離する
date: 2026-08-29T13:48:57+09:00
branch: feature/generator-composition-http-adapters
---

## 1. Decision

1. `CURSOR_API_KEY` が担う3つの知識を所有者ごとに分離する。N1（generator が親環境から読む key 名）は `internal/config` が configuration boundary の読取契約として所有し続ける（`config.CursorAPIKeyEnv`）。
2. N2（Cursor CLI child が API key を期待する env 変数名）は `internal/infrastructure/manuscript/cursorcli` が Cursor CLI の呼び出し仕様として所有する。argv flag 名（`--model` 等）と同じ層の vendor 知識。
3. N3（child process へ親から継承を許す env 名の allowlist）は `internal/infrastructure/commandlaunch` が契約 package の定数として所有する。POSIX child process 一般の最小継承要件であり、特定 launcher 実装にも特定 vendor にも依存しない。
4. 直近 Decision（`docs/decisions/2026-08-29T10-50-44-feature-generator-composition-http-adapters.md`）結論5「child env 名の SSoT を `internal/config` の env key 定数へ一本化」は、N1 に限定して継承する。N2 は Cursor CLI 仕様として別所有とし、N1 と同一視しない。

## 2. Reason

結論1〜2 の核: `CURSOR_API_KEY` は「generator が読む key 名」と「Cursor CLI child が期待する env 名」で**変更理由が違う**。Cursor CLI が将来 `CURSOR_TOKEN` へ改名したら N2 だけが変わるべきで、`config/names.go`（configuration boundary の読取契約）は無関係。今は両者を `config.CursorAPIKeyEnv` 1個へ潰しているため、Cursor CLI 仕様変更が configuration boundary の変更として現れる（SRP 違反）。HTTP 経路との対称性も根拠: `gemini` Adapter は「API key を `x-goog-api-key` header に載せる」という名前を自分で持ち、誰も Composition へ上げていない。command 経路の inject env 名は HTTP 経路の header 名と同型であり、`cursorcli` が持つのが対称。

結論2 の補足: D4（`docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md`）§5「vendor Adapter は秘密名を知らない」の「秘密名」が指していたのは、当時 `SecretName` opaque token が担った**置き場をまたぐ識別子**である。AgentSecrets が不採用になり置き場が process env 一択へ潰れた後（`docs/decisions/2026-08-27T12-17-00-*`）、残る「名前」は Cursor CLI protocol 上の env 変数名だけであり、これは D4 §5 の禁止対象ではない。D4 §5 の文言は生きているが指示対象が消えている。

結論3: N3 allowlist（`PATH`/`HOME`/`TMPDIR`）は「何を通すか」という集合の定義であり、OS の事実として実装非依存。`agentsecrets` launcher が復活しても値の解決経路が keychain 経由に変わるだけで、通す3つの名前は不変。実装差し替えで不変なものは実装の default ではなく契約側の知識。`processenv` 実装に閉じると、2つ目の実装が現れた瞬間に同じ3定数が2箇所へ複製される（DRY違反、D4 §Reason 3「変わり方の軸で切る」）。Cursor 実行要件でもない（`PATH`/`HOME`/`TMPDIR` に Cursor 固有要素がゼロ、別 CLI へ差し替えても同一）。`docs/decisions/2026-08-22T11-55-22-feature-generator-cursor-text-writer.md` §3「CLI の実行に必要な最小限の環境変数（PATH 等）」も AgentSecrets 前提の文脈だが内容は実装非依存として書かれており、契約側配置と整合する。

N3 を `commandlaunch` へ入れる判断が呼ぶ前提確認: `docs/decisions/2026-08-22T18-35-00-feature-generator-infras-all-narrow-integration.md` §2 は「A の固定契約は `commandlaunch/contract.go` を正とする」と契約 package の存在を確定した。この確定は AgentSecrets 2軸（置き場 local/remote を選ぶための契約）を前提にしていた。2軸は失効したが、`contract.go` の存在根拠は Dependency Rule 単独で立ち直る: `cursorcli`（vendor Adapter）が `commandlaunch/processenv`（runtime 実装 package）を import しない経路を、契約 interface が提供する。実装が1つでもこの根拠は実装数に依存しない（直近 Decision §Rejected で既に明文化済み）。加えて N3 定数が `commandlaunch` へ入ることで、契約 package は `Command` / `SecretEnv` / `Launcher` / allowlist 定数を持ち、「畳めるほど薄い」状態ではなくなる。なお契約 package と実装 package の分離は維持する。根拠は変わらない。

結論4: 直近結論5 が正しかったのは「`secretnames` package の削除」と「N1 の SSoT は `config`」。誤りは N1 と N2 を同一と扱った点だけ。N1 限定で継承する。

## 3. Rejected

1. N3 allowlist を `cursorcli`（vendor Adapter）が所有する案。`PATH`/`HOME`/`TMPDIR` は Cursor CLI 固有ではなく POSIX child process 一般の最小要件であり、別 vendor CLI へ差し替えても同一。vendor Adapter に置くと、Adapter が増えるたび同じ3定数と「何を漏らさないか」の Least Privilege 判断が分散する。

2. N3 allowlist を `commandlaunch/processenv` 実装の default に閉じる案。実装が2つ以上になった時点で同じ3定数が各実装へ複製される。「何を通すか」は実装非依存であり、解決手段（`os.LookupEnv`）だけが `processenv` 固有。

3. `commandlaunch.Launcher` を `internal/application/port` の Driven Port へ上げる案。`Launch(ctx, Command{Program, Args, Stdin})` は OS process 実行の語彙そのもので、`port.TextWriter` が vendor CLI envelope / stdout / exit code を露出しない invariant と同じ理由で Application に置けない（D4 §2 / §Rejected 2 を再確認）。構造的にも `Launcher` を呼ぶのは `cursorcli`（infrastructure）で Application ではなく、`ports-adapters` §2「Port は application core 側が所有する」の所有根拠を満たさない。

4. `commandlaunch/contract.go` を `commandlaunch/processenv` へ畳む案。package 名 `processenv` を `cursorcli` が import した時点で「Cursor Adapter は process environment という置き場を知る」と import graph 上で宣言され、D4 §Rejected 4「vendor Adapter が `processenv` の具象へ直接依存する案」の禁止対象になる。実装数ではなく「Adapter が抽象に依存する経路を作るか」で分離の要否が決まる（直近 Decision §Rejected 4）。

5. 過去 D4 / `2026-08-22T18-35-00` の本文を現在の結論へ書き換える案。当時の 2軸前提での判断の identity と時系列を壊す。supersede を明示した新 Decision を正とする（`decisions.md` §8）。
