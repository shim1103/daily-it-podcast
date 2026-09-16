//go:build !r2locals3

package r2locals3

import "context"

// start は Sociable Unit / 既定 build 向け Fake。wrangler を起動しない。
func start(ctx context.Context) (*Peer, func(), error) {
	_ = ctx
	return &Peer{}, func() {}, nil
}
