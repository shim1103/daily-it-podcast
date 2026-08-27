---
name: Generatorのruntime configを一度だけ読みcapability別Configとして渡す
date: 2026-08-27T13:56:13+09:00
branch: docs/env-secret-management-reconsider
---

## 1. Decision

Generatorのruntime configは`apps/generator/internal/config`をconfiguration boundaryとし、startup時にprocess environmentから一度だけload・validateする。以後はcapability別にgroup化したtyped ConfigをCompositionから必要な依存先へ渡す。field、environment key、loader、validation errorの境界契約は同packageのA artifactを正とし、本Decisionへ複製しない。

## 2. Reason

runtime中の各Adapterが個別にenvironmentを読むと、設定値を読む時点、未設定の扱い、validation順序がAdapterごとに分散する。同じprocess内で設定の見え方が変わり、外部I/Oを始めた後に一部だけ設定不備へ到達しうる。startupの単一境界で一度だけ確定すれば、外部I/O前に設定全体の不備を返し、その後のApplicationとInfrastructureは検証済み入力として扱える。

VariablesとSecretsは保存時の保護区分であり、Generatorが実行する仕事の分割ではない。application内部までその区分を持ち込むと、同じcapabilityに必要な識別子とcredentialが別構造へ散り、利用側が保存方式を知る。capability単位なら、各依存先は自身の仕事に必要な設定だけを受け取り、GitHub Actions上の保存区分を知らずに済む。

process environmentはGitHub Actionsから直接注入できる標準境界である。別のfile loaderを加えても、localでcredential付き実operationを行わない現在の運用には利用経路がなく、設定源と依存だけを増やす。

## 3. Rejected

1. 各Adapterが必要時にenvironmentを読む案。validationの所有と実行時点が分散し、外部I/O開始前に設定全体を確定できない。
2. Compositionがenvironment読取とvalidationを直接所有する案。結線とconfiguration parsingが同じmoduleへ入り、境界契約を独立して参照できない。
3. ConfigをVariablesとSecretsへgroup化する案。保存時の機密区分を利用側へ漏らし、capabilityごとの依存を表現できない。
4. `.env` loaderを併用する案。process environmentだけで本番注入を満たせるため、使わないlocal secret経路とsource precedenceを追加する。
