package client

import (
	"encoding/base64"
	"fmt"
	"github.com/dgrijalva/jwt-go"
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

	decodedSecret, err := base64.StdEncoding.DecodeString(secret)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, payload)
	token.Header["alg"] = "HS256"
	token.Header["sign_type"] = "SIGN"
	signedToken, err := token.SignedString(decodedSecret)
	if err != nil {
		return "", err
	}
	return signedToken, nil
}
