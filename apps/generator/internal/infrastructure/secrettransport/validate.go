package secrettransport

import (
	"fmt"
	"net/url"
	"strings"
)

// ValidateAbsoluteHTTPSURL は targetURL が host を持つ絶対 https URL であることを検証する。
//
// @ensure targetURL が空白のみ、パース不能、scheme が https 以外、または host が空のとき error を返す。
func ValidateAbsoluteHTTPSURL(targetURL string) error {
	trimmed := strings.TrimSpace(targetURL)
	if trimmed == "" {
		return fmt.Errorf("TargetURL is empty")
	}
	u, err := url.Parse(trimmed)
	if err != nil || u.Scheme != "https" || u.Host == "" {
		return fmt.Errorf("TargetURL must be absolute https URL with host")
	}
	return nil
}
