package errors

import "fmt"

// EpisodeIDMismatch は原稿内 episodeId と保存対象の stem が異なる。
type EpisodeIDMismatch struct {
	Expected string
	Actual   string
}

func (e *EpisodeIDMismatch) Error() string {
	if e == nil {
		return "episode id does not match manuscript"
	}
	return fmt.Sprintf("episode id %q does not match manuscript episodeId %q", e.Expected, e.Actual)
}
