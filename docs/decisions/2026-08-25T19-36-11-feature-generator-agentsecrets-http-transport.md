---
name: AgentSecrets HTTP proxy の正本は secrettransport/agentsecrets へ吸収する
date: 2026-08-25T19:36:11
branch: feature/generator-agentsecrets-http-transport
---

## 1. Decision

1. AgentSecrets HTTP（proxy）の runtime 正本は `secrettransport/agentsecrets` に置く。`secrettransport.Client`（SecretRef / BindingResolver）を満たす実装が唯一の HTTP 出口実装であり、秘密名文字列 API の別 Client を並行で残さない。
2. `infrastructure/agentsecrets` の HTTP 側（proxy Client / Request / Inject と、それにだけ属する test）は上記へ吸収したうえで削除する。PROXY プロトコル知識の所有を1箇所に閉じる。
3. `infrastructure/agentsecrets` の command 側（`EnvWrapper` / Cursor project path 規約）は本判断の対象外とし、command 出口 Issue が所有する。HTTP 吸収のために command 側を消さない・混ぜない。
4. 一時的な wrap（旧 proxy Client を内包する adapter）は結線復元の到達手段として許容するが、HTTP 出口の最終形ではない。最終形は吸収後の単一実装である。
5. この判断の実施は **mv/rename ではない**。`philosophy`（DRY / Orthogonality / DIP / KISS）に沿って所有境界を再設計することが本体であり、path 付け替えはその結果にすぎない。

## 2. Reason

1. 出口契約の runtime 配置は `secrettransport/processenv` と同型にするのが先行 2軸 Decision（`2026-08-25T13-53-55`）の読みである。HTTP × AgentSecrets だけが `infrastructure/agentsecrets` に名前 API Client を残し、その上に SecretRef adapter を載せる形は、正本が2つある状態であり「同じ問いに同じ答え」が返せない。
2. Decision `2026-08-25T08-03-00` は `agentsecrets.Client` が `secrettransport.Client` を満たすことを前提に書いている。満たす主体が旧名前 API Client のままでは前提と実装が食い違う。満たす主体を `secrettransport/agentsecrets` に固定し、旧 Client を消すと前提と配置が一致する。
3. wrap を恒久化すると、SecretRef 解決と PROXY header 組み立てが package 境界をまたぎ、失敗時の error prefix が二重化する。吸収すれば契約実装と I/O 詳細が1 package に閉じ、processenv 側と同じ読み方ができる（`philosophy` §2-2 DRY、§4-5 一貫性）。
4. command 側（EnvWrapper）を同時に移すと出口軸 Issue 分割（`2026-08-25T14-20-18`）を壊す。HTTP 正本の吸収と command 未結線資産の維持は直交する。

## 3. Rejected

1. wrap 層を最終形として残し、旧 `infrastructure/agentsecrets` proxy Client を下層として恒久同居させる案。正本が2つ残り、Decision `2026-08-25T08-03-00` の「`agentsecrets.Client` が `secrettransport.Client` を満たす」がどの型を指すか曖昧なままになる。
2. proxy を吸収せず、旧 package の Client を in-place で `secrettransport.Client` に作り替える案。runtime 配置が `secrettransport/processenv` と非対称になり、次の読み手が「processenv は契約配下、agentsecrets は別 dir」と再判断する。
3. HTTP 吸収と EnvWrapper / command launcher を同一 Issue にまとめる案。出口・契約・AC が同居し、出口軸分割（`2026-08-25T14-20-18`）に反する。
4. 旧 proxy Client を削除せず「使われていなければ放置」する案。未使用の公開 API と test が残り、次 agent がどちらを直すべきか迷う。
5. file を移して import を直すだけの「移管完了」案。所有境界・契約充足・正本の単一化を設計し直さないまま path だけ変えると、wrap と同じ問題（正本がどこか読めない）が形を変えて残る。
