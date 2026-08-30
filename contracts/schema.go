package contracts

import _ "embed"

// why: go:embed は module を越えられない。JSON と同 dir に byte 列だけ公開する。
//
//go:embed manuscript.schema.json
var ManuscriptSchema []byte
