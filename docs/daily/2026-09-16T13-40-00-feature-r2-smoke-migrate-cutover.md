---
name: 列6（R2疎通・人手移行・本番同着切替）の実装差分調査から結線切替commitまで
date: 2026-09-16T13:40:00
session_id: none
branch: feature/r2-smoke-migrate-cutover
prev: なし
---

## 1. Summary

target Issue `r2-smoke-migrate-cutover.md`（列6）の記述が最新codebaseと不整合であることが判明した。「Adapter本実装済み・切替だけ残る」という前提に対し、実際は generator（`newProduceEpisode`）・playback（本番 route）とも本番結線自体がDrive固定のままで、R2 Adapter本体・factory関数は実装済みだが未呼出だった。この差分を調査・Decision化した上で、結線をR2へ完全置換し、疎通専用`workflow_dispatch`を新設した。

## 2. Changes

1. generator/playback双方のcomposition/route層を調査し、R2 Adapter本体は実装済みだが本番呼出経路が無いことを特定した
2. Drive→R2移送の手動/script化判断（移動・確認とも人手のまま）と、結線方式・疎通dispatch範囲の判断をDecision化した
3. executorへ結線切替（generator/playback）と疎通専用yml新設を委譲した
4. 委譲後の実装をtesting-strategy（naming-and-layout・gwt・levels）に照らして査読し、配置・GWT構造・test名の3点を修正した
5. 疎通専用helper（List/Get/Delete相当）が本番infrastructure packageへbuild tag無しで追加されていたため、`r2smoke`タグを追加して本番buildから除外した
6. `r2-smoke-migrate-cutover.md`（列6）・`r2-post-cutover-verify-oauth.md`（列7）を実装差分に揃えた
7. 全変更を1 concern単位で5commitに分割しpushした

### Commits

- `959fb8e`
- `537e32f`
- `fe831c2`
- `7e35a5d`
- `d19f4ed`
