package main

import (
	"strings"
	"testing"
)

func TestAssertCompletedSetsOnly_returnsNil_whenOnlyMatchingJsonMp3SetsExist(t *testing.T) {
	// Given: stem 一致の json+mp3 だけが並ぶ
	files := []listedFile{
		{ID: "j1", Name: "a.json"},
		{ID: "m1", Name: "a.mp3"},
		{ID: "j2", Name: "b.json"},
		{ID: "m2", Name: "b.mp3"},
	}

	// When: assertCompletedSetsOnly を呼ぶ
	err := assertCompletedSetsOnly(files)

	// Then: error なし
	if err != nil {
		t.Fatalf("assertCompletedSetsOnly: err = %v, want nil", err)
	}
}

func TestAssertCompletedSetsOnly_returnsNil_whenFolderEmpty(t *testing.T) {
	// Given: 空一覧
	files := []listedFile{}

	// When: assertCompletedSetsOnly を呼ぶ
	err := assertCompletedSetsOnly(files)

	// Then: error なし（違反 object が無い）
	if err != nil {
		t.Fatalf("assertCompletedSetsOnly: err = %v, want nil", err)
	}
}

func TestAssertCompletedSetsOnly_returnsError_whenNonJsonMp3ObjectExists(t *testing.T) {
	// Given: 完成ペアに加えて別 suffix の object がある
	files := []listedFile{
		{ID: "j1", Name: "a.json"},
		{ID: "m1", Name: "a.mp3"},
		{ID: "x1", Name: "notes.txt"},
	}

	// When: assertCompletedSetsOnly を呼ぶ
	err := assertCompletedSetsOnly(files)

	// Then: 完成ペア以外を示す error
	if err == nil {
		t.Fatal("assertCompletedSetsOnly: err = nil, want 完成ペア以外 error")
	}
	if !strings.Contains(err.Error(), "notes.txt") {
		t.Fatalf("assertCompletedSetsOnly: err = %v, want notes.txt を含む", err)
	}
}

func TestAssertCompletedSetsOnly_returnsError_whenWavRemains(t *testing.T) {
	// Given: 完成ペアに加えて wav がある
	files := []listedFile{
		{ID: "j1", Name: "a.json"},
		{ID: "m1", Name: "a.mp3"},
		{ID: "w1", Name: "a.wav"},
	}

	// When: assertCompletedSetsOnly を呼ぶ
	err := assertCompletedSetsOnly(files)

	// Then: wav 名を含む error
	if err == nil {
		t.Fatal("assertCompletedSetsOnly: err = nil, want wav 残存 error")
	}
	if !strings.Contains(err.Error(), "a.wav") {
		t.Fatalf("assertCompletedSetsOnly: err = %v, want a.wav を含む", err)
	}
}

func TestAssertCompletedSetsOnly_returnsError_whenMp3MissingForJson(t *testing.T) {
	// Given: json だけで同 stem の mp3 が無い
	files := []listedFile{{ID: "j1", Name: "a.json"}}

	// When: assertCompletedSetsOnly を呼ぶ
	err := assertCompletedSetsOnly(files)

	// Then: mp3 欠落を示す error
	if err == nil {
		t.Fatal("assertCompletedSetsOnly: err = nil, want mp3 欠落 error")
	}
	if !strings.Contains(err.Error(), "a.json") {
		t.Fatalf("assertCompletedSetsOnly: err = %v, want a.json を含む", err)
	}
}

func TestAssertCompletedSetsOnly_returnsError_whenJsonMissingForMp3(t *testing.T) {
	// Given: mp3 だけで同 stem の json が無い
	files := []listedFile{{ID: "m1", Name: "a.mp3"}}

	// When: assertCompletedSetsOnly を呼ぶ
	err := assertCompletedSetsOnly(files)

	// Then: json 欠落を示す error
	if err == nil {
		t.Fatal("assertCompletedSetsOnly: err = nil, want json 欠落 error")
	}
	if !strings.Contains(err.Error(), "a.mp3") {
		t.Fatalf("assertCompletedSetsOnly: err = %v, want a.mp3 を含む", err)
	}
}

func TestWavIDs_returnsOnlyWavFileIDs(t *testing.T) {
	// Given: json / mp3 / wav が混在する一覧
	files := []listedFile{
		{ID: "j1", Name: "a.json"},
		{ID: "m1", Name: "a.mp3"},
		{ID: "w1", Name: "a.wav"},
		{ID: "w2", Name: "b.wav"},
	}

	// When: wavIDs を呼ぶ
	got := wavIDs(files)

	// Then: wav の ID だけが順に返る
	if len(got) != 2 || got[0] != "w1" || got[1] != "w2" {
		t.Fatalf("wavIDs = %#v, want [w1 w2]", got)
	}
}
