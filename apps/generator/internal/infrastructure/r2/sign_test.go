package r2

import (
	"net/http"
	"testing"
	"time"
)

func TestSignV4Put_setsGoldenAuthorizationAndPayloadHash_whenNowAndCredentialsFixed(t *testing.T) {
	t.Parallel()

	fixed := time.Date(2026, 3, 15, 12, 30, 45, 0, time.UTC)
	payload := []byte(`{"episodeId":"golden-ep"}`)
	req, err := http.NewRequest(
		http.MethodPut,
		"https://golden-account-id.r2.cloudflarestorage.com/golden-bucket/golden-ep.json",
		nil,
	)
	if err != nil {
		t.Fatalf("NewRequest: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if err := signV4Put(req, payload, "golden-access-key-id", "golden-secret-access-key", fixed); err != nil {
		t.Fatalf("signV4Put: %v", err)
	}

	const wantPayloadHash = "6ecb86db6d0d92a73a6e4b2b792d1238d7443dda85ae4358999ec42ac52954e5"
	const wantAuthorization = "AWS4-HMAC-SHA256 Credential=golden-access-key-id/20260315/auto/s3/aws4_request, SignedHeaders=content-type;host;x-amz-content-sha256;x-amz-date, Signature=314b3c3c70cd031755f407f699a37f2bfda3d738808d546a526f9573384588f9"

	if got := req.Header.Get("X-Amz-Content-Sha256"); got != wantPayloadHash {
		t.Fatalf("X-Amz-Content-Sha256 = %q, want %q", got, wantPayloadHash)
	}
	if got := req.Header.Get("Authorization"); got != wantAuthorization {
		t.Fatalf("Authorization = %q, want %q", got, wantAuthorization)
	}
}
