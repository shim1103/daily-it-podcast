---
name: "原稿TextWriterをどう実装するか？"
description: ApplicationにDomainとUseCaseを残し、AdapterはCloudflare AI Gatewayへの薄いHTTPに落とす。Cursor Cloud Agents RESTは捨て、model列・retry・fallbackはGateway設定側に置く。
---

## 1. Decision

1. 原稿（`port.TextWriter`）の実装境界を次に固定する。
   1. **Application**: Domain 仕様（`ManuscriptDraft` 検証、`buildFn`、invalid-draft 再試行の進行）と、取得元切替 UseCase。
   2. **Adapter（薄い）**: Gateway へ text 呼び出し 1 種と、全滅時の番兵／infra error への写像だけ。
   3. **Gateway（設定）**: model 列、fallback、retry、analytics / logs。Go の `writerretry` や Cursor/Gemini 個別の status 分類表は持たない。
2. 呼び出しは **Cloudflare AI Gateway** 経由にする。hosting / R2 と同じ Cloudflare アカウントに揃える。
3. 原稿経路の **Cursor Cloud Agents REST（`cursorapi`）は捨てる**。model は「安くて楽」を優先し、Gateway が出せる text モデルに限る。具体 model id は Adapter 定数または Gateway 設定を正とし、本 Decision に写さない。
4. 本 Decision は、自前 multi-provider（`geminiapi` + `cursorapi` + `writerretry`）を主とする先行判断のうち、**transport / retry / fallback / model 追従を自前に置く部分**を置き換える。Domain 検証を Application に置く判断と、切替 UseCase を Application に置く判断は維持する。
5. TTS 側の境界方針は `generator-tts-implementation` と同型にする（本 file は原稿側だけを答える）。

## 2. Reason

### なぜ薄い Adapter + 厚い Application か

長期保守で重いのは、model 版落ち・429 分類・Cursor SSE の opaque ERROR・無料/有料枠の監視を Go と Decision で抱え続けることである。AI Gateway は retry / fallback / analytics を既に持ち、R2 と同じ Cloudflare に寄せると dashboard・secret・運用面が一本になる。`ManuscriptDraft` の検証失敗の持ち越しは product 固有であり、Gateway に渡すと「安い text を返す管」以上の意味を持たせることになり、vendor 設定と Domain が混ざる。

### なぜ Cursor 原稿を捨てるか

原稿は brief→text（+ structured）で足り、Cursor の agent workflow（create / SSE / run 終端）は自前面が厚い。model を問わない前提では、opaque な Cursor 経路を残す長期コストが効用を上回る。

### なぜ retry / fallback の正を Gateway に移すか

自前の status 分類表と `writerretry` を残したまま Gateway も併用すると二重 SSOT になる。寄せるなら分類の正は Gateway 側に移す。

## 3. Rejected

1. **Cursor Cloud Agents REST を primary / fallback に残す案** — model 非固定方針と矛盾し、SSE/run 終端の自前面が残る。
2. **Cloudflare Agents SDK（Durable Object / Think）ごと載せる案** — 日次の brief→text には過剰。Gateway の HTTP + fallback で足りる。
3. **retry/fallback 分類を Go の `writerretry` に残したまま Gateway も併用する案** — 二重 SSOT。
4. **TextWriter だけ Gateway に寄せ、TTS は別境界のまま残す案** — 枠監視と Adapter 肉が二系統残り、寄せる目的が半減する（TTS 側は `generator-tts-implementation`）。
