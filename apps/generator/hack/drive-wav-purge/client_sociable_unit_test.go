package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

type stubTokens struct {
	token string
}

func (s stubTokens) Token(context.Context) (string, error) {
	return s.token, nil
}

func TestPurgeClient_purgeWavAndVerify_deletesWavThenRequiresJsonMp3Sets(t *testing.T) {
	// Given: list 1 回目は json+mp3+wav、delete 後の list は同 stem の json+mp3 のみ。Bearer tok-1
	var deleted []string
	var gotAuth []string
	listCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = append(gotAuth, r.Header.Get("Authorization"))
		switch {
		case r.Method == http.MethodGet && strings.Contains(r.URL.Path, "/drive/v3/files"):
			listCalls++
			w.Header().Set("Content-Type", "application/json")
			if listCalls == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"files": []map[string]string{
						{"id": "j1", "name": "ep.json"},
						{"id": "m1", "name": "ep.mp3"},
						{"id": "w1", "name": "ep.wav"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": []map[string]string{
					{"id": "j1", "name": "ep.json"},
					{"id": "m1", "name": "ep.mp3"},
				},
			})
		case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/drive/v3/files/"):
			id := strings.TrimPrefix(r.URL.Path, "/drive/v3/files/")
			id, _ = url.PathUnescape(id)
			deleted = append(deleted, id)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected %s %s", r.Method, r.URL.String())
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)
	client := &purgeClient{
		http:     srv.Client(),
		tokens:   stubTokens{token: "tok-1"},
		folderID: "folder-1",
		filesURL: srv.URL + "/drive/v3/files",
	}

	// When: purgeWavAndVerify を呼ぶ
	err := client.purgeWavAndVerify(context.Background())

	// Then: error なし。list 2 回。wav id だけ delete。全要求が Bearer tok-1
	if err != nil {
		t.Fatalf("purgeWavAndVerify: err = %v, want nil", err)
	}
	if listCalls != 2 {
		t.Fatalf("listCalls = %d, want 2", listCalls)
	}
	if len(deleted) != 1 || deleted[0] != "w1" {
		t.Fatalf("deleted = %#v, want [w1]", deleted)
	}
	for i, auth := range gotAuth {
		if auth != "Bearer tok-1" {
			t.Fatalf("gotAuth[%d] = %q, want Bearer tok-1", i, auth)
		}
	}
}

func TestPurgeClient_purgeWavAndVerify_returnsError_whenSetIncompleteAfterDelete(t *testing.T) {
	// Given: list 1 回目は json+wav、delete 後も mp3 が無く json だけ
	listCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			listCalls++
			w.Header().Set("Content-Type", "application/json")
			if listCalls == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"files": []map[string]string{
						{"id": "j1", "name": "ep.json"},
						{"id": "w1", "name": "ep.wav"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": []map[string]string{
					{"id": "j1", "name": "ep.json"},
				},
			})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)
	client := &purgeClient{
		http:     srv.Client(),
		tokens:   stubTokens{token: "tok"},
		folderID: "folder",
		filesURL: srv.URL + "/drive/v3/files",
	}

	// When: purgeWavAndVerify を呼ぶ
	err := client.purgeWavAndVerify(context.Background())

	// Then: 完成ペア欠落の error
	if err == nil {
		t.Fatal("purgeWavAndVerify: err = nil, want 完成ペア欠落")
	}
	if !strings.Contains(err.Error(), "ep.json") {
		t.Fatalf("purgeWavAndVerify: err = %v, want ep.json を含む", err)
	}
}

func TestPurgeClient_purgeWavAndVerify_returnsError_whenForeignObjectRemainsAfterDelete(t *testing.T) {
	// Given: delete 後の list に json+mp3 以外（notes.txt）が残る
	listCalls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodGet:
			listCalls++
			w.Header().Set("Content-Type", "application/json")
			if listCalls == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"files": []map[string]string{
						{"id": "j1", "name": "ep.json"},
						{"id": "m1", "name": "ep.mp3"},
						{"id": "w1", "name": "ep.wav"},
						{"id": "x1", "name": "notes.txt"},
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"files": []map[string]string{
					{"id": "j1", "name": "ep.json"},
					{"id": "m1", "name": "ep.mp3"},
					{"id": "x1", "name": "notes.txt"},
				},
			})
		case r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)
	client := &purgeClient{
		http:     srv.Client(),
		tokens:   stubTokens{token: "tok"},
		folderID: "folder",
		filesURL: srv.URL + "/drive/v3/files",
	}

	// When: purgeWavAndVerify を呼ぶ
	err := client.purgeWavAndVerify(context.Background())

	// Then: 完成ペア以外を示す error
	if err == nil {
		t.Fatal("purgeWavAndVerify: err = nil, want 完成ペア以外 error")
	}
	if !strings.Contains(err.Error(), "notes.txt") {
		t.Fatalf("purgeWavAndVerify: err = %v, want notes.txt を含む", err)
	}
}
