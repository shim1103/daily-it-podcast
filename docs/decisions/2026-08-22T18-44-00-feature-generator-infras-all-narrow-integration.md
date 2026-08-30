---
name: Generator の production credential runtime は process environment を使う
date: 2026-08-22T18:44:00
branch: feature/generator-infras-all-narrow-integration
---

## 1. Decision

1. Generator の production credential runtime は `processenv` implementation とする。`processenv` は Composition の `SecretRef` binding を使い、HTTP request の header、form、JSON field と Cursor child process の allowlist env へ必要な値だけを供給する。
2. GitHub Actions は process environment を供給する caller であり、Generator code は GitHub Actions 固有 API や workflow 設定へ依存しない。
3. production runtime で secret 値を process memory から完全に排除することは要求しない。ただし値を argv、stdin、error、log、child process の allowlist 外 environment へ出さない。
4. production workflow は、選択した source provider と実行経路に必要な secret だけを process environment へ渡す。

## 2. Reason

1. GitHub Actions の secret は job / step の process environment へ供給される。HTTP header・form・JSON body へ直接注入する runtime は、この境界をそのまま消費できる。AgentSecrets proxy を CI 内で再構成する必要はない。
2. `processenv` を GHA 固有 package にしないことで、同じ Generator binary を別の scheduler や container runtime からも同じ contract で起動できる（`philosophy` §4-2）。
3. local AgentSecrets は値を Go process へ渡さないが、CI の process environment はその性質を提供しない。両者の runtime 制約を同一化せず、value の出力・継承を禁止する境界 contract で露出を最小化する（`philosophy` §4-3）。
4. source provider の代替 credential を同時に inject すると、選ばれない provider の秘密まで process が保持する。必要な secret だけを渡すことで最小権限を保つ。

## 3. Rejected

1. production GHA で AgentSecrets proxy / project を運用する案。CI secret management と別の project binding を重ね、runtime の構成と障害点を増やす。
2. Generator が GitHub Actions API を import して secret を取得する案。application が deployer 固有の実装へ依存し、他 runtime で同じ binary を使えない。
3. 全 provider の secret を常に job env へ inject する案。選ばれない provider の秘密まで process へ渡り、Least Privilege に反する。
