package config

// secretRedaction はSecretの表示時にraw valueの代わりへ出力する固定表記である。
const secretRedaction = "[REDACTED]"

// secretValue はSecret契約を満たすopaqueなcredential保持体である。
//
// @invariant raw valueはRevealでだけ取り出せ、String / GoStringは露出しない。
type secretValue struct {
	raw string
}

// newSecret はraw valueを保持するSecretを生成する。
//
// @ensure 戻り値のRevealはrawと一致し、String / GoStringはrawを含まない。
func newSecret(raw string) Secret {
	return secretValue{raw: raw}
}

// String はraw valueを露出せず固定のredaction表記を返す。
func (s secretValue) String() string {
	return secretRedaction
}

// GoString はraw valueを露出せず固定のredaction表記を返す。
func (s secretValue) GoString() string {
	return secretRedaction
}

// Reveal はraw valueを返す。
func (s secretValue) Reveal() string {
	return s.raw
}
