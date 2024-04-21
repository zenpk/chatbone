package util

import (
	"crypto/rsa"
	"encoding/json"

	"github.com/cristalhq/jwt/v5"
)

type Jwk struct {
	Kty string `json:"kty"`
	E   string `json:"e"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	N   string `json:"n"`
}

type Claims struct {
	jwt.RegisteredClaims
	Uuid     string `json:"uuid"`
	Username string `json:"username"`
	ClientId string `json:"clientId"`
}

func VerifyAndParse(token string, key *rsa.PublicKey) (*Claims, error) {
	verifier, err := jwt.NewVerifierRS(jwt.RS256, key)
	if err != nil {
		return nil, err
	}
	parsedToken, err := jwt.Parse([]byte(token), verifier)
	if err != nil {
		return nil, err
	}
	newClaims := new(Claims)
	if err := json.Unmarshal(parsedToken.Claims(), newClaims); err != nil {
		return nil, err
	}
	return newClaims, nil
}
