---
name: root.ts coverage 修正・diagrams 色付け・generator lane 整理・seek bug 根本修正・Preview URL 無効化
date: 2026-09-04T23:20:00
session_id: none
branch: feature/playback-e2e-redeploy-master
prev: 2026-09-04T19-45-00-feature-playback-e2e-redeploy-master.md
---

## 1. Summary

`~/projects/daily-it-podcast` の local `develop` を誤って直接 ff-merge し、その後の shim の `git pull` で分岐・破損したのを `origin/develop` へ reset して復元した。`origin/develop`（PR #129）を feature branch へ merge。`root.ts` の func coverage 66.66% を、unit で到達しない fetch ラッパを外して 100% にした。`generator-system-e2e-produce-episode.md` 削除時に参照が宙に浮いていた `generator-lane.md` を修正。`apps/diagrams/icons.py` の custom icon（Cloudflare・Google Drive・Gemini・Cursor・Hono）へ公式ブランドカラーを注入し `runtime.png` を再生成。shim から本番で「seek が 0:00 に戻る」と再報告があり、2 回の的外れな修正（`readyState` 待ち・`seeked` 待ち）を経て、Chrome devtools で実機の event を直接観測し、**worker の音声配信が HTTP Range 未対応で `<audio>.currentTime` 代入がブラウザに無視されていた**ことを特定・修正した。「停止中の topic seek で再生が始まる」bug も別途修正。`wrangler deploy` の Preview URL warning を config で恒久解消。Decision 3 本を起票。

## 2. Changes

1. `~/projects/daily-it-podcast` の local `develop` を誤って `origin/develop` より先に進めた（本 branch を直接 ff-merge）。直後に PR #129 が `origin/develop` へ merge され、local `develop` は origin と分岐 → shim の `git pull` が rebase 失敗 → `ours` strategy merge で壊れた状態に。`git reset --hard origin/develop` で復元。
2. `origin/develop`（PR #129, `8cf60c7`）を feature branch へ merge（`3f48fd9`）。merge 直後 `docs/lessons/index.md` が 1 行（`# lessons` のみ）だったのを conflict 解消の破損と誤診断し、union 復元を試みた。実際は shim が session 外（`139e22e`/`53c725a`）で意図的に空へ整理した状態が正しく、union 復元は誤り。shim の指摘（「0 line is SSoT」）で working tree を `HEAD`（空）へ戻した。
3. `apps/playback/worker/src/composition/root.ts` の func coverage が 66.66% だった原因を特定: drive 分岐が `fetch` を `(input, init) => fetch(input, init)` ラッパ越しに渡しており、unit test では `GoogleDriveEpisodeRepository` が実 HTTP を打たないためラッパが一度も呼ばれない。narrow integration が `fetch: globalThis.fetch` 直渡しで動く実績から、ラッパを外して直渡しに変更（100% 達成）。
4. `docs/tasks/todo/generator-system-e2e-produce-episode.md` を削除した際、未完了の rate 計測 follow-up 2 件（TTS rate 実 dispatch・draft 尺 A/B）への参照が `generator-lane.md` 側に宙に浮いていた（child file 削除後もリンクが残存）ことに気付き、lane の D 表（未決 index）へ統合し直した。
5. `apps/diagrams/icons.py` の custom icon（Cloudflare / Google Drive / Gemini / Cursor / Hono）が Simple Icons の無 fill svg をそのまま raster していたためモノクロだった。cdn.simpleicons.org の既定 fill 値（各社ブランドカラー）を catalog へ持たせ、取得直後の svg へ注入するよう変更。`runtime.png` を再生成。
6. **seek bug（本番で再発）**: shim から「本番で seek が 0:00 に戻る」と再報告。1 回目の修正（`readyState` 未達なら `loadedmetadata` を待つ）だけでは直らず、E2E は `currentTime≈26`（自然再生で timeout 前に進んだ位置）で失敗。2 回目の修正で `seeked` event を待ってから `play()` する形に変えたが、CI では `currentTime=0` のまま timeout。Chrome devtools で実機の `seeked` event を直接観測し、`currentTime` 代入後に `seeked` は発火するがその時点の `currentTime` が代入値ではなく `0` のままであることを確認。**真因は worker の音声配信が HTTP Range request に応えておらず（常に `200` で 24MB 全体を返す）、ブラウザが「seek 不可能な配信」と判断して `currentTime` 代入を無視していたこと**。`audio-response.ts` に Range 対応（`206`/`416`/`Accept-Ranges`）を実装し、deploy 後に devtools で `seeked` 時の `currentTime` が期待値と一致することを確認、E2E も 32 秒 timeout → 3.8 秒 pass に短縮。
7. shim から「停止中に topic bar を押すと再生が始まってしまう」と別途指摘。`seek()` が常に `shouldPlay:true` で呼ばれていたのを、呼んだ時点で実際に再生進行中（`phase:loading/playing`）かどうかで分岐するよう変更。停止中（idle・paused・ended・error）は位置だけ動かし `phase:paused` を維持。
8. `wrangler deploy` の「Preview URL が既定で有効になる」warning を `apps/playback/wrangler.jsonc` に `preview_urls: false` を明示して解消。
9. Decision 3 本を起票: Preview URL 恒久無効化（`23-15-00`）、音声配信の Range 対応（`23-16-00`）、停止中 seek の方針（`23-17-00`、先行の「押した時点の state に関わらず再生する」記述を上書き）。

### Commits

- `82defe1`
- `fcc1e19`
- `34e68b8`
- `3f48fd9`
- `0a96bcd`
- `fe465f2`
- `ed9f223`
- `139e22e`
- `df79d2d`
- `35aa7dd`
- `7b162bc`
- `8ffe9c8`
- `da9af58`
- `34b8099`
- `79dd607`
