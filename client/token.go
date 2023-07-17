package client

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

func GenerateToken(apikey string, expSeconds int) (string, error) {
	parts := strings.Split(apikey, ".")
	if len(parts) != 2 {
		return "", fmt.Errorf("apikey format error")
	}
	id := parts[0]
	secret := parts[1]
	now := time.Now().UnixMilli()
	payload := jwt.MapClaims{
		"api_key":   id,
		"timestamp": now,
		"exp":       now + int64(expSeconds)*1000,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
