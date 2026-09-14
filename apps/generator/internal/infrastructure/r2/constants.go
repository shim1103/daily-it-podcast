package r2

const (
	jsonMIME = "application/json"
	mp3MIME  = "audio/mpeg"

	jsonExt = ".json"
	mp3Ext  = ".mp3"

	// maxPutAttempts は 1 object の put 試行上限。初回 + 1 回 retry（情報源 Adapter と同型）。
	maxPutAttempts = 2

	// awsRegion / awsService は R2 の S3 互換 SigV4 固定値。
	awsRegion  = "auto"
	awsService = "s3"
)
