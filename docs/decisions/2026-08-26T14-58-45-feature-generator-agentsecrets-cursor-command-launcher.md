---
name: Composition 内は bindings（表）と runtime（Client/Launcher）を file 分割する
date: 2026-08-26T14:58:45
branch: feature/generator-agentsecrets-cursor-command-launcher
---

## 1. Decision

1. Composition package 内の credential 結線は、次の **2 file** に分ける。`secret_bindings.go` は `SecretRef`↔秘密名の対応表（および Resolve）だけを持つ。`runtime.go` は置き場×出口の契約実装（`secrettransport.Client` / `commandlaunch.Launcher`）の組み立てと、その組み立てに必要な Cursor command 方針（allowlist・ProjectDir・SecretKeys 宣言）を持つ。
2. each-composition file（`cursorcli.go` / `gemini.go` 等）は Port（または Gate UseCase）への Adapter 結線だけを持ち、表の内部や runtime 具象の組み立て詳細を所有しない。runtime factory を呼ぶ。
3. Composition の test は対象 file の責務に合わせる。表の対応は bindings 側 test、runtime factory / 方針は runtime 側 test、vendor Port 結線は each-composition 側（必要なときだけ）。型assert による「実装が間違っていない」回帰 guard や、下位 Scope が所有する分岐の再assertは置かない。
4. package 全体が「結線を所有する」という先行 Decision（`2026-08-25T13-53-55`、`2026-08-22T18-35-00`）は維持する。本 Decision は Composition **内**の file SRP だけを追加する。

## 2. Reason

1. `secret_bindings` という名前の file に Client / Launcher factory まで同居させると、読み手は「表の SSoT」と「runtime 実装の選択」を同一変更理由だと誤認する。変更理由が「秘密名の対応」と「どの置き場実装を New するか」で分かれるなら file も分ける（`philosophy` §3-1 S、§2-1 Orthogonality）。
2. factory を each-composition へ散らすと、同一の `processenv.NewClient(bindings, …)` が vendor 数だけ重複し、local/remote 切替時に漏れやすい。共有スロットは `runtime.go` 1箇所に閉じ、each は Port 結線に薄くする（`philosophy` §2-2 DRY と §3-1 S の両立）。
3. allowlist / ProjectDir / SecretKeys は「表の一行」ではなく Cursor command runtime を閉じるための方針であり、Launcher 組み立てと同じ変更理由を持つ。表 file に残すと再び bindings が肥大する。
4. test を file 責務へ揃えないと、bindings test が runtime factory の nil 否定を抱え、runtime の変更で無関係な test file が赤くなる。Fault Isolation（`testing-strategy` levels）と minimization（下位詳細の再assert禁止）に反する。

## 3. Rejected

1. bindings に表も factory も置く案。file 名と責務が不一致のまま残り、今回の問いが再発する。
2. factory を each-composition へ全部戻す案。HTTP Client の同一組み立てが vendor file に散り、置き場一括切替の変更理由が N 箇所になる。
3. bindings を表のみにし allowlist / ProjectDir も表に残す案。方針と表が同居し、Launcher だけ runtime へ移す中途半端な分割になる。
4. 型assert で production/local 具象を Composition test に固定する案。振る舞いでなく実装構造を見て、結線層に分岐が無いのに test だけ増える（先行の shaving 判断と同じ却下）。
