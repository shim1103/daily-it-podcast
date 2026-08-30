package composition

import (
	"crypto/rand"
	"encoding/hex"
)

// newEpisodeID は RFC4122 version 4（乱数）UUID を xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx 形式の
// 小文字 hex 文字列で返す。ProduceEpisode へ注入する不透明 episodeId 発行器の production 実体である。
//
// @ensure 戻りは RFC4122 v4 書式（version nibble = 4、variant nibble ∈ 8..b）の 36 文字。
// @invariant crypto/rand.Read の error は panic する。起動後の乱数枯渇は回復不能で、
//
//	ここは Composition Root なので fallback を持たない。
func newEpisodeID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic("composition: crypto/rand read failed: " + err.Error())
	}
	b[6] = (b[6] & 0x0f) | 0x40 // version 4
	b[8] = (b[8] & 0x3f) | 0x80 // variant 10

	var buf [36]byte
	hex.Encode(buf[0:8], b[0:4])
	buf[8] = '-'
	hex.Encode(buf[9:13], b[4:6])
	buf[13] = '-'
	hex.Encode(buf[14:18], b[6:8])
	buf[18] = '-'
	hex.Encode(buf[19:23], b[8:10])
	buf[23] = '-'
	hex.Encode(buf[24:36], b[10:16])
	return string(buf[:])
}
