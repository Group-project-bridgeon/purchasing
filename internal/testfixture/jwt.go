package testfixture

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Signer is synthetic test material, not Auth issuance or normal provisioning.
type Signer struct {
	Key        *rsa.PrivateKey
	PublicPath string
	UserID     uuid.UUID
}

func NewSigner(t *testing.T) *Signer {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal("synthetic RSA generation failed")
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatal("synthetic public key encoding failed")
	}
	path := filepath.Join(t.TempDir(), "public.pem")
	if err = os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der}), 0600); err != nil {
		t.Fatal("synthetic public key write failed")
	}
	return &Signer{Key: key, PublicPath: path, UserID: uuid.New()}
}
func (s *Signer) Token(t *testing.T, permissions []string, change func(jwt.MapClaims)) string {
	t.Helper()
	if permissions == nil {
		permissions = []string{}
	}
	now := time.Now().Unix()
	claims := jwt.MapClaims{"iss": "erp-auth", "aud": []string{"erp-api"}, "sub": s.UserID.String(), "user_id": s.UserID.String(), "email": "synthetic@example.invalid", "token_version": 0, "permissions": permissions, "iat": now, "nbf": now, "exp": now + 300}
	if change != nil {
		change(claims)
	}
	encoded, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.Key)
	if err != nil {
		t.Fatal("synthetic JWT signing failed")
	}
	return encoded
}
