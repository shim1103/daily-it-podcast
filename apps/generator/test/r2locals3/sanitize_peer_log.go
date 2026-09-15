package r2locals3

import "strings"

// sanitizePeerLog は wrangler 起動 log から local peer 実値を伏せ、長さを制限する。
func sanitizePeerLog(raw string) string {
	out := raw
	for _, secret := range []string{localAccessKeyID, localSecretAccessKey, localBucket} {
		if secret == "" {
			continue
		}
		out = strings.ReplaceAll(out, secret, "[redacted]")
	}
	runes := []rune(out)
	if len(runes) > peerLogMaxRunes {
		out = string(runes[len(runes)-peerLogMaxRunes:])
	}
	return out
}
