---
name: AgentSecrets command 素材の正本は commandlaunch/agentsecrets であり infrastructure/agentsecrets は置かない
date: 2026-08-26T12:14:00
branch: feature/generator-agentsecrets-http-proxy-absorb
---

## 1. Decision

1. AgentSecrets × command 出口の素材（`EnvWrapper` / project path 規約）の正本は `commandlaunch/agentsecrets` に置く。
2. `infrastructure/agentsecrets` package は置かない。HTTP 側は `secrettransport/agentsecrets`、command 側は `commandlaunch/agentsecrets` に閉じ、出口を知らない共有袋を残さない。
3. 本判断は置き場の所有境界だけを固定する。`commandlaunch.Launcher` 実装と Composition 結線は command 出口 Issue が所有する。

## 2. Reason

1. 秘密まわりは置き場×出口の2軸である（`2026-08-25T13-53-55`）。processenv は既に出口ごとに runtime が分かれている。AgentSecrets だけ出口横断の共有 package を残すと、同じ問いに「どちらが正本か」が答えられない。
2. HTTP 吸収 Decision（`2026-08-25T19-36-11`）は command 側を HTTP Issue に混ぜないことを決めた。所有を `commandlaunch/agentsecrets` へ寄せることは、その分割を壊さず、袋 package を空にして消すための配置である。
3. `EnvWrapper` を `secrettransport/agentsecrets` へ同居させると、HTTP 契約配下の package が CLI argv / project dir 規約を知る。出口軸の Orthogonality が崩れる。

## 3. Rejected

1. `EnvWrapper` を `secrettransport/agentsecrets` に同居させる案。HTTP 出口 package が command 知識を持ち、2軸 Decision に反する。
2. 未結線のまま `infrastructure/agentsecrets` を残す案。読み手が HTTP 正本と command 素材のどちらを直すか迷い、正本が袋に戻る。
3. path 移動だけを「設計完了」とみなす案。command 出口の契約吸収（`Launcher`）は別 Issue の責務であり、所有境界の固定と契約実装の完了を同一視しない。
