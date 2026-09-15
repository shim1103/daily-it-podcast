// Package r2locals3 は generator Narrow Integration 用の experimental local S3 peer 契約である。
// 本番 Adapter ではない。
//
// 正: docs/decisions/2026-09-15T12-02-48-feature-generator-r2-write-adapter.md
//
// 既定 build（!r2locals3）の Start は Fake。実 wrangler 起動は `-tags r2locals3` の別 integration。
package r2locals3

import "context"

// Peer は local S3 互換 endpoint へ Writer / Lookup が刺すための注入値である。
//
// BaseURL は path-style S3 API の root（例: http://127.0.0.1:PORT/cdn-cgi/local/r2/s3）。
//
// @invariant AccessKey / Secret / AccountID / Bucket / BaseURL の実値を error message へ載せない。
type Peer struct {
	BaseURL         string
	AccessKeyID     string
	SecretAccessKey string
	AccountID       string
	Bucket          string
}

// Start は experimental local S3 peer を用意する。
//
// @require ctx は呼び出し元の寿命に従う。
// @ensure 既定 build は Fake Peer と no-op cleanup を返す（wrangler 未起動）。
// @ensure `-tags r2locals3` では実 endpoint の Peer と、必ず呼ぶ cleanup を返す。起動失敗は error（黙って skip しない）。
func Start(ctx context.Context) (peer *Peer, cleanup func(), err error) {
	return start(ctx)
}
