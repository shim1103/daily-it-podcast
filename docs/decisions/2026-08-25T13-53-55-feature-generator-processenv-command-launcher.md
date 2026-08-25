---
name: Generator 秘密境界は置き場×出口の2軸であり commandlaunch/secrettransport は結線ではない
date: 2026-08-25T13:53:55
branch: feature/generator-processenv-command-launcher
---

## 1. Decision

1. Generator の秘密まわりは次の **2軸**で切る。軸Aは秘密の**置き場 / 供給 runtime**（local = AgentSecrets、remote = process environment。GHA は env を供給する caller であり Generator code は GHA 固有 API を知らない）。軸Bは**出口**（command = Cursor CLI、HTTP = 他 vendor）。
2. 理想最終系の部品は「独立 infra が4つ」ではない。出口の**契約が2**（`commandlaunch`・`secrettransport`）、置き場の**runtime 実装が2系統**（`processenv`・`agentsecrets`）、掛け算で runtime 実装スロットが最大4、それに **vendor Adapter が N**、**結線は Composition 1箇所だけ**である。
3. **結線オンリー**なのは Composition だけである。Composition は `SecretRef`↔秘密名 binding、Cursor child env allowlist、必要なら AgentSecrets project、および local/remote runtime の選択を所有し、vendor Adapter へ契約の実装だけを注入する。
4. `commandlaunch.Launcher` と `secrettransport.Client` は結線ではない。それぞれ command 起動・秘密参照付き HTTP という**出口の I/O 契約**である。Application の Port ではない（UseCase の語彙は `TextWriter` / `ItemSource` 等）。契約 package は Infrastructure に置き、vendor Adapter は契約の抽象にだけ依存し、runtime 具象（`processenv` / `agentsecrets`）には依存しない。
5. vendor Adapter（例: `cursorcli`）は出口の話し方（argv・envelope・URL・protocol）だけを知る。秘密値・秘密名・allowlist・project・local/remote のどれか・runtime 具象を知らない。runtime 実装は vendor flag / JSON / URL path を知らない。
6. 先行決定との関係: 契約の存在と Composition 所有は `2026-08-22T18-35-00`、production の置き場が process environment であることは `2026-08-22T18-44-00` を正とする。本 Decision は「何が結線で何が契約か」「2軸と理想配置」「Application へ上げない」を固定し、同一の迷いを再発させない。

## 2. Reason

1. 「infra 同士の依存は常に不正」「`commandlaunch` は結線だから Composition / Application 固有」という読みは、Dependency Rule（内側依存）と DIP（実装詳細への依存禁止）を混同する。同じ外側 ring 内で、vendor I/O と credential runtime という**変わり方の軸が違う**境界を抽象へ依存させるのは DIP の適用であり、Clean の Dependency Rule 違反ではない（`philosophy` §3-1 D、`architecture` ring-model / ports-adapters）。
2. Application に `Launcher` を上げると UseCase が command 起動を知る。それは Application Port の所有範囲を超え、Least Privilege / SRP に反する（`philosophy` §3-1 S・§4-3）。Application が所有する Driven Port は業務語彙の接点（`TextWriter` 等）に留める。
3. 置き場と出口を1軸に畳むと、runtime 追加のたびに vendor Adapter が増え、同一 protocol 知識が重複する。2軸に分けると runtime 差分は契約実装の差し替えに閉じ、vendor 差分は Adapter に閉じる（`philosophy` §2-1 Orthogonality、先行 `2026-08-22T18-35-00`）。
4. Composition 以外に「結線だけ」の層を Infra へ置くと、I/O 契約と組み立てが同名のまま混ざり、次の読み手が契約を削除・移動しやすい。結線の定義を Composition に固定すると迷いの再発を止める。

## 3. Rejected

1. `commandlaunch` / `secrettransport` を Composition 固有の結線モジュールとみなす案。これらは I/O 契約であり、結線（実装選択・binding）ではない。
2. `commandlaunch.Launcher` を Application Port へ上げる案。UseCase が command 起動という手段語彙を持ち、Port が肥大し ISP / Least Privilege に反する。
3. 置き場×出口の各マスを「別々の独立 infra が4つ」と数え、契約なしに vendor×runtime を直結する案。知識が掛け算で増え Orthogonality / DRY に反する。
4. vendor Adapter が `processenv` または `agentsecrets` の具象へ直接依存する案。runtime 追加が全 Adapter に波及する（先行 Decision と同じ却下）。
5. Generator code が GHA API / workflow 設定を import して秘密を取る案。置き場の caller 境界を壊し、他 scheduler で同じ binary を使えなくする（`2026-08-22T18-44-00` と同旨）。
