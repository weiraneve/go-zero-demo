package main

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type TokenBuilder struct {
	config struct {
		AccessSecret string
		AccessExpire time.Duration
	}
}

func (b *TokenBuilder) GenerateToken(uid string) (string, int64, error) {
	tok := jwt.New(jwt.SigningMethodHS256)
	claims := tok.Claims.(jwt.MapClaims)
	claims["uid"] = uid
	expiredAt := time.Now().Add(b.config.AccessExpire).Unix()
	claims["exp"] = expiredAt

	tkr, err := tok.SignedString([]byte(b.config.AccessSecret))
	return tkr, expiredAt, err
}

func main() {
	builder := &TokenBuilder{}
	builder.config.AccessSecret = "e6808875f980b78126a4a2c59c99ef77"
	builder.config.AccessExpire = 168 * time.Hour // 7天

	token, exp, err := builder.GenerateToken("80000878")
	if err != nil {
		panic(err)
	}
	fmt.Println("Token:", token)
	fmt.Println("Expires At:", exp)
}
