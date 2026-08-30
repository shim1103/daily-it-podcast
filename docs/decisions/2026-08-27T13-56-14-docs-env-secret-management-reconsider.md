---
name: HTTP Adapterを標準Clientと必要なconfigだけへ依存させる
date: 2026-08-27T13:56:14+09:00
branch: docs/env-secret-management-reconsider
---

## 1. Decision

GeneratorのHTTP Adapterは標準の`*http.Client`と、そのAdapterが必要とするruntime config / credentialだけへ依存する。汎用の`secrettransport`、`SecretRef`、`BindingResolver`はtarget architectureから除く。

## 2. Reason

HTTP Adapterはvendor固有のURL、header、body、認証方式を知る境界である。汎用transportへsecret注入方式を移すと、標準HTTP requestとは別のrequest表現、参照とenvironment keyの対応表、解決失敗の規則が必要になる。結果としてvendor Adapterだけで完結する認証知識が複数moduleへ分散し、変更時にAdapterと汎用transportの両方を直す。

runtime configをstartupで検証済みの型へ確定すれば、Adapterへ未解決のsecret参照を渡す理由はなくなる。Adapterは必要なcredentialだけを受け取り、標準Clientでrequestを送れば足りる。標準Clientを境界にすると、timeout、Transport差し替え、test serverとの接続にGo標準の仕組みをそのまま使え、独自HTTP abstractionの契約を維持しなくてよい。

この依存形でもcredentialの保存元はAdapterから隠せる。保存元を隠す責務はconfiguration boundaryが持ち、HTTP送信の責務へsecret名の解決まで混ぜない。

## 3. Rejected

1. `secrettransport`を残し、process environment実装だけへ縮小する案。保存元の切替が不要になっても、独自request表現とsecret参照解決の二重境界が残る。
2. vendor Adapterごとに独自Client interfaceを作る案。標準Clientで置換とtestができる範囲までinterfaceを増やし、必要のない抽象を維持する。
3. Adapterがenvironment keyを受け取り、自分で値を読む案。HTTP送信へconfiguration loadingを混ぜ、startup時の一括validationを崩す。
4. raw credentialを汎用mapで渡す案。必要な値の型と所有が不明確になり、Adapterが無関係なcredentialへaccessできる。
