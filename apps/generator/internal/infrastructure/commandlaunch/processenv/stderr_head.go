package processenv

import (
	"bytes"
	"fmt"
	"strings"
)

// what: 失敗診断に載せる child stderr の先頭上限（byte）。
const stderrHeadLimit = 300

func withStderrHead(err error, stderr []byte, secretValue string, stdin []byte) error {
	head := stderrHead(stderr, secretValue, stdin)
	if head == "" {
		return err
	}
	return fmt.Errorf("%w: stderr_head=%s", err, head)
}

func stderrHead(stderr []byte, secretValue string, stdin []byte) string {
	if len(stderr) == 0 {
		return ""
	}
	if secretValue != "" && bytes.Contains(stderr, []byte(secretValue)) {
		return ""
	}
	if len(stdin) > 0 && bytes.Contains(stderr, stdin) {
		return ""
	}
	head := stderr
	if len(head) > stderrHeadLimit {
		head = head[:stderrHeadLimit]
	}
	return oneLineBytes(head)
}

func oneLineBytes(b []byte) string {
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	return s
}
