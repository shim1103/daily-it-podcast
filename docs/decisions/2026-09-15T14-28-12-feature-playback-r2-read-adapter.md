---
name: Cloudflare Workers native binding契約のAdapterはNarrow Integrationを持たずsociable unit testへ寄せる
date: 2026-09-15T14:28:12
branch: feature/playback-r2-read-adapter
---

## 1. Decision

1. R2 binding（`R2Bucket`）のようなCloudflare Workers native binding契約（HTTPではなくJS関数呼び出し契約）に依存するDriven Adapterは、実HTTPサーバーを立てる形のNarrow Integration testを持たない。
2. binding契約のin-memory fakeを使ったtestは、実物のI/O境界（network・process境界）を通過しないため、sociable unit testと同一scopeとして扱い、別種別のtest fileへ分離しない。
3. なぜbinding契約Adapterがreal境界を持てないかは、Adapter class doc commentへwhy commentとして明示する。

## 2. Reason

1. Google Drive Adapterのような外部HTTP API依存のAdapterは、custom fetchでURL hostをlocal listenerへ書き換え、実際にnode:httpサーバーへHTTPリクエストが飛ぶ形でNarrow Integration testを構成できる（`testing-strategy`の「実物はnetwork/TLS境界」を満たす）。
2. R2 bindingはCloudflare Workers runtimeが提供するJS object契約であり、HTTP越しではない。したがってDrive同様の手法（実サーバーへの実際のHTTP到達）は原理的に適用できない。
3. Cloudflareの`miniflare`packageは`wrangler`のtransitive dependencyとしてnode_modules内に存在し、実際に`R2Bucket`相当のbinding実装を起動できることを検証した。しかし現行版の公開`new Miniflare(opts)`は独自nested config schemaを要求し、従来の平坦なoptionsでは動かない。動かすには非公開寄りの互換shim関数を経由する必要があり、安定したpublic APIとは言えない。
4. `miniflare`をproject依存へ加えると、`package.json`未記載のまま`wrangler`のhoisting構造へ暗黙依存するphantom dependencyになるか、明示devDependency化しても`wrangler`同梱版と別バージョン管理になり2つのCloudflare runtime実装がCIに乗る不安定要因が増える。
5. 実際にR2 binding のin-memory fakeを使ったNarrow Integration testを作成し、既存のsociable unit testと比較した結果、全ケースが完全に重複していた（実物境界を持たないtestは、fakeを直接使うsociable unit testと抽象度が同一のため）。

## 3. Rejected

1. Miniflareを導入し実際のR2 binding runtime実装に対してtestする案 — 非公開APIへの依存、phantom dependencyまたは二重runtime管理という導入コストが、得られる実境界検証の価値に見合わないため却下。
2. in-memory fakeのままNarrow Integration testを維持する案 — sociable unit testと完全重複し、「Narrow Integration」の名で虚偽のscope表示になるため却下。虚偽のscope表示は`testing-strategy`のtest種別ラベルの信頼性を損なう。
