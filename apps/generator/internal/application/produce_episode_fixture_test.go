package application_test

import (
	"encoding/binary"
	"encoding/json"
	"math"
	"strings"
	"testing"

	"github.com/shim1103/daily-it-podcast/apps/generator/internal/entities/constants"
)

// wireTopic は buildWireJSON が組む topic 1 件分の素材。
type wireTopic struct {
	Title   string `json:"title"`
	Preface string `json:"preface"`
	Detail  string `json:"detail"`
}

// jaRunes は日本語 rune（ひらがな）を n 個並べた文字列を返す。
func jaRunes(n int) string {
	return strings.Repeat("あ", n)
}

// jaSentence は日本語 n rune + 末尾句点（合計 n+1 rune）の朗読 field を返す。
func jaSentence(n int) string {
	return jaRunes(n) + string(constants.DraftSentenceSuffixRune)
}

// validWireTopicCount は buildValidWireJSON が常に用いる topic 数。
// DraftTopicCountMax にすると各朗読 field を min 長のままでも合計文字数が total range の
// 下限を満たせる（Min-1 等の逸脱を避けつつ「複数 topic」の性質も持つ）。
const validWireTopicCount = constants.DraftTopicCountMax

// buildValidWireJSON は ManuscriptDraftFromWriterOutput の検証を通る wire JSON を組む。
// 各朗読 field は min 長ちょうど。topic 数は validWireTopicCount。
// title / preface / detail に topic ごとの識別 suffix を入れ、Run が topic 順を保つことを test 側で識別できるようにする。
func buildValidWireJSON() string {
	introRunes := constants.DraftIntroMinLen - 1     // + prefix でちょうど min
	closingRunes := constants.DraftClosingMinLen - 1 // 同上
	prefaceRunes := constants.DraftTopicPrefaceMinLen - 5
	// why: preface / detail 下限を下げたあと、各 field を min 付近にすると total 下限に届かない。
	// detail を足して total min を満たす（validateTotalChars を fixture が通るための調整）。
	detailPad := 40
	detailRunes := constants.DraftTopicDetailMinLen - 6 + detailPad

	topics := make([]wireTopic, validWireTopicCount)
	for i := 0; i < validWireTopicCount; i++ {
		suffix := string(rune('０' + i)) // 全角数字 1 rune で識別
		topics[i] = wireTopic{
			Title:   jaRunes(constants.DraftTopicTitleMinLen-1) + suffix,
			Preface: "まえおき" + suffix + jaSentence(prefaceRunes),
			Detail:  "しょうさい" + suffix + jaSentence(detailRunes),
		}
	}
	doc := map[string]any{
		"title":          jaRunes(constants.DraftTitleMinLen + 2),
		"intro":          "どうにゅう" + jaSentence(introRunes),
		"topics":         topics,
		"closingSummary": "まとめ" + jaSentence(closingRunes),
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	return string(raw)
}

// fixedWav の PCM パラメータ（16-bit PCM / mono / 8000Hz）。
const (
	fixtureSampleRate    = 8000
	fixtureChannels      = 1
	fixtureBitsPerSample = 16
)

// fixedWavOfDuration は指定秒数のゼロ埋め data を持つ 44 byte header の RIFF/WAVE を返す。
// 全 segment で同一 PCM パラメータ・既知尺にするので startSec を算術で予測できる。
func fixedWavOfDuration(t *testing.T, durationSec float64) []byte {
	t.Helper()
	blockAlign := fixtureChannels * fixtureBitsPerSample / 8
	byteRate := fixtureSampleRate * blockAlign
	dataLen := int(math.Round(durationSec*float64(fixtureSampleRate))) * blockAlign

	out := make([]byte, 44+dataLen)
	copy(out[0:4], "RIFF")
	binary.LittleEndian.PutUint32(out[4:8], uint32(36+dataLen))
	copy(out[8:12], "WAVE")
	copy(out[12:16], "fmt ")
	binary.LittleEndian.PutUint32(out[16:20], 16)
	binary.LittleEndian.PutUint16(out[20:22], 1)
	binary.LittleEndian.PutUint16(out[22:24], fixtureChannels)
	binary.LittleEndian.PutUint32(out[24:28], fixtureSampleRate)
	binary.LittleEndian.PutUint32(out[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(out[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(out[34:36], fixtureBitsPerSample)
	copy(out[36:40], "data")
	binary.LittleEndian.PutUint32(out[40:44], uint32(dataLen))
	return out
}

// manuscriptDoc は Run が WriteEpisode へ渡した manuscript bytes の検証用 shape。
type manuscriptDoc struct {
	EpisodeID   string  `json:"episodeId"`
	Date        string  `json:"date"`
	Title       string  `json:"title"`
	DurationSec float64 `json:"durationSec"`
	Body        struct {
		Opening string `json:"opening"`
		Topics  []struct {
			Title    string  `json:"title"`
			Preface  string  `json:"preface"`
			Detail   string  `json:"detail"`
			StartSec float64 `json:"startSec"`
		} `json:"topics"`
		Closing string `json:"closing"`
	} `json:"body"`
}

// wireFieldsOf は wire JSON string から検証で参照する top-level 朗読 field を取り出す。
func wireFieldsOf(t *testing.T, wire string) (title, closingSummary string) {
	t.Helper()
	var v struct {
		Title          string `json:"title"`
		ClosingSummary string `json:"closingSummary"`
	}
	if err := json.Unmarshal([]byte(wire), &v); err != nil {
		t.Fatalf("wire Unmarshal: %v", err)
	}
	return v.Title, v.ClosingSummary
}

// wireTitleOf は wire JSON string から title field を取り出す。
func wireTitleOf(t *testing.T, wire string) string {
	t.Helper()
	title, _ := wireFieldsOf(t, wire)
	return title
}

// wireClosingSummaryOf は wire JSON string から closingSummary field を取り出す。
func wireClosingSummaryOf(t *testing.T, wire string) string {
	t.Helper()
	_, closing := wireFieldsOf(t, wire)
	return closing
}

func unmarshalManuscript(t *testing.T, raw []byte) manuscriptDoc {
	t.Helper()
	var m manuscriptDoc
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatalf("manuscript bytes Unmarshal: %v\n%s", err, raw)
	}
	return m
}
