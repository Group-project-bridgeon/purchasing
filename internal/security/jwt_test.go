package security

import (
	"testing"
	"time"

	"github.com/Group-project-bridgeon/purchasing/internal/testfixture"
	"github.com/golang-jwt/jwt/v5"
)

func TestHumanJWTContract(t *testing.T) {
	s := testfixture.NewSigner(t)
	v, err := Load(s.PublicPath, "erp-auth", "erp-api")
	if err != nil {
		t.Fatal("load synthetic public verifier")
	}
	if _, err = v.Verify(s.Token(t, []string{"purchasing.purchase-orders:read"}, nil)); err != nil {
		t.Fatal("compatible human token rejected")
	}
	for name, change := range map[string]func(jwt.MapClaims){
		"issuer": func(c jwt.MapClaims) { c["iss"] = "other" }, "audience": func(c jwt.MapClaims) { c["aud"] = []string{"inventory-reservations"} },
		"multiple_audiences": func(c jwt.MapClaims) { c["aud"] = []string{"erp-api", "other"} },
		"missing_identity":   func(c jwt.MapClaims) { delete(c, "user_id") }, "subject": func(c jwt.MapClaims) { c["sub"] = "machine:" + s.UserID.String() },
		"missing_version": func(c jwt.MapClaims) { delete(c, "token_version") }, "negative_version": func(c jwt.MapClaims) { c["token_version"] = -1 },
		"missing_iat": func(c jwt.MapClaims) { delete(c, "iat") }, "missing_nbf": func(c jwt.MapClaims) { delete(c, "nbf") }, "missing_exp": func(c jwt.MapClaims) { delete(c, "exp") },
		"numeric_string": func(c jwt.MapClaims) { c["iat"] = "1234" }, "overlong": func(c jwt.MapClaims) { c["exp"] = time.Now().Unix() + 901 },
		"expired": func(c jwt.MapClaims) {
			c["iat"] = time.Now().Unix() - 600
			c["nbf"] = c["iat"]
			c["exp"] = time.Now().Unix() - 1
		},
		"future":           func(c jwt.MapClaims) { c["iat"] = time.Now().Unix() + 60; c["nbf"] = c["iat"] },
		"permissions_null": func(c jwt.MapClaims) { c["permissions"] = nil }, "malformed_permission": func(c jwt.MapClaims) { c["permissions"] = []string{"purchasing.purchase-orders"} },
		"machine_kind": func(c jwt.MapClaims) { c["token_kind"] = "machine" }, "machine_client": func(c jwt.MapClaims) { c["client_id"] = s.UserID.String() },
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := v.Verify(s.Token(t, nil, change)); err == nil {
				t.Fatal("invalid human JWT accepted")
			}
		})
	}
	other := testfixture.NewSigner(t)
	if _, err = v.Verify(other.Token(t, nil, nil)); err == nil {
		t.Fatal("wrong signature accepted")
	}
	hs, err := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{"iss": "erp-auth"}).SignedString([]byte("synthetic-test-only"))
	if err != nil {
		t.Fatal("synthetic HS signing failed")
	}
	if _, err = v.Verify(hs); err == nil {
		t.Fatal("wrong algorithm accepted")
	}
	for _, path := range []string{"", s.PublicPath + "-missing"} {
		if _, err := Load(path, "erp-auth", "erp-api"); err == nil {
			t.Fatal("missing security configuration accepted")
		}
	}
}
