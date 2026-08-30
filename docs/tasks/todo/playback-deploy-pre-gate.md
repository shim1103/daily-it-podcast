## 1. Summary

初回 `wrangler deploy` の直前に、account・bundle・Google OAuth consent を確認する deploy 前ゲート。本番 hostname へ traffic を載せる前の最終チェックリスト。

## 2. Context

1. C-03（runtime config 4 key）と C-04（Access Application + Allow）は完了済み。
2. wrangler toolchain と `deploy:dry-run` は C-1/C-2 で repo 内完了。
3. 初回 deploy 本体と OTP 実検証は `docs/tasks/todo/playback-lane.md` Phase 2（deploy + `DEPLOY.md` §7）。

## 3. Canonical Sources

1. `DEPLOY.md` — 公開形・Access・Verification の運用 SSOT。
2. `apps/playback/wrangler.jsonc` — Worker 名・assets・`/episodes*` 先回り（値は本 file に写さない）。
3. `docs/decisions/2026-08-25T17-10-00-feature-playback-worker-deploy.md` — Access 先行・同一 origin。
4. `docs/tasks/todo/playback-lane.md` — Phase 2 以降の進捗 index。

## 4. Scope

### In Scope

1. `npx wrangler whoami` で意図した Cloudflare account であること。
2. `npm run deploy:dry-run`（`apps/playback`）が exit 0。
3. Google Cloud OAuth consent screen が **Testing 以外**（Production 推奨。refresh_token の短期失効回避）。
4. （任意）account の workers.dev 公開を **Disabled** にしてから Phase 2 deploy し、Access 保存後に Enable（露出 window 緩和）。

### Out of Scope

1. 初回 `wrangler deploy` と §7 Verification（lane Phase 2）。
2. rollback 文書化・logging（lane Phase 3）。
3. CI へ `npm run build` を載せる code 変更。

## 5. Contract

1. 上記 In Scope を満たしてから Phase 2 に進む。
2. secret / variable の実値は repo に書かない。

## 6. Constraints

1. GitHub Issue 化しない。本 file が契約の正。
2. `DEPLOY.md` に手順の二重定義を増やさない（参照のみ）。

## 7. Acceptance Criteria

1. [ ] `npx wrangler whoami` が意図した account を示す。
2. [ ] `npm run deploy:dry-run` が exit 0。
3. [ ] OAuth consent screen が Production（または Testing でない状態）であることを確認した。
4. [ ] （任意）workers.dev Disabled → Phase 2 → Enable の手順を取る場合、その順序を守った。

## 8. Verification

```bash
cd apps/playback
npx wrangler whoami
npm run deploy:dry-run
```

OAuth consent: Google Cloud Console → APIs & Services → OAuth consent screen → Publishing status。

## 9. Dependencies

1. C-03 runtime config 4 key 投入済み。
2. C-04 Access Application + Allow 保存済み。

## 10. Risks

1. Testing のまま deploy すると refresh_token が短期失効しうる。
2. workers.dev を Enabled のまま Phase 2 すると、Access 未効の短い露出 window が理論上ありうる（任意緩和を使う）。

## 11. Notes

1. Variable 2 つは Dashboard（Workers → Settings → Variables）。Secret 2 つは `npx wrangler secret list` で名前確認。
