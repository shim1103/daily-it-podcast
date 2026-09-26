package gemini

import (
	"net/http"
	"testing"
)

// TestNewSpeechSynthesizerWithTuning_fallsBackToDefaultTuning_whenZeroValueFields は
// Tuning のゼロ値 field が既定値へフォールバックすることを検証する。
func TestNewSpeechSynthesizerWithTuning_fallsBackToDefaultTuning_whenZeroValueFields(t *testing.T) {
	// Given: 空 Tuning を注入した Synthesizer
	synth := NewSpeechSynthesizerWithTuning(&http.Client{}, "gemini-fake-key", TierFree, Tuning{}, nil)

	// Then: 各 tuning field が既定値
	if synth.callGap != defaultCallGap {
		t.Fatalf("callGap = %v, want %v", synth.callGap, defaultCallGap)
	}
	if synth.retryBackoffBase != defaultRetryBackoffBase {
		t.Fatalf("retryBackoffBase = %v, want %v", synth.retryBackoffBase, defaultRetryBackoffBase)
	}
	if synth.retryBackoffMax != defaultRetryBackoffMax {
		t.Fatalf("retryBackoffMax = %v, want %v", synth.retryBackoffMax, defaultRetryBackoffMax)
	}
}

// TestNewSpeechSynthesizer_usesDefaultTuning_whenConstructedPlainly は
// 既定 constructor が既定 tuning を使う（挙動不変）ことを検証する。
func TestNewSpeechSynthesizer_usesDefaultTuning_whenConstructedPlainly(t *testing.T) {
	// Given / When: 既定 constructor
	synth := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierFree, nil)

	// Then: tuning field はすべて既定値
	if synth.callGap != defaultCallGap || synth.retryBackoffBase != defaultRetryBackoffBase || synth.retryBackoffMax != defaultRetryBackoffMax {
		t.Fatalf("tuning = {%v, %v, %v}, want defaults {%v, %v, %v}",
			synth.callGap, synth.retryBackoffBase, synth.retryBackoffMax,
			defaultCallGap, defaultRetryBackoffBase, defaultRetryBackoffMax)
	}
}

// TestNewSpeechSynthesizer_storesGivenTier_whenConstructed は
// 渡した Tier がそのまま保持されることを検証する（config 変数名からの暗黙判定をやめ
// 明示引数にした Decision 2026-09-16T11-41-59 §1-5 の契約）。
func TestNewSpeechSynthesizer_storesGivenTier_whenConstructed(t *testing.T) {
	// Given / When: TierPaid を渡して構築する
	synth := NewSpeechSynthesizer(&http.Client{}, "gemini-fake-key", TierPaid, nil)

	// Then: 渡した Tier がそのまま保持される
	if synth.tier != TierPaid {
		t.Fatalf("tier = %v, want %v", synth.tier, TierPaid)
	}
}
