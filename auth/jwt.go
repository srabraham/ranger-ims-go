package auth

import (
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"log"
	"strings"
	"time"
)

type JWTer struct {
	SecretKey string
}

func (j JWTer) CreateJWT(rangerName string, duration time.Duration) string {
	//secret := conf.Cfg.Core.JWTSecret
	//log.Println("secret:", secret)
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			// TODO
			"exp": time.Now().Add(duration).Unix(),
			"iat": time.Now().Unix(),
			"iss": "ranger-ims-go",

			// TODO
			"preferred_username": rangerName,
			"ranger_on_site":     true,
			"ranger_positions":   "Dirt - Green Dot,Green Dot Lead,Tech Ops,Green Dot Sanctuary,Operator,Tech On Call,Green Dot Lead Intern",
			"ranger_teams":       "Green Dot Team,Operator Team,Tech Cadre",
			"sub":                rangerName,
		},
	).SignedString([]byte(j.SecretKey))
	if err != nil {
		log.Panic(err)
	}
	return token
}

func (j JWTer) AuthenticateJWT(authHeader string) (jwt.MapClaims, error) {
	authHeader = strings.TrimPrefix(authHeader, "Bearer ")
	claims := jwt.MapClaims{}
	tok, err := jwt.ParseWithClaims(authHeader, &claims, func(token *jwt.Token) (any, error) {
		return []byte(j.SecretKey), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Name}))
	if err != nil {
		return nil, fmt.Errorf("[jwt.Parse]: %w", err)
	}
	if tok == nil {
		return nil, fmt.Errorf("token is nil")
	}
	if !tok.Valid {
		return nil, fmt.Errorf("token is invalid")
	}
	return claims, nil
}
