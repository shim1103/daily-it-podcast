package main

import (
	"fmt"
	"strings"
)

type listedFile struct {
	ID   string
	Name string
}

const (
	wavExt  = ".wav"
	mp3Ext  = ".mp3"
	jsonExt = ".json"
)

func wavIDs(files []listedFile) []string {
	var ids []string
	for _, f := range files {
		if strings.HasSuffix(f.Name, wavExt) {
			ids = append(ids, f.ID)
		}
	}
	return ids
}

// assertCompletedSetsOnly は folder 直下が「同 stem の .json+.mp3」set だけであることを検証する。
//
// @require files は同一 folder の一覧想定。
// @ensure .json / .mp3 以外が1件でもあれば error。同 stem のペア欠落も error。
func assertCompletedSetsOnly(files []listedFile) error {
	jsonStems := make(map[string]struct{})
	mp3Stems := make(map[string]struct{})
	for _, f := range files {
		switch {
		case strings.HasSuffix(f.Name, jsonExt):
			stem := strings.TrimSuffix(f.Name, jsonExt)
			if stem == "" {
				return fmt.Errorf("完成ペア以外の object がある: %s", f.Name)
			}
			jsonStems[stem] = struct{}{}
		case strings.HasSuffix(f.Name, mp3Ext):
			stem := strings.TrimSuffix(f.Name, mp3Ext)
			if stem == "" {
				return fmt.Errorf("完成ペア以外の object がある: %s", f.Name)
			}
			mp3Stems[stem] = struct{}{}
		default:
			return fmt.Errorf("完成ペア以外の object がある: %s", f.Name)
		}
	}
	for stem := range jsonStems {
		if _, ok := mp3Stems[stem]; !ok {
			return fmt.Errorf("完成ペア欠落: %s.json に対応する .mp3 が無い", stem)
		}
	}
	for stem := range mp3Stems {
		if _, ok := jsonStems[stem]; !ok {
			return fmt.Errorf("完成ペア欠落: %s.mp3 に対応する .json が無い", stem)
		}
	}
	return nil
}
