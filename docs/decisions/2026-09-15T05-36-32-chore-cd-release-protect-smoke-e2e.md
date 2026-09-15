---
name: playback-smokeは「Access以外の本当のe2e」1本に統合する。UI smokeとcredential smokeの2本立ては廃止する
date: 2026-09-15T05:36:32
branch: chore/cd-release-protect-smoke-e2e
---

## 1. Decision

1. `playback-smoke`（先行案：UI smoke[Playwright+dummy backend fixture] / credential smoke[vitest+TEST_*credentialでNode直接疎通]の2本立て）を廃止し、**1本のPlaywright test**（`test/smoke/playback.smoke.spec.ts`）へ統合する。
2. 統合したsmokeは、`web/vite.smoke.config.ts`（新設）が提供するVite dev server上で、`createApp()`（override無し。既存本番composition経路そのまま）をTEST専用（`TEST_GOOGLE_OAUTH_*` / `TEST_DRIVE_FOLDER_ID`）credential入りの`env`で呼ぶ。専用middleware関数は新設せず、Hono appの`fetch(req, env)`第2引数にTEST_*credentialを直接注入するだけで済ませる。
3. `web/vite.config.ts`（既存、fixture固定のdummy backend。開発用`npm run dev`）は変更しない。`web/vite.smoke.config.ts`は別fileとして新設し、originはlocalhost:3000（既存devと同じ形）のまま、Vite dev server自体が"local runtime"を兼ねる（`wrangler dev`は使わない）。
4. TEST Drive folderは`generator-system.yml`が都度書込・削除する運用のため、smokeはepisode件数を固定assertしない。「一覧応答がruntime config errorにならない」ことを常時確認し、「episodeが1件以上あるときだけ選択・再生の操作が例外にならない」ことを追加確認する。

## 2. Reason

1. shimの指摘「vite local dev, hono local devでoriginをlocalhost:3000で、TEST secret使えば、Accessを除く本当のe2e-testができるんではん？」の通り、Hono appの`fetch(request, env)`はWorkers専用APIに依存しないため、Vite dev server内で直接呼べる。`wrangler dev`のような追加runtime層は不要（既存`web/vite.config.ts`の`dummy-backend-api` middlewareが、この形が既に動作することを示していた）。
2. shimの指摘「env GOOGLE_OAUTH_CLIENT_SECRETとして、secret.TEST_*を渡せば、専用関数は不要では？」の通り、`createDummyBackendMiddleware`のような専用関数を増やす必要は無い。`createApp()`（override無し）へ渡す`env` objectの中身をTEST_*credentialにするだけで、既存の`validatePlaybackEnv` → `mode: "drive"`選択ロジックがそのまま働く。
3. shimの指摘「smoke-testがAccess以外の本当のe2e1本として書き直せ」の通り、UI機能確認とcredential疎通確認を分けて2本持つ設計（本session内で一度実装したが、commit・公開前に本Decisionへ書き直した。supersede対象のdecision fileは存在しない）は、実際には1本で両方を満たせるにもかかわらず不必要に分割していた。実browserでUIを操作しながら、裏では実TEST Drive folderへ実際に疎通する1本の方が、issue3の核心（「Access以外の本番相当経路が生きているか」をUI込みで検証する）に忠実である。
4. TEST Drive folderの中身は`generator-system.yml`（`docs/decisions/2026-08-30T16-23-00`）が実行のたびに書込・検証後削除する運用であり、folderが空の瞬間が起こりうる。件数を固定assertするとfolderの状態次第でsmokeがflakyになるため、「疎通が成立しているか」と「データがあるときの操作」を分けて確認する。

## 3. Rejected

1. **UI smoke（credential無し・dummy backend fixture）とcredential smoke（vitest・Node直接疎通）の2本立てを維持する案**（本session内で一度実装したが、commit・公開前に本Decisionへ書き直したためsupersede対象のdecision fileは存在しない） — shimの指摘により、1本で両方を満たせることが判明した。2本立てはUI操作と実疎通を別々にしか見ておらず、「UIを実際に操作した結果、実Driveのデータが正しく描画されるか」という統合的な確認ができていなかった。
2. **`wrangler dev`をTEST_*credentialで起動する案**（過去に一度検討し却下、今回も再考したが変更なし） — Vite dev server上で`createApp().fetch(req, env)`を直接呼べば足り、`wrangler dev`という追加プロセス・追加設定（`.dev.vars`等）は不要（YAGNI）。
3. **TEST用に専用middleware関数（例: `createTestDriveBackendMiddleware`）を新設する案** — `env`引数の中身を変えるだけで済むため、関数を増やす理由が無い。既存`createApp()`のシグネチャを何も変更せずに使える。
4. **TEST Drive folderに安定fixture episodeを常設し、件数・内容を固定assertする案** — `generator-system.yml`が同じfolderへ都度書込・削除するため、fixtureを常設すると`docs/decisions/2026-08-30T16-23-00`のSystem test postcondition（実行前後の差分で1 stemを見る）と衝突しうる。固定assertはfolderの運用実態と合わない。
