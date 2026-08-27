---
name: runtime configを機密性に基づいてVariablesとSecretsへ分ける
date: 2026-08-27T12:17:01+09:00
branch: docs/env-secret-management-reconsider
---

## 1. Decision

1. OAuth client IDとDrive folder IDは、secretではないruntime configとして扱う。
2. GitHub ActionsではOAuth client IDとDrive folder IDをVariablesから注入する。
3. Cloudflare WorkersではOAuth client IDとDrive folder IDをVariablesとして注入する。
4. OAuth client secret、OAuth refresh token、API keyはSecretsとして注入する。
5. VariablesとSecretsは保存時の保護区分を表す。runtimeへ注入された後は、applicationが受け取るruntime configとして同じconfiguration boundaryで検証する。
6. environmentごとに変わる実値はsource codeの定数へhardcodeしない。

## 2. Reason

OAuth client IDはOAuth clientを識別する公開可能な識別子であり、それだけでは認証できない。Drive folder IDも保存先resourceを識別する値であり、それだけではDriveへのaccess権限を与えない。両者を機密値として扱うと、何を漏洩から守る必要があるかが不明確になり、secret inventoryとrotation対象を過剰にする。

OAuth client secret、refresh token、API keyは、値を取得した主体が認証済みoperationを実行できる。これらは漏洩時に権限行使へつながるため、Secretsの保護・表示制限・rotation対象とする。

VariablesとSecretsは保存・管理時の性質が異なる。一方、GitHub Actionsのstep environmentまたはCloudflare Worker bindingへ注入された後は、applicationにとって外部から届いた未検証のruntime configである。保存区分をapplication内部の別々のbusiness interfaceへ持ち込まず、runtimeごとのconfiguration boundaryで型と値を検証する。

OAuth client IDとfolder IDはdeploy先によって変わりうる。source codeへ固定するとcode変更とdeploy設定変更が結合するため、runtimeから供給する。

## 3. Rejected

1. runtime configをすべてSecretsへ入れる案。注入経路は揃うが、識別子とcredentialの機密性を区別できず、secret管理対象を過剰にする。
2. OAuth client IDとfolder IDをsource codeの定数へhardcodeする案。environment固有値の変更にcode変更を要求し、runtime configとcodeを結合する。
3. folder IDだけをSecretへ入れる案。folder ID単体はaccess権限を与えず、保護区分を上げる根拠にならない。
4. Variablesを検証せず、そのままAdapterへ渡す案。保存時に非secretであっても、未設定・空文字・誤形式がruntime failureを起こす。
