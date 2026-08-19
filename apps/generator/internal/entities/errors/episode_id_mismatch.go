package errors

// EpisodeIDMismatch は JSON 内 episodeId とファイル stem が一致しない。
type EpisodeIDMismatch struct {
	Stem      string
	EpisodeID string
}

func (e *EpisodeIDMismatch) Error() string {
	return "episodeId " + e.EpisodeID + " does not match stem " + e.Stem
}
