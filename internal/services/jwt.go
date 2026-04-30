package services

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// MintHostJWT creates a 1-year ES256 JWT signed with the EC P-256 private key.
//
// privKeyPEM must be PKCS8 EC PEM (produced by `openssl genpkey -algorithm EC
// -pkeyopt ec_paramgen_curve:P-256`). hostID becomes the `sub` claim; the
// scheduler uses it as the key into its host→stream registry.
//
// caps lists the capabilities this host is authorized to advertise in the
// coold Hello frame. Always includes "coold"; hosts that accept builds also
// carry "builder". The scheduler cross-checks the advertised Hello capability
// set against this claim and rejects streams that try to elevate.
func MintHostJWT(privKeyPEM []byte, hostID string, caps []string) (string, error) {
	block, _ := pem.Decode(privKeyPEM)
	if block == nil {
		return "", fmt.Errorf("no PEM block found in private key")
	}
	raw, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("parse PKCS8 private key: %w", err)
	}
	ecKey, ok := raw.(*ecdsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("expected EC private key, got %T", raw)
	}
	if len(caps) == 0 {
		caps = []string{"coold"}
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  hostID,
		"aud":  "coold",
		"caps": caps,
		"iat":  now.Unix(),
		"exp":  now.Add(365 * 24 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	return token.SignedString(ecKey)
}
