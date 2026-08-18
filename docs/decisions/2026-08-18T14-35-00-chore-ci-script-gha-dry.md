---
name: hook と GHA は同じ scripts を呼ぶ。GHA は Unit も実行する
date: 2026-08-18T14:35:00
branch: chore/ci-script-gha-dry
---

## 1. Decision

1. test 手順の正は `scripts/generator/` と `scripts/playback/` の片系入口。root の `check-static.sh` / `test-integration.sh` はそれらを呼ぶだけ。root の `test-unit.sh` は composer 契約を実行してから片系 unit を呼ぶ
2. pre-commit は root の static と Unit、pre-push は root の Integration を呼ぶ。GHA も同じ root 入口を呼ぶ。YAML と hook に `golangci-lint run` や `npm run test` を書かない
3. GHA は Integration 専用にしない。Unit（+ 現行 static）用 workflow と Integration 用 workflow を分ける。知識の単一化と runner の複数化は別である
4. 本 decision は `2026-08-15T23-17-00-chore-test-and-ci` の「GHA は Integration のみ」「push CI に Unit を併記するな」を上書きする。`2026-08-17T14-45-00-chore-test-and-ci-coverage-layer` の「GHA に coverage / 層 lint を載せない」のうち、Unit 入口を GHA で呼ばない部分も上書きする。Integration job に Unit 閾値を載せることはしない
5. 新しい linter / formatter / playback の層検知 / playback coverage は入れない

## 2. Reason

1. DRY が禁じるのは手順の複製（hook と YAML に同じ command を書くこと）であり、同じ script を local と remote で実行することではない
2. pre-commit と pre-push を同一 script に畳むと Test Pyramid の Scope 分離が壊れる。caller を揃える対象は GHA と hook の組であり、commit と push の組ではない
3. GHA は hook を bypass した変更を拾う runner である。hook 側を DRY にしたことは、GHA で Unit を回さない理由にならない
4. 片系入口は generator と playback の直交性を保つ。集約は pre-commit / GHA 用の薄い合成に限る

## 3. Rejected

1. GHA を Integration のみに保つ案（hook の DRY を理由にする。知識の単一化と実行場所の複数化を混同する）
2. YAML に `npm run test:unit` や `golangci-lint run` を直書きする案（手順の正が二箇所になる）
3. pre-commit と pre-push を同一入口にする案（Unit と Integration の gate が潰れる）
4. 旧 `chore/test-and-ci` へ積み増す案（base が `origin/develop` から遅れ、完了済み PR の branch を再利用する）
