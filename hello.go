package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"fmt"
	"math/big"
	mrand "math/rand"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

func makeCert() (*ecdsa.PrivateKey, string) {
	key, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	tmpl := &x509.Certificate{
		SerialNumber:          new(big.Int).SetInt64(time.Now().UnixNano()),
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		BasicConstraintsValid: true,
	}
	der, _ := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	return key, base64.StdEncoding.EncodeToString(der)
}

func sign(key *ecdsa.PrivateKey, x5c string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	token.Header["x5c"] = []string{x5c}
	s, _ := token.SignedString(key)
	return s
}

func main() {
	key, x5c := makeCert()
	tx := fmt.Sprintf("tx-%d", mrand.Int63())

	onetimeClaims := jwt.MapClaims{
		"bundleId":      "com.example.local",
		"environment":   "Sandbox",
		"productId":     "purchase_basic",
		"transactionId": tx,
		"iat":           time.Now().Unix(),
	}
	fmt.Println("ONETIME JWS:\n", sign(key, x5c, onetimeClaims))

	subClaims := jwt.MapClaims{
		"bundleId":      "com.example.local",
		"environment":   "Sandbox",
		"productId":     "subscription_premium",
		"transactionId": fmt.Sprintf("tx-%d", mrand.Int63()),
		"iat":           time.Now().Unix(),
	}
	fmt.Println("SUB JWS:\n", sign(key, x5c, subClaims))
}
