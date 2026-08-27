---
name: Generatorのcredential付き実operationをGitHub Actionsへ限定する
date: 2026-08-27T12:17:00+09:00
branch: docs/env-secret-management-reconsider
---

## 1. Decision

1. Generatorのcredential付き実operationはGitHub Actions runnerだけで実行する。
2. 通常のlocal開発と自動testは実serviceを呼ばず、local secretを必要としない。
3. Generatorのproduction runtimeはprocess environmentからruntime configを受け取る。
4. このprojectのlocal secret供給手段としてAgentSecretsを採用しない。
5. 本Decisionは、local AgentSecrets採用を決めた過去のDecisionを変更せず、2026-08-27時点からの判断として置き換える。

## 2. Reason

通常のlocal開発と自動testが実serviceを呼ばないなら、そのexecution environmentへcredentialを配置する目的がない。利用しないsecretをlocal machineへ保存すると、漏洩・誤操作・rotation・復元の対象だけが増え、必要最小限の権限にならない。

AgentSecretsはlocalでsecretが必要な場合には値の露出を抑えられる。一方、このprojectではlocal実operationを要件から外したため、AgentSecretsのproxy、command wrapper、project分割、keychain運用を維持しても実行上の価値がない。

Generatorの本番実行場所はGitHub Actionsである。GitHub Actionsが必要値をprocess environmentへ注入すれば、Generatorは保存元のproduct固有APIを知らずに受け取れる。実行場所を増やさず、既存のproduction経路へ限定する方が構成とcredential配布先を小さくできる。

過去のDecisionは、その時点でlocal AgentSecretsを採用した理由を保存する記録である。過去本文を書き換えず、新しい前提と結論を本Decisionへ記録することで、時間ごとの判断を保持する。

## 3. Rejected

1. local実operationを維持してAgentSecretsを使い続ける案。現在のlocal開発要件に実service呼び出しがなく、不要なruntime・secret・運用を残す。
2. `.env.local`へcredentialを保存する案。保存形式を変えても、local environmentがsecretを必要としないという判断は変わらない。
3. localとGitHub Actionsの両方へcredentialを配布する案。実行場所と漏洩面を増やすが、現在必要なoperationはGitHub Actionsだけで完了する。
4. GitHub Actions固有APIをGenerator codeから直接読む案。Generatorを特定schedulerへ結合し、process environmentという標準境界を失う。
