package oauth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const tokenURL = "https://oauth2.googleapis.com/token"

// TokenSource は Google OAuth refresh で access token を取得する。
type TokenSource struct {
	client       *http.Client
	clientID     string
	clientSecret string
	refreshToken string
}

// NewTokenSource は Google OAuth TokenSource を返す。
//
// @require httpClient != nil。
// @ensure OAuth credential は refresh form の組み立てにだけ使い、保存元の知識は持たない。
func NewTokenSource(
	httpClient *http.Client,
	clientID, clientSecret, refreshToken string,
) *TokenSource {
	return &TokenSource{
		client:       httpClient,
		clientID:     clientID,
		clientSecret: clientSecret,
		refreshToken: refreshToken,
	}
}

// Token は refresh token を使って非空の access token を返す。
//
// @require ctx != nil。client != nil。
// @ensure 失敗時は元の原因を cause chain に保持した *Error を返す。
func (s *TokenSource) Token(ctx context.Context) (string, error) {
	if s == nil || s.client == nil {
		return "", infraErr("refresh", fmt.Errorf("client is nil"))
	}

	form := url.Values{
		"grant_type":    []string{"refresh_token"},
		"client_id":     []string{s.clientID},
		"client_secret": []string{s.clientSecret},
		"refresh_token": []string{s.refreshToken},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", infraErr("refresh_build", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	res, err := s.client.Do(req)
	if err != nil {
		return "", infraErr("refresh_do", err)
	}
	defer func() { _ = res.Body.Close() }()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return "", infraErr("refresh_read", err)
	}
	if res.StatusCode != http.StatusOK {
		return "", infraErr("refresh_http", fmt.Errorf("status %d", res.StatusCode))
	}

	var response struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.Unmarshal(raw, &response); err != nil {
		return "", infraErr("refresh_decode", err)
	}
	if strings.TrimSpace(response.AccessToken) == "" {
		return "", infraErr("refresh_decode", fmt.Errorf("access token is empty"))
	}
	return response.AccessToken, nil
}
