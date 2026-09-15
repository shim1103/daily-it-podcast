package r2

import "strings"

// endpointOverride は Writer / Lookup が共有する S3 互換 endpoint base 上書き。
type endpointOverride struct {
	endpointBase string
}

func (e endpointOverride) withBase(base string) endpointOverride {
	return endpointOverride{endpointBase: strings.TrimSpace(base)}
}
