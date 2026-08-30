// Package errors は Generator の Domain Error を単一型で表現する。
package errors

// why: Domain は package 単位でなく層単位の分類なので層名を固定 prefix にする。
const errorPrefix = "generator domain"

// why: typo は compile で捕まらないので Op の語彙を定数へ固定する。
const (
	OpEmptyEpisodeID         = "empty_episode_id"
	OpEmptyAudio             = "empty_audio"
	OpInvalidManuscript      = "invalid_manuscript"
	OpInvalidManuscriptDraft = "invalid_manuscript_draft"
	OpNoSourceItems          = "no_source_items"
	OpEpisodeIDMismatch      = "episode_id_mismatch"
	OpCorruptSpeechAudio     = "corrupt_speech_audio"
	// OpInconsistentEpisodeAssembly は episode 組み立て時の segment 数・topic 数などの内部不整合。
	// build helper が検出する。
	OpInconsistentEpisodeAssembly = "inconsistent_episode_assembly"
)

type Error struct {
	Op  string
	Err error
}

// Error は "<prefix>: <op>: <詳細>" 形式を返す。cause が無ければ詳細を省く。
func (e *Error) Error() string {
	if e == nil {
		return errorPrefix + ": <nil>"
	}
	if e.Err == nil {
		return errorPrefix + ": " + e.Op
	}
	return errorPrefix + ": " + e.Op + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// why: application 層が Domain Error を構築するため exported。infraErr / configErr と <層>Err で対称。
func DomainErr(op string, err error) *Error {
	return &Error{Op: op, Err: err}
}
