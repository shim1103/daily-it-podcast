package geminiapi

// Tier は TextWriter が使う Gemini API key の課金区分を明示する識別子である。
// config の変数名からの暗黙判定を避け、Composition が呼び出し時に明示で渡す
// （gemini（TTS）の Tier と同じ意図。Decision 2026-09-16T11-41-59）。
// 429 の枯渇分類は課金区分で変えない。
type Tier int

const (
	TierFree Tier = iota
	TierPaid
)
