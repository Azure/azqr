// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package pipeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Azure/azqr/internal/models"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/golang-jwt/jwt/v5"
)

// fakeCredential is a minimal azcore.TokenCredential test double that either returns a
// fixed token or a fixed error, letting tests force both the success and failure paths of
// InitializationStage's tenant ID capture without any network access.
type fakeCredential struct {
	token string
	err   error
}

func (f *fakeCredential) GetToken(ctx context.Context, opts policy.TokenRequestOptions) (azcore.AccessToken, error) {
	if f.err != nil {
		return azcore.AccessToken{}, f.err
	}
	return azcore.AccessToken{Token: f.token, ExpiresOn: time.Now().Add(time.Hour)}, nil
}

// unverifiedJWT builds a syntactically valid, unsigned JWT carrying the given claims, which is
// all captureTenantID needs since it never verifies the token's signature.
func unverifiedJWT(t *testing.T, claims jwt.MapClaims) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	signed, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}
	return signed
}

func newInitStageCtx(cred azcore.TokenCredential) *ScanContext {
	stages := models.NewStageConfigsWithDefaults()
	return &ScanContext{
		Ctx:    context.Background(),
		Cred:   cred,
		Params: &models.ScanParams{Stages: stages},
	}
}

func TestInitializationStageCapturesTenantIDFromToken(t *testing.T) {
	s := NewInitializationStage()
	ctx := newInitStageCtx(&fakeCredential{token: unverifiedJWT(t, jwt.MapClaims{"tid": "tenant-123"})})

	got := s.captureTenantID(ctx)
	if got != "tenant-123" {
		t.Errorf("captureTenantID() = %q, want %q", got, "tenant-123")
	}
}

func TestInitializationStageTenantCaptureFailureIsNonFatal(t *testing.T) {
	s := NewInitializationStage()

	ctx := newInitStageCtx(&fakeCredential{err: errors.New("token acquisition failed")})
	if got := s.captureTenantID(ctx); got != "" {
		t.Errorf("captureTenantID() with token error = %q, want empty", got)
	}

	ctx2 := newInitStageCtx(&fakeCredential{token: "not-a-jwt"})
	if got := s.captureTenantID(ctx2); got != "" {
		t.Errorf("captureTenantID() with malformed token = %q, want empty", got)
	}

	ctx3 := newInitStageCtx(&fakeCredential{token: unverifiedJWT(t, jwt.MapClaims{"sub": "no-tid-here"})})
	if got := s.captureTenantID(ctx3); got != "" {
		t.Errorf("captureTenantID() with missing tid claim = %q, want empty", got)
	}
}
