package errors

// EmptyEpisodeID は episodeID（stem）が空である。
type EmptyEpisodeID struct{}

func (e *EmptyEpisodeID) Error() string {
	return "episode id is empty"
}
