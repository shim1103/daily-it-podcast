---
name: 進捗D1の一過性疎通を実施してPASSを確認し、疎通専用の実装を削除した
date: 2026-09-30T16:30:00
session_id: 5e60167a-d898-4f60-ab70-6695f6b66416
branch: feature/playback-progress-d1-smoke
prev: なし
---

## 1. Summary

進捗 D1 の一過性疎通（issue #195）を、`workflow_dispatch` で TEST D1 に対して実施し、PASS を確認した後に疎通専用の実装一式を削除した。疎通は D1 の往復に加えて、binding が local の模擬でなく remote へ届くことも見た。共有 config の D1・R2 が local の模擬に落ちていたことを実測し、`remote: true` を明記して解消した。R2 の疎通は #195 の Scope 外だが、同じ dispatch に含めた。

## 2. Changes

1. dispatch は7回実行した。結果は次のとおり。
   1. run 36672173668: `CLOUDFLARE_ACCOUNT_ID` が空で前提確認が停止した。Variable `TEST_R2_ACCOUNT_ID` は未登録で、登録済みは `R2_ACCOUNT_ID` だけだった。
   2. run 36672401613・36672511307・36672606929: D1 API が code 7403 を返した。診断（`whoami`）を足して、token が見ている account は正しいことを確かめた。
   3. run 36673101017・36673201073: 権限を変更した後も 7403 で、`d1 list`（読取）も code 10000 だった。shim が token の更新 button を押し忘れていて、変更が反映されていなかった。
   4. run 36673616434: 更新 button の押下後、migration が remote の TEST D1 へ適用された。疎通 test は 6 件とも失敗した（D1 は `no such table`、R2 は remote に object が無い）。binding が local の模擬を見ていた。
   5. run 36673790780: 共有 config へ `remote: true` を足した後、6/6 が PASS した。
2. yml を master へ載せる PR #208 と、削除する PR #209 は、shim が merge した。tmp の worktree と branch は、local・remote とも削除した。branch の新規作成が hook に止められたため、既存の tmp branch に削除 commit を積んで PR にした。
3. 疎通 PASS 後に、疎通専用の実装（yml、script、`test/dispatch/`）を削除した。migration file と、共有 config の `migrations_dir`・`remote: true` は恒久物として残した。
4. #208・#209 の master push で CD 連鎖が2回走り、`playback-e2e` が失敗した。原因は Access session の失効で、shim が確定した。本 session の変更ではない（deploy されたコードは同一）。
5. sub-feature の branch へ master を取り込んだのは不要だった（shim の指摘）。production の merge commit が branch の履歴に入っている。tree の差分は相殺されるため、履歴は変えていない。
6. 別件（Scope 外）として、e2e の session 失効時の復旧手順（登録 script、期限確認 step、失敗時の画面保存、DEPLOY.md、fixture README の R2 化）を作った。d1-smoke の merge 後に別 branch で扱うため、commit せず worktree に残している。`access-session-cli.ts` に test は無い（偽の JWT で分岐を手元で確認しただけ）。
7. wiki #192 を更新した。疎通で解決した Open questions を削除し、未確認事項を追加した。

### Commits

- `1c6d12c`
- `1f1d1a0`
- `dd0b389`
- `779e36d`
- `79980e4`
- `f1f61f1`
- `7fd55c7`
- `74edadf`
- `a5f72cf`
- `fbba973`
- `5fb9cd0`
- `02d2adc`
- `f1ca511`
- `df4273e`
- `92d342e`
- `10c7ab7`
- `1bd38b5`
- `82fb25c`
- `8700218`
- `ace28a1`
- `09dddb8`
