---
name: Issue #218 の branch で、時刻の brand 型を共有の最内側 domain/ に置く判断から、domain の語彙・音声の内容検査・Issue の title 規約・skill の正本の更新まで進めた
date: 2026-10-16T21:01:19
session_id: f6f031c7-50a4-48ac-b4a4-03a1cd8726ed
branch: feature/playback-progress-versioned-persistence
prev: 2026-10-08T00-12-56-feature-playback-progress-versioned-persistence.md
---

## 1. Summary

#218 の実装後の整理として、web・worker・contracts が共有する `apps/playback/domain/` を新設し、時刻の brand 型 `UtcInstant` と domain の語彙の型を置いた。UseCase の束は domain の語彙から合成し、音声の内容検査 `verifyAudio` を UseCase に足した。あわせて、skill の正本（`agent-standards`）と GitHub の Issue（#193 の本文と title、子 12 件の title）を、この session の判断に合わせて更新した。PR は作っていない。

## 2. Changes

1. 検証の最終結果は次のとおり。typecheck は exit 0、lint と format は緑、層 lint は 126 modules で違反 0、unit test は 72 file・626 test で branches 100%、integration test は 59 test が緑だった。integration は、miniflare の registry の書込先を `WRANGLER_REGISTRY_PATH` と `XDG_CONFIG_HOME` で `$TMPDIR` に向けると、sandbox 内で通った。
2. 層 lint に `domain-no-outer`・`domain-no-npm`・`contracts-only-domain`・`worker-entities-no-outer`・`web-primitive-strict` の変更を足し、違反行を一時的に足して検知を実測した。entities から domain への import は、後から許した。
3. 実測した事実。zod の `z.iso.datetime({ offset: true })` は、2月30日・24時・60秒・`+24:00`・offset なし・日付のみを拒否する。Node 22.23 に `Temporal` は無い。`z.fromJSONSchema` は検証が動くが、出力型は推論されない。小数が 4 桁以上の時刻は、Node で切り捨てられた（workerd は未実測で、契約が 3 桁へ切ってから `Date` に渡す）。
4. 時刻の入口は、Opus 5.5 の相談を受けて、domain が文法まで手書きで持つ形から、`Date` を受ける形に縮めた。計画は Opus 5.5 の査読を受けて改訂してから実行した（理由と Rejected は Decision が正）。
5. 事故が 4 つあった。
   1. 相談役の sub-agent が read-only の契約に反して、repo 内に probe の script を作った。`rm` は hook が拒否し、shim が削除した。
   2. 実装の sub-agent が `biome check --write` で担当外を含む約 30 file を書き換えた。HEAD の内容で戻し、最終の作業木を `git status` と gate で確認した。
   3. 私が `git restore --staged` で、shim が stage 済みの `.codex/config.toml` の追跡解除を一度取り消し、すぐ戻した。
   4. shim が `docs/task.md` の項目を削除していた。確認済みとして受け入れ、書き足さなかった。
6. sandbox の制約。`gh` は keychain と証明書の検証に失敗し、sandbox の解除で通った。`agent-standards` への書込み（`.git/index.lock`）も拒否された。解除した shell は一時 dir の path が異なるので、共有する file は scratchpad の絶対 path で渡した。
7. GitHub を更新した。#193 の本文を新しい template（Sub-features の章を無くし、AC が子の一覧）に書き換えた。子 12 件と #193 の title を `{kind}({branch名}): {要約}` に更新した。branch 名は、着手済みの 4 件は実際の branch、未着手の 8 件はこの session で決めた。epic へ merged 済みの子 PR（#210・#214・#215）が閉じる #195〜#197 は、元から closed だった。
8. `agent-standards` の commit は `d21f6c2`・`54fbd9c`（この session の skill 編集を shim が commit）、`8ff8f67`・`d7709c4`・`23aeccb`・`448d919`（この session の `/commit --repo`）。title の規約、`worktree-setup`、`issue/SKILL.md` の掃除、新規 `coding-terms/boundary-trust.md` と各 terms の差分は未 commit。
9. 実音声の署名の検査（`verifyAudio`）が実 R2 の mp3 を弾かないかは、未確認として wiki #192 の Q11 へ移した。
10. MEMORY は空だった。

### Commits

- `86e7e39`
- `9778ce8`
- `741414f`
- `18dc12c`
- `6d95fad`
- `744b622`
- `fb45a06`
- `f8ff913`
- `9669bd0`
- `991789d`
- `84d3492`
- `e35e24b`
- `2cbd65a`
- `4c60a28`
- `14a4b96`
- `9885bae`
- `0bc0ed3`
- `bebe208`
