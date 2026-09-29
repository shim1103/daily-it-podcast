package contracts

import _ "embed"

// why: go:embed は module を越えられない。JSON と同 dir に byte 列だけ公開する。
// WriterOutput の responseSchema は generator の models.WriterOutput（型から生成）。
// contracts は完成 manuscript など generator / playback 共有 wire だけを持つ。

//go:embed manuscript.schema.json
var ManuscriptSchema []byte
