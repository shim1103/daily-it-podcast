package r2locals3

// local S3 peer の注入値。本番 credential ではない。
const (
	localAccessKeyID     = "local-access-key-id"
	localSecretAccessKey = "local-secret-access-key"
	localAccountID       = "local"
	localBucket          = "generator-r2-local-s3"
	s3APIPathPrefix      = "/cdn-cgi/local/r2/s3"
	peerLogMaxRunes      = 400
)
