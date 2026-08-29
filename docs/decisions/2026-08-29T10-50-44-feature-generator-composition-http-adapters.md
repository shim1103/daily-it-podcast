---
name: 置き場一択化に伴う commandlaunch 契約と composition 結線の KISS 化
date: 2026-08-29T10:50:44+09:00
branch: feature/generator-composition-http-adapters
superseded_by: 2026-08-29T13-48-57-feature-generator-composition-http-adapters.md
---

> **上書き**: 結論5 のうち N2（Cursor CLI inject env 名）を N1 と同一視する部分は `docs/decisions/2026-08-29T13-48-57-feature-generator-composition-http-adapters.md` が正。N1 の SSoT を `internal/config` とする側面と `secretnames` package 削除は同 Decision が維持する。

## 1. Decision

1. `commandlaunch` の出口契約は `Command` と `Launcher` だけにする。`SecretName`（opaque token）と `SecretBinding`（解決 interface）を削除する。
2. `processenv.Launcher` は秘密の解決を行わない。Composition が configuration boundary（`internal/config`）で検証済みの Cursor API key 値と、その child env 名を渡す。`launcher` の environment 参照は child env allowlist（`PATH`/`HOME`/`TMPDIR`）の透過引き回しに限る。
3. Composition 内の `secret_bindings.go`（`SecretName`↔秘密名 対応表 file）を廃止する。D5（docs/decisions/2026-08-26T14-58-45-feature-generator-agentsecrets-cursor-command-launcher.md）の file 分割根拠（対応表と置き場実装選択が別の変更理由を持つ）は、置き場が1系統に潰れ対応表が構造的に消えたため失効する。
4. Cursor CLI command 経路の結線を、他3 capability（getxapi / gemini / gdrive）と同じ「Composition が `config.Config` の capability 別 field を Adapter へ渡す」対称形にする。`internal/config` の Cursor capability field を実際に consume する。
5. child env 名の SSoT を `internal/config` の env key 定数へ一本化し、`internal/infrastructure/secretnames` package を削除する。

## 2. Reason

結論1〜2: D4（docs/decisions/2026-08-25T13-53-55-feature-generator-processenv-command-launcher.md）の `SecretName`/`SecretBinding` は「軸A=置き場（local AgentSecrets / remote process env の2系統）」を Composition が選ぶための間接だった。D3 で軸Aが process env 一択に潰れた時点で、opaque token が隠していたのは秘密値ではなく child env 名（`CURSOR_API_KEY`。GitHub Actions workflow に平文で書かれ、秘密ではない）であり、守る保護対象が存在しない。間接を残すと、`launcher` に文字列1個を渡すために token 生成 + interface + map 1エントリの3層を経由し、読み手に「置き場が複数あり選択している」という消えた前提を示唆し続ける。Orthogonality / KISS に抵触。

結論2 の validation timing: 現状 `config.Load` が startup で `CURSOR_API_KEY` 未設定/空文字を violation として検出する一方、`launcher` が実行時に `os.LookupEnv` で同じ env を再度読む。同一 env を別 package が別 timing で二重に読み、`launcher` の実行時 lookup は startup の一括 validation を迂回する。これは D1 への抵触である。Composition が検証済み値を渡せば、Cursor 経路も「外部 I/O 前に設定全体の不備を返す」D1 の保証内に入る（D2 が HTTP Adapter に与えたのと同じ扱いを command 経路へ広げる）。

結論3: D5（docs/decisions/2026-08-26T14-58-45-feature-generator-agentsecrets-cursor-command-launcher.md）は「`SecretRef`↔秘密名の対応表」と「どの置き場実装を New するか」が別の変更理由を持つことを file 分割の根拠にした。対応表が構造的に消え、置き場実装が1つに固定された今、`secret_bindings.go` を独立 file で維持する理由がない。1判断1ファイルと同様に、消えた変更理由に対応する file も畳む（DRY）。

結論4: 現状 `produce_episode.go` の結線5行のうち Cursor の行だけが `cfg` を受け取らず、`cfg.Cursor` が load・validate されて未使用の dead field になっている。他3 capability と対称にすれば、結線層の各行が同一の形（`config.Config` の capability field を渡す）になり、次の読み手が「Cursor だけなぜ違うのか」を再度問わずに済む（configuration-boundary §3、Orthogonality）。

結論5: `CURSOR_API_KEY` という env 名は configuration boundary の知識であり、`internal/config` が既にその boundary の定義者である。`secretnames` package は D4 の2軸時代に「置き場をまたぐ共通名」として必要だったが、D3 で置き場が1系統に潰れた時点で存在理由が消えた。実際 `secretnames` の7定数のうち Cursor 用1個以外は現在どこからも参照されない dead code であり、HTTP Adapter 移行時に env 名の役割が `internal/config` へ移った取りこぼしである。SSoT を1つにする。

## 3. Rejected

1. `commandlaunch` を「child env 名を文字列1個受け取る」だけに縮小し、`launcher` が引き続き `os.LookupEnv` で値を読む案（相談時の案A）。間接3層は消えるが、`cfg.Cursor` が dead field のまま残り、同一 env の二重読取と startup validation の迂回（D1 抵触）を解消しない。後日「なぜ Cursor だけ config を通らないか」を再び議論することになる。
2. opaque token だけ捨て、`SecretBinding` interface を文字列 key に変える案（相談時の案C）。map 1エントリの間接が残り、D5 の file 分割根拠が失効したまま `secret_bindings.go` が存続する。症状の一部だけ消して病因を残す。
3. `config.Secret` 型（opaque）のまま `launcher` へ渡す案。`internal/infrastructure/commandlaunch/processenv` が `internal/config` へ依存し、`Reveal()` 済み primitive だけを受け取る他3 HTTP Adapter と非対称になる。`Reveal()` の呼び出し点が Composition Root にあること自体が境界の表明であり、そこから先は primitive でよい。
4. `internal/infrastructure/commandlaunch/processenv` サブ package を畳んで実装を Composition へ寄せる案。実装が1つでも、`cursorcli` Adapter が具象 `*processenv.Launcher` に依存すると Dependency Rule を破る。contract package と impl package の分離は実装数ではなく「Adapter が抽象に依存する経路を作るか」で決まり、test double の差し込み口としても必要。
5. 過去の D4 / D5 本文を現在の結論へ書き換える案。当時 AgentSecrets 2軸を前提に置いた判断の identity と時系列を壊す。supersede を明示した新 Decision を正とする（decisions.md §8）。
