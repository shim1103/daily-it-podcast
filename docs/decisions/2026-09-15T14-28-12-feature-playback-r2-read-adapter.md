---
name: Cloudflare Workers native binding契約のAdapterはsociable unit testを基本とし、実binding peerでしか観測できない意味（SQLの意味、手書きbinding型と実bindingの構造的一致）を持つAdapterだけNarrow Integrationを足す
date: 2026-09-15T14:28:12
branch: feature/playback-r2-read-adapter
---

## 1. Decision

1. R2 binding（`R2Bucket`）・D1 binding のようなCloudflare Workers native binding契約（HTTPではなくJS関数呼び出し契約）に依存するDriven Adapterは、実HTTPサーバーを立てる形のNarrow Integration testを持たない。
2. binding契約のin-memory fakeや手書きdoubleを使ったtestは、実物のI/O境界（network・process境界）を通過しないため、sociable unit testと同一scopeとして扱い、別種別のtest fileへ分離しない。fakeを刺すだけのNarrow Integrationは持たない。
3. 実binding peer（`getPlatformProxy` の local binding。正は `2026-09-15T12-02-48-feature-generator-r2-write-adapter.md`・`2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md`）があり、Adapterの意味がその実体でしか観測できないときだけ、sociable unitに加えて実proxyを刺すNarrow Integrationを持つ。D1の `ProgressRepository` adapterがこれに当たり、binding呼び出しの形・Error写像・再試行なしはsociable unit、実SQLiteで通るSQL（保存した行の読み戻し・差分取得）・時刻offset混在の保存形・bind数上限はNarrow Integrationが所有する。勝ち側の判定（merge）はadapterでなくDomainが持つ（`2026-10-01T18-54-16-feature-playback-progress-d1-local-peer.md`）ので、Domainのunitが所有する。手書きのbinding型（`R2BucketBinding`・`D1DatabaseBinding`）と実bindingの構造的な一致も、実peerでしか観測できない意味に当たる。よってD1 adapterに加え、R2 adapterもNarrow Integrationの対象とする。
4. 実proxyを起動するNarrow Integrationは、unit projectから除外する。
5. 実peerを持てないAdapterについて、なぜreal境界を持てないかは、Adapter class doc commentへwhy commentとして明示する。

## 2. Reason

1. Google Drive Adapterのような外部HTTP API依存のAdapterは、custom fetchでURL hostをlocal listenerへ書き換え、実際にnode:httpサーバーへHTTPリクエストが飛ぶ形でNarrow Integration testを構成できる（`testing-strategy`の「実物はnetwork/TLS境界」を満たす）。
2. R2・D1 bindingはCloudflare Workers runtimeが提供するJS object契約であり、HTTP越しではない。したがってDrive同様の手法（実サーバーへの実際のHTTP到達）は原理的に適用できない。
3. Cloudflareの`miniflare`packageは`wrangler`のtransitive dependencyとしてnode_modules内に存在し、実際に`R2Bucket`相当のbinding実装を起動できることを検証した。しかし現行版の公開`new Miniflare(opts)`は独自nested config schemaを要求し、従来の平坦なoptionsでは動かない。動かすには非公開寄りの互換shim関数を経由する必要があり、安定したpublic APIとは言えない。
4. `miniflare`をproject依存へ加えると、`package.json`未記載のまま`wrangler`のhoisting構造へ暗黙依存するphantom dependencyになるか、明示devDependency化しても`wrangler`同梱版と別バージョン管理になり2つのCloudflare runtime実装がCIに乗る不安定要因が増える。実peerの手段は、この直接導入ではなく`getPlatformProxy`（上記2 Decision）で得る。
5. 実際にR2 binding のin-memory fakeを使ったNarrow Integration testを作成し、既存のsociable unit testと比較した結果、全ケースが完全に重複していた（実物境界を持たないtestは、fakeを直接使うsociable unit testと抽象度が同一のため）。したがってfakeを刺すNarrow Integrationは持たない。
6. 実binding peerが無かった時点では、Adapterのtestは全てfakeに寄るほかなく、Narrow Integrationを持たない答えで足りた。D1 adapterの価値は、実SQLiteで通るSQL（行の保存・差分取得・列の対応）、offset混在時刻の保存形、1 queryあたりのbind数上限にある。勝ち側の判定はDomainが持つため、この価値に含まない。これらは、呼び出し形を記録する手書きdoubleや、読取は空・書込はchanges 0を返すだけのFakeでは観測できない。sociable unitが緑でも、SQLの誤りを見逃す。
7. 手書きのbinding型と実bindingの構造的な一致は、`tsc`が検査しない。`D1Database`・`R2Bucket`が`apps/playback`の型scopeに無いためである（`tsconfig.json`の`types`が`["vite/client","node"]`で、`tsc --noEmit`で`TS2552`・`TS2304`になることを確認済み）。実bindingから手書き型への代入は、型検査のどこにも現れない。sociable unitは手書き型を満たすdoubleを刺すだけで、実物との一致は観測できない。実bindingを刺すNarrow Integrationだけが、この一致を守る。この点でR2 adapterとD1 adapterは対称であり、R2もSQLの有無に関わらずNarrow Integrationの対象になる。
8. 実proxyの起動をunit（Fast）へ混ぜないため、実peerを使うNarrow Integrationはunit projectから除外する（`2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md` の維持条）。

## 3. Rejected

1. Miniflareを導入し実際のR2 binding runtime実装に対してtestする案 — 非公開APIへの依存、phantom dependencyまたは二重runtime管理という導入コストが、得られる実境界検証の価値に見合わないため却下。
2. in-memory fakeのままNarrow Integration testを維持する案 — sociable unit testと完全重複し、「Narrow Integration」の名で虚偽のscope表示になるため却下。虚偽のscope表示は`testing-strategy`のtest種別ラベルの信頼性を損なう。
3. 実binding peerがあっても、binding契約Adapterは一律にNarrow Integrationを持たない（旧答え）— peerが無かった時の答えで、peerを得た今は、D1 adapterのSQL意味がsociable unitで観測できずSQLの誤りを見逃すため採らなくなった。さらに、手書きのbinding型と実bindingの一致は`tsc`もsociable unitも検査しないため、R2 adapterについても採らない。
4. SQLの意味を手書きdoubleで再現し、sociable unitへ寄せる案 — 検証したい意味をtest側へ複製することになり、複製が実SQLiteと一致する保証が無い。実peerなら複製が要らない。
5. 実proxyのNarrow Integrationをunit projectに含める案 — 実proxy起動がunitのFastを壊す（`2026-09-16T00-20-08-feature-generator-r2-test-peer-scope.md` で却下済み）。
