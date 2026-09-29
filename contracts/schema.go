package contracts

import _ "embed"

// why: go:embed は module を越えられない。JSON と同 dir に byte 列だけ公開する。
// WriterOutput schema は generator の entities/models 正本（models.WriterOutputSchema）。
// contracts は完成 manuscript など generator / playback 共有 wire だけを持つ。

//go:embed manuscript.schema.json
var ManuscriptSchema []byte
