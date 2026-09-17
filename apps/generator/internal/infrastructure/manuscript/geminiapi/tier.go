package geminiapi

// Tier は TextWriter が使う Gemini API key の課金区分を明示する識別子である。
// config の変数名からの暗黙判定を避け、Composition が呼び出し時に明示で渡す
// （gemini（TTS）の Tier と同じ意図。Decision 2026-09-16T11-41-59）。現時点では
// TierPaid が挙動を変えるわけではなく、constructor が受け取る識別子として保持
// するだけ（scope-split A：境界契約のみ）。
type Tier int

const (
	TierFree Tier = iota
	TierPaid
)
