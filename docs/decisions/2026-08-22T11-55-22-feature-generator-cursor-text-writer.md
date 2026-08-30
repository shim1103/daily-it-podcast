---
name: CLI 境界の秘密供給は agentsecrets env -- の exec wrapper に定める
date: 2026-08-22T11:55:22
branch: feature/generator-cursor-text-writer
---

## 1. Decision

1. Cursor CLI のような **CLI 境界**の秘密供給は `agentsecrets env -- <command> [args...]` の exec wrapper 方式に定める。Go process は秘密値を一度も保持せず、生成する argv も秘密値を含まない。値は AgentSecrets が keychain から解決し、wrapper が起動する子 process の memory にだけ存在する。
2. **HTTP 境界**（generator の既存 4 adapter）は従来どおり AgentSecrets HTTP proxy のキー名注入を使う。機構は境界の種類で分かれるが、**Go process が秘密値を見ないという zero-knowledge の思想は両者で共通**である。HTTP には proxy、CLI には exec wrapper という対応で、同じ思想を各境界の形に合わせて実現する。実行主体（shim か agent か）で経路を分けることは引き続き行わない（`2026-08-16T00-06-30-docs-agentsecrets-secret-export.md`）。
3. 子 process へ渡る環境変数は、その呼び出しが必要とするものだけに限定する。親 process の環境を暗黙に全継承させない。Cursor CLI 呼び出しの場合、AgentSecrets が注入する Cursor 用の秘密と、CLI の実行に必要な最小限の環境変数（PATH 等）だけを渡す。generator が保持する他 vendor の秘密が Cursor CLI へ到達してはならない。
4. `docs/decisions/2026-08-18T16-30-00-feature-cursor-text-writer.md` §4 は、この decision が上書きする。旧 §4 は「実行環境で `CURSOR_API_KEY` が供給されている前提で、Adapter は env の存在だけを期待する」と定めたが、これは AgentSecrets が HTTP proxy しか持たないという誤った前提に立つ判断だった。実際には `agentsecrets env` subcommand が存在し、CLI 境界にも zero-knowledge な供給経路がある。**旧 file は書き換えず、CLI 境界の秘密供給についてはこの decision を現在の正とする**。旧 decision の §1・§2・§3・§5 は引き続き有効である。

## 2. Reason

1. `agentsecrets env -- <command>` は「秘密を解決して子 process へ渡す」という 1 つのことだけを行い、任意の command を入力として受け取る。これを自前の取得 API と env 組み立てで再実装せず、既存の wrapper へそのまま委譲するのが `philosophy` §4-1 の直接の適用である。
2. 子 process への env 暗黙全継承は、Cursor CLI が正当な目的（原稿断片の生成）に必要としない他 vendor の秘密まで到達させる。これは `philosophy` §4-3 に反する。渡す範囲を必要最小へ絞ることでこの違反を解消する。
3. Go process が秘密値を保持しない構造は、値が log・error message・panic trace・agent context へ漏れる経路そのものを設計から消す。取得後に注意深く扱う運用ではなく、取得しない構造で担保する。
4. HTTP 境界と CLI 境界で機構が異なることは、思想の分裂ではない。どちらも「値を見る主体を最小化する」という同一の目的に対し、各境界が提供する最も制約の強い手段を選んだ結果である（`philosophy` §4-2）。この対応を明示しておくことで、境界ごとにどちらを選ぶかの判断が毎回同じ結論に落ちる（`philosophy` §4-5）。

## 3. Rejected

1. Go process が AgentSecrets から秘密値を取得し、`exec.Cmd.Env` へ載せる案。Go が値を見る時点で zero-knowledge が崩れ、`2026-08-15T17-48-16-docs-agentsecrets-secret-management.md` で vault 型 CLI を却下した理由（取得後の値が agent memory に載る）を自ら再現する。加えて `internal/infrastructure/agentsecrets/proxy.go` の公開 API は `Do` のみで、値を取り出す機構を意図的に持たない。この案は proxy へ取り出し API を足すことを要求し、既存の設計意図を壊す。
2. 現状維持（`exec.CommandContext` による env 暗黙全継承）の案。呼び出し側が env を何も指定しないため一見単純だが、Cursor CLI へ渡る秘密の範囲が親 process の環境全体という暗黙の広さになる。渡す範囲が code から読み取れず、generator に秘密が 1 つ増えるたびに Cursor CLI の権限が黙って広がる。`philosophy` §4-3 違反であり、違反が静かに拡大する点で単なる現状の不備より悪い。
3. Cursor 用の秘密だけを別経路（専用の env file や CI secret）で供給し、AgentSecrets の外へ出す案。秘密の所在が二重化し、local 秘密は AgentSecrets に一本化するという `2026-08-15T17-48-16-docs-agentsecrets-secret-management.md` の決定と矛盾する。
