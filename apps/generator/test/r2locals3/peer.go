// Package r2locals3 は generator Narrow Integration 用の experimental local S3 peer 契約である。
// 本番 Adapter ではない。起動本体・gate 配線は C1 Issue の C が埋める。
//
// 正: docs/decisions/2026-09-15T12-02-48-feature-generator-r2-write-adapter.md
package r2locals3

import "context"

// Peer は local S3 互換 endpoint へ Writer / Lookup が刺すための注入値である。
//
// @invariant AccessKey / Secret / AccountID / Bucket / BaseURL の実値を error message へ載せない（C 実装時も維持）。
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
// @ensure 本 stub は zero Peer と no-op cleanup と nil error を返す（wrangler 未起動）。
// @ensure C は実 endpoint を返す Peer と、必ず呼ぶ cleanup を返す。
func Start(ctx context.Context) (peer *Peer, cleanup func(), err error) {
	_ = ctx
	return &Peer{}, func() {}, nil
}
