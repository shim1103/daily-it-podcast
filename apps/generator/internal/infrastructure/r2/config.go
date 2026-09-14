package r2

// Credentials は R2（S3 互換）SigV4 用 access key ペア。
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
}

// Endpoint は path-style PutObject の Account ID と bucket。
type Endpoint struct {
	AccountID string
	Bucket    string
}
